package discovery

import (
	"context"
)

// ConnectorCapabilities describes what a data source connector supports.
type ConnectorCapabilities struct {
	CanDiscover             bool `json:"can_discover"`
	CanSample               bool `json:"can_sample"`
	CanDelete               bool `json:"can_delete"`
	CanUpdate               bool `json:"can_update"`
	CanExport               bool `json:"can_export"`
	SupportsStreaming       bool `json:"supports_streaming"`
	SupportsIncremental     bool `json:"supports_incremental"`
	SupportsSchemaDiscovery bool `json:"supports_schema_discovery"`
	SupportsDataSampling    bool `json:"supports_data_sampling"`
	SupportsParallelScan    bool `json:"supports_parallel_scan"`
	MaxConcurrency          int  `json:"max_concurrency"`
}

// Connector defines the universal interface for data source connectors.
// implementations reside in internal/infrastructure/connector.
type Connector interface {
	// Connect establishes a connection to the data source.
	Connect(ctx context.Context, ds *DataSource) error

	// DiscoverSchema returns the schema/structure of the data source.
	// It returns the inventory stats and a list of entities (tables/files).
	DiscoverSchema(ctx context.Context) (*DataInventory, []DataEntity, error)

	// GetFields returns the fields (columns) for a specific entity.
	GetFields(ctx context.Context, entityID string) ([]DataField, error) // entityID is the name or ID from DataEntity

	// SampleData retrieves sample values from a specific entity/field.
	SampleData(ctx context.Context, entity, field string, limit int) ([]string, error)

	// Capabilities returns what operations this connector supports.
	Capabilities() ConnectorCapabilities

	// Close releases the connection.
	Close() error
}
