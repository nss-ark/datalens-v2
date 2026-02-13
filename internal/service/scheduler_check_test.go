package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/mock"
)

// Local Mocks

type LocalMockTenantRepo struct {
	mock.Mock
}

func (m *LocalMockTenantRepo) Create(ctx context.Context, t *identity.Tenant) error {
	return m.Called(ctx, t).Error(0)
}
func (m *LocalMockTenantRepo) GetByID(ctx context.Context, id types.ID) (*identity.Tenant, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*identity.Tenant), args.Error(1)
}
func (m *LocalMockTenantRepo) GetByDomain(ctx context.Context, domain string) (*identity.Tenant, error) {
	args := m.Called(ctx, domain)
	return args.Get(0).(*identity.Tenant), args.Error(1)
}
func (m *LocalMockTenantRepo) GetAll(ctx context.Context) ([]identity.Tenant, error) {
	args := m.Called(ctx)
	return args.Get(0).([]identity.Tenant), args.Error(1)
}

func (m *LocalMockTenantRepo) Update(ctx context.Context, t *identity.Tenant) error {
	return m.Called(ctx, t).Error(0)
}
func (m *LocalMockTenantRepo) Delete(ctx context.Context, id types.ID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *LocalMockTenantRepo) Search(ctx context.Context, filter identity.TenantFilter) ([]identity.Tenant, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]identity.Tenant), args.Int(1), args.Error(2)
}
func (m *LocalMockTenantRepo) GetStats(ctx context.Context) (*identity.TenantStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.TenantStats), args.Error(1)
}

type LocalMockDataSourceRepo struct {
	mock.Mock
}

func (m *LocalMockDataSourceRepo) Create(ctx context.Context, ds *discovery.DataSource) error {
	return m.Called(ctx, ds).Error(0)
}
func (m *LocalMockDataSourceRepo) GetByID(ctx context.Context, id types.ID) (*discovery.DataSource, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*discovery.DataSource), args.Error(1)
}
func (m *LocalMockDataSourceRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]discovery.DataSource, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]discovery.DataSource), args.Error(1)
}
func (m *LocalMockDataSourceRepo) Update(ctx context.Context, ds *discovery.DataSource) error {
	return m.Called(ctx, ds).Error(0)
}
func (m *LocalMockDataSourceRepo) Delete(ctx context.Context, id types.ID) error {
	return m.Called(ctx, id).Error(0)
}

type LocalMockScanOrchestrator struct {
	mock.Mock
}

func (m *LocalMockScanOrchestrator) EnqueueScan(ctx context.Context, dataSourceID types.ID, tenantID types.ID, scanType discovery.ScanType) (*discovery.ScanRun, error) {
	args := m.Called(ctx, dataSourceID, tenantID, scanType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*discovery.ScanRun), args.Error(1)
}
func (m *LocalMockScanOrchestrator) GetScan(ctx context.Context, id types.ID) (*discovery.ScanRun, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*discovery.ScanRun), args.Error(1)
}
func (m *LocalMockScanOrchestrator) GetHistory(ctx context.Context, dataSourceID types.ID) ([]discovery.ScanRun, error) {
	args := m.Called(ctx, dataSourceID)
	return args.Get(0).([]discovery.ScanRun), args.Error(1)
}

func TestCheckSchedules_TriggersScan(t *testing.T) {
	// Setup
	tenantRepo := new(LocalMockTenantRepo)
	dsRepo := new(LocalMockDataSourceRepo)
	scanSvc := new(LocalMockScanOrchestrator)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	scheduler := NewSchedulerService(dsRepo, tenantRepo, nil, scanSvc, nil, logger)

	ctx := context.Background()
	tenantID := types.NewID()

	// Expectations
	activeTenant := identity.Tenant{
		BaseEntity: types.BaseEntity{ID: tenantID},
		Status:     identity.TenantActive,
	}
	tenantRepo.On("GetAll", mock.Anything).Return([]identity.Tenant{activeTenant}, nil)

	schedule := "0 * * * *"
	lastRun := time.Now().Add(-2 * time.Hour) // Definitely due
	dsID := types.NewID()

	dataSource := discovery.DataSource{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: dsID}, TenantID: tenantID},
		ScanSchedule: &schedule,
		Status:       discovery.ConnectionStatusConnected,
		LastSyncAt:   &lastRun,
	}
	dsRepo.On("GetByTenant", mock.Anything, tenantID).Return([]discovery.DataSource{dataSource}, nil)

	scanSvc.On("EnqueueScan", mock.Anything, dsID, tenantID, discovery.ScanTypeFull).Return(&discovery.ScanRun{}, nil)

	// Execute
	scheduler.checkSchedules(ctx)

	// Verify
	tenantRepo.AssertExpectations(t)
	dsRepo.AssertExpectations(t)
	scanSvc.AssertExpectations(t)
}

func TestCheckSchedules_NotDue(t *testing.T) {
	// Setup
	tenantRepo := new(LocalMockTenantRepo)
	dsRepo := new(LocalMockDataSourceRepo)
	scanSvc := new(LocalMockScanOrchestrator)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	scheduler := NewSchedulerService(dsRepo, tenantRepo, nil, scanSvc, nil, logger)

	ctx := context.Background()
	tenantID := types.NewID()

	// Expectations
	activeTenant := identity.Tenant{
		BaseEntity: types.BaseEntity{ID: tenantID},
		Status:     identity.TenantActive,
	}
	tenantRepo.On("GetAll", mock.Anything).Return([]identity.Tenant{activeTenant}, nil)

	schedule := "0 * * * *"
	lastRun := time.Now() // Just ran, not due
	dsID := types.NewID()

	dataSource := discovery.DataSource{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: dsID}, TenantID: tenantID},
		ScanSchedule: &schedule,
		Status:       discovery.ConnectionStatusConnected,
		LastSyncAt:   &lastRun,
	}

	dsRepo.On("GetByTenant", mock.Anything, tenantID).Return([]discovery.DataSource{dataSource}, nil)

	// EnqueueScan should NOT be called

	// Execute
	scheduler.checkSchedules(ctx)

	// Verify
	tenantRepo.AssertExpectations(t)
	dsRepo.AssertExpectations(t)
	scanSvc.AssertNotCalled(t, "EnqueueScan")
}
