package detection_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/internal/service/ai"
	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/types"
)

var dbPool *pgxpool.Pool

// projectRoot returns the absolute path to the project root by walking up
// from the current source file until it finds go.mod.
func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find project root (go.mod)")
		}
		dir = parent
	}
}

func setupIntegrationDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { postgres.Terminate(ctx) })

	host, err := postgres.Host(ctx)
	require.NoError(t, err)
	port, err := postgres.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := "postgres://testuser:testpassword@" + host + ":" + port.Port() + "/testdb?sslmode=disable"
	dbPool, err = pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	// Run actual migration file
	migrationPath := filepath.Join(projectRoot(), "migrations", "001_initial_schema.sql")
	sql, err := os.ReadFile(migrationPath)
	require.NoError(t, err, "Could not read migration file")
	_, err = dbPool.Exec(ctx, string(sql))
	require.NoError(t, err, "Migration failed")
}

func TestDetectionWorkflow_Integration(t *testing.T) {
	setupIntegrationDB(t)
	ctx := context.Background()

	// 1. Setup Repos
	tenantRepo := repository.NewTenantRepo(dbPool)
	dsRepo := repository.NewDataSourceRepo(dbPool)
	invRepo := repository.NewDataInventoryRepo(dbPool)
	entityRepo := repository.NewDataEntityRepo(dbPool)
	fieldRepo := repository.NewDataFieldRepo(dbPool)
	piiRepo := repository.NewPIIClassificationRepo(dbPool)

	// 2. Build Data Hierarchy: Tenant -> DataSource -> Inventory -> Entity -> Field
	tenant := &identity.Tenant{
		Name:   "DetectionCo",
		Domain: "detect-" + types.NewID().String()[:8] + ".com",
		Status: identity.TenantActive,
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	ds := &discovery.DataSource{
		Name:   "Prod DB",
		Type:   types.DataSourcePostgreSQL,
		Status: discovery.ConnectionStatusConnected,
	}
	ds.TenantID = tenant.ID
	require.NoError(t, dsRepo.Create(ctx, ds))

	inv := &discovery.DataInventory{DataSourceID: ds.ID}
	require.NoError(t, invRepo.Create(ctx, inv))

	entity := &discovery.DataEntity{
		InventoryID: inv.ID,
		Name:        "customers",
		Type:        discovery.EntityTypeTable,
	}
	require.NoError(t, entityRepo.Create(ctx, entity))

	field := &discovery.DataField{
		EntityID: entity.ID,
		Name:     "email",
		DataType: "VARCHAR",
	}
	require.NoError(t, fieldRepo.Create(ctx, field))

	// 3. Setup Detector with Mock AI Strategy
	mockGateway := new(MockGateway)
	aiStrategy := detection.NewAIStrategy(mockGateway, 0.7)
	detector := detection.NewComposableDetector(aiStrategy)

	mockGateway.On("DetectPII", ctx, mock.Anything).Return(&ai.PIIDetectionResult{
		IsPII:       true,
		Category:    types.PIICategoryContact,
		Type:        types.PIITypeEmail,
		Sensitivity: types.SensitivityMedium,
		Confidence:  0.98,
		Reasoning:   "It is an email",
	}, nil)

	// 4. Run Detection
	input := detection.Input{
		TableName:  "customers",
		ColumnName: "email",
		Samples:    []string{"test@example.com"},
	}

	report, err := detector.Detect(ctx, input)
	require.NoError(t, err)
	require.True(t, report.IsPII)
	require.NotEmpty(t, report.Detections)

	// 5. Persist Results (simulating the service layer)
	for _, det := range report.Detections {
		classification := &discovery.PIIClassification{
			FieldID:         field.ID,
			DataSourceID:    ds.ID,
			EntityName:      input.TableName,
			FieldName:       input.ColumnName,
			Category:        det.Category,
			Type:            det.Type,
			Sensitivity:     det.Sensitivity,
			Confidence:      det.FinalConfidence,
			DetectionMethod: det.Methods[0],
			Status:          types.VerificationPending,
			Reasoning:       det.Reasoning,
		}
		err := piiRepo.Create(ctx, classification)
		require.NoError(t, err)
	}

	// 6. Verify Persistence
	page, err := piiRepo.GetByDataSource(ctx, ds.ID, types.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, page.Total)
	assert.Equal(t, types.PIITypeEmail, page.Items[0].Type)
	assert.Equal(t, types.DetectionMethodAI, page.Items[0].DetectionMethod)

	// 7. Verify Pending Flow
	pending, err := piiRepo.GetPending(ctx, tenant.ID, types.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, pending.Total)
}
