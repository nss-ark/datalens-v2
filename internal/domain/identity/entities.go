// Package identity defines the domain entities for multi-tenant
// organization management, user accounts, roles, and permissions.
package identity

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Tenant — An organization using DataLens
// =============================================================================

// Tenant represents a customer organization.
type Tenant struct {
	types.BaseEntity
	Name     string         `json:"name" db:"name"`
	Domain   string         `json:"domain" db:"domain"`
	Industry string         `json:"industry" db:"industry"`
	Country  string         `json:"country" db:"country"`
	Plan     PlanType       `json:"plan" db:"plan"`
	Status   TenantStatus   `json:"status" db:"status"`
	Settings TenantSettings `json:"settings" db:"settings"`
}

// PlanType classifies the subscription tier.
type PlanType string

const (
	PlanFree         PlanType = "FREE"
	PlanStarter      PlanType = "STARTER"
	PlanProfessional PlanType = "PROFESSIONAL"
	PlanEnterprise   PlanType = "ENTERPRISE"
)

// TenantStatus tracks tenant lifecycle.
type TenantStatus string

const (
	TenantActive    TenantStatus = "ACTIVE"
	TenantSuspended TenantStatus = "SUSPENDED"
	TenantDeleted   TenantStatus = "DELETED"
)

// TenantSettings holds configurable tenant options.
type TenantSettings struct {
	DefaultRegulation  string   `json:"default_regulation"`
	EnabledRegulations []string `json:"enabled_regulations"`
	RetentionDays      int      `json:"retention_days"`
	EnableAI           bool     `json:"enable_ai"`
	AIProvider         string   `json:"ai_provider"`
}

// =============================================================================
// User — A platform user
// =============================================================================

// User represents an authenticated user of the platform.
type User struct {
	types.TenantEntity
	Email       string     `json:"email" db:"email"`
	Name        string     `json:"name" db:"name"`
	Password    string     `json:"-" db:"password"` // Hashed, never exposed
	Status      UserStatus `json:"status" db:"status"`
	RoleIDs     []types.ID `json:"role_ids" db:"role_ids"`
	MFAEnabled  bool       `json:"mfa_enabled" db:"mfa_enabled"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

// UserStatus tracks user account state.
type UserStatus string

const (
	UserActive    UserStatus = "ACTIVE"
	UserInvited   UserStatus = "INVITED"
	UserSuspended UserStatus = "SUSPENDED"
)

// =============================================================================
// Role — Permission grouping
// =============================================================================

// Role defines a named set of permissions.
type Role struct {
	types.BaseEntity
	TenantID    *types.ID    `json:"tenant_id,omitempty" db:"tenant_id"` // nil = system role
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Permissions []Permission `json:"permissions" db:"permissions"`
	IsSystem    bool         `json:"is_system" db:"is_system"`
}

// Permission defines access to a specific resource and actions.
type Permission struct {
	Resource string   `json:"resource"` // DSR, CONSENT, PII, SETTINGS, etc.
	Actions  []string `json:"actions"`  // READ, WRITE, DELETE, VERIFY, ADMIN
}

// System role names.
const (
	RoleAdmin         = "ADMIN"
	RoleDPO           = "DPO"
	RoleAnalyst       = "ANALYST"
	RoleViewer        = "VIEWER"
	RolePlatformAdmin = "PLATFORM_ADMIN"
)

// TenantFilter defines criteria for searching tenants.
type TenantFilter struct {
	Status *TenantStatus
	Limit  int
	Offset int
}

// TenantStats holds aggregate counts.
type TenantStats struct {
	TotalTenants  int64
	ActiveTenants int64
}

// =============================================================================
// Repository Interfaces
// =============================================================================

// TenantRepository defines persistence for tenants.
type TenantRepository interface {
	Create(ctx context.Context, t *Tenant) error
	GetByID(ctx context.Context, id types.ID) (*Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*Tenant, error)
	GetAll(ctx context.Context) ([]Tenant, error)
	Search(ctx context.Context, filter TenantFilter) ([]Tenant, int, error)
	GetStats(ctx context.Context) (*TenantStats, error)
	Update(ctx context.Context, t *Tenant) error
	Delete(ctx context.Context, id types.ID) error
}

// UserRepository defines persistence for users.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id types.ID) (*User, error)
	GetByEmail(ctx context.Context, tenantID types.ID, email string) (*User, error)
	GetByEmailGlobal(ctx context.Context, email string) (*User, error)
	GetByTenant(ctx context.Context, tenantID types.ID) ([]User, error)
	CountGlobal(ctx context.Context) (int64, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id types.ID) error
}

// RoleRepository defines persistence for roles.
type RoleRepository interface {
	Create(ctx context.Context, r *Role) error
	GetByID(ctx context.Context, id types.ID) (*Role, error)
	GetSystemRoles(ctx context.Context) ([]Role, error)
	GetByTenant(ctx context.Context, tenantID types.ID) ([]Role, error)
	Update(ctx context.Context, r *Role) error
}
