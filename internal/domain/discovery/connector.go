package discovery

import (
	"context"
	"time"
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

// DiscoveryInput contains options for the discovery process.
type DiscoveryInput struct {
	// ChangedSince, if set, requests the connector to only return entities
	// modified after this time. If the connector does not support incremental
	// discovery, it may ignore this field and return all entities.
	ChangedSince time.Time
}

// Connector defines the universal interface for data source connectors.
// implementations reside in internal/infrastructure/connector.
type Connector interface {
	// Connect establishes a connection to the data source.
	Connect(ctx context.Context, ds *DataSource) error

	// DiscoverSchema returns the schema/structure of the data source.
	// It returns the inventory stats and a list of entities (tables/files).
	DiscoverSchema(ctx context.Context, input DiscoveryInput) (*DataInventory, []DataEntity, error)

	// GetFields returns the fields (columns) for a specific entity.
	GetFields(ctx context.Context, entityID string) ([]DataField, error) // entityID is the name or ID from DataEntity

	// SampleData retrieves sample values from a specific entity/field.
	SampleData(ctx context.Context, entity, field string, limit int) ([]string, error)

	// Capabilities returns what operations this connector supports.
	Capabilities() ConnectorCapabilities

	// Close releases the connection.
	Close() error
}

// ScannableConnector is an optional interface for connectors that support
// custom scanning logic (e.g. streaming file usage) instead of standard sampling.
type ScannableConnector interface {
	Connector
	// Scan performs a scan on the data source and invokes the callback for each PII finding.
	// This allows the connector to optimize traversal (e.g. streaming) and use internal detection.
	Scan(ctx context.Context, ds *DataSource, onFinding func(PIIClassification)) error
}
