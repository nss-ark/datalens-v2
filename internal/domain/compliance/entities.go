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
type DSR struct {
	types.TenantEntity
	SubjectID types.ID       `json:"subject_id" db:"subject_id"`
	Type      types.DSRType  `json:"type" db:"type"`
	Status    DSRStatus      `json:"status" db:"status"`
	Priority  types.Priority `json:"priority" db:"priority"`

	// Request context
	RegulationRef string    `json:"regulation_ref" db:"regulation_ref"`
	RequestedAt   time.Time `json:"requested_at" db:"requested_at"`
	Deadline      time.Time `json:"deadline" db:"deadline"`
	RequestedBy   string    `json:"requested_by" db:"requested_by"`
	RequestNotes  string    `json:"request_notes,omitempty" db:"request_notes"`

	// Assignment
	AssignedTo *types.ID `json:"assigned_to,omitempty" db:"assigned_to"`

	// Completion
	CompletedAt     *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CompletionNotes string     `json:"completion_notes,omitempty" db:"completion_notes"`
	EvidenceID      *types.ID  `json:"evidence_id,omitempty" db:"evidence_id"`
}

// DSRStatus tracks DSR lifecycle.
type DSRStatus string

const (
	DSRStatusPending      DSRStatus = "PENDING"
	DSRStatusVerifying    DSRStatus = "VERIFYING" // Identity verification
	DSRStatusInProgress   DSRStatus = "IN_PROGRESS"
	DSRStatusExecuting    DSRStatus = "EXECUTING" // Tasks running
	DSRStatusCompleted    DSRStatus = "COMPLETED"
	DSRStatusFailed       DSRStatus = "FAILED"
	DSRStatusRejected     DSRStatus = "REJECTED"      // Failed verification
	DSRStatusAutoVerified DSRStatus = "AUTO_VERIFIED" // Post-execution check passed
)

// DSRTask is an actionable unit within a DSR, targeting one data source.
type DSRTask struct {
	types.BaseEntity
	DSRID        types.ID   `json:"dsr_id" db:"dsr_id"`
	DataSourceID types.ID   `json:"data_source_id" db:"data_source_id"`
	Action       DSRAction  `json:"action" db:"action"`
	Status       TaskStatus `json:"status" db:"status"`
	ExecutedAt   *time.Time `json:"executed_at,omitempty" db:"executed_at"`
	Result       *string    `json:"result,omitempty" db:"result"`
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"`
}

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

const (
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusCompleted TaskStatus = "COMPLETED"
	TaskStatusFailed    TaskStatus = "FAILED"
	TaskStatusSkipped   TaskStatus = "SKIPPED"
)

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
// Grievance — Complaint from a data subject
// =============================================================================

// Grievance represents a formal complaint.
type Grievance struct {
	types.TenantEntity
	SubjectID     *types.ID       `json:"subject_id,omitempty" db:"subject_id"`
	Type          GrievanceType   `json:"type" db:"type"`
	Description   string          `json:"description" db:"description"`
	ReceivedAt    time.Time       `json:"received_at" db:"received_at"`
	ReceivedVia   string          `json:"received_via" db:"received_via"`
	Status        GrievanceStatus `json:"status" db:"status"`
	AssignedTo    *types.ID       `json:"assigned_to,omitempty" db:"assigned_to"`
	Response      *string         `json:"response,omitempty" db:"response"`
	ResolvedAt    *time.Time      `json:"resolved_at,omitempty" db:"resolved_at"`
	Deadline      time.Time       `json:"deadline" db:"deadline"`
	RegulationRef string          `json:"regulation_ref" db:"regulation_ref"`
}

// GrievanceType classifies the complaint.
type GrievanceType string

const (
	GrievanceConsent      GrievanceType = "CONSENT"
	GrievanceDataAccuracy GrievanceType = "DATA_ACCURACY"
	GrievanceBreach       GrievanceType = "BREACH"
	GrievanceServiceIssue GrievanceType = "SERVICE_ISSUE"
	GrievanceOther        GrievanceType = "OTHER"
)

// GrievanceStatus tracks grievance lifecycle.
type GrievanceStatus string

const (
	GrievanceStatusPending   GrievanceStatus = "PENDING"
	GrievanceStatusAssigned  GrievanceStatus = "ASSIGNED"
	GrievanceStatusResolved  GrievanceStatus = "RESOLVED"
	GrievanceStatusEscalated GrievanceStatus = "ESCALATED"
)

// =============================================================================
// Repository Interfaces
// =============================================================================

// DSRRepository defines persistence for DSR operations.
type DSRRepository interface {
	Create(ctx context.Context, dsr *DSR) error
	GetByID(ctx context.Context, id types.ID) (*DSR, error)
	GetByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[DSR], error)
	GetPending(ctx context.Context, tenantID types.ID) ([]DSR, error)
	GetOverdue(ctx context.Context, tenantID types.ID) ([]DSR, error)
	Update(ctx context.Context, dsr *DSR) error
}

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
