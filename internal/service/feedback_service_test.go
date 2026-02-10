package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

func TestFeedbackService_SubmitFeedback_Verified(t *testing.T) {
	// Setup
	fbRepo := newMockDetectionFeedbackRepo()
	piiRepo := newMockPIIClassificationRepo()
	eb := newMockEventBus()
	svc := NewFeedbackService(fbRepo, piiRepo, eb, slog.Default())

	ctx := context.Background()
	tenantID := types.NewID()
	userID := types.NewID()
	classificationID := types.NewID()

	// Seed Classification
	existing := &discovery.PIIClassification{
		BaseEntity: types.BaseEntity{
			ID: classificationID,
		},
		Category:        types.PIICategoryFinancial,
		Type:            types.PIITypeCreditCard,
		Confidence:      0.8,
		DetectionMethod: types.DetectionMethodAI,
		Status:          types.VerificationPending,
		FieldName:       "cc_num",
		EntityName:      "payments",
	}
	require.NoError(t, piiRepo.Create(ctx, existing))

	// Execute
	input := SubmitFeedbackInput{
		ClassificationID: classificationID,
		FeedbackType:     discovery.FeedbackVerified,
		Notes:            "Looks correct",
	}
	resp, err := svc.SubmitFeedback(ctx, tenantID, userID, input)
	require.NoError(t, err)

	// Verify Response
	assert.Equal(t, discovery.FeedbackVerified, resp.Feedback.FeedbackType)
	assert.Equal(t, types.VerificationVerified, resp.Classification.Status)

	// Verify Persistence (Feedback)
	fb, err := fbRepo.GetByID(ctx, resp.Feedback.ID)
	require.NoError(t, err)
	assert.Equal(t, discovery.FeedbackVerified, fb.FeedbackType)
	assert.Equal(t, classificationID, fb.ClassificationID)

	// Verify Persistence (Classification Update)
	updated, err := piiRepo.GetByID(ctx, classificationID)
	require.NoError(t, err)
	assert.Equal(t, types.VerificationVerified, updated.Status)
	assert.Equal(t, userID, *updated.VerifiedBy)

	// Verify Event
	assert.Len(t, eb.Events, 1)
	assert.Equal(t, eventbus.EventPIIVerified, eb.Events[0].Type)
}

func TestFeedbackService_SubmitFeedback_Corrected(t *testing.T) {
	// Setup
	fbRepo := newMockDetectionFeedbackRepo()
	piiRepo := newMockPIIClassificationRepo()
	eb := newMockEventBus()
	svc := NewFeedbackService(fbRepo, piiRepo, eb, slog.Default())

	ctx := context.Background()
	tenantID := types.NewID()
	userID := types.NewID()
	classificationID := types.NewID()

	existing := &discovery.PIIClassification{
		BaseEntity: types.BaseEntity{
			ID: classificationID,
		},
		Category:   types.PIICategoryFinancial,
		Type:       types.PIITypeCreditCard,
		Confidence: 0.8,
		Status:     types.VerificationPending,
	}
	require.NoError(t, piiRepo.Create(ctx, existing))

	// Execute (Correct to Email)
	newCat := types.PIICategoryContact
	newType := types.PIITypeEmail
	input := SubmitFeedbackInput{
		ClassificationID:  classificationID,
		FeedbackType:      discovery.FeedbackCorrected,
		CorrectedCategory: &newCat,
		CorrectedType:     &newType,
		Notes:             "Actually an email",
	}
	resp, err := svc.SubmitFeedback(ctx, tenantID, userID, input)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, discovery.FeedbackCorrected, resp.Feedback.FeedbackType)
	assert.Equal(t, types.VerificationVerified, resp.Classification.Status)
	assert.Equal(t, types.PIITypeEmail, resp.Classification.Type)
	assert.Equal(t, 1.0, resp.Classification.Confidence)
	assert.Equal(t, types.DetectionMethodManual, resp.Classification.DetectionMethod)

	// Event
	assert.Len(t, eb.Events, 1)
	assert.Equal(t, eventbus.EventPIICorrected, eb.Events[0].Type)
}

func TestFeedbackService_SubmitFeedback_Rejected(t *testing.T) {
	// Setup
	fbRepo := newMockDetectionFeedbackRepo()
	piiRepo := newMockPIIClassificationRepo()
	eb := newMockEventBus()
	svc := NewFeedbackService(fbRepo, piiRepo, eb, slog.Default())

	ctx := context.Background()
	tenantID := types.NewID()
	userID := types.NewID()
	classificationID := types.NewID()

	existing := &discovery.PIIClassification{
		BaseEntity: types.BaseEntity{
			ID: classificationID,
		},
		Status:     types.VerificationPending,
		Confidence: 0.5,
	}
	require.NoError(t, piiRepo.Create(ctx, existing))

	// Execute
	input := SubmitFeedbackInput{
		ClassificationID: classificationID,
		FeedbackType:     discovery.FeedbackRejected,
		Notes:            "Not PII",
	}
	resp, err := svc.SubmitFeedback(ctx, tenantID, userID, input)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, types.VerificationRejected, resp.Classification.Status)

	// Event
	assert.Len(t, eb.Events, 1)
	assert.Equal(t, eventbus.EventPIIRejected, eb.Events[0].Type)
}

func TestFeedbackService_SubmitFeedback_Validation(t *testing.T) {
	svc := NewFeedbackService(nil, nil, nil, slog.Default())
	ctx := context.Background()

	// Missing ID
	_, err := svc.SubmitFeedback(ctx, types.NewID(), types.NewID(), SubmitFeedbackInput{
		FeedbackType: discovery.FeedbackVerified,
	})
	assert.Error(t, err)

	// Invalid Enum
	_, err = svc.SubmitFeedback(ctx, types.NewID(), types.NewID(), SubmitFeedbackInput{
		ClassificationID: types.NewID(),
		FeedbackType:     "INVALID_TYPE",
	})
	assert.Error(t, err)

	// Corrected without details
	_, err = svc.SubmitFeedback(ctx, types.NewID(), types.NewID(), SubmitFeedbackInput{
		ClassificationID: types.NewID(),
		FeedbackType:     discovery.FeedbackCorrected,
		// Missing CorrectedCategory/Type
	})
	assert.Error(t, err)
}
