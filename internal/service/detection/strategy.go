// Package detection implements the composable PII detection engine.
// Multiple detection strategies (AI, regex, heuristic, industry) are
// chained together by the ComposableDetector, which merges their
// results using weighted confidence scoring.
package detection

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Detection Strategy Interface
// =============================================================================

// Strategy defines a single PII detection method. Each strategy
// independently analyzes input and produces detection results.
// The ComposableDetector chains multiple strategies.
type Strategy interface {
	// Name returns a human-readable strategy identifier.
	Name() string

	// Method returns the DetectionMethod enum for this strategy.
	Method() types.DetectionMethod

	// Detect analyzes the input and returns zero or more detections.
	Detect(ctx context.Context, input Input) ([]Result, error)

	// Weight returns the confidence weight for this strategy (0.0–1.0).
	// Used by ComposableDetector to merge results across strategies.
	Weight() float64
}

// =============================================================================
// Detection Input / Output Types
// =============================================================================

// Input contains all information available for PII detection on a single field.
type Input struct {
	// Column metadata
	ColumnName string `json:"column_name"`
	DataType   string `json:"data_type"`
	Nullable   bool   `json:"nullable"`
	TableName  string `json:"table_name,omitempty"`
	SchemaName string `json:"schema_name,omitempty"`

	// Sample values (sanitized for AI strategies, raw for regex/heuristic)
	Samples          []string `json:"samples,omitempty"`
	SanitizedSamples []string `json:"sanitized_samples,omitempty"`

	// Context from adjacent columns (helps AI understand relationships)
	AdjacentColumns []ColumnContext `json:"adjacent_columns,omitempty"`

	// Industry context for sector-specific detection
	Industry string `json:"industry,omitempty"`
}

// ColumnContext provides minimal metadata about a neighboring column.
type ColumnContext struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

// Result holds a single PII detection finding from one strategy.
type Result struct {
	Category    types.PIICategory      `json:"category"`
	Type        types.PIIType          `json:"type"`
	Sensitivity types.SensitivityLevel `json:"sensitivity"`
	Confidence  float64                `json:"confidence"` // 0.0–1.0
	Method      types.DetectionMethod  `json:"method"`
	Reasoning   string                 `json:"reasoning"`
}

// =============================================================================
// Composable Detector Interface
// =============================================================================

// Detector is the top-level PII detection engine. It chains multiple
// strategies and merges their results into a unified detection report.
type Detector interface {
	// Detect runs all registered strategies against the input and
	// produces a merged, deduplicated, confidence-scored report.
	Detect(ctx context.Context, input Input) (*Report, error)
}

// Report is the final output of the composable detector for one field.
type Report struct {
	ColumnName string            `json:"column_name"`
	IsPII      bool              `json:"is_pii"`
	Detections []MergedDetection `json:"detections"`
	TopMatch   *MergedDetection  `json:"top_match,omitempty"`
	Duration   time.Duration     `json:"duration"`
	Strategies []StrategyOutcome `json:"strategies"`
}

// MergedDetection is a deduplicated detection result after merging
// across multiple strategies.
type MergedDetection struct {
	Category        types.PIICategory       `json:"category"`
	Type            types.PIIType           `json:"type"`
	Sensitivity     types.SensitivityLevel  `json:"sensitivity"`
	FinalConfidence float64                 `json:"final_confidence"` // Weighted merge of all strategies
	Methods         []types.DetectionMethod `json:"methods"`          // Which strategies found this
	Reasoning       string                  `json:"reasoning"`
	RequiresReview  bool                    `json:"requires_review"` // true if confidence < 0.80
}

// StrategyOutcome records what a single strategy found (or didn't find).
type StrategyOutcome struct {
	Name     string                `json:"name"`
	Method   types.DetectionMethod `json:"method"`
	Found    bool                  `json:"found"`
	Results  int                   `json:"results"`
	Duration time.Duration         `json:"duration"`
	Error    string                `json:"error,omitempty"`
}

// =============================================================================
// Confidence Routing
// =============================================================================

// ConfidenceLevel classifies a detection's confidence for routing.
type ConfidenceLevel string

const (
	// ConfidenceAutoVerify means ≥ 0.95 — no human review needed.
	ConfidenceAutoVerify ConfidenceLevel = "AUTO_VERIFY"
	// ConfidenceQuickVerify means 0.80–0.95 — human confirms with one click.
	ConfidenceQuickVerify ConfidenceLevel = "QUICK_VERIFY"
	// ConfidenceManualReview means 0.50–0.80 — human must inspect.
	ConfidenceManualReview ConfidenceLevel = "MANUAL_REVIEW"
	// ConfidenceLow means < 0.50 — flagged for investigation.
	ConfidenceLow ConfidenceLevel = "LOW_CONFIDENCE"
)

// RouteByConfidence determines the verification workflow for a detection.
func RouteByConfidence(confidence float64) ConfidenceLevel {
	switch {
	case confidence >= 0.95:
		return ConfidenceAutoVerify
	case confidence >= 0.80:
		return ConfidenceQuickVerify
	case confidence >= 0.50:
		return ConfidenceManualReview
	default:
		return ConfidenceLow
	}
}
