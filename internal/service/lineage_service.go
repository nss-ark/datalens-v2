package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

type LineageService struct {
	repo     governance.LineageRepository
	dsRepo   discovery.DataSourceRepository
	eventBus eventbus.EventBus
	logger   *slog.Logger
}

func NewLineageService(
	repo governance.LineageRepository,
	dsRepo discovery.DataSourceRepository,
	eventBus eventbus.EventBus,
	logger *slog.Logger,
) *LineageService {
	return &LineageService{
		repo:     repo,
		dsRepo:   dsRepo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// TrackFlow records a new data movement.
func (s *LineageService) TrackFlow(ctx context.Context, flow *governance.DataFlow) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}

	// Set system fields
	flow.ID = types.NewID()
	flow.TenantID = tenantID
	flow.CreatedAt = time.Now()
	// Although the handler sets UpdatedAt, we ensure it here too
	if flow.UpdatedAt.IsZero() {
		flow.UpdatedAt = time.Now()
	}
	if flow.Status == "" {
		flow.Status = governance.FlowStatusActive
	}

	// Validate Source and Destination exist (and belong to tenant)
	source, err := s.dsRepo.GetByID(ctx, flow.SourceID)
	if err != nil {
		return fmt.Errorf("get source: %w", err)
	}
	if source.TenantID != tenantID {
		return types.NewForbiddenError("source data source not found")
	}

	dest, err := s.dsRepo.GetByID(ctx, flow.DestinationID)
	if err != nil {
		return fmt.Errorf("get destination: %w", err)
	}
	if dest.TenantID != tenantID {
		return types.NewForbiddenError("destination data source not found")
	}

	if err := s.repo.Create(ctx, flow); err != nil {
		return err
	}

	// Publish event
	event := eventbus.NewEvent("governance.lineage.flow_tracked", "governance", tenantID, flow)
	return s.eventBus.Publish(ctx, event)
}

// GetGraph builds the lineage graph for the tenant.
func (s *LineageService) GetGraph(ctx context.Context, tenantID types.ID) (*governance.LineageGraph, error) {
	flows, err := s.repo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get flows: %w", err)
	}

	graph := &governance.LineageGraph{
		Nodes: []governance.GraphNode{},
		Edges: []governance.GraphEdge{},
	}

	nodeMap := make(map[string]bool)

	// Helper to add node if not exists
	var addNode func(id types.ID) error
	addNode = func(id types.ID) error {
		if nodeMap[id.String()] {
			return nil
		}
		ds, err := s.dsRepo.GetByID(ctx, id)
		if err != nil {
			s.logger.Warn("Failed to fetch data source for lineage graph", "id", id, "error", err)
			// Add a placeholder node if DS is missing (soft delete?)
			graph.Nodes = append(graph.Nodes, governance.GraphNode{
				ID:    id.String(),
				Label: "Unknown Source",
				Type:  "UNKNOWN",
			})
			nodeMap[id.String()] = true
			return nil
		}

		graph.Nodes = append(graph.Nodes, governance.GraphNode{
			ID:    ds.ID.String(),
			Label: ds.Name,
			Type:  string(ds.Type),
			Data: map[string]interface{}{
				"connection_status": ds.Status,
			},
		})
		nodeMap[id.String()] = true
		return nil
	}

	for _, flow := range flows {
		// Add Nodes
		if err := addNode(flow.SourceID); err != nil {
			continue
		}
		if err := addNode(flow.DestinationID); err != nil {
			continue
		}

		// Add Edge
		graph.Edges = append(graph.Edges, governance.GraphEdge{
			ID:       flow.ID.String(),
			Source:   flow.SourceID.String(),
			Target:   flow.DestinationID.String(),
			Label:    flow.DataType,
			Animated: true,
			FlowID:   flow.ID.String(),
		})
	}

	return graph, nil
}
