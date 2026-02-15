package compliance

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// RetentionPolicyStatus represents the status of a retention policy.
type RetentionPolicyStatus string

const (
	RetentionPolicyActive RetentionPolicyStatus = "ACTIVE"
	RetentionPolicyPaused RetentionPolicyStatus = "PAUSED"
)

// RetentionPolicy defines rules for data retention and erasure.
// Implements DPDP Rule R8(1-5).
type RetentionPolicy struct {
	ID               types.ID              `json:"id"`
	TenantID         types.ID              `json:"tenant_id"`
	PurposeID        types.ID              `json:"purpose_id"` // Links to the Purpose being governed
	MaxRetentionDays int                   `json:"max_retention_days"`
	DataCategories   []string              `json:"data_categories"` // e.g., ["contact", "financial"]
	Status           RetentionPolicyStatus `json:"status"`
	AutoErase        bool                  `json:"auto_erase"` // If true, triggers erasure DSR automatically
	Description      string                `json:"description,omitempty"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

// RetentionLog tracks retention and erasure actions for audit and proof.
// Implements DPDP Rule R8(4) - Proof of erasure.
type RetentionLog struct {
	ID        types.ID  `json:"id"`
	TenantID  types.ID  `json:"tenant_id"`
	PolicyID  types.ID  `json:"policy_id"`
	Action    string    `json:"action"` // e.g., "ERASED", "NOTIFIED_PROCESSOR", "EXPIRED"
	Target    string    `json:"target"` // Description of what was erased (e.g. "User u_123", "File f_456")
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// RetentionPolicyRepository defines the persistence interface for retention policies.
type RetentionPolicyRepository interface {
	Create(ctx context.Context, policy *RetentionPolicy) error
	GetByID(ctx context.Context, id types.ID) (*RetentionPolicy, error)
	// GetByTenant retrieves all policies for a tenant.
	GetByTenant(ctx context.Context, tenantID types.ID) ([]RetentionPolicy, error)
	Update(ctx context.Context, policy *RetentionPolicy) error
	Delete(ctx context.Context, id types.ID) error

	// Log methods
	CreateLog(ctx context.Context, log *RetentionLog) error
	GetLogs(ctx context.Context, tenantID types.ID, policyID *types.ID, pagination types.Pagination) (*types.PaginatedResult[RetentionLog], error)
}
