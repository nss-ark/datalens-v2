package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Mock Connector
// =============================================================================

type DSRExecutorMockConnector struct {
	mock.Mock
}

func (m *DSRExecutorMockConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	args := m.Called(ctx, ds)
	return args.Error(0)
}

func (m *DSRExecutorMockConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*discovery.DataInventory), args.Get(1).([]discovery.DataEntity), args.Error(2)
}

func (m *DSRExecutorMockConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	args := m.Called(ctx, entityID)
	return args.Get(0).([]discovery.DataField), args.Error(1)
}

func (m *DSRExecutorMockConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	args := m.Called(ctx, entity, field, limit)
	return args.Get(0).([]string), args.Error(1)
}

func (m *DSRExecutorMockConnector) Capabilities() discovery.ConnectorCapabilities {
	args := m.Called()
	return args.Get(0).(discovery.ConnectorCapabilities)
}

func (m *DSRExecutorMockConnector) Close() error {
	args := m.Called()
	return args.Error(0)
}

// =============================================================================
// Settings
// =============================================================================

func setupExecutorTest(t *testing.T) (*DSRExecutor, *mockDSRRepository, *mockDataSourceRepo, *mockPIIClassificationRepo, *DSRExecutorMockConnector, *mockEventBus) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	dsrRepo := newMockDSRRepository()
	dsRepo := newMockDataSourceRepo()
	piiRepo := newMockPIIClassificationRepo()
	eb := newMockEventBus()

	// Registry & Mock Connector
	mockConn := new(DSRExecutorMockConnector)
	registry := connector.NewConnectorRegistry()
	registry.Register(types.DataSourcePostgreSQL, func() discovery.Connector {
		return mockConn
	})

	executor := NewDSRExecutor(dsrRepo, dsRepo, piiRepo, registry, eb, logger)
	return executor, dsrRepo, dsRepo, piiRepo, mockConn, eb
}

// =============================================================================
// Tests
// =============================================================================

func TestExecuteDSR_Access(t *testing.T) {
	// Setup
	executor, dsrRepo, dsRepo, piiRepo, mockConn, eb := setupExecutorTest(t)
	ctx := context.Background()

	tenantID := types.NewID()
	dsID := types.NewID()
	dsrID := types.NewID()
	taskID := types.NewID()

	// 1. Create DSR
	dsr := &compliance.DSR{
		ID:                 dsrID,
		TenantID:           tenantID,
		RequestType:        compliance.RequestTypeAccess,
		Status:             compliance.DSRStatusApproved, // Start from approved
		SubjectName:        "John Doe",
		SubjectEmail:       "john@example.com",
		SubjectIdentifiers: map[string]string{"email": "john@example.com"},
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	dsrRepo.Create(ctx, dsr)

	// 2. Create Task
	task := &compliance.DSRTask{
		ID:           taskID,
		DSRID:        dsrID,
		DataSourceID: dsID,
		TenantID:     tenantID,
		TaskType:     compliance.RequestTypeAccess,
		Status:       compliance.TaskStatusPending,
		CreatedAt:    time.Now(),
	}
	dsrRepo.CreateTask(ctx, task)

	// 3. Create Data Source
	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: dsID}, TenantID: tenantID},
		Name:         "Postgres DB",
		Type:         types.DataSourcePostgreSQL,
	}
	dsRepo.Create(ctx, ds)

	// 4. Create PII Classifications
	pii1 := &discovery.PIIClassification{
		BaseEntity:   types.BaseEntity{ID: types.NewID()},
		DataSourceID: dsID,
		EntityName:   "users",
		FieldName:    "email",
		Category:     types.PIICategoryContact,
		Confidence:   0.9,
	}
	piiRepo.Create(ctx, pii1)

	// 5. Mock Interactions
	// Connect
	mockConn.On("Connect", ctx, mock.AnythingOfType("*discovery.DataSource")).Return(nil)
	// Sample Data (Success match)
	mockConn.On("SampleData", ctx, "users", "email", 100).Return([]string{"john@example.com", "jane@example.com"}, nil)
	// Close
	mockConn.On("Close").Return(nil)

	// Execute
	err := executor.ExecuteDSR(ctx, dsrID)

	// Verify
	require.NoError(t, err)

	// Check DSR Status
	updatedDSR, _ := dsrRepo.GetByID(ctx, dsrID)
	assert.Equal(t, compliance.DSRStatusCompleted, updatedDSR.Status)
	assert.NotNil(t, updatedDSR.CompletedAt)

	// Check Task Status
	tasks, _ := dsrRepo.GetTasksByDSR(ctx, dsrID)
	require.Len(t, tasks, 1)
	assert.Equal(t, compliance.TaskStatusCompleted, tasks[0].Status)

	// Check Task Result
	result, ok := tasks[0].Result.(map[string]interface{})
	require.True(t, ok)
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	users, ok := data["users"].(map[string]interface{})
	require.True(t, ok)
	emails, ok := users["email"].([]string)
	require.True(t, ok)
	assert.Contains(t, emails, "john@example.com")
	assert.NotContains(t, emails, "jane@example.com") // Should be filtered out

	// Check Events
	require.Len(t, eb.Events, 1)
	assert.Equal(t, eventbus.EventDSRCompleted, eb.Events[0].Type)
}

func TestExecuteDSR_Erasure(t *testing.T) {
	// Setup
	executor, dsrRepo, dsRepo, piiRepo, mockConn, eb := setupExecutorTest(t)
	ctx := context.Background()

	tenantID := types.NewID()
	dsID := types.NewID()
	dsrID := types.NewID()
	taskID := types.NewID()

	// 1. Create DSR
	dsr := &compliance.DSR{
		ID:                 dsrID,
		TenantID:           tenantID,
		RequestType:        compliance.RequestTypeErasure,
		Status:             compliance.DSRStatusApproved,
		SubjectIdentifiers: map[string]string{"email": "john@example.com"},
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	dsrRepo.Create(ctx, dsr)

	// 2. Create Task
	task := &compliance.DSRTask{
		ID:           taskID,
		DSRID:        dsrID,
		DataSourceID: dsID,
		TenantID:     tenantID,
		TaskType:     compliance.RequestTypeErasure,
		Status:       compliance.TaskStatusPending,
		CreatedAt:    time.Now(),
	}
	dsrRepo.CreateTask(ctx, task)

	// 3. Create Data Source
	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: dsID}, TenantID: tenantID},
		Name:         "Postgres DB",
		Type:         types.DataSourcePostgreSQL,
	}
	dsRepo.Create(ctx, ds)

	// 4. Create PII Classifications
	pii1 := &discovery.PIIClassification{
		BaseEntity:   types.BaseEntity{ID: types.NewID()},
		DataSourceID: dsID,
		EntityName:   "users",
		FieldName:    "email",
	}
	piiRepo.Create(ctx, pii1)

	// 5. Mock Interactions
	mockConn.On("Connect", ctx, mock.AnythingOfType("*discovery.DataSource")).Return(nil)
	mockConn.On("Close").Return(nil)

	// Execute
	err := executor.ExecuteDSR(ctx, dsrID)

	// Verify
	require.NoError(t, err)

	// Check DSR Status
	updatedDSR, _ := dsrRepo.GetByID(ctx, dsrID)
	assert.Equal(t, compliance.DSRStatusCompleted, updatedDSR.Status)

	// Check Task Result (Erasure Log)
	tasks, _ := dsrRepo.GetTasksByDSR(ctx, dsrID)
	require.Len(t, tasks, 1)
	result, ok := tasks[0].Result.(map[string]interface{})
	require.True(t, ok)
	deletions, ok := result["deletions"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, deletions, 1)
	assert.Equal(t, "users", deletions[0]["entity"])

	// Check Events (Should emit dsr.data_deleted)
	// DSRExecutor emits DSRCompleted at end, AND custom "dsr.data_deleted" inside executeErasureRequest
	// So we expect 2 events?
	// Looking at code: executeErasureRequest publishes "dsr.data_deleted". ExecuteDSR publishes "DSRCompleted".
	require.Len(t, eb.Events, 2)
	assert.Equal(t, "dsr.data_deleted", eb.Events[0].Type)
	assert.Equal(t, eventbus.EventDSRCompleted, eb.Events[1].Type)
}

func TestExecuteDSR_Correction_Stub(t *testing.T) {
	// Setup
	executor, dsrRepo, dsRepo, _, _, _ := setupExecutorTest(t)
	ctx := context.Background()

	tenantID := types.NewID()
	dsID := types.NewID()
	dsrID := types.NewID()
	taskID := types.NewID()

	dsr := &compliance.DSR{
		ID:          dsrID,
		TenantID:    tenantID,
		RequestType: compliance.RequestTypeCorrection,
		Status:      compliance.DSRStatusApproved,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	dsrRepo.Create(ctx, dsr)

	task := &compliance.DSRTask{
		ID:           taskID,
		DSRID:        dsrID,
		DataSourceID: dsID,
		TenantID:     tenantID,
		TaskType:     compliance.RequestTypeCorrection,
		Status:       compliance.TaskStatusPending,
		CreatedAt:    time.Now(),
	}
	dsrRepo.CreateTask(ctx, task)

	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: dsID}, TenantID: tenantID},
		Name:         "Postgres DB",
		Type:         types.DataSourcePostgreSQL,
	}
	dsRepo.Create(ctx, ds)

	// Execute
	err := executor.ExecuteDSR(ctx, dsrID)

	// Verify
	require.NoError(t, err)
	tasks, _ := dsrRepo.GetTasksByDSR(ctx, dsrID)
	require.Len(t, tasks, 1)
	assert.Equal(t, compliance.TaskStatusCompleted, tasks[0].Status)

	result, _ := tasks[0].Result.(map[string]interface{})
	assert.Contains(t, result["note"], "Correction capability requires connector Update() method")
}

func TestExecuteDSR_MultipleSources(t *testing.T) {
	// Setup
	executor, dsrRepo, dsRepo, piiRepo, mockConn, _ := setupExecutorTest(t)
	ctx := context.Background()

	tenantID := types.NewID()
	dsrID := types.NewID()

	dsr := &compliance.DSR{
		ID:                 dsrID,
		TenantID:           tenantID,
		RequestType:        compliance.RequestTypeAccess,
		Status:             compliance.DSRStatusApproved,
		SubjectIdentifiers: map[string]string{"email": "john@example.com"},
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	dsrRepo.Create(ctx, dsr)

	// Create 3 tasks for 3 data sources
	for i := 0; i < 3; i++ {
		dsID := types.NewID()
		ds := &discovery.DataSource{
			TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: dsID}, TenantID: tenantID},
			Name:         "DB",
			Type:         types.DataSourcePostgreSQL,
		}
		dsRepo.Create(ctx, ds)

		task := &compliance.DSRTask{
			ID:           types.NewID(),
			DSRID:        dsrID,
			DataSourceID: dsID,
			TenantID:     tenantID,
			TaskType:     compliance.RequestTypeAccess,
			Status:       compliance.TaskStatusPending,
		}
		dsrRepo.CreateTask(ctx, task)

		piiRepo.Create(ctx, &discovery.PIIClassification{
			BaseEntity:   types.BaseEntity{ID: types.NewID()},
			DataSourceID: dsID,
			EntityName:   "users",
			FieldName:    "email",
		})
	}

	// Mock Expect calls (3 times)
	mockConn.On("Connect", ctx, mock.Anything).Return(nil).Times(3)
	mockConn.On("SampleData", ctx, "users", "email", 100).Return([]string{"john@example.com"}, nil).Times(3)
	mockConn.On("Close").Return(nil).Times(3)

	// Execute
	err := executor.ExecuteDSR(ctx, dsrID)

	// Verify
	require.NoError(t, err)
	tasks, _ := dsrRepo.GetTasksByDSR(ctx, dsrID)
	assert.Len(t, tasks, 3)
	for _, task := range tasks {
		assert.Equal(t, compliance.TaskStatusCompleted, task.Status)
	}
}

func TestExecuteDSR_PartialFailure(t *testing.T) {
	// Setup
	executor, dsrRepo, dsRepo, piiRepo, mockConn, eb := setupExecutorTest(t)
	ctx := context.Background()

	tenantID := types.NewID()
	dsrID := types.NewID()

	dsr := &compliance.DSR{
		ID:                 dsrID,
		TenantID:           tenantID,
		RequestType:        compliance.RequestTypeAccess,
		Status:             compliance.DSRStatusApproved,
		SubjectIdentifiers: map[string]string{"email": "john@example.com"},
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	dsrRepo.Create(ctx, dsr)

	// Task 1: Success
	dsID1 := types.NewID()
	dsRepo.Create(ctx, &discovery.DataSource{TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: dsID1}, TenantID: tenantID}, Type: types.DataSourcePostgreSQL})
	dsrRepo.CreateTask(ctx, &compliance.DSRTask{ID: types.NewID(), DSRID: dsrID, DataSourceID: dsID1, TaskType: compliance.RequestTypeAccess, Status: compliance.TaskStatusPending})
	piiRepo.Create(ctx, &discovery.PIIClassification{BaseEntity: types.BaseEntity{ID: types.NewID()}, DataSourceID: dsID1, EntityName: "users", FieldName: "email"})

	// Task 2: Failure (Connect error)
	dsID2 := types.NewID()
	dsRepo.Create(ctx, &discovery.DataSource{TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: dsID2}, TenantID: tenantID}, Type: types.DataSourcePostgreSQL})
	dsrRepo.CreateTask(ctx, &compliance.DSRTask{ID: types.NewID(), DSRID: dsrID, DataSourceID: dsID2, TaskType: compliance.RequestTypeAccess, Status: compliance.TaskStatusPending})

	// Mock Expects
	// Task 1 calls
	mockConn.On("Connect", ctx, mock.MatchedBy(func(ds *discovery.DataSource) bool { return ds.ID == dsID1 })).Return(nil)
	mockConn.On("SampleData", ctx, "users", "email", 100).Return([]string{"john@example.com"}, nil)
	mockConn.On("Close").Return(nil)

	// Task 2 calls (Fail connect)
	mockConn.On("Connect", ctx, mock.MatchedBy(func(ds *discovery.DataSource) bool { return ds.ID == dsID2 })).Return(errors.New("connection failed"))

	// Execute
	err := executor.ExecuteDSR(ctx, dsrID)

	// Verify
	require.NoError(t, err) // Main execution logic does NOT return error, it updates status

	updatedDSR, _ := dsrRepo.GetByID(ctx, dsrID)
	assert.Equal(t, compliance.DSRStatusFailed, updatedDSR.Status)
	assert.Contains(t, updatedDSR.Reason, "1 task(s) failed")

	tasks, _ := dsrRepo.GetTasksByDSR(ctx, dsrID)
	require.Len(t, tasks, 2)

	for _, task := range tasks {
		if task.DataSourceID == dsID1 {
			assert.Equal(t, compliance.TaskStatusCompleted, task.Status)
		} else {
			assert.Equal(t, compliance.TaskStatusFailed, task.Status)
			assert.Contains(t, task.Error, "connect: connection failed")
		}
	}

	// Check Events (DSRFailed)
	assert.Equal(t, eventbus.EventDSRFailed, eb.Events[0].Type)
}
