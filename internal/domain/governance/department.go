package governance

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// Department represents an organizational unit within a tenant.
type Department struct {
	ID                  types.ID  `json:"id" db:"id"`
	TenantID            types.ID  `json:"tenant_id" db:"tenant_id"`
	Name                string    `json:"name" db:"name"`
	Description         string    `json:"description,omitempty" db:"description"`
	OwnerID             *types.ID `json:"owner_id,omitempty" db:"owner_id"`
	OwnerName           string    `json:"owner_name,omitempty" db:"owner_name"`
	OwnerEmail          string    `json:"owner_email,omitempty" db:"owner_email"`
	Responsibilities    []string  `json:"responsibilities" db:"responsibilities"`
	NotificationEnabled bool      `json:"notification_enabled" db:"notification_enabled"`
	IsActive            bool      `json:"is_active" db:"is_active"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// DepartmentRepository defines persistence for departments.
type DepartmentRepository interface {
	Create(ctx context.Context, d *Department) error
	GetByID(ctx context.Context, id types.ID) (*Department, error)
	GetByTenant(ctx context.Context, tenantID types.ID) ([]Department, error)
	GetByOwner(ctx context.Context, ownerID types.ID) ([]Department, error)
	Update(ctx context.Context, d *Department) error
	Delete(ctx context.Context, id types.ID) error
}
