package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// FeedbackService handles the verify/correct/reject workflow for PII detections.
type FeedbackService struct {
	feedbackRepo discovery.DetectionFeedbackRepository
	piiRepo      discovery.PIIClassificationRepository
	eventBus     eventbus.EventBus
	logger       *slog.Logger
}

// NewFeedbackService creates a new FeedbackService.
func NewFeedbackService(
	feedbackRepo discovery.DetectionFeedbackRepository,
	piiRepo discovery.PIIClassificationRepository,
	eb eventbus.EventBus,
	logger *slog.Logger,
) *FeedbackService {
	return &FeedbackService{
		feedbackRepo: feedbackRepo,
		piiRepo:      piiRepo,
		eventBus:     eb,
		logger:       logger.With("service", "feedback"),
	}
}

// =============================================================================
// DTOs
// =============================================================================

// SubmitFeedbackInput is the DTO for submitting feedback on a PII detection.
type SubmitFeedbackInput struct {
	ClassificationID  types.ID               `json:"classification_id"`
	FeedbackType      discovery.FeedbackType `json:"feedback_type"`
	CorrectedCategory *types.PIICategory     `json:"corrected_category,omitempty"`
	CorrectedType     *types.PIIType         `json:"corrected_type,omitempty"`
	Notes             string                 `json:"notes"`
}

// Validate checks the SubmitFeedbackInput for correctness.
func (in *SubmitFeedbackInput) Validate() error {
	ve := &types.ValidationErrors{}

	if in.ClassificationID == (types.ID{}) {
		ve.Add("classification_id", "classification_id is required")
	}

	switch in.FeedbackType {
	case discovery.FeedbackVerified, discovery.FeedbackCorrected, discovery.FeedbackRejected:
		// valid
	default:
		ve.Add("feedback_type", "must be VERIFIED, CORRECTED, or REJECTED")
	}

	if in.FeedbackType == discovery.FeedbackCorrected {
		if in.CorrectedCategory == nil {
			ve.Add("corrected_category", "required when feedback_type is CORRECTED")
		}
		if in.CorrectedType == nil {
			ve.Add("corrected_type", "required when feedback_type is CORRECTED")
		}
	}

	if ve.HasErrors() {
		return ve.ToDomainError()
	}
	return nil
}

// FeedbackResponse wraps a feedback record with the updated classification.
type FeedbackResponse struct {
	Feedback       *discovery.DetectionFeedback `json:"feedback"`
	Classification *discovery.PIIClassification `json:"classification"`
}

// =============================================================================
// Operations
// =============================================================================

// SubmitFeedback processes a verify/correct/reject action on a PII classification.
// It records the feedback, updates the classification status, and publishes events.
func (s *FeedbackService) SubmitFeedback(ctx context.Context, tenantID, userID types.ID, in SubmitFeedbackInput) (*FeedbackResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	// 1. Fetch the classification being reviewed
	classification, err := s.piiRepo.GetByID(ctx, in.ClassificationID)
	if err != nil {
		return nil, err
	}

	// 2. Build the feedback record
	now := time.Now().UTC()
	feedback := &discovery.DetectionFeedback{
		ClassificationID:   in.ClassificationID,
		TenantID:           tenantID,
		FeedbackType:       in.FeedbackType,
		OriginalCategory:   classification.Category,
		OriginalType:       classification.Type,
		OriginalConfidence: classification.Confidence,
		OriginalMethod:     classification.DetectionMethod,
		CorrectedCategory:  in.CorrectedCategory,
		CorrectedType:      in.CorrectedType,
		CorrectedBy:        userID,
		CorrectedAt:        now,
		Notes:              in.Notes,
		ColumnName:         classification.FieldName,
		TableName:          classification.EntityName,
	}

	// 3. Persist the feedback
	if err := s.feedbackRepo.Create(ctx, feedback); err != nil {
		return nil, fmt.Errorf("create feedback: %w", err)
	}

	// 4. Update the classification based on feedback type
	switch in.FeedbackType {
	case discovery.FeedbackVerified:
		classification.Status = types.VerificationVerified
		classification.VerifiedBy = &userID
		classification.VerifiedAt = &now

	case discovery.FeedbackCorrected:
		classification.Status = types.VerificationVerified
		classification.VerifiedBy = &userID
		classification.VerifiedAt = &now
		// Apply the corrected values
		if in.CorrectedCategory != nil {
			classification.Category = *in.CorrectedCategory
		}
		if in.CorrectedType != nil {
			classification.Type = *in.CorrectedType
		}
		classification.Confidence = 1.0 // Human-verified = 100% confidence
		classification.DetectionMethod = types.DetectionMethodManual

	case discovery.FeedbackRejected:
		classification.Status = types.VerificationRejected
		classification.VerifiedBy = &userID
		classification.VerifiedAt = &now
	}

	if err := s.piiRepo.Update(ctx, classification); err != nil {
		return nil, fmt.Errorf("update classification: %w", err)
	}

	// 5. Publish event
	eventType := s.feedbackToEventType(in.FeedbackType)
	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(
		eventType, "discovery", tenantID,
		map[string]any{
			"feedback_id":       feedback.ID,
			"classification_id": in.ClassificationID,
			"feedback_type":     string(in.FeedbackType),
			"field_name":        classification.FieldName,
			"entity_name":       classification.EntityName,
		},
	))

	s.logger.InfoContext(ctx, "feedback submitted",
		slog.String("tenant_id", tenantID.String()),
		slog.String("classification_id", in.ClassificationID.String()),
		slog.String("feedback_type", string(in.FeedbackType)),
		slog.String("field_name", classification.FieldName),
	)

	return &FeedbackResponse{
		Feedback:       feedback,
		Classification: classification,
	}, nil
}

// GetByClassification returns all feedback for a specific classification.
func (s *FeedbackService) GetByClassification(ctx context.Context, classificationID types.ID) ([]discovery.DetectionFeedback, error) {
	return s.feedbackRepo.GetByClassification(ctx, classificationID)
}

// ListByTenant returns paginated feedback for a tenant.
func (s *FeedbackService) ListByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[discovery.DetectionFeedback], error) {
	return s.feedbackRepo.GetByTenant(ctx, tenantID, pagination)
}

// GetAccuracyStats returns detection accuracy metrics for a given method.
func (s *FeedbackService) GetAccuracyStats(ctx context.Context, tenantID types.ID, method types.DetectionMethod) (*discovery.AccuracyStats, error) {
	return s.feedbackRepo.GetAccuracyStats(ctx, tenantID, method)
}

func (s *FeedbackService) feedbackToEventType(ft discovery.FeedbackType) string {
	switch ft {
	case discovery.FeedbackVerified:
		return eventbus.EventPIIVerified
	case discovery.FeedbackCorrected:
		return eventbus.EventPIICorrected
	case discovery.FeedbackRejected:
		return eventbus.EventPIIRejected
	default:
		return eventbus.EventPIIVerified
	}
}
