package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Mock DSR Repository
// =============================================================================

type mockDSRRepository struct {
	dsrs  map[types.ID]*compliance.DSR
	tasks map[types.ID][]compliance.DSRTask
}

func newMockDSRRepository() *mockDSRRepository {
	return &mockDSRRepository{
		dsrs:  make(map[types.ID]*compliance.DSR),
		tasks: make(map[types.ID][]compliance.DSRTask),
	}
}

func (m *mockDSRRepository) Create(_ context.Context, dsr *compliance.DSR) error {
	m.dsrs[dsr.ID] = dsr
	return nil
}

func (m *mockDSRRepository) GetByID(_ context.Context, id types.ID) (*compliance.DSR, error) {
	dsr, ok := m.dsrs[id]
	if !ok {
		return nil, types.NewNotFoundError("DSR", id)
	}
	return dsr, nil
}

func (m *mockDSRRepository) GetByTenant(_ context.Context, tenantID types.ID, pagination types.Pagination, statusFilter *compliance.DSRStatus) (*types.PaginatedResult[compliance.DSR], error) {
	var items []compliance.DSR
	for _, dsr := range m.dsrs {
		if dsr.TenantID == tenantID {
			if statusFilter == nil || dsr.Status == *statusFilter {
				items = append(items, *dsr)
			}
		}
	}
	return &types.PaginatedResult[compliance.DSR]{
		Items:      items,
		Total:      len(items),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: 1,
	}, nil
}

func (m *mockDSRRepository) GetOverdue(_ context.Context, tenantID types.ID) ([]compliance.DSR, error) {
	var items []compliance.DSR
	now := time.Now()
	for _, dsr := range m.dsrs {
		if dsr.TenantID == tenantID && dsr.Status == compliance.DSRStatusPending && dsr.SLADeadline.Before(now) {
			items = append(items, *dsr)
		}
	}
	return items, nil
}

func (m *mockDSRRepository) Update(_ context.Context, dsr *compliance.DSR) error {
	m.dsrs[dsr.ID] = dsr
	return nil
}

func (m *mockDSRRepository) CreateTask(_ context.Context, task *compliance.DSRTask) error {
	m.tasks[task.DSRID] = append(m.tasks[task.DSRID], *task)
	return nil
}

func (m *mockDSRRepository) GetTasksByDSR(_ context.Context, dsrID types.ID) ([]compliance.DSRTask, error) {
	return m.tasks[dsrID], nil
}

func (m *mockDSRRepository) UpdateTask(_ context.Context, task *compliance.DSRTask) error {
	tasks := m.tasks[task.DSRID]
	for i, t := range tasks {
		if t.ID == task.ID {
			tasks[i] = *task
			break
		}
	}
	return nil
}

// =============================================================================
// Mock DSR Queue
// =============================================================================

type mockDSRQueue struct{}

func newMockDSRQueue() *mockDSRQueue                              { return &mockDSRQueue{} }
func (m *mockDSRQueue) Enqueue(_ context.Context, _ string) error { return nil }
func (m *mockDSRQueue) Subscribe(_ context.Context, _ func(ctx context.Context, dsrID string) error) error {
	return nil
}

// =============================================================================
// Tests
// =============================================================================

func TestDSRService_CreateDSR_Success(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dsrRepo := newMockDSRRepository()
	dsRepo := newMockDataSourceRepo()
	dsrQueue := newMockDSRQueue()
	eb := newMockEventBus()
	svc := NewDSRService(dsrRepo, dsRepo, dsrQueue, eb, logger)

	tenantID := types.NewID()
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, tenantID)

	// Execute
	req := CreateDSRRequest{
		RequestType:        compliance.RequestTypeAccess,
		SubjectName:        "John Doe",
		SubjectEmail:       "john@example.com",
		SubjectIdentifiers: map[string]string{"user_id": "u_123"},
		Priority:           "HIGH",
	}

	dsr, err := svc.CreateDSR(ctx, req)

	// Verify
	require.NoError(t, err)
	assert.NotNil(t, dsr)
	assert.Equal(t, compliance.DSRStatusPending, dsr.Status)
	assert.Equal(t, "John Doe", dsr.SubjectName)
	assert.Equal(t, "john@example.com", dsr.SubjectEmail)

	// Verify SLA calculation (30 days from now)
	expectedSLA := time.Now().AddDate(0, 0, 30)
	assert.WithinDuration(t, expectedSLA, dsr.SLADeadline, 5*time.Second)

	// Verify event was published
	require.Len(t, eb.Events, 1)
	event := eb.Events[0]
	assert.Equal(t, eventbus.EventDSRCreated, event.Type)
	data, ok := event.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, dsr.ID, data["dsr_id"])
}

func TestDSRService_ApproveDSR_Success(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dsrRepo := newMockDSRRepository()
	dsRepo := newMockDataSourceRepo()
	dsrQueue := newMockDSRQueue()
	eb := newMockEventBus()
	svc := NewDSRService(dsrRepo, dsRepo, dsrQueue, eb, logger)

	tenantID := types.NewID()
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, tenantID)

	// Create pending DSR
	dsr := &compliance.DSR{
		ID:           types.NewID(),
		TenantID:     tenantID,
		RequestType:  compliance.RequestTypeAccess,
		Status:       compliance.DSRStatusPending,
		SubjectName:  "Alice",
		SubjectEmail: "alice@example.com",
		SLADeadline:  time.Now().AddDate(0, 0, 30),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	dsrRepo.Create(ctx, dsr)

	// Create 2 data sources for the tenant
	ds1 := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID()},
			TenantID:   tenantID,
		},
		Name: "DB1",
	}
	ds2 := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID()},
			TenantID:   tenantID,
		},
		Name: "DB2",
	}
	dsRepo.Create(ctx, ds1)
	dsRepo.Create(ctx, ds2)

	// Execute
	approved, err := svc.ApproveDSR(ctx, dsr.ID)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, compliance.DSRStatusApproved, approved.Status)

	// Verify tasks were created (one per data source)
	tasks, _ := dsrRepo.GetTasksByDSR(ctx, dsr.ID)
	assert.Len(t, tasks, 2, "Should create a task for each data source")
	assert.Equal(t, ds1.ID, tasks[0].DataSourceID)
	assert.Equal(t, ds2.ID, tasks[1].DataSourceID)
	assert.Equal(t, compliance.TaskStatusPending, tasks[0].Status)

	// Verify event was published
	require.Len(t, eb.Events, 1)
	assert.Equal(t, eventbus.EventDSRExecuting, eb.Events[0].Type)
}

func TestDSRService_RejectDSR_Success(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dsrRepo := newMockDSRRepository()
	dsRepo := newMockDataSourceRepo()
	dsrQueue := newMockDSRQueue()
	eb := newMockEventBus()
	svc := NewDSRService(dsrRepo, dsRepo, dsrQueue, eb, logger)

	tenantID := types.NewID()
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, tenantID)

	// Create pending DSR
	dsr := &compliance.DSR{
		ID:          types.NewID(),
		TenantID:    tenantID,
		RequestType: compliance.RequestTypeErasure,
		Status:      compliance.DSRStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	dsrRepo.Create(ctx, dsr)

	// Execute
	reason := "Invalid identity verification"
	rejected, err := svc.RejectDSR(ctx, dsr.ID, reason)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, compliance.DSRStatusRejected, rejected.Status)
	assert.Equal(t, reason, rejected.Reason)
	assert.NotNil(t, rejected.CompletedAt)

	// Verify event was published
	require.Len(t, eb.Events, 1)
	event := eb.Events[0]
	assert.Equal(t, eventbus.EventDSRRejected, event.Type)
	data, ok := event.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, dsr.ID, data["dsr_id"])
	assert.Equal(t, reason, data["reason"])
}

func TestDSRService_ApproveDSR_InvalidTransition(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dsrRepo := newMockDSRRepository()
	dsRepo := newMockDataSourceRepo()
	dsrQueue := newMockDSRQueue()
	eb := newMockEventBus()
	svc := NewDSRService(dsrRepo, dsRepo, dsrQueue, eb, logger)

	tenantID := types.NewID()
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, tenantID)

	// Create COMPLETED DSR (invalid for approval)
	completedAt := time.Now()
	dsr := &compliance.DSR{
		ID:          types.NewID(),
		TenantID:    tenantID,
		RequestType: compliance.RequestTypeAccess,
		Status:      compliance.DSRStatusCompleted,
		CompletedAt: &completedAt,
		CreatedAt:   time.Now().Add(-48 * time.Hour),
		UpdatedAt:   time.Now(),
	}
	dsrRepo.Create(ctx, dsr)

	// Execute
	_, err := svc.ApproveDSR(ctx, dsr.ID)

	// Verify
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transition")
}

func TestDSRService_GetDSRs_WithFilter(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dsrRepo := newMockDSRRepository()
	dsRepo := newMockDataSourceRepo()
	dsrQueue := newMockDSRQueue()
	eb := newMockEventBus()
	svc := NewDSRService(dsrRepo, dsRepo, dsrQueue, eb, logger)

	tenantID := types.NewID()
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, tenantID)

	// Create multiple DSRs with different statuses
	pending := &compliance.DSR{ID: types.NewID(), TenantID: tenantID, Status: compliance.DSRStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	approved := &compliance.DSR{ID: types.NewID(), TenantID: tenantID, Status: compliance.DSRStatusApproved, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	rejected := &compliance.DSR{ID: types.NewID(), TenantID: tenantID, Status: compliance.DSRStatusRejected, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	dsrRepo.Create(ctx, pending)
	dsrRepo.Create(ctx, approved)
	dsrRepo.Create(ctx, rejected)

	// Execute: Filter by PENDING
	statusFilter := compliance.DSRStatusPending
	result, err := svc.GetDSRs(ctx, types.Pagination{Page: 1, PageSize: 10}, &statusFilter)

	// Verify
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, compliance.DSRStatusPending, result.Items[0].Status)

	// Execute: Get all (no filter)
	resultAll, err := svc.GetDSRs(ctx, types.Pagination{Page: 1, PageSize: 10}, nil)

	// Verify
	require.NoError(t, err)
	assert.Len(t, resultAll.Items, 3)
}
