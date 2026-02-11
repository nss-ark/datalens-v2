package governance

import (
	"context"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Data Lineage — Tracking data movement
// =============================================================================

// DataFlow represents a movement of data from a source to a description.
type DataFlow struct {
	types.TenantEntity
	SourceID      types.ID   `json:"source_id" db:"source_id"`           // DataSource ID
	DestinationID types.ID   `json:"destination_id" db:"destination_id"` // DataSource ID
	DataType      string     `json:"data_type" db:"data_type"`           // "TABLE", "COLUMN", "FILE"
	DataPath      string     `json:"data_path" db:"data_path"`           // "schema.table" or "bucket/key"
	PurposeID     *types.ID  `json:"purpose_id,omitempty" db:"purpose_id"`
	Status        FlowStatus `json:"status" db:"status"`
	Description   string     `json:"description" db:"description"`
}

type FlowStatus string

const (
	FlowStatusActive   FlowStatus = "ACTIVE"
	FlowStatusInactive FlowStatus = "INACTIVE"
	FlowStatusProposed FlowStatus = "PROPOSED"
)

// LineageGraph represents the visualization structure for the UI.
// Compatible with React Flow or Recharts Sanctey.
type LineageGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	ID    string                 `json:"id"`
	Label string                 `json:"label"`
	Type  string                 `json:"type"` // "DATA_SOURCE", "PROCESS", "THIRD_PARTY"
	Data  map[string]interface{} `json:"data,omitempty"`
}

type GraphEdge struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Label    string `json:"label,omitempty"`
	Animated bool   `json:"animated,omitempty"`
	FlowID   string `json:"flowId"` // Reference to DataFlow ID
}

// LineageRepository defines persistence for data flows.
type LineageRepository interface {
	Create(ctx context.Context, flow *DataFlow) error
	GetByTenant(ctx context.Context, tenantID types.ID) ([]DataFlow, error)
	GetBySource(ctx context.Context, sourceID types.ID) ([]DataFlow, error)
	GetByDestination(ctx context.Context, destID types.ID) ([]DataFlow, error)
	// GetGraphNodes returns unique data sources involved in flows for a tenant
	// This might be a service-level composition, but repo support helps.
}
