package discovery

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// DetectionFeedback — Human feedback on PII detection results
// =============================================================================

// FeedbackType classifies the kind of human feedback.
type FeedbackType string

const (
	// FeedbackVerified means the detection was correct as-is.
	FeedbackVerified FeedbackType = "VERIFIED"
	// FeedbackCorrected means the detection was partially wrong and was corrected.
	FeedbackCorrected FeedbackType = "CORRECTED"
	// FeedbackRejected means the detection was entirely wrong.
	FeedbackRejected FeedbackType = "REJECTED"
)

// DetectionFeedback records human review of a PIIClassification result.
// This feeds back into the learning loop to improve detection accuracy.
type DetectionFeedback struct {
	types.BaseEntity
	ClassificationID types.ID     `json:"classification_id" db:"classification_id"`
	TenantID         types.ID     `json:"tenant_id" db:"tenant_id"`
	FeedbackType     FeedbackType `json:"feedback_type" db:"feedback_type"`

	// Original detection result (stored for learning comparison)
	OriginalCategory   types.PIICategory     `json:"original_category" db:"original_category"`
	OriginalType       types.PIIType         `json:"original_type" db:"original_type"`
	OriginalConfidence float64               `json:"original_confidence" db:"original_confidence"`
	OriginalMethod     types.DetectionMethod `json:"original_method" db:"original_method"`

	// Corrected values (populated only for CORRECTED feedback)
	CorrectedCategory *types.PIICategory `json:"corrected_category,omitempty" db:"corrected_category"`
	CorrectedType     *types.PIIType     `json:"corrected_type,omitempty" db:"corrected_type"`

	// Metadata
	CorrectedBy types.ID  `json:"corrected_by" db:"corrected_by"`
	CorrectedAt time.Time `json:"corrected_at" db:"corrected_at"`
	Notes       string    `json:"notes,omitempty" db:"notes"`

	// Context for learning — column/table info for pattern extraction
	ColumnName string `json:"column_name" db:"column_name"`
	TableName  string `json:"table_name" db:"table_name"`
	DataType   string `json:"data_type" db:"data_type"`
}

// =============================================================================
// Repository Interface
// =============================================================================

// DetectionFeedbackRepository defines persistence for detection feedback.
type DetectionFeedbackRepository interface {
	Create(ctx context.Context, feedback *DetectionFeedback) error
	GetByID(ctx context.Context, id types.ID) (*DetectionFeedback, error)
	GetByClassification(ctx context.Context, classificationID types.ID) ([]DetectionFeedback, error)
	GetByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[DetectionFeedback], error)

	// GetCorrectionPatterns returns feedback where column names match a pattern,
	// used by the learning service to extract rules from repeated corrections.
	GetCorrectionPatterns(ctx context.Context, tenantID types.ID, columnPattern string) ([]DetectionFeedback, error)

	// GetAccuracyStats returns feedback counts by type for a strategy,
	// used to track per-strategy accuracy metrics.
	GetAccuracyStats(ctx context.Context, tenantID types.ID, method types.DetectionMethod) (*AccuracyStats, error)
}

// AccuracyStats summarizes detection accuracy based on human feedback.
type AccuracyStats struct {
	Method    types.DetectionMethod `json:"method"`
	Total     int                   `json:"total"`
	Verified  int                   `json:"verified"`
	Corrected int                   `json:"corrected"`
	Rejected  int                   `json:"rejected"`
	Accuracy  float64               `json:"accuracy"` // Verified / Total
}
