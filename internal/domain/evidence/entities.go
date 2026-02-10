// Package evidence defines the domain entities for immutable audit
// trails, evidence generation, and tamper-proof record keeping.
//
// Every action in DataLens produces evidence. This context ensures
// the evidence is legally admissible and tamper-proof.
package evidence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// AuditEvent — Immutable record of every significant action
// =============================================================================

// AuditEvent captures a single action within the system, forming
// a hash-chained audit trail for tamper detection.
type AuditEvent struct {
	types.BaseEntity
	TenantID  types.ID `json:"tenant_id" db:"tenant_id"`
	EventType string   `json:"event_type" db:"event_type"`

	// Who
	ActorID   types.ID  `json:"actor_id" db:"actor_id"`
	ActorType ActorType `json:"actor_type" db:"actor_type"`

	// What
	ResourceType string   `json:"resource_type" db:"resource_type"`
	ResourceID   types.ID `json:"resource_id" db:"resource_id"`
	Action       string   `json:"action" db:"action"`

	// State change
	Before json.RawMessage `json:"before,omitempty" db:"before"`
	After  json.RawMessage `json:"after,omitempty" db:"after"`

	// Metadata
	Metadata  types.Metadata `json:"metadata,omitempty" db:"metadata"`
	IPAddress string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent string         `json:"user_agent,omitempty" db:"user_agent"`

	// Integrity (hash chain)
	PreviousHash string `json:"previous_hash" db:"previous_hash"`
	Hash         string `json:"hash" db:"hash"`
	Signature    string `json:"signature" db:"signature"`
}

// ActorType classifies who performed an action.
type ActorType string

const (
	ActorUser   ActorType = "USER"
	ActorSystem ActorType = "SYSTEM"
	ActorAPI    ActorType = "API"
	ActorAgent  ActorType = "AGENT"
)

// =============================================================================
// EvidencePackage — Compiled proof bundle
// =============================================================================

// EvidencePackage bundles related audit events into a legally
// admissible evidence collection.
type EvidencePackage struct {
	types.BaseEntity
	TenantID types.ID     `json:"tenant_id" db:"tenant_id"`
	Type     EvidenceType `json:"type" db:"type"`
	Title    string       `json:"title" db:"title"`
	Summary  string       `json:"summary" db:"summary"`

	// Content
	EventIDs  []types.ID `json:"event_ids" db:"event_ids"`
	Documents []Document `json:"documents,omitempty" db:"documents"`

	// Context
	GeneratedFor string     `json:"generated_for" db:"generated_for"`
	GeneratedAt  time.Time  `json:"generated_at" db:"generated_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty" db:"expires_at"`

	// Integrity
	Hash        string `json:"hash" db:"hash"`
	Signature   string `json:"signature" db:"signature"`
	StoragePath string `json:"storage_path" db:"storage_path"`
}

// EvidenceType classifies the evidence package.
type EvidenceType string

const (
	EvidenceDSRCompletion   EvidenceType = "DSR_COMPLETION"
	EvidenceConsentProof    EvidenceType = "CONSENT_PROOF"
	EvidenceBreachResponse  EvidenceType = "BREACH_RESPONSE"
	EvidenceScanReport      EvidenceType = "SCAN_REPORT"
	EvidenceComplianceAudit EvidenceType = "COMPLIANCE_AUDIT"
)

// Document represents a supporting document in an evidence package.
type Document struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	MimeType  string `json:"mime_type"`
	Hash      string `json:"hash"`
	SizeBytes int64  `json:"size_bytes"`
}

// =============================================================================
// Repository Interfaces
// =============================================================================

// AuditEventRepository defines persistence for audit events.
type AuditEventRepository interface {
	// Create stores a new audit event and returns it with hash computed.
	Create(ctx context.Context, event *AuditEvent) error

	// GetByID retrieves a specific event.
	GetByID(ctx context.Context, id types.ID) (*AuditEvent, error)

	// GetByResource retrieves events for a specific resource.
	GetByResource(ctx context.Context, resourceType string, resourceID types.ID) ([]AuditEvent, error)

	// GetByTenant retrieves events with pagination and optional filtering.
	GetByTenant(ctx context.Context, tenantID types.ID, filter AuditFilter, pagination types.Pagination) (*types.PaginatedResult[AuditEvent], error)

	// GetLatestHash returns the hash of the most recent event for chain linking.
	GetLatestHash(ctx context.Context, tenantID types.ID) (string, error)

	// VerifyChain validates the integrity of the audit chain.
	VerifyChain(ctx context.Context, tenantID types.ID) (bool, error)
}

// AuditFilter provides filtering options for audit queries.
type AuditFilter struct {
	EventType    *string
	ActorID      *types.ID
	ResourceType *string
	StartTime    *time.Time
	EndTime      *time.Time
}

// EvidencePackageRepository defines persistence for evidence packages.
type EvidencePackageRepository interface {
	Create(ctx context.Context, pkg *EvidencePackage) error
	GetByID(ctx context.Context, id types.ID) (*EvidencePackage, error)
	GetByTenant(ctx context.Context, tenantID types.ID) ([]EvidencePackage, error)
}
