// Package eventbus defines the event bus interface for publishing
// and subscribing to domain events across bounded contexts.
//
// Events are the primary mechanism for inter-context communication,
// ensuring loose coupling between domain modules.
package eventbus

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Event Types
// =============================================================================

// Event represents a domain event that occurred in the system.
type Event struct {
	ID        types.ID       `json:"id"`
	Type      string         `json:"type"`
	Source    string         `json:"source"` // Which context emitted it
	TenantID  types.ID       `json:"tenant_id"`
	Timestamp time.Time      `json:"timestamp"`
	Data      any            `json:"data"`
	Metadata  types.Metadata `json:"metadata,omitempty"`
}

// NewEvent creates a new event with a generated ID and current timestamp.
func NewEvent(eventType, source string, tenantID types.ID, data any) Event {
	return Event{
		ID:        types.NewID(),
		Type:      eventType,
		Source:    source,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

// =============================================================================
// Event Type Constants
// =============================================================================

const (
	// PII Events
	EventPIIDiscovered = "pii.discovered"
	EventPIIVerified   = "pii.verified"
	EventPIICorrected  = "pii.corrected"
	EventPIIRejected   = "pii.rejected"
	EventPIIClassified = "pii.classified"

	// Scan Events
	EventScanStarted   = "scan.started"
	EventScanProgress  = "scan.progress"
	EventScanCompleted = "scan.completed"
	EventScanFailed    = "scan.failed"

	// DSR Events
	EventDSRCreated            = "dsr.created"
	EventDSRExecuting          = "dsr.executing"
	EventDSRCompleted          = "dsr.completed"
	EventDSRFailed             = "dsr.failed"
	EventDSRVerified           = "dsr.verified"
	EventDSRVerificationFailed = "dsr.verification_failed"
	EventDSRRejected           = "dsr.rejected"

	// Consent Events
	EventConsentGranted          = "consent.granted"
	EventConsentWithdrawn        = "consent.withdrawn"
	EventConsentWidgetCreated    = "consent.widget_created"
	EventConsentExpiring         = "consent.expiring"
	EventConsentExpired          = "consent.expired"
	EventConsentReceiptGenerated = "consent.receipt_generated"

	// Breach Events
	EventBreachDetected = "breach.detected"
	EventBreachNotified = "breach.notified"
	EventBreachResolved = "breach.resolved"

	// Policy Events
	EventPolicyViolation  = "policy.violation"
	EventPolicyRemediated = "policy.remediated"

	// Data Source Events
	EventDataSourceConnected    = "datasource.connected"
	EventDataSourceDisconnected = "datasource.disconnected"
	EventDataSourceError        = "datasource.error"
	EventDataSourceCreated      = "datasource.created"
	EventDataSourceUpdated      = "datasource.updated"
	EventDataSourceDeleted      = "datasource.deleted"

	// Policy / Purpose Events
	EventPolicyCreated = "policy.created"
	EventPolicyUpdated = "policy.updated"
	EventPolicyDeleted = "policy.deleted"

	// User / Auth Events
	EventUserRegistered = "user.registered"
	EventUserLoggedIn   = "user.logged_in"
	EventTenantCreated  = "tenant.created"

	// Breach Events (Additional)
	EventBreachIncidentCreated = "breach.incident_created"
	EventBreachIncidentUpdated = "breach.incident_updated"

	// Grievance Events
	EventGrievanceSubmitted = "compliance.grievance_submitted"
	EventGrievanceAssigned  = "compliance.grievance_assigned"
	EventGrievanceResolved  = "compliance.grievance_resolved"
	EventGrievanceEscalated = "compliance.grievance_escalated"

	// Translation Events
	EventNoticeTranslated      = "consent.notice_translated"
	EventTranslationOverridden = "consent.translation_overridden"

	// DSR Events (Additional)
	EventDSRDataAccessed           = "dsr.data_accessed"
	EventDSRManualDeletionRequired = "dsr.manual_deletion_required"
	EventDSRDataDeleted            = "dsr.data_deleted"

	// DPR Events
	EventDPRSubmitted = "dpr.submitted"

	// Governance Events (Additional)
	EventLineageFlowTracked      = "governance.lineage.flow_tracked"
	EventGovernancePolicyCreated = "governance.policy_created"
)

// =============================================================================
// EventBus Interface
// =============================================================================

// EventHandler processes a received event.
type EventHandler func(ctx context.Context, event Event) error

// Subscription represents an active event subscription.
type Subscription interface {
	// Unsubscribe removes this subscription.
	Unsubscribe() error
}

// EventBus defines the contract for publishing and subscribing to events.
type EventBus interface {
	// Publish sends an event to all subscribers.
	Publish(ctx context.Context, event Event) error

	// Subscribe registers a handler for events matching the pattern.
	// Pattern supports wildcards: "pii.*" matches all PII events.
	Subscribe(ctx context.Context, pattern string, handler EventHandler) (Subscription, error)

	// Close gracefully shuts down the event bus.
	Close() error
}
