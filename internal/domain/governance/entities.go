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

// Purpose represents a reason for processing personal data.
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
	Rules       []PolicyRule   `json:"rules" db:"rules"`
	Severity    types.Severity `json:"severity" db:"severity"`
	Actions     []PolicyAction `json:"actions" db:"actions"`
	IsActive    bool           `json:"is_active" db:"is_active"`
}

// PolicyType classifies the governance rule.
type PolicyType string

const (
	PolicyTypeRetention PolicyType = "RETENTION"
	PolicyTypeAccess    PolicyType = "ACCESS"
	PolicyTypeTransfer  PolicyType = "TRANSFER"
	PolicyTypeAlert     PolicyType = "ALERT"
	PolicyTypeConsent   PolicyType = "CONSENT"
)

// PolicyRule defines a single condition within a policy.
type PolicyRule struct {
	ID          types.ID `json:"id"`
	Field       string   `json:"field"`
	Operator    string   `json:"operator"`
	Value       any      `json:"value"`
	Description string   `json:"description"`
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
	Update(ctx context.Context, p *Policy) error
	Delete(ctx context.Context, id types.ID) error
}
