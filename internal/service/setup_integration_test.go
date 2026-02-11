//go:build integration

package service

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	var connStr string
	var container *tcpostgres.PostgresContainer
	var err error

	// Check for CI environment or existing DB
	if os.Getenv("DATABASE_URL") != "" {
		connStr = os.Getenv("DATABASE_URL")
	} else {
		// Try to reuse existing container if possible (not easy with testcontainers-go without fixed name)
		// For now, start new one.
		container, err = tcpostgres.Run(ctx,
			"postgres:16-alpine",
			tcpostgres.WithDatabase("datalens_test_service"),
			tcpostgres.WithUsername("test"),
			tcpostgres.WithPassword("test"),
			testcontainers.WithWaitStrategy(
				wait.ForListeningPort("5432/tcp").
					WithStartupTimeout(60*time.Second),
			),
		)
		if err != nil {
			panic("failed to start postgres container: " + err.Error())
		}

		// Map port? No, ConnectionString handles it.

		connStr, err = container.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			panic("failed to get connection string: " + err.Error())
		}
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}

	// Apply migrations
	if err := applyMigrations(ctx, testPool); err != nil {
		// If migrations fail, we might want to panic, but let's log.
		// panic("failed to apply migrations: " + err.Error())
		// Panic is better to stop early.
		panic("failed to apply migrations: " + err.Error())
	}

	code := m.Run()

	testPool.Close()
	if container != nil {
		container.Terminate(ctx)
	}

	os.Exit(code)
}

func setupPostgres(t *testing.T) *pgxpool.Pool {
	if testPool == nil {
		t.Fatal("Global testPool is nil. Did TestMain run?")
	}
	return testPool
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Find migrations directory relative to this test file.
	// Assuming this file is in internal/service/
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	filesToCheck := []string{"001_initial_schema.sql", "002_api_keys.sql", "003_detection_feedback.sql", "004_dsr.sql", "005_consent.sql", "006_governance_violations.sql"}

	for _, f := range filesToCheck {
		path := filepath.Join(migrationsDir, f)
		// Check if file exists first to avoid error if I added a hypothetical one
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		sql, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return err
		}
	}
	return nil
}
