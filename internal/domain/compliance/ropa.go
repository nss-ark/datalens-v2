package compliance

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// RoPAVersion represents a versioned snapshot of a Record of Processing Activities.
type RoPAVersion struct {
	ID            types.ID    `json:"id" db:"id"`
	TenantID      types.ID    `json:"tenant_id" db:"tenant_id"`
	Version       string      `json:"version" db:"version"`
	GeneratedBy   string      `json:"generated_by" db:"generated_by"` // "auto" | user_id
	Status        RoPAStatus  `json:"status" db:"status"`
	Content       RoPAContent `json:"content" db:"content"`
	ChangeSummary string      `json:"change_summary,omitempty" db:"change_summary"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
}

// RoPAStatus defines the lifecycle status of a RoPA version.
type RoPAStatus string

const (
	RoPAStatusDraft     RoPAStatus = "DRAFT"
	RoPAStatusPublished RoPAStatus = "PUBLISHED"
	RoPAStatusArchived  RoPAStatus = "ARCHIVED"
)

// RoPAContent holds the full snapshot of a Record of Processing Activities.
type RoPAContent struct {
	OrganizationName  string           `json:"organization_name"`
	GeneratedAt       time.Time        `json:"generated_at"`
	Purposes          []RoPAPurpose    `json:"purposes"`
	DataSources       []RoPADataSource `json:"data_sources"`
	RetentionPolicies []RoPARetention  `json:"retention_policies"`
	ThirdParties      []RoPAThirdParty `json:"third_parties"`
	DataCategories    []string         `json:"data_categories"`
}

// RoPAPurpose is a snapshot of a Purpose for the RoPA.
type RoPAPurpose struct {
	ID          types.ID `json:"id"`
	Name        string   `json:"name"`
	Code        string   `json:"code"`
	LegalBasis  string   `json:"legal_basis"`
	Description string   `json:"description"`
	IsActive    bool     `json:"is_active"`
}

// RoPADataSource is a snapshot of a DataSource for the RoPA.
type RoPADataSource struct {
	ID       types.ID `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	IsActive bool     `json:"is_active"`
}

// RoPARetention is a snapshot of a RetentionPolicy for the RoPA.
type RoPARetention struct {
	ID               types.ID `json:"id"`
	PurposeName      string   `json:"purpose_name"`
	MaxRetentionDays int      `json:"max_retention_days"`
	DataCategories   []string `json:"data_categories"`
	AutoErase        bool     `json:"auto_erase"`
}

// RoPAThirdParty is a snapshot of a ThirdParty for the RoPA.
type RoPAThirdParty struct {
	ID      types.ID `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Country string   `json:"country"`
}

// RoPARepository defines persistence for RoPA versions.
type RoPARepository interface {
	Create(ctx context.Context, version *RoPAVersion) error
	GetLatest(ctx context.Context, tenantID types.ID) (*RoPAVersion, error)
	GetByVersion(ctx context.Context, tenantID types.ID, version string) (*RoPAVersion, error)
	ListVersions(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[RoPAVersion], error)
	UpdateStatus(ctx context.Context, id types.ID, status RoPAStatus) error
}
