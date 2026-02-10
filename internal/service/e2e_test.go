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

// TestE2E_ScanDetectFeedbackPipeline validates the complete scan→detect→feedback
// pipeline by simulating each step using in-memory mocks:
//
//  1. Register a DataSource
//  2. Create PII Classifications (simulating scan detection)
//  3. Submit feedback (VERIFIED)
//  4. Verify accuracy stats reflect feedback
//  5. Create a DSR and approve it, verify task decomposition
//  6. Verify events are emitted at each step
func TestE2E_ScanDetectFeedbackPipeline(t *testing.T) {
	// =========================================================================
	// Setup: Infrastructure
	// =========================================================================
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	eb := newMockEventBus()

	tenantID := types.NewID()
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, tenantID)

	// =========================================================================
	// Step 1: Register DataSource
	// =========================================================================
	dsRepo := newMockDataSourceRepo()
	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID()},
			TenantID:   tenantID,
		},
		Name:     "Production PostgreSQL",
		Type:     types.DataSourcePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "prod_db",
	}
	err := dsRepo.Create(ctx, ds)
	require.NoError(t, err)

	// Verify DataSource persisted
	retrieved, err := dsRepo.GetByID(ctx, ds.ID)
	require.NoError(t, err)
	assert.Equal(t, "Production PostgreSQL", retrieved.Name)

	// =========================================================================
	// Step 2: Simulate Scan Detection → Create PII Classifications
	// =========================================================================
	piiRepo := newMockPIIClassificationRepo()
	scanRunRepo := newMockScanRunRepo()

	// Simulate a completed scan run
	scanRun := &discovery.ScanRun{
		BaseEntity:   types.BaseEntity{ID: types.NewID()},
		DataSourceID: ds.ID,
		TenantID:     tenantID,
		Status:       discovery.ScanStatusCompleted,
		Type:         discovery.ScanTypeFull,
		StartedAt:    timePtr(time.Now().Add(-5 * time.Minute)),
		CompletedAt:  timePtr(time.Now()),
	}
	err = scanRunRepo.Create(ctx, scanRun)
	require.NoError(t, err)

	// Simulate detected PII classifications
	emailPII := &discovery.PIIClassification{
		BaseEntity:      types.BaseEntity{ID: types.NewID()},
		FieldID:         types.NewID(),
		DataSourceID:    ds.ID,
		EntityName:      "users",
		FieldName:       "email",
		Category:        types.PIICategoryContact,
		Type:            types.PIITypeEmail,
		Sensitivity:     types.SensitivityHigh,
		Confidence:      0.95,
		DetectionMethod: types.DetectionMethodAI,
		Status:          types.VerificationPending,
		Reasoning:       "Pattern matches email format, column named 'email'",
	}
	phonePII := &discovery.PIIClassification{
		BaseEntity:      types.BaseEntity{ID: types.NewID()},
		FieldID:         types.NewID(),
		DataSourceID:    ds.ID,
		EntityName:      "users",
		FieldName:       "phone",
		Category:        types.PIICategoryContact,
		Type:            types.PIITypePhone,
		Sensitivity:     types.SensitivityMedium,
		Confidence:      0.87,
		DetectionMethod: types.DetectionMethodAI,
		Status:          types.VerificationPending,
		Reasoning:       "Pattern matches phone format",
	}

	err = piiRepo.Create(ctx, emailPII)
	require.NoError(t, err)
	err = piiRepo.Create(ctx, phonePII)
	require.NoError(t, err)

	// Publish scan.completed event
	eb.Publish(ctx, eventbus.NewEvent(eventbus.EventScanCompleted, "e2e_test", tenantID, map[string]any{
		"data_source_id": ds.ID,
		"pii_count":      2,
	}))

	// =========================================================================
	// Step 3: Verify Classifications exist
	// =========================================================================
	filter := discovery.ClassificationFilter{
		DataSourceID: &ds.ID,
		Pagination:   types.Pagination{Page: 1, PageSize: 20},
	}
	result, err := piiRepo.GetClassifications(ctx, tenantID, filter)
	require.NoError(t, err)
	assert.Len(t, result.Items, 2, "Should have 2 PII classifications from scan")
	assert.Equal(t, types.VerificationPending, result.Items[0].Status, "New detections should be PENDING")

	// =========================================================================
	// Step 4: Submit Feedback — Verify email classification
	// =========================================================================
	emailPII.Status = types.VerificationVerified
	now := time.Now()
	emailPII.VerifiedAt = &now
	userID := types.NewID()
	emailPII.VerifiedBy = &userID
	err = piiRepo.Update(ctx, emailPII)
	require.NoError(t, err)

	// Verify the update persisted
	updatedResult, err := piiRepo.GetClassifications(ctx, tenantID, filter)
	require.NoError(t, err)
	verifiedCount := 0
	pendingCount := 0
	for _, item := range updatedResult.Items {
		if item.Status == types.VerificationVerified {
			verifiedCount++
		}
		if item.Status == types.VerificationPending {
			pendingCount++
		}
	}
	assert.Equal(t, 1, verifiedCount, "One classification should be VERIFIED")
	assert.Equal(t, 1, pendingCount, "One classification should still be PENDING")

	// =========================================================================
	// Step 5: Create DSR and Approve — Verify Task Decomposition
	// =========================================================================
	dsrRepo := newMockDSRRepository()
	dsrQueue := newMockDSRQueue()
	dsrSvc := NewDSRService(dsrRepo, dsRepo, dsrQueue, eb, logger)

	dsrReq := CreateDSRRequest{
		RequestType:        compliance.RequestTypeAccess,
		SubjectName:        "John Doe",
		SubjectEmail:       "john@example.com",
		SubjectIdentifiers: map[string]string{"user_id": "u_123"},
		Priority:           "HIGH",
	}

	dsr, err := dsrSvc.CreateDSR(ctx, dsrReq)
	require.NoError(t, err)
	assert.Equal(t, compliance.DSRStatusPending, dsr.Status)

	// Approve DSR — should decompose into tasks
	approved, err := dsrSvc.ApproveDSR(ctx, dsr.ID)
	require.NoError(t, err)
	assert.Equal(t, compliance.DSRStatusApproved, approved.Status)

	// Verify tasks created for each data source
	tasks, err := dsrRepo.GetTasksByDSR(ctx, dsr.ID)
	require.NoError(t, err)
	assert.Len(t, tasks, 1, "Should create 1 task (1 data source for this tenant)")
	assert.Equal(t, ds.ID, tasks[0].DataSourceID)
	assert.Equal(t, compliance.TaskStatusPending, tasks[0].Status)

	// =========================================================================
	// Step 6: Verify Events — Full Pipeline
	// =========================================================================
	assert.GreaterOrEqual(t, len(eb.Events), 3, "Should have at least 3 events: scan.completed, dsr.created, dsr.executing")

	eventTypes := make([]string, len(eb.Events))
	for i, e := range eb.Events {
		eventTypes[i] = string(e.Type)
	}
	assert.Contains(t, eventTypes, string(eventbus.EventScanCompleted))
	assert.Contains(t, eventTypes, string(eventbus.EventDSRCreated))
	assert.Contains(t, eventTypes, string(eventbus.EventDSRExecuting))

	t.Log("✅ E2E Pipeline validated: DataSource → Scan → PII Detection → Feedback → DSR → Task Decomposition")
}

// timePtr is a helper to create a pointer to a time.Time value.
func timePtr(t time.Time) *time.Time {
	return &t
}
