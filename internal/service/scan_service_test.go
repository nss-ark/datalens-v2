package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// Mocks

type MockScanQueue struct {
	mock.Mock
}

func (m *MockScanQueue) Enqueue(ctx context.Context, jobID string) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func (m *MockScanQueue) Subscribe(ctx context.Context, handler func(ctx context.Context, jobID string) error) error {
	args := m.Called(ctx, handler)
	return args.Error(0)
}

type MockDiscoveryOrchestrator struct {
	mock.Mock
}

func (m *MockDiscoveryOrchestrator) ScanDataSource(ctx context.Context, dataSourceID types.ID) error {
	args := m.Called(ctx, dataSourceID)
	return args.Error(0)
}

func (m *MockDiscoveryOrchestrator) TestConnection(ctx context.Context, dataSourceID types.ID) error {
	args := m.Called(ctx, dataSourceID)
	return args.Error(0)
}

func (m *MockDiscoveryOrchestrator) GetClassifications(ctx context.Context, tenantID types.ID, filter discovery.ClassificationFilter) (*types.PaginatedResult[discovery.PIIClassification], error) {
	args := m.Called(ctx, tenantID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PaginatedResult[discovery.PIIClassification]), args.Error(1)
}

type MockScanRunRepo struct {
	mock.Mock
}

func (m *MockScanRunRepo) Create(ctx context.Context, run *discovery.ScanRun) error {
	args := m.Called(ctx, run)
	return args.Error(0)
}

func (m *MockScanRunRepo) GetByID(ctx context.Context, id types.ID) (*discovery.ScanRun, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*discovery.ScanRun), args.Error(1)
}

func (m *MockScanRunRepo) GetByDataSource(ctx context.Context, dataSourceID types.ID) ([]discovery.ScanRun, error) {
	args := m.Called(ctx, dataSourceID)
	return args.Get(0).([]discovery.ScanRun), args.Error(1)
}

func (m *MockScanRunRepo) GetActive(ctx context.Context, tenantID types.ID) ([]discovery.ScanRun, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]discovery.ScanRun), args.Error(1)
}

func (m *MockScanRunRepo) GetRecent(ctx context.Context, tenantID types.ID, limit int) ([]discovery.ScanRun, error) {
	args := m.Called(ctx, tenantID, limit)
	return args.Get(0).([]discovery.ScanRun), args.Error(1)
}

func (m *MockScanRunRepo) Update(ctx context.Context, run *discovery.ScanRun) error {
	args := m.Called(ctx, run)
	return args.Error(0)
}

type MockDataSourceRepo struct {
	mock.Mock
}

func (m *MockDataSourceRepo) Create(ctx context.Context, ds *discovery.DataSource) error {
	args := m.Called(ctx, ds)
	return args.Error(0)
}

func (m *MockDataSourceRepo) GetByID(ctx context.Context, id types.ID) (*discovery.DataSource, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*discovery.DataSource), args.Error(1)
}

func (m *MockDataSourceRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]discovery.DataSource, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]discovery.DataSource), args.Error(1)
}

func (m *MockDataSourceRepo) Update(ctx context.Context, ds *discovery.DataSource) error {
	args := m.Called(ctx, ds)
	return args.Error(0)
}

func (m *MockDataSourceRepo) Delete(ctx context.Context, id types.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Tests

func TestScanService_EnqueueScan_Success(t *testing.T) {
	// Setup
	scanRepo := new(MockScanRunRepo)
	dsRepo := new(MockDataSourceRepo)
	queue := new(MockScanQueue)
	discoverySvc := new(MockDiscoveryOrchestrator)
	svc := NewScanService(scanRepo, dsRepo, queue, discoverySvc, slog.Default())

	ctx := context.Background()
	tenantID := types.NewID()
	dsID := types.NewID()

	// Mock DS
	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: dsID},
			TenantID:   tenantID,
		},
		Name: "Test DB",
	}
	dsRepo.On("GetByID", ctx, dsID).Return(ds, nil)

	// Mock Active Scans (0)
	scanRepo.On("GetActive", ctx, tenantID).Return([]discovery.ScanRun{}, nil)

	// Mock Create
	// Use mock.MatchedBy to ignore dynamic fields like ID/CreatedAt
	scanRepo.On("Create", ctx, mock.MatchedBy(func(run *discovery.ScanRun) bool {
		return run.DataSourceID == dsID && run.TenantID == tenantID && run.Status == discovery.ScanStatusPending
	})).Return(nil)

	// Mock Queue
	queue.On("Enqueue", ctx, mock.AnythingOfType("string")).Return(nil)

	// Execute
	run, err := svc.EnqueueScan(ctx, dsID, tenantID, discovery.ScanTypeFull)

	// Verify
	require.NoError(t, err)
	assert.NotNil(t, run)
	assert.Equal(t, discovery.ScanStatusPending, run.Status)
	assert.Equal(t, dsID, run.DataSourceID)

	scanRepo.AssertExpectations(t)
	queue.AssertExpectations(t)
}

func TestScanService_EnqueueScan_ConcurrencyLimit(t *testing.T) {
	// Setup
	scanRepo := new(MockScanRunRepo)
	dsRepo := newMockDataSourceRepo()
	queue := new(MockScanQueue)
	discoverySvc := new(MockDiscoveryOrchestrator)
	svc := NewScanService(scanRepo, dsRepo, queue, discoverySvc, slog.Default())

	ctx := context.Background()
	tenantID := types.NewID()
	dsID := types.NewID()

	// Mock Active Scans (3 - limit is 3)
	activeScans := []discovery.ScanRun{{}, {}, {}}
	scanRepo.On("GetActive", ctx, tenantID).Return(activeScans, nil)

	// Execute
	_, err := svc.EnqueueScan(ctx, dsID, tenantID, discovery.ScanTypeFull)

	// Verify
	assert.Error(t, err)
	// Check for quota exceeded error
	assert.Contains(t, err.Error(), "max concurrent scans")
}

func TestScanService_ProcessScanJob_Success(t *testing.T) {
	// Setup
	scanRepo := new(MockScanRunRepo)
	dsRepo := newMockDataSourceRepo()
	queue := new(MockScanQueue)
	discoverySvc := new(MockDiscoveryOrchestrator)
	svc := NewScanService(scanRepo, dsRepo, queue, discoverySvc, slog.Default())

	ctx := context.Background()
	runID := types.NewID()
	dsID := types.NewID()

	// 1. Pending Run
	pendingRun := &discovery.ScanRun{
		BaseEntity:   types.BaseEntity{ID: runID},
		DataSourceID: dsID,
		Status:       discovery.ScanStatusPending,
	}
	scanRepo.On("GetByID", ctx, runID).Return(pendingRun, nil)

	// 2. Update to Running
	scanRepo.On("Update", ctx, mock.MatchedBy(func(run *discovery.ScanRun) bool {
		return run.ID == runID && run.Status == discovery.ScanStatusRunning
	})).Return(nil)

	// 3. Discovery Service Scan
	discoverySvc.On("ScanDataSource", ctx, dsID).Return(nil)

	// 4. Update to Completed
	scanRepo.On("Update", ctx, mock.MatchedBy(func(run *discovery.ScanRun) bool {
		return run.ID == runID && run.Status == discovery.ScanStatusCompleted && run.Progress == 100
	})).Return(nil)

	// Execute
	err := svc.ProcessScanJob(ctx, runID.String())

	// Verify
	require.NoError(t, err)
	scanRepo.AssertExpectations(t)
	discoverySvc.AssertExpectations(t)
}

func TestScanService_ProcessScanJob_Failure(t *testing.T) {
	// Setup
	scanRepo := new(MockScanRunRepo)
	dsRepo := newMockDataSourceRepo()
	queue := new(MockScanQueue)
	discoverySvc := new(MockDiscoveryOrchestrator)
	svc := NewScanService(scanRepo, dsRepo, queue, discoverySvc, slog.Default())

	ctx := context.Background()
	runID := types.NewID()
	dsID := types.NewID()

	// 1. Pending Run
	pendingRun := &discovery.ScanRun{
		BaseEntity:   types.BaseEntity{ID: runID},
		DataSourceID: dsID,
		Status:       discovery.ScanStatusPending,
	}
	scanRepo.On("GetByID", ctx, runID).Return(pendingRun, nil)

	// 2. Update to Running
	scanRepo.On("Update", ctx, mock.MatchedBy(func(run *discovery.ScanRun) bool {
		return run.ID == runID && run.Status == discovery.ScanStatusRunning
	})).Return(nil)

	// 3. Discovery Service Scan FAILS
	scanErr := errors.New("connection failed")
	discoverySvc.On("ScanDataSource", ctx, dsID).Return(scanErr)

	// 4. Update to Failed
	scanRepo.On("Update", ctx, mock.MatchedBy(func(run *discovery.ScanRun) bool {
		return run.ID == runID && run.Status == discovery.ScanStatusFailed && *run.ErrorMessage == "connection failed"
	})).Return(nil)

	// Execute
	err := svc.ProcessScanJob(ctx, runID.String())

	// Verify
	require.NoError(t, err) // Worker should not error out (it handled the failure)
	scanRepo.AssertExpectations(t)
}
