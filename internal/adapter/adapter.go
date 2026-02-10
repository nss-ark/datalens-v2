// Package adapter defines the compliance regulation adapter interface.
//
// Each regulation (DPDPA, GDPR, CCPA) is implemented as an adapter
// that plugs into the regulation-agnostic core. This keeps the core
// engine clean while allowing regulation-specific behavior.
package adapter

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// ComplianceAdapter Interface
// =============================================================================

// ComplianceAdapter defines regulation-specific behavior that plugs
// into the regulation-agnostic core engine.
type ComplianceAdapter interface {
	// Name returns the human-readable regulation name.
	Name() string

	// Code returns the regulation identifier (e.g., "DPDPA", "GDPR").
	Code() string

	// --- DSR Configuration ---

	// SupportedDSRTypes returns which DSR types this regulation supports.
	SupportedDSRTypes() []types.DSRType

	// DSRDeadline computes the deadline for a given DSR type.
	DSRDeadline(dsrType types.DSRType, receivedAt time.Time) time.Time

	// DSRRequiresVerification returns true if identity verification is mandatory.
	DSRRequiresVerification(dsrType types.DSRType) bool

	// --- Consent Configuration ---

	// ConsentRequirements returns consent configuration for this regulation.
	ConsentRequirements() ConsentConfig

	// --- Breach Configuration ---

	// BreachNotificationDeadline returns the authority notification deadline.
	BreachNotificationDeadline(detectedAt time.Time) time.Time

	// BreachRequiresSubjectNotification returns true if subjects must be notified.
	BreachRequiresSubjectNotification(severity types.Severity) bool

	// --- Classification ---

	// PIICategories returns the PII categories relevant to this regulation.
	PIICategories() []types.PIICategory

	// SensitiveCategories returns categories that require enhanced protection.
	SensitiveCategories() []types.PIICategory

	// --- Rights ---

	// DataSubjectRights returns all rights under this regulation.
	DataSubjectRights() []Right

	// --- Validation ---

	// ValidateCompliance checks if current state meets regulation requirements.
	ValidateCompliance(ctx context.Context, tenantID types.ID) (*ComplianceReport, error)
}

// =============================================================================
// Supporting Types
// =============================================================================

// ConsentConfig defines regulation-specific consent requirements.
type ConsentConfig struct {
	RequiresExplicit bool   `json:"requires_explicit"`
	GranularRequired bool   `json:"granular_required"`
	MaxAgeMonths     int    `json:"max_age_months"`
	WithdrawalMustBe string `json:"withdrawal_must_be"` // "EASY", "ANY_TIME"
	MinAge           int    `json:"min_age"`            // Age for minor consent
	GuardianRequired bool   `json:"guardian_required"`
}

// Right defines a data subject right under a specific regulation.
type Right struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Mandatory   bool   `json:"mandatory"`
}

// ComplianceReport summarizes compliance status.
type ComplianceReport struct {
	Regulation   string            `json:"regulation"`
	OverallScore float64           `json:"overall_score"` // 0.0 to 1.0
	Status       ComplianceStatus  `json:"status"`
	Issues       []ComplianceIssue `json:"issues"`
	GeneratedAt  time.Time         `json:"generated_at"`
}

// ComplianceStatus classifies overall compliance.
type ComplianceStatus string

const (
	ComplianceCompliant    ComplianceStatus = "COMPLIANT"
	CompliancePartial      ComplianceStatus = "PARTIAL"
	ComplianceNonCompliant ComplianceStatus = "NON_COMPLIANT"
)

// ComplianceIssue describes a specific compliance gap.
type ComplianceIssue struct {
	Category    string         `json:"category"`
	Description string         `json:"description"`
	Severity    types.Severity `json:"severity"`
	Remediation string         `json:"remediation"`
}
