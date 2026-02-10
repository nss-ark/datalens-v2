// Package discovery defines the domain entities for data source
// management, PII detection, and scanning operations.
//
// This context is responsible for discovering and cataloging personal
// data across all connected data sources.
package discovery

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// DataSource — Represents any connected data repository
// =============================================================================

// DataSource is a database, file system, cloud storage, or Control Centre
// that contains data to be scanned for PII.
type DataSource struct {
	types.TenantEntity
	Name         string               `json:"name" db:"name"`
	Type         types.DataSourceType `json:"type" db:"type"`
	Description  string               `json:"description" db:"description"`
	Host         string               `json:"-" db:"host"` // Encrypted at rest
	Port         int                  `json:"-" db:"port"`
	Database     string               `json:"database" db:"database"`                     // Database name for relational DBs, bucket for S3
	Credentials  string               `json:"-" db:"credentials"`                         // Encrypted
	Config       string               `json:"config" db:"config"`                         // JSON config specific to connector type
	ScanSchedule *string              `json:"scan_schedule,omitempty" db:"scan_schedule"` // Cron expression for automated scans
	Status       ConnectionStatus     `json:"status" db:"status"`
	LastSyncAt   *time.Time           `json:"last_sync_at" db:"last_sync_at"`
	ErrorMessage *string              `json:"error_message,omitempty" db:"error_message"`
}

// ConnectionStatus tracks the data source connection state.
type ConnectionStatus string

const (
	ConnectionStatusConnected    ConnectionStatus = "CONNECTED"
	ConnectionStatusDisconnected ConnectionStatus = "DISCONNECTED"
	ConnectionStatusError        ConnectionStatus = "ERROR"
	ConnectionStatusTesting      ConnectionStatus = "TESTING"
)

// =============================================================================
// DataInventory — Discovered data structure from a source
// =============================================================================

// DataInventory represents the cataloged schema of a data source.
type DataInventory struct {
	types.BaseEntity
	DataSourceID   types.ID  `json:"data_source_id" db:"data_source_id"`
	TotalEntities  int       `json:"total_entities" db:"total_entities"`
	TotalFields    int       `json:"total_fields" db:"total_fields"`
	PIIFieldsCount int       `json:"pii_fields_count" db:"pii_fields_count"`
	LastScannedAt  time.Time `json:"last_scanned_at" db:"last_scanned_at"`
	SchemaVersion  string    `json:"schema_version" db:"schema_version"`
}

// =============================================================================
// DataEntity — A table, collection, folder, or file
// =============================================================================

// DataEntity is a logical grouping of data fields (e.g., table, collection).
type DataEntity struct {
	types.BaseEntity
	InventoryID   types.ID   `json:"inventory_id" db:"inventory_id"`
	Name          string     `json:"name" db:"name"`
	Schema        string     `json:"schema" db:"schema"`
	Type          EntityType `json:"type" db:"type"`
	RowCount      *int64     `json:"row_count,omitempty" db:"row_count"`
	PIIConfidence float64    `json:"pii_confidence" db:"pii_confidence"`
}

// EntityType classifies the kind of data entity.
type EntityType string

const (
	EntityTypeTable      EntityType = "TABLE"
	EntityTypeView       EntityType = "VIEW"
	EntityTypeCollection EntityType = "COLLECTION"
	EntityTypeFolder     EntityType = "FOLDER"
	EntityTypeFile       EntityType = "FILE"
)

// =============================================================================
// DataField — A column, property, or data element
// =============================================================================

// DataField represents a single data element within an entity.
type DataField struct {
	types.BaseEntity
	EntityID     types.ID `json:"entity_id" db:"entity_id"`
	Name         string   `json:"name" db:"name"`
	DataType     string   `json:"data_type" db:"data_type"`
	Nullable     bool     `json:"nullable" db:"nullable"`
	IsPrimaryKey bool     `json:"is_primary_key" db:"is_primary_key"`
	IsForeignKey bool     `json:"is_foreign_key" db:"is_foreign_key"`
}

// =============================================================================
// PIIClassification — Detection result for a data field
// =============================================================================

// PIIClassification holds the PII analysis result for a specific field.
type PIIClassification struct {
	types.BaseEntity
	FieldID         types.ID                 `json:"field_id" db:"field_id"`
	DataSourceID    types.ID                 `json:"data_source_id" db:"data_source_id"`
	EntityName      string                   `json:"entity_name" db:"entity_name"`
	FieldName       string                   `json:"field_name" db:"field_name"`
	Category        types.PIICategory        `json:"category" db:"category"`
	Type            types.PIIType            `json:"type" db:"type"`
	Sensitivity     types.SensitivityLevel   `json:"sensitivity" db:"sensitivity"`
	Confidence      float64                  `json:"confidence" db:"confidence"`
	DetectionMethod types.DetectionMethod    `json:"detection_method" db:"detection_method"`
	Status          types.VerificationStatus `json:"status" db:"status"`
	VerifiedBy      *types.ID                `json:"verified_by,omitempty" db:"verified_by"`
	VerifiedAt      *time.Time               `json:"verified_at,omitempty" db:"verified_at"`
	Reasoning       string                   `json:"reasoning" db:"reasoning"`
}

// =============================================================================
// ScanRun — A single scanning operation
// =============================================================================

// ScanRun tracks the execution of a scan against a data source.
type ScanRun struct {
	types.BaseEntity
	DataSourceID types.ID   `json:"data_source_id" db:"data_source_id"`
	TenantID     types.ID   `json:"tenant_id" db:"tenant_id"`
	Type         ScanType   `json:"type" db:"type"`
	Status       ScanStatus `json:"status" db:"status"`
	Progress     int        `json:"progress" db:"progress"`
	StartedAt    *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	Stats        ScanStats  `json:"stats" db:"stats"`
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"`
}

// ScanType classifies the kind of scan.
type ScanType string

const (
	ScanTypeFull        ScanType = "FULL"
	ScanTypeIncremental ScanType = "INCREMENTAL"
	ScanTypeTargeted    ScanType = "TARGETED"
)

// ScanStatus tracks scan execution state.
type ScanStatus string

const (
	ScanStatusPending   ScanStatus = "PENDING"
	ScanStatusRunning   ScanStatus = "RUNNING"
	ScanStatusCompleted ScanStatus = "COMPLETED"
	ScanStatusFailed    ScanStatus = "FAILED"
	ScanStatusCancelled ScanStatus = "CANCELLED"
)

// ScanStats holds aggregated scan metrics.
type ScanStats struct {
	EntitiesScanned int           `json:"entities_scanned"`
	FieldsScanned   int           `json:"fields_scanned"`
	PIIDetected     int           `json:"pii_detected"`
	Duration        time.Duration `json:"duration"`
	BytesProcessed  int64         `json:"bytes_processed"`
}

// =============================================================================
// Repository Interfaces
// =============================================================================

// DataSourceRepository defines persistence operations for data sources.
type DataSourceRepository interface {
	Create(ctx context.Context, ds *DataSource) error
	GetByID(ctx context.Context, id types.ID) (*DataSource, error)
	GetByTenant(ctx context.Context, tenantID types.ID) ([]DataSource, error)
	Update(ctx context.Context, ds *DataSource) error
	Delete(ctx context.Context, id types.ID) error
}

// ClassificationFilter defines criteria for filtering PII classifications.
type ClassificationFilter struct {
	DataSourceID    *types.ID
	Status          *types.VerificationStatus
	DetectionMethod *types.DetectionMethod
	Pagination      types.Pagination
}

// PIICounts holds aggregated PII statistics.
type PIICounts struct {
	Total      int            `json:"total"`
	ByCategory map[string]int `json:"by_category"`
}

// PIIClassificationRepository defines persistence for PII findings.
type PIIClassificationRepository interface {
	Create(ctx context.Context, c *PIIClassification) error
	GetByID(ctx context.Context, id types.ID) (*PIIClassification, error)
	GetByDataSource(ctx context.Context, dataSourceID types.ID, pagination types.Pagination) (*types.PaginatedResult[PIIClassification], error)
	GetPending(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[PIIClassification], error)
	GetClassifications(ctx context.Context, tenantID types.ID, filter ClassificationFilter) (*types.PaginatedResult[PIIClassification], error)
	GetCounts(ctx context.Context, tenantID types.ID) (*PIICounts, error)
	Update(ctx context.Context, c *PIIClassification) error
	BulkCreate(ctx context.Context, classifications []PIIClassification) error
}

// ScanRunRepository defines persistence for scan operations.
type ScanRunRepository interface {
	Create(ctx context.Context, run *ScanRun) error
	GetByID(ctx context.Context, id types.ID) (*ScanRun, error)
	GetByDataSource(ctx context.Context, dataSourceID types.ID) ([]ScanRun, error)
	GetActive(ctx context.Context, tenantID types.ID) ([]ScanRun, error)
	GetRecent(ctx context.Context, tenantID types.ID, limit int) ([]ScanRun, error)
	Update(ctx context.Context, run *ScanRun) error
}

// DataInventoryRepository defines persistence for data inventories.
type DataInventoryRepository interface {
	Create(ctx context.Context, inv *DataInventory) error
	GetByID(ctx context.Context, id types.ID) (*DataInventory, error)
	GetByDataSource(ctx context.Context, dataSourceID types.ID) (*DataInventory, error)
	Update(ctx context.Context, inv *DataInventory) error
}

// DataEntityRepository defines persistence for data entities.
type DataEntityRepository interface {
	Create(ctx context.Context, entity *DataEntity) error
	GetByID(ctx context.Context, id types.ID) (*DataEntity, error)
	GetByInventory(ctx context.Context, inventoryID types.ID) ([]DataEntity, error)
	Update(ctx context.Context, entity *DataEntity) error
	Delete(ctx context.Context, id types.ID) error
}

// DataFieldRepository defines persistence for data fields.
type DataFieldRepository interface {
	Create(ctx context.Context, field *DataField) error
	GetByID(ctx context.Context, id types.ID) (*DataField, error)
	GetByEntity(ctx context.Context, entityID types.ID) ([]DataField, error)
	Update(ctx context.Context, field *DataField) error
	Delete(ctx context.Context, id types.ID) error
}

// =============================================================================
// Connector Interface
// =============================================================================
