package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/pkg/types"
)

func TestDSRRepo_CRUD(t *testing.T) {
	dsrRepo := repository.NewDSRRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	// 1. Setup Tenant
	tenant := &identity.Tenant{
		Name:     "DSRTestCo",
		Domain:   "dsrtest-" + types.NewID().String()[:8] + ".com",
		Plan:     identity.PlanFree,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{DefaultRegulation: "GDPR"},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	// 2. Create DSR
	dsr := &compliance.DSR{
		ID:                 types.NewID(),
		TenantID:           tenant.ID,
		RequestType:        compliance.RequestTypeErasure,
		Status:             compliance.DSRStatusPending,
		SubjectName:        "John Doe",
		SubjectEmail:       "john@example.com",
		SubjectIdentifiers: map[string]string{"user_id": "u_123"},
		Priority:           "HIGH",
		SLADeadline:        time.Now().Add(30 * 24 * time.Hour).UTC(),
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	err := dsrRepo.Create(ctx, dsr)
	require.NoError(t, err)

	// 3. GetByID
	got, err := dsrRepo.GetByID(ctx, dsr.ID)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", got.SubjectName)
	assert.Equal(t, compliance.RequestTypeErasure, got.RequestType)

	// 4. Update
	got.Status = compliance.DSRStatusApproved
	got.AssignedTo = &types.ID{} // Just a placeholder ID or nil
	err = dsrRepo.Update(ctx, got)
	require.NoError(t, err)

	got2, err := dsrRepo.GetByID(ctx, dsr.ID)
	require.NoError(t, err)
	assert.Equal(t, compliance.DSRStatusApproved, got2.Status)

	// 5. ListByTenant (Pagination & Filtering)
	// Add another DSR with different status
	dsr2 := &compliance.DSR{
		ID:           types.NewID(),
		TenantID:     tenant.ID,
		RequestType:  compliance.RequestTypeAccess,
		Status:       compliance.DSRStatusRejected,
		SubjectName:  "Jane Doe",
		SubjectEmail: "jane@example.com",
		SLADeadline:  time.Now().Add(30 * 24 * time.Hour).UTC(),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	require.NoError(t, dsrRepo.Create(ctx, dsr2))

	// Filter by APPROVED (should find dsr1)
	statusFilter := compliance.DSRStatusApproved
	page, err := dsrRepo.GetByTenant(ctx, tenant.ID, types.Pagination{Page: 1, PageSize: 10}, &statusFilter, nil)
	require.NoError(t, err)
	assert.Len(t, page.Items, 1)
	assert.Equal(t, dsr.ID, page.Items[0].ID)

	// Filter by REJECTED (should find dsr2)
	statusFilter2 := compliance.DSRStatusRejected
	page2, err := dsrRepo.GetByTenant(ctx, tenant.ID, types.Pagination{Page: 1, PageSize: 10}, &statusFilter2, nil)
	require.NoError(t, err)
	assert.Len(t, page2.Items, 1)
	assert.Equal(t, dsr2.ID, page2.Items[0].ID)

	// No filter (should find both)
	pageAll, err := dsrRepo.GetByTenant(ctx, tenant.ID, types.Pagination{Page: 1, PageSize: 10}, nil, nil)
	require.NoError(t, err)
	assert.Len(t, pageAll.Items, 2)
}

func TestDSRRepo_GetOverdue(t *testing.T) {
	dsrRepo := repository.NewDSRRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	tenant := &identity.Tenant{
		Name:     "OverdueCo",
		Domain:   "overdue-" + types.NewID().String()[:8] + ".com",
		Plan:     identity.PlanFree,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{DefaultRegulation: "GDPR"},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	// 1. Normal active DSR
	active := &compliance.DSR{
		ID:           types.NewID(),
		TenantID:     tenant.ID,
		RequestType:  compliance.RequestTypeAccess,
		Status:       compliance.DSRStatusPending,
		SubjectName:  "Active User",
		SubjectEmail: "active@example.com",
		SLADeadline:  time.Now().Add(24 * time.Hour).UTC(), // Future
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	require.NoError(t, dsrRepo.Create(ctx, active))

	// 2. Overdue DSR
	overdue := &compliance.DSR{
		ID:           types.NewID(),
		TenantID:     tenant.ID,
		RequestType:  compliance.RequestTypeAccess,
		Status:       compliance.DSRStatusPending, // Still pending
		SubjectName:  "Overdue User",
		SubjectEmail: "overdue@example.com",
		SLADeadline:  time.Now().Add(-24 * time.Hour).UTC(), // Past
		CreatedAt:    time.Now().Add(-35 * 24 * time.Hour).UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	require.NoError(t, dsrRepo.Create(ctx, overdue))

	// 3. Completed DSR (past deadline but completed, so NOT overdue)
	completed := &compliance.DSR{
		ID:           types.NewID(),
		TenantID:     tenant.ID,
		RequestType:  compliance.RequestTypeAccess,
		Status:       compliance.DSRStatusCompleted,
		SubjectName:  "Completed User",
		SubjectEmail: "completed@example.com",
		SLADeadline:  time.Now().Add(-24 * time.Hour).UTC(), // Past
		CreatedAt:    time.Now().Add(-35 * 24 * time.Hour).UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	require.NoError(t, dsrRepo.Create(ctx, completed))

	// Test GetOverdue
	results, err := dsrRepo.GetOverdue(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, overdue.ID, results[0].ID)
}

func TestDSRRepo_TaskCRUD(t *testing.T) {
	dsrRepo := repository.NewDSRRepo(testPool)
	dsRepo := repository.NewDataSourceRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	// Setup Tenant & DSR & DataSource
	tenant := &identity.Tenant{
		Name:     "TaskTestCo",
		Domain:   "tasktest-" + types.NewID().String()[:8] + ".com",
		Plan:     identity.PlanFree,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{DefaultRegulation: "GDPR"},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	ds := &discovery.DataSource{
		Name:     "Task DB",
		Type:     types.DataSourcePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Status:   discovery.ConnectionStatusConnected,
	}
	ds.TenantID = tenant.ID
	require.NoError(t, dsRepo.Create(ctx, ds))

	dsr := &compliance.DSR{
		ID:           types.NewID(),
		TenantID:     tenant.ID,
		RequestType:  compliance.RequestTypeErasure,
		Status:       compliance.DSRStatusApproved, // Tasks usually created after approval
		SubjectName:  "Task User",
		SubjectEmail: "task@example.com",
		SLADeadline:  time.Now().Add(30 * 24 * time.Hour).UTC(),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	require.NoError(t, dsrRepo.Create(ctx, dsr))

	// 1. Create Task
	task := &compliance.DSRTask{
		ID:           types.NewID(),
		DSRID:        dsr.ID,
		DataSourceID: ds.ID,
		TenantID:     tenant.ID,
		TaskType:     compliance.RequestTypeErasure,
		Status:       compliance.TaskStatusPending,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	err := dsrRepo.CreateTask(ctx, task)
	require.NoError(t, err)

	// 2. GetTasksByDSR
	tasks, err := dsrRepo.GetTasksByDSR(ctx, dsr.ID)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, task.ID, tasks[0].ID)
	assert.Equal(t, compliance.TaskStatusPending, tasks[0].Status)

	// 3. Update Task
	task.Status = compliance.TaskStatusCompleted
	task.Result = map[string]any{"rows_deleted": 5}
	completedAt := time.Now().UTC()
	task.CompletedAt = &completedAt

	err = dsrRepo.UpdateTask(ctx, task)
	require.NoError(t, err)

	tasks2, err := dsrRepo.GetTasksByDSR(ctx, dsr.ID)
	require.NoError(t, err)
	assert.Equal(t, compliance.TaskStatusCompleted, tasks2[0].Status)
	assert.NotNil(t, tasks2[0].CompletedAt)
}
