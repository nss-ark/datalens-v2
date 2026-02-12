//go:build integration

package audit_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/complyark/datalens/internal/domain/audit"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/pkg/types"
)

func setupPostgres(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	// Use specific wait strategy
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("datalens_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	// Determine the input directory relative to this file
	// We assume the test is running from the project root or we can find the migration file
	_, filename, _, _ := runtime.Caller(0)
	// filename is .../internal/domain/audit/audit_integration_test.go
	// adjusting relative path to reach internal/database/migrations
	projectRoot := filepath.Join(filepath.Dir(filename), "../../..")
	migrationFile := filepath.Join(projectRoot, "internal/database/migrations/009_audit_logs.sql")

	runMigration(t, pool, migrationFile)

	return pool
}

func runMigration(t *testing.T, pool *pgxpool.Pool, migrationPath string) {
	content, err := os.ReadFile(migrationPath)
	require.NoError(t, err, "failed to read migration file: %s", migrationPath)

	_, err = pool.Exec(context.Background(), string(content))
	require.NoError(t, err, "failed to execute migration")
}

func TestAuditRepository_CreateAndQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool := setupPostgres(t)
	repo := repository.NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// 1. Create Tenant and User IDs (random UUIDs)
	tenantID := types.NewID()
	userID := types.NewID()
	resourceID := types.NewID()

	// 2. Create Audit Log
	logEntry := &audit.AuditLog{
		ID:           types.NewID(),
		TenantID:     tenantID,
		UserID:       userID,
		Action:       "TEST_ACTION",
		ResourceType: "TEST_RESOURCE",
		ResourceID:   resourceID,
		OldValues:    map[string]any{"status": "inactive"},
		NewValues:    map[string]any{"status": "active"},
		IPAddress:    "127.0.0.1",
		UserAgent:    "TestAgent/1.0",
		CreatedAt:    time.Now().UTC(),
	}

	err := repo.Create(ctx, logEntry)
	require.NoError(t, err)

	// 3. Query Logs
	logs, err := repo.GetByTenant(ctx, tenantID, 10)
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logEntry.ID, logs[0].ID)
	assert.Equal(t, logEntry.Action, logs[0].Action)
	assert.Equal(t, logEntry.ResourceType, logs[0].ResourceType)

	// Assert JSONB fields
	assert.Equal(t, "inactive", logs[0].OldValues["status"])
	assert.Equal(t, "active", logs[0].NewValues["status"])
}

func TestAuditRepository_GetByTenant_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool := setupPostgres(t)
	repo := repository.NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// Random tenant with no logs
	logs, err := repo.GetByTenant(ctx, types.NewID(), 10)
	require.NoError(t, err)
	assert.Empty(t, logs)
}
