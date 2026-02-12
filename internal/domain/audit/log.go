package audit

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// AuditLog represents a single audit entry for a sensitive action.
type AuditLog struct {
	ID           types.ID       `json:"id"`
	TenantID     types.ID       `json:"tenant_id"` // Maps to client_id in DB
	UserID       types.ID       `json:"user_id"`   // Maps to user_id in DB
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   types.ID       `json:"resource_id"`
	OldValues    map[string]any `json:"old_values,omitempty"`
	NewValues    map[string]any `json:"new_values,omitempty"`
	IPAddress    string         `json:"ip_address,omitempty"`
	UserAgent    string         `json:"user_agent,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

// Repository defines the interface for persisting audit logs.
type Repository interface {
	// Create persists a new audit log entry.
	Create(ctx context.Context, log *AuditLog) error

	// GetByTenant retrieves audit logs for a tenant with optional filtering.
	// For MVP, we might just list them, but filtering is good to have in interface.
	GetByTenant(ctx context.Context, tenantID types.ID, limit int) ([]AuditLog, error)
}
