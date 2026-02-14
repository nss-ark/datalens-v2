package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector"
	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Mock Connector
// =============================================================================

// MockConnector is defined in mocks_test.go

// =============================================================================
// Mock Strategy
// =============================================================================

type MockStrategy struct {
	mock.Mock
}

func (m *MockStrategy) Name() string                  { return "MockStrategy" }
func (m *MockStrategy) Method() types.DetectionMethod { return types.DetectionMethodAI }
func (m *MockStrategy) Weight() float64               { return 1.0 }
func (m *MockStrategy) Detect(ctx context.Context, input detection.Input) ([]detection.Result, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]detection.Result), args.Error(1)
}

// =============================================================================
// Tests
// =============================================================================

func TestDiscoveryService_ScanDataSource_Success(t *testing.T) {
	// Setup Mocks
	dsRepo := newMockDataSourceRepo()
	invRepo := newMockDataInventoryRepo()
	entityRepo := newMockDataEntityRepo()
	fieldRepo := newMockDataFieldRepo()
	piiRepo := newMockPIIClassificationRepo()
	connectorMock := new(MockConnector)
	eb := newMockEventBus()

	// Setup Detector with Mock Strategy
	mockStrategy := new(MockStrategy)
	detector := detection.NewComposableDetector(mockStrategy)

	// Setup Registry with Mock Connector
	registry := connector.NewConnectorRegistry(&config.Config{}, detector, nil)
	testDSType := types.DataSourceType("TEST_MOCK")
	registry.Register(testDSType, func() discovery.Connector {
		return connectorMock
	})

	// Setup Service
	scanRunRepo := newMockScanRunRepo()
	svc := NewDiscoveryService(dsRepo, invRepo, entityRepo, fieldRepo, piiRepo, scanRunRepo, registry, detector, eb, slog.Default())

	ctx := context.Background()
	tenantID := types.NewID()

	// Setup Data Source
	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{
				ID: types.NewID(),
			},
			TenantID: tenantID,
		},
		Name: "Test DB",
		Type: types.DataSourceType("TEST_MOCK"),
	}
	require.NoError(t, dsRepo.Create(ctx, ds))

	// Mock Connector Behavior
	// Connect is called with the DS
	connectorMock.On("Connect", ctx, mock.Anything).Return(nil)
	connectorMock.On("Close").Return(nil)

	// Mock Schema Discovery
	inv := &discovery.DataInventory{DataSourceID: ds.ID}
	entities := []discovery.DataEntity{
		{Name: "users", Type: discovery.EntityTypeTable},
	}
	connectorMock.On("DiscoverSchema", ctx, mock.Anything).Return(inv, entities, nil)

	// Mock GetFields
	fields := []discovery.DataField{
		{Name: "email", DataType: "varchar"},
	}
	connectorMock.On("GetFields", ctx, "users").Return(fields, nil)

	// Mock SampleData
	samples := []string{"test@example.com"}
	connectorMock.On("SampleData", ctx, "users", "email", 10).Return(samples, nil)

	// Mock Detection Strategy
	expectedDetection := []detection.Result{
		{
			Category:    types.PIICategoryContact,
			Type:        types.PIITypeEmail,
			Sensitivity: types.SensitivityMedium,
			Confidence:  0.95,
			Method:      types.DetectionMethodAI,
			Reasoning:   "Looks like email",
		},
	}
	mockStrategy.On("Detect", ctx, mock.Anything).Return(expectedDetection, nil)

	// Execute
	err := svc.ScanDataSource(ctx, ds.ID)
	require.NoError(t, err)

	// Verify Persistence
	// Inventory created?
	savedInv, err := invRepo.GetByDataSource(ctx, ds.ID)
	require.NoError(t, err)
	assert.NotNil(t, savedInv)

	// Entity created?
	savedEntities, err := entityRepo.GetByInventory(ctx, savedInv.ID)
	require.NoError(t, err)
	assert.Len(t, savedEntities, 1)
	assert.Equal(t, "users", savedEntities[0].Name)

	// Field created?
	savedFields, err := fieldRepo.GetByEntity(ctx, savedEntities[0].ID)
	require.NoError(t, err)
	assert.Len(t, savedFields, 1)
	assert.Equal(t, "email", savedFields[0].Name)

	// Classification created?
	classifications, err := piiRepo.GetByDataSource(ctx, ds.ID, types.Pagination{})
	require.NoError(t, err)
	assert.Len(t, classifications.Items, 1)
	assert.Equal(t, types.PIITypeEmail, classifications.Items[0].Type)

	connectorMock.AssertExpectations(t)
	mockStrategy.AssertExpectations(t)
}
