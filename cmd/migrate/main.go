// DataLens 2.0 — Database Migration Runner
//
// Runs SQL migrations against the PostgreSQL database.
// Usage: go run ./cmd/migrate [up|down|status]
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/pkg/logging"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logging.New(cfg.App.Env, cfg.App.LogLevel).
		WithComponent("migrate")

	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down|status]")
		os.Exit(1)
	}

	command := os.Args[1]

	log.Info("Running migration",
		"command", command,
		"database", cfg.DB.Name,
		"host", cfg.DB.Host,
	)

	// Connect to database
	connStr := cfg.DB.DSN()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Error("Database ping failed", "error", err)
		os.Exit(1)
	}
	log.Info("Connected to database")

	// Ensure schema_migrations table exists
	if err := ensureMigrationsTable(ctx, pool); err != nil {
		log.Error("Failed to create migrations table", "error", err)
		os.Exit(1)
	}

	switch command {
	case "up":
		if err := migrateUp(ctx, pool, log); err != nil {
			log.Error("Migration UP failed", "error", err)
			os.Exit(1)
		}
		log.Info("All migrations applied successfully")

	case "down":
		if err := migrateDown(ctx, pool, log); err != nil {
			log.Error("Migration DOWN failed", "error", err)
			os.Exit(1)
		}
		log.Info("Last migration rolled back")

	case "status":
		if err := migrateStatus(ctx, pool, log); err != nil {
			log.Error("Failed to get migration status", "error", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown command: %s\nUsage: migrate [up|down|status]\n", command)
		os.Exit(1)
	}
}

func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     VARCHAR(255) PRIMARY KEY,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`
	_, err := pool.Exec(ctx, query)
	return err
}

func getAppliedMigrations(ctx context.Context, pool *pgxpool.Pool) (map[string]time.Time, error) {
	rows, err := pool.Query(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]time.Time)
	for rows.Next() {
		var version string
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		applied[version] = appliedAt
	}
	return applied, rows.Err()
}

func getMigrationFiles() ([]string, error) {
	// Look for migrations directory relative to working directory
	migrationsDir := "migrations"
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func migrateUp(ctx context.Context, pool *pgxpool.Pool, log *logging.Logger) error {
	applied, err := getAppliedMigrations(ctx, pool)
	if err != nil {
		return fmt.Errorf("get applied: %w", err)
	}

	files, err := getMigrationFiles()
	if err != nil {
		return err
	}

	pending := 0
	for _, file := range files {
		version := strings.TrimSuffix(file, ".sql")
		if _, ok := applied[version]; ok {
			continue
		}

		log.Info("Applying migration", "version", version)

		sql, err := os.ReadFile(filepath.Join("migrations", file))
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("exec %s: %w", file, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", file, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", file, err)
		}

		log.Info("Migration applied", "version", version)
		pending++
	}

	if pending == 0 {
		log.Info("No pending migrations")
	} else {
		log.Info("Migrations complete", "applied", pending)
	}
	return nil
}

func migrateDown(ctx context.Context, pool *pgxpool.Pool, log *logging.Logger) error {
	applied, err := getAppliedMigrations(ctx, pool)
	if err != nil {
		return fmt.Errorf("get applied: %w", err)
	}

	if len(applied) == 0 {
		log.Info("No migrations to roll back")
		return nil
	}

	// Find the latest applied migration
	var versions []string
	for v := range applied {
		versions = append(versions, v)
	}
	sort.Strings(versions)
	latest := versions[len(versions)-1]

	log.Info("Rolling back migration", "version", latest)

	// Remove from tracking (we don't have down migration files yet — just untrack)
	_, err = pool.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", latest)
	if err != nil {
		return fmt.Errorf("remove %s: %w", latest, err)
	}

	log.Info("Migration rolled back", "version", latest)
	return nil
}

func migrateStatus(ctx context.Context, pool *pgxpool.Pool, log *logging.Logger) error {
	applied, err := getAppliedMigrations(ctx, pool)
	if err != nil {
		return fmt.Errorf("get applied: %w", err)
	}

	files, err := getMigrationFiles()
	if err != nil {
		return err
	}

	fmt.Println("\n  Migration Status")
	fmt.Println("  ────────────────────────────────────────────────")
	for _, file := range files {
		version := strings.TrimSuffix(file, ".sql")
		if t, ok := applied[version]; ok {
			fmt.Printf("  ✅ %s  (applied %s)\n", version, t.Format(time.RFC3339))
		} else {
			fmt.Printf("  ⬜ %s  (pending)\n", version)
		}
	}
	fmt.Println()

	log.Info("Status checked", "total", len(files), "applied", len(applied), "pending", len(files)-len(applied))
	return nil
}
