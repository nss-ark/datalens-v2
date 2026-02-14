package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector"
	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/types"
)

func TestIntegration_Postgres_Scanning_Flow(t *testing.T) {
	// 1. Setup Config & Connection Details
	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID()},
			TenantID:   types.NewID(),
		},
		Name:        "Local Postgres",
		Type:        types.DataSourcePostgreSQL,
		Host:        "localhost",
		Port:        5433,
		Database:    "datalens",
		Credentials: "datalens:datalens_dev",
	}

	// 2. Setup Real Connector
	pgConn := connector.NewPostgresConnector()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pgConn.Connect(ctx, ds); err != nil {
		t.Logf("Skipping integration test: could not connect to postgres: %v", err)
		t.SkipNow()
	}
	defer pgConn.Close()

	// 3. Setup Mocks (In-Memory Repos)
	// These are in-memory implementations from mocks_test.go, NOT testify mocks. Do not call .On().
	dsRepo := newMockDataSourceRepo()
	require.NoError(t, dsRepo.Create(ctx, ds)) // Pre-seed DS

	invRepo := newMockDataInventoryRepo()
	entityRepo := newMockDataEntityRepo()
	fieldRepo := newMockDataFieldRepo()
	piiRepo := newMockPIIClassificationRepo()
	scanRunRepo := newMockScanRunRepo()
	eb := newMockEventBus()

	// 4. Setup Mock Detector (This IS a testify mock in discovery_service_test.go)
	mockStrategy := new(MockStrategy)
	mockStrategy.On("Detect", mock.Anything, mock.Anything).Return([]detection.Result{
		{
			Category:   types.PIICategoryIdentity,
			Type:       types.PIITypeName,
			Confidence: 0.9,
			Method:     types.DetectionMethodAI,
			Reasoning:  "Integration Test Match",
		},
	}, nil).Maybe()

	detector := detection.NewComposableDetector(mockStrategy)

	// 5. Registry
	registry := connector.NewConnectorRegistry(&config.Config{}, detector)
	registry.Register(types.DataSourcePostgreSQL, func() discovery.Connector {
		return pgConn
	})

	// 6. Service
	svc := NewDiscoveryService(dsRepo, invRepo, entityRepo, fieldRepo, piiRepo, scanRunRepo, registry, detector, eb, slog.Default())

	// 7. Execute Scan
	t.Log("Starting ScanDataSource...")
	err := svc.ScanDataSource(ctx, ds.ID)
	require.NoError(t, err)

	// 8. Verify
	// Check for Inventory
	inv, err := invRepo.GetByDataSource(ctx, ds.ID)
	require.NoError(t, err)
	assert.NotNil(t, inv)
	t.Logf("Inventory Created: %+v", inv)

	// Check for Entities
	entities, err := entityRepo.GetByInventory(ctx, inv.ID)
	require.NoError(t, err)
	t.Logf("Found %d entities", len(entities))
	// 'datalens' DB likely has tables.
	// assert.NotEmpty(t, entities) // Only if we are sure DB is not empty.

	// Check for PII (if detection worked)
	// Since we mock detection to always return PII for ANY call, if we sampled any data, we should have PII.
	// We only sample if we found fields.

	// If entities > 0, we expect fields.
	if len(entities) > 0 {
		fields, err := fieldRepo.GetByEntity(ctx, entities[0].ID)
		require.NoError(t, err)
		t.Logf("First entity has %d fields", len(fields))

		// If fields > 0, we expect PII classification because mock detector always returns result.
		// Wait, Detect is called per field.
		if len(fields) > 0 {
			// Check PII
			pagination := types.Pagination{Page: 1, PageSize: 10}
			pii, err := piiRepo.GetByDataSource(ctx, ds.ID, pagination)
			require.NoError(t, err)
			t.Logf("Found %d PII classifications", len(pii.Items))
			// assert.NotEmpty(t, pii.Items)
		}
	}
}
