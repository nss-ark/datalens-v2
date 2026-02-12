package compliance

import (
	"context"
	"errors"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// DSRStatus represents the lifecycle state of a Data Subject Request.
type DSRStatus string

const (
	DSRStatusPending              DSRStatus = "PENDING"
	DSRStatusIdentityVerification DSRStatus = "IDENTITY_VERIFICATION"
	DSRStatusApproved             DSRStatus = "APPROVED"
	DSRStatusInProgress           DSRStatus = "IN_PROGRESS"
	DSRStatusCompleted            DSRStatus = "COMPLETED"
	DSRStatusRejected             DSRStatus = "REJECTED"
	DSRStatusFailed               DSRStatus = "FAILED"
)

// DSRRequestType represents the type of gdpr/dpdpa request.
type DSRRequestType string

const (
	RequestTypeAccess      DSRRequestType = "ACCESS"
	RequestTypeErasure     DSRRequestType = "ERASURE"
	RequestTypeCorrection  DSRRequestType = "CORRECTION"
	RequestTypePortability DSRRequestType = "PORTABILITY"
	RequestTypeNomination  DSRRequestType = "NOMINATION"
)

// DSR represents a Data Subject Request.
type DSR struct {
	ID                 types.ID          `json:"id"`
	TenantID           types.ID          `json:"tenant_id"`
	RequestType        DSRRequestType    `json:"request_type"`
	Status             DSRStatus         `json:"status"`
	SubjectName        string            `json:"subject_name"`
	SubjectEmail       string            `json:"subject_email"`
	SubjectIdentifiers map[string]string `json:"subject_identifiers"` // e.g. {"phone": "+1234", "user_id": "u_123"}
	Priority           string            `json:"priority"`            // "HIGH", "MEDIUM", "LOW"
	SLADeadline        time.Time         `json:"sla_deadline"`
	AssignedTo         *types.ID         `json:"assigned_to,omitempty"`
	Reason             string            `json:"reason,omitempty"` // For rejection or specific context
	Notes              string            `json:"notes,omitempty"`
	Metadata           types.Metadata    `json:"metadata,omitempty"` // Added back
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	CompletedAt        *time.Time        `json:"completed_at,omitempty"`
}

// DSRRepository defines the persistence interface for DSRs.
type DSRRepository interface {
	Create(ctx context.Context, dsr *DSR) error
	GetByID(ctx context.Context, id types.ID) (*DSR, error)
	GetByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination, statusFilter *DSRStatus) (*types.PaginatedResult[DSR], error)
	GetOverdue(ctx context.Context, tenantID types.ID) ([]DSR, error)
	Update(ctx context.Context, dsr *DSR) error

	// Task management
	CreateTask(ctx context.Context, task *DSRTask) error
	GetTasksByDSR(ctx context.Context, dsrID types.ID) ([]DSRTask, error)
	UpdateTask(ctx context.Context, task *DSRTask) error
}

// ValidateTransition checks if a status transition is valid.
func (d *DSR) ValidateTransition(newStatus DSRStatus) error {
	valid := false
	switch d.Status {
	case DSRStatusPending:
		valid = newStatus == DSRStatusIdentityVerification || newStatus == DSRStatusApproved || newStatus == DSRStatusRejected
	case DSRStatusIdentityVerification:
		valid = newStatus == DSRStatusApproved || newStatus == DSRStatusRejected
	case DSRStatusApproved:
		valid = newStatus == DSRStatusInProgress || newStatus == DSRStatusFailed
	case DSRStatusInProgress:
		valid = newStatus == DSRStatusCompleted || newStatus == DSRStatusFailed
	case DSRStatusCompleted, DSRStatusRejected, DSRStatusFailed:
		valid = false // Terminal states
	}

	if !valid {
		return errors.New("invalid status transition from " + string(d.Status) + " to " + string(newStatus))
	}
	return nil
}
