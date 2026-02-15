// Package compliance defines the domain entities for DSR management,
// consent handling, breach tracking, and grievance resolution.
//
// This context handles all compliance-specific operations that involve
// data subjects and their rights under applicable regulations.
package compliance

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// DataSubject — An individual whose data is processed
// =============================================================================

// DataSubject represents a natural person whose personal data is processed.
type DataSubject struct {
	types.TenantEntity
	Identifier     string         `json:"identifier" db:"identifier"`
	IdentifierType IdentifierType `json:"identifier_type" db:"identifier_type"`
	DisplayName    string         `json:"display_name,omitempty" db:"display_name"`
	FirstSeenAt    time.Time      `json:"first_seen_at" db:"first_seen_at"`
	LastActivityAt time.Time      `json:"last_activity_at" db:"last_activity_at"`
	Status         SubjectStatus  `json:"status" db:"status"`
}

// IdentifierType classifies the subject's primary identifier.
type IdentifierType string

const (
	IdentifierEmail  IdentifierType = "EMAIL"
	IdentifierPhone  IdentifierType = "PHONE"
	IdentifierCustom IdentifierType = "CUSTOM"
)

// SubjectStatus tracks the data subject's lifecycle.
type SubjectStatus string

const (
	SubjectActive     SubjectStatus = "ACTIVE"
	SubjectDeleted    SubjectStatus = "DELETED"
	SubjectAnonymized SubjectStatus = "ANONYMIZED"
)

// =============================================================================
// DSR — Data Subject Request
// =============================================================================

// DSR is a regulation-agnostic data subject request.

// DSRTask is an actionable unit within a DSR, targeting one data source.

// DSRAction defines what action to take on data.
type DSRAction string

const (
	DSRActionExport DSRAction = "EXPORT"
	DSRActionDelete DSRAction = "DELETE"
	DSRActionUpdate DSRAction = "UPDATE"
	DSRActionBlock  DSRAction = "BLOCK"
)

// TaskStatus tracks DSR task execution state.
type TaskStatus string

// =============================================================================
// Consent — Data subject's permission for processing
// =============================================================================

// Consent records a data subject's permission for data processing.
type Consent struct {
	types.TenantEntity
	SubjectID        types.ID               `json:"subject_id" db:"subject_id"`
	PurposeIDs       []types.ID             `json:"purpose_ids" db:"purpose_ids"`
	GrantedAt        time.Time              `json:"granted_at" db:"granted_at"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty" db:"expires_at"`
	Mechanism        types.ConsentMechanism `json:"mechanism" db:"mechanism"`
	Status           ConsentStatus          `json:"status" db:"status"`
	WithdrawnAt      *time.Time             `json:"withdrawn_at,omitempty" db:"withdrawn_at"`
	WithdrawalReason *string                `json:"withdrawal_reason,omitempty" db:"withdrawal_reason"`
	RegulationRef    string                 `json:"regulation_ref" db:"regulation_ref"`
	Proof            ConsentProof           `json:"proof" db:"proof"`
}

// ConsentStatus tracks consent lifecycle.
type ConsentStatus string

const (
	ConsentStatusActive    ConsentStatus = "ACTIVE"
	ConsentStatusWithdrawn ConsentStatus = "WITHDRAWN"
	ConsentStatusExpired   ConsentStatus = "EXPIRED"
)

// ConsentProof captures evidence of consent.
type ConsentProof struct {
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	Timestamp      time.Time `json:"timestamp"`
	Method         string    `json:"method"`
	ScreenshotHash *string   `json:"screenshot_hash,omitempty"`
	Signature      string    `json:"signature"`
}

// =============================================================================
// Breach — Data breach incident
// =============================================================================

// Breach represents a data security incident.
type Breach struct {
	types.TenantEntity
	DetectedAt          time.Time      `json:"detected_at" db:"detected_at"`
	DetectedBy          string         `json:"detected_by" db:"detected_by"`
	Type                BreachType     `json:"type" db:"type"`
	Severity            types.Severity `json:"severity" db:"severity"`
	Description         string         `json:"description" db:"description"`
	AffectedRecords     *int64         `json:"affected_records,omitempty" db:"affected_records"`
	Status              BreachStatus   `json:"status" db:"status"`
	ContainedAt         *time.Time     `json:"contained_at,omitempty" db:"contained_at"`
	ResolvedAt          *time.Time     `json:"resolved_at,omitempty" db:"resolved_at"`
	AuthorityNotifiedAt *time.Time     `json:"authority_notified_at,omitempty" db:"authority_notified_at"`
	SubjectsNotifiedAt  *time.Time     `json:"subjects_notified_at,omitempty" db:"subjects_notified_at"`
	RootCause           string         `json:"root_cause,omitempty" db:"root_cause"`
	Remediation         string         `json:"remediation,omitempty" db:"remediation"`
	EvidenceID          *types.ID      `json:"evidence_id,omitempty" db:"evidence_id"`
}

// BreachType classifies the incident.
type BreachType string

const (
	BreachUnauthorizedAccess BreachType = "UNAUTHORIZED_ACCESS"
	BreachDataLoss           BreachType = "DATA_LOSS"
	BreachDataExfiltration   BreachType = "DATA_EXFILTRATION"
	BreachSystemCompromise   BreachType = "SYSTEM_COMPROMISE"
	BreachInsiderThreat      BreachType = "INSIDER_THREAT"
)

// BreachStatus tracks incident lifecycle.
type BreachStatus string

const (
	BreachStatusDetected      BreachStatus = "DETECTED"
	BreachStatusInvestigating BreachStatus = "INVESTIGATING"
	BreachStatusContained     BreachStatus = "CONTAINED"
	BreachStatusResolved      BreachStatus = "RESOLVED"
)

// =============================================================================
// DPO Contact — Data Protection Officer details
// =============================================================================

// DPOContact represents the contact details of the Data Protection Officer.
// Required by DPDPA S8(10) and R9 Schedule II.
type DPOContact struct {
	types.TenantEntity
	OrgName    string    `json:"org_name" db:"org_name"`
	DPOName    string    `json:"dpo_name" db:"dpo_name"`
	DPOEmail   string    `json:"dpo_email" db:"dpo_email"`
	DPOPhone   *string   `json:"dpo_phone,omitempty" db:"dpo_phone"`
	Address    *string   `json:"address,omitempty" db:"address"`
	WebsiteURL *string   `json:"website_url,omitempty" db:"website_url"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// =============================================================================
// Repository Interfaces
// =============================================================================

// DPOContactRepository defines persistence for DPO contact details.
type DPOContactRepository interface {
	// Upsert creates or updates the DPO contact for a tenant.
	Upsert(ctx context.Context, contact *DPOContact) error
	// Get retrieves the DPO contact for a tenant.
	Get(ctx context.Context, tenantID types.ID) (*DPOContact, error)
}

// DSRRepository defines persistence for DSR operations.

// ConsentRepository defines persistence for consent records.
type ConsentRepository interface {
	Create(ctx context.Context, consent *Consent) error
	GetByID(ctx context.Context, id types.ID) (*Consent, error)
	GetBySubject(ctx context.Context, subjectID types.ID) ([]Consent, error)
	GetActive(ctx context.Context, subjectID types.ID) ([]Consent, error)
	GetExpiring(ctx context.Context, tenantID types.ID, before time.Time) ([]Consent, error)
	Update(ctx context.Context, consent *Consent) error
}

// BreachRepository defines persistence for breach incidents.
type BreachRepository interface {
	Create(ctx context.Context, breach *Breach) error
	GetByID(ctx context.Context, id types.ID) (*Breach, error)
	GetByTenant(ctx context.Context, tenantID types.ID) ([]Breach, error)
	GetActive(ctx context.Context, tenantID types.ID) ([]Breach, error)
	Update(ctx context.Context, breach *Breach) error
}
