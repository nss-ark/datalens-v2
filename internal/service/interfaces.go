package service

import (
	"context"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// DiscoveryOrchestrator defines the interface for discovery operations.
type DiscoveryOrchestrator interface {
	ScanDataSource(ctx context.Context, dataSourceID types.ID) (*discovery.ScanStats, error)
	TestConnection(ctx context.Context, dataSourceID types.ID) error
	GetClassifications(ctx context.Context, tenantID types.ID, filter discovery.ClassificationFilter) (*types.PaginatedResult[discovery.PIIClassification], error)
}

// ScanOrchestrator defines the interface for scan job management.
type ScanOrchestrator interface {
	EnqueueScan(ctx context.Context, dataSourceID types.ID, tenantID types.ID, scanType discovery.ScanType) (*discovery.ScanRun, error)
	GetScan(ctx context.Context, id types.ID) (*discovery.ScanRun, error)
	GetHistory(ctx context.Context, dataSourceID types.ID) ([]discovery.ScanRun, error)
}
