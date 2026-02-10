package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// DataSourceService handles data source lifecycle operations.
type DataSourceService struct {
	repo     discovery.DataSourceRepository
	eventBus eventbus.EventBus
	logger   *slog.Logger
}

// NewDataSourceService creates a new DataSourceService.
func NewDataSourceService(repo discovery.DataSourceRepository, eb eventbus.EventBus, logger *slog.Logger) *DataSourceService {
	return &DataSourceService{
		repo:     repo,
		eventBus: eb,
		logger:   logger.With("service", "data_source"),
	}
}

// CreateDataSourceInput holds the fields required to create a data source.
type CreateDataSourceInput struct {
	TenantID    types.ID
	Name        string
	Type        types.DataSourceType
	Description string
	Host        string
	Port        int
	Database    string
	Credentials string
}

// Create validates and persists a new data source, then publishes an event.
func (s *DataSourceService) Create(ctx context.Context, in CreateDataSourceInput) (*discovery.DataSource, error) {
	if in.Name == "" {
		return nil, types.NewValidationError("name is required", nil)
	}
	if in.Type == "" {
		return nil, types.NewValidationError("type is required", nil)
	}

	ds := &discovery.DataSource{
		Name:        in.Name,
		Type:        in.Type,
		Description: in.Description,
		Host:        in.Host,
		Port:        in.Port,
		Database:    in.Database,
		Credentials: in.Credentials,
		Status:      discovery.ConnectionStatusDisconnected,
	}
	ds.TenantID = in.TenantID

	if err := s.repo.Create(ctx, ds); err != nil {
		return nil, fmt.Errorf("create data source: %w", err)
	}

	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(
		eventbus.EventDataSourceCreated, "discovery", in.TenantID,
		map[string]any{"id": ds.ID, "name": ds.Name, "type": string(ds.Type)},
	))

	s.logger.InfoContext(ctx, "data source created", "id", ds.ID, "name", ds.Name)
	return ds, nil
}

// GetByID retrieves a data source by ID.
func (s *DataSourceService) GetByID(ctx context.Context, id types.ID) (*discovery.DataSource, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByTenant retrieves all data sources for a tenant.
func (s *DataSourceService) ListByTenant(ctx context.Context, tenantID types.ID) ([]discovery.DataSource, error) {
	return s.repo.GetByTenant(ctx, tenantID)
}

// UpdateDataSourceInput holds updatable fields.
type UpdateDataSourceInput struct {
	ID          types.ID
	Name        string
	Description string
	Host        string
	Port        *int
	Database    string
	Credentials string
}

// Update modifies an existing data source.
func (s *DataSourceService) Update(ctx context.Context, in UpdateDataSourceInput) (*discovery.DataSource, error) {
	ds, err := s.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if in.Name != "" {
		ds.Name = in.Name
	}
	if in.Description != "" {
		ds.Description = in.Description
	}
	if in.Host != "" {
		ds.Host = in.Host
	}
	if in.Port != nil {
		ds.Port = *in.Port
	}
	if in.Database != "" {
		ds.Database = in.Database
	}
	if in.Credentials != "" {
		ds.Credentials = in.Credentials
	}

	if err := s.repo.Update(ctx, ds); err != nil {
		return nil, err
	}

	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(
		eventbus.EventDataSourceUpdated, "discovery", ds.TenantID,
		map[string]any{"id": ds.ID, "name": ds.Name},
	))

	return ds, nil
}

// Delete soft-deletes a data source.
func (s *DataSourceService) Delete(ctx context.Context, id types.ID) error {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(
		eventbus.EventDataSourceDeleted, "discovery", ds.TenantID,
		map[string]any{"id": id, "name": ds.Name},
	))

	s.logger.InfoContext(ctx, "data source deleted", "id", id)
	return nil
}
