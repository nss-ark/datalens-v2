//go:build integration

package service

import (
	"context"
	"fmt"
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
	_, filename, _, _ := runtime.Caller(0)
	serviceDir := filepath.Dir(filename)

	// Define migration directories relative to this file (internal/service)
	// 1. Root migrations: ../../migrations
	// 2. Internal migrations: ../database/migrations
	dirs := []string{
		filepath.Join(serviceDir, "..", "..", "migrations"),
		filepath.Join(serviceDir, "..", "database", "migrations"),
	}

	for _, dir := range dirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
		if err != nil {
			return err
		}

		// Sort files to ensure order (001, 002, ...)
		// filepath.Glob returns sorted matches, but no harm in verifying or relying on it.
		// (Glob output is sorted)

		for _, path := range files {
			sql, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if _, err := pool.Exec(ctx, string(sql)); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", filepath.Base(path), err)
			}
		}
	}
	return nil
}
