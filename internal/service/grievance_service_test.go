package service

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrievanceService_Lifecycle(t *testing.T) {
	repo := newMockGrievanceRepo()
	eventBus := newMockEventBus()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewGrievanceService(repo, eventBus, logger)

	ctx := context.Background()
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	subjectID := types.NewID()

	// 1. Submit
	req := CreateGrievanceRequest{
		Subject:       "Consent Issue",
		Description:   "I cannot withdraw consent.",
		Category:      "CONSENT",
		DataSubjectID: subjectID.String(),
	}

	grievance, err := svc.SubmitGrievance(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, compliance.GrievanceStatusOpen, grievance.Status)
	assert.NotNil(t, grievance.DueDate)
	// Verify SLA: DueDate should be roughly 30 days from now
	expectedDue := time.Now().Add(30 * 24 * time.Hour)
	assert.WithinDuration(t, expectedDue, *grievance.DueDate, 5*time.Second)

	// 2. Assign
	dpoID := types.NewID()
	err = svc.AssignGrievance(ctx, grievance.ID, dpoID)
	require.NoError(t, err)

	updated, _ := svc.GetGrievance(ctx, grievance.ID)
	assert.Equal(t, compliance.GrievanceStatusInProgress, updated.Status)
	assert.Equal(t, dpoID, *updated.AssignedTo)

	// 3. Resolve
	err = svc.ResolveGrievance(ctx, grievance.ID, "Fixed the button.")
	require.NoError(t, err)

	resolved, _ := svc.GetGrievance(ctx, grievance.ID)
	assert.Equal(t, compliance.GrievanceStatusResolved, resolved.Status)
	assert.NotNil(t, resolved.ResolvedAt)
	assert.Equal(t, "Fixed the button.", *resolved.Resolution)

	// 4. Feedback
	err = svc.SubmitFeedback(ctx, grievance.ID, 5, "Great service")
	require.NoError(t, err)

	closed, _ := svc.GetGrievance(ctx, grievance.ID)
	assert.Equal(t, compliance.GrievanceStatusClosed, closed.Status)
	assert.Equal(t, 5, *closed.FeedbackRating)
}

func TestGrievanceService_Escalate(t *testing.T) {
	repo := newMockGrievanceRepo()
	svc := NewGrievanceService(repo, newMockEventBus(), slog.Default())
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, types.NewID())

	g, err := svc.SubmitGrievance(ctx, CreateGrievanceRequest{Subject: "S", Description: "D", DataSubjectID: types.NewID().String()})
	require.NoError(t, err)

	err = svc.EscalateGrievance(ctx, g.ID, "DPBI (Data Protection Board of India)")
	require.NoError(t, err)

	updated, _ := svc.GetGrievance(ctx, g.ID)
	assert.Equal(t, compliance.GrievanceStatusEscalated, updated.Status)
	assert.Equal(t, "DPBI (Data Protection Board of India)", *updated.EscalatedTo)
}

func TestGrievanceService_InvalidTransitions(t *testing.T) {
	repo := newMockGrievanceRepo()
	svc := NewGrievanceService(repo, newMockEventBus(), slog.Default())
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, types.NewID())

	g, _ := svc.SubmitGrievance(ctx, CreateGrievanceRequest{Subject: "S", Description: "D", DataSubjectID: types.NewID().String()})

	// Try to submit feedback on OPEN grievance (should fail, needs RESOLVED)
	err := svc.SubmitFeedback(ctx, g.ID, 1, "Bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only submit feedback for resolved")
}

func TestGrievanceService_TenantIsolation(t *testing.T) {
	repo := newMockGrievanceRepo()
	svc := NewGrievanceService(repo, newMockEventBus(), slog.Default())

	tenant1 := types.NewID()
	tenant2 := types.NewID()

	// Tenant 1 creates grievance
	ctx1 := context.WithValue(context.Background(), types.ContextKeyTenantID, tenant1)
	g1, _ := svc.SubmitGrievance(ctx1, CreateGrievanceRequest{Subject: "T1", Description: "D", DataSubjectID: types.NewID().String()})

	// Tenant 2 tries to access it
	ctx2 := context.WithValue(context.Background(), types.ContextKeyTenantID, tenant2)
	_, err := svc.GetGrievance(ctx2, g1.ID)
	require.Error(t, err)
	assert.True(t, types.IsNotFoundError(err)) // Should be Not Found, not Forbidden/Unauthorized for isolation
}
