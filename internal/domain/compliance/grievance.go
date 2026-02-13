package compliance

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// Grievance represents a formal complaint lodged by a data principal.
type Grievance struct {
	types.TenantEntity
	DataSubjectID   types.ID        `json:"data_subject_id" db:"data_subject_id"`
	Subject         string          `json:"subject" db:"subject"`
	Description     string          `json:"description" db:"description"`
	Category        string          `json:"category" db:"category"` // CONSENT, DATA_PROCESSING, DSR_DISSATISFACTION, BREACH_CONCERN, OTHER
	Status          GrievanceStatus `json:"status" db:"status"`
	Priority        int             `json:"priority" db:"priority"`
	AssignedTo      *types.ID       `json:"assigned_to,omitempty" db:"assigned_to"`
	Resolution      *string         `json:"resolution,omitempty" db:"resolution"`
	SubmittedAt     time.Time       `json:"submitted_at" db:"submitted_at"`
	DueDate         *time.Time      `json:"due_date,omitempty" db:"due_date"` // 30-day SLA per DPDPA
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty" db:"resolved_at"`
	EscalatedTo     *string         `json:"escalated_to,omitempty" db:"escalated_to"` // DPA authority name
	FeedbackRating  *int            `json:"feedback_rating,omitempty" db:"feedback_rating"`
	FeedbackComment *string         `json:"feedback_comment,omitempty" db:"feedback_comment"`
}

// GrievanceStatus tracks the lifecycle of a grievance.
type GrievanceStatus string

const (
	GrievanceStatusOpen       GrievanceStatus = "OPEN"
	GrievanceStatusInProgress GrievanceStatus = "IN_PROGRESS"
	GrievanceStatusResolved   GrievanceStatus = "RESOLVED"
	GrievanceStatusEscalated  GrievanceStatus = "ESCALATED"
	GrievanceStatusClosed     GrievanceStatus = "CLOSED" // Closed after feedback or expiry
)

// GrievanceRepository defines persistence for grievances.
type GrievanceRepository interface {
	Create(ctx context.Context, g *Grievance) error
	GetByID(ctx context.Context, id types.ID) (*Grievance, error)
	ListByTenant(ctx context.Context, tenantID types.ID, filters map[string]any, pagination types.Pagination) (*types.PaginatedResult[Grievance], error)
	ListBySubject(ctx context.Context, tenantID, subjectID types.ID) ([]Grievance, error)
	Update(ctx context.Context, g *Grievance) error
}
