// Package governance defines the domain entities for data governance,
// including purposes, policies, data mapping, and third-party management.
//
// This context manages the "why" and "how" of personal data processing.
package governance

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Purpose — Why personal data is collected/processed
// =============================================================================

// =============================================================================
// Purpose — Why personal data is collected/processed
// =============================================================================
type Purpose struct {
	types.TenantEntity
	Code            string           `json:"code" db:"code"`
	Name            string           `json:"name" db:"name"`
	Description     string           `json:"description" db:"description"`
	LegalBasis      types.LegalBasis `json:"legal_basis" db:"legal_basis"`
	RetentionDays   int              `json:"retention_days" db:"retention_days"`
	IsActive        bool             `json:"is_active" db:"is_active"`
	RequiresConsent bool             `json:"requires_consent" db:"requires_consent"`
}

// =============================================================================
// DataMapping — Links discovered PII to governance
// =============================================================================

// DataMapping connects a PII classification to its governance context.
type DataMapping struct {
	types.TenantEntity
	ClassificationID types.ID             `json:"classification_id" db:"classification_id"`
	PurposeIDs       []types.ID           `json:"purpose_ids" db:"purpose_ids"`
	RetentionDays    int                  `json:"retention_days" db:"retention_days"`
	ThirdPartyIDs    []types.ID           `json:"third_party_ids,omitempty" db:"third_party_ids"`
	Notes            string               `json:"notes,omitempty" db:"notes"`
	MappedBy         types.ID             `json:"mapped_by" db:"mapped_by"`
	MappedAt         time.Time            `json:"mapped_at" db:"mapped_at"`
	CrossBorder      *CrossBorderTransfer `json:"cross_border,omitempty" db:"cross_border"`
}

// CrossBorderTransfer documents international data flows.
type CrossBorderTransfer struct {
	DestinationCountry string `json:"destination_country"`
	LegalMechanism     string `json:"legal_mechanism"`
	Documentation      string `json:"documentation"`
	ApprovalStatus     string `json:"approval_status"`
}

// =============================================================================
// Policy — Governance rules and enforcement
// =============================================================================

// Policy defines a data governance rule with automated enforcement.
type Policy struct {
	types.TenantEntity
	Name        string         `json:"name" db:"name"`
	Description string         `json:"description" db:"description"`
	Type        PolicyType     `json:"type" db:"type"`
	Rules       []PolicyRule   `json:"rules" db:"rules"` // Stored as JSONB
	Severity    types.Severity `json:"severity" db:"severity"`
	Actions     []PolicyAction `json:"actions" db:"actions"` // Stored as JSONB
	IsActive    bool           `json:"is_active" db:"is_active"`
}

// PolicyType classifies the governance rule.
type PolicyType string

const (
	PolicyTypeRetention    PolicyType = "RETENTION"
	PolicyTypeAccess       PolicyType = "ACCESS"
	PolicyTypeTransfer     PolicyType = "TRANSFER"
	PolicyTypeAlert        PolicyType = "ALERT"
	PolicyTypeConsent      PolicyType = "CONSENT"
	PolicyTypeMinimization PolicyType = "MINIMIZATION"
)

// PolicyRule defines a single condition within a policy.
type PolicyRule struct {
	ID          types.ID `json:"id"`
	Field       string   `json:"field"`       // e.g., "sensitivity", "data_type", "retention_days"
	Operator    string   `json:"operator"`    // e.g., "EQ", "GT", "CONTAINS"
	Value       any      `json:"value"`       // e.g., "HIGH", 365
	Description string   `json:"description"` // e.g., "Sensitivity is HIGH"
}

// PolicyAction defines what happens when a policy is violated.
type PolicyAction string

const (
	PolicyActionAlert         PolicyAction = "ALERT"
	PolicyActionBlock         PolicyAction = "BLOCK"
	PolicyActionAutoRemediate PolicyAction = "AUTO_REMEDIATE"
	PolicyActionLog           PolicyAction = "LOG"
)

// =============================================================================
// Violation — Policy enforcement results
// =============================================================================

// Violation represents a detected breach of a governance policy.
type Violation struct {
	types.TenantEntity
	PolicyID     types.ID        `json:"policy_id" db:"policy_id"`
	DataSourceID types.ID        `json:"data_source_id" db:"data_source_id"`
	EntityName   string          `json:"entity_name" db:"entity_name"` // Table/Collection name
	FieldName    string          `json:"field_name" db:"field_name"`   // Column/Field name
	Status       ViolationStatus `json:"status" db:"status"`
	Severity     types.Severity  `json:"severity" db:"severity"`
	DetectedAt   time.Time       `json:"detected_at" db:"detected_at"`
	ResolvedAt   *time.Time      `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy   *types.ID       `json:"resolved_by,omitempty" db:"resolved_by"`
	Resolution   *string         `json:"resolution,omitempty" db:"resolution"`
}

// ViolationStatus tracks the lifecycle of a violation.
type ViolationStatus string

const (
	ViolationStatusOpen       ViolationStatus = "OPEN"
	ViolationStatusInProgress ViolationStatus = "IN_PROGRESS"
	ViolationStatusResolved   ViolationStatus = "RESOLVED"
	ViolationStatusIgnored    ViolationStatus = "IGNORED"
)

// =============================================================================
// ThirdParty — External data processors/controllers
// =============================================================================

// ThirdParty represents an external entity that processes data.
type ThirdParty struct {
	types.TenantEntity
	Name       string         `json:"name" db:"name"`
	Type       ThirdPartyType `json:"type" db:"type"`
	Country    string         `json:"country" db:"country"`
	DPADocPath *string        `json:"dpa_doc_path,omitempty" db:"dpa_doc_path"`
	IsActive   bool           `json:"is_active" db:"is_active"`
	PurposeIDs []types.ID     `json:"purpose_ids" db:"purpose_ids"`
}

// ThirdPartyType classifies the third party's role.
type ThirdPartyType string

const (
	ThirdPartyProcessor  ThirdPartyType = "PROCESSOR"
	ThirdPartyController ThirdPartyType = "CONTROLLER"
	ThirdPartyVendor     ThirdPartyType = "VENDOR"
)

// =============================================================================
// Repository Interfaces
// =============================================================================

// PurposeRepository defines persistence for purposes.
type PurposeRepository interface {
	Create(ctx context.Context, p *Purpose) error
	GetByID(ctx context.Context, id types.ID) (*Purpose, error)
	GetByTenant(ctx context.Context, tenantID types.ID) ([]Purpose, error)
	GetByCode(ctx context.Context, tenantID types.ID, code string) (*Purpose, error)
	Update(ctx context.Context, p *Purpose) error
	Delete(ctx context.Context, id types.ID) error
}

// DataMappingRepository defines persistence for data mappings.
type DataMappingRepository interface {
	Create(ctx context.Context, dm *DataMapping) error
	GetByID(ctx context.Context, id types.ID) (*DataMapping, error)
	GetByClassification(ctx context.Context, classificationID types.ID) (*DataMapping, error)
	GetUnmapped(ctx context.Context, tenantID types.ID) ([]types.ID, error)
	Update(ctx context.Context, dm *DataMapping) error
}

// PolicyRepository defines persistence for policies.
type PolicyRepository interface {
	Create(ctx context.Context, p *Policy) error
	GetByID(ctx context.Context, id types.ID) (*Policy, error)
	GetActive(ctx context.Context, tenantID types.ID) ([]Policy, error)
	GetByType(ctx context.Context, tenantID types.ID, policyType PolicyType) ([]Policy, error)
	Update(ctx context.Context, p *Policy) error
	Delete(ctx context.Context, id types.ID) error
}

// ViolationRepository defines persistence for violations.
type ViolationRepository interface {
	Create(ctx context.Context, v *Violation) error
	GetByID(ctx context.Context, id types.ID) (*Violation, error)
	GetByTenant(ctx context.Context, tenantID types.ID, status *ViolationStatus) ([]Violation, error)
	GetByPolicy(ctx context.Context, policyID types.ID) ([]Violation, error)
	GetByDataSource(ctx context.Context, dataSourceID types.ID) ([]Violation, error)
	UpdateStatus(ctx context.Context, id types.ID, status ViolationStatus, resolvedBy *types.ID, resolution *string) error
}

// =============================================================================
// Purpose Suggestion — Transient entity for automation
// =============================================================================

// PurposeSuggestion represents a suggested purpose for a data element.
type PurposeSuggestion struct {
	Table                string    `json:"table"`
	Column               string    `json:"column"`
	SuggestedPurposeID   *types.ID `json:"suggested_purpose_id,omitempty"`   // ID if purpose exists
	SuggestedPurposeCode string    `json:"suggested_purpose_code,omitempty"` // Code to look up or create
	Confidence           float64   `json:"confidence"`
	Reason               string    `json:"reason"`
	Strategy             string    `json:"strategy"` // "PATTERN", "HEURISTIC", "AI"
}
