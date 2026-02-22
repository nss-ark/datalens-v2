package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector"
	m365 "github.com/complyark/datalens/internal/infrastructure/connector/m365"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// DataSourceService handles data source lifecycle operations.
type DataSourceService struct {
	repo     discovery.DataSourceRepository
	registry *connector.ConnectorRegistry
	eventBus eventbus.EventBus
	logger   *slog.Logger
}

// NewDataSourceService creates a new DataSourceService.
func NewDataSourceService(repo discovery.DataSourceRepository, registry *connector.ConnectorRegistry, eb eventbus.EventBus, logger *slog.Logger) *DataSourceService {
	return &DataSourceService{
		repo:     repo,
		registry: registry,
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
	Config      string
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
		Type:        types.NormalizeDataSourceType(string(in.Type)),
		Description: in.Description,
		Host:        in.Host,
		Port:        in.Port,
		Database:    in.Database,
		Credentials: in.Credentials,
		Config:      in.Config,
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

// CreateFromFileInput holds input for creating a file-based data source.
type CreateFromFileInput struct {
	TenantID types.ID
	Name     string
	Filename string
	Content  io.Reader
}

// CreateFromFile saves an uploaded file and creates a data source record.
func (s *DataSourceService) CreateFromFile(ctx context.Context, in CreateFromFileInput) (*discovery.DataSource, error) {
	if in.Name == "" {
		return nil, types.NewValidationError("name is required", nil)
	}
	if in.Filename == "" {
		return nil, types.NewValidationError("filename is required", nil)
	}

	// Ensure upload directory exists
	uploadDir := filepath.Join("uploads", fmt.Sprintf("tenant_%s", in.TenantID))
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	// Generate safe filename
	safeFilename := fmt.Sprintf("%s_%s", types.NewID(), filepath.Base(in.Filename))
	filePath := filepath.Join(uploadDir, safeFilename)

	// Save file
	outFile, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, in.Content); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	// Create JSON config with path
	configMap := map[string]string{"path": filePath}
	configBytes, _ := json.Marshal(configMap)

	// Create DataSource
	return s.Create(ctx, CreateDataSourceInput{
		TenantID:    in.TenantID,
		Name:        in.Name,
		Type:        types.DataSourceFileUpload,
		Description: fmt.Sprintf("Uploaded file: %s", in.Filename),
		Config:      string(configBytes),
		// Other fields empty
	})
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
	Config      string
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
	if in.Config != "" {
		ds.Config = in.Config
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

// SetSchedule sets a cron expression for automated scanning.
func (s *DataSourceService) SetSchedule(ctx context.Context, id types.ID, cronExpr string) (*discovery.DataSource, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	ds.ScanSchedule = &cronExpr

	if err := s.repo.Update(ctx, ds); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "scan schedule set", "id", id, "cron", cronExpr)
	return ds, nil
}

// ClearSchedule removes the scan schedule from a data source.
func (s *DataSourceService) ClearSchedule(ctx context.Context, id types.ID) error {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	ds.ScanSchedule = nil

	if err := s.repo.Update(ctx, ds); err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "scan schedule cleared", "id", id)
	return nil
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

	// If it's a file upload, clean up the file
	if ds.Type == types.DataSourceFileUpload {
		var config map[string]string
		if err := json.Unmarshal([]byte(ds.Config), &config); err == nil {
			if path, ok := config["path"]; ok {
				// We ignore errors here as the DB record is already deleted (soft deleted actually).
				// If we want to be strict, we could log it.
				if err := os.Remove(path); err != nil {
					s.logger.WarnContext(ctx, "failed to delete file", "path", path, "error", err)
				}
			}
		}
	}

	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(
		eventbus.EventDataSourceDeleted, "discovery", ds.TenantID,
		map[string]any{"id": id, "name": ds.Name},
	))

	s.logger.InfoContext(ctx, "data source deleted", "id", id)
	return nil
}

// ListM365Users retrieves a list of users from an M365 data source.
// It requires the data source to be of type Microsoft365 (or OneDrive/Outlook).
func (s *DataSourceService) ListM365Users(ctx context.Context, id types.ID) ([]m365.User, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validation
	if ds.Type != types.DataSourceMicrosoft365 && ds.Type != types.DataSourceOneDrive && ds.Type != types.DataSourceOutlook {
		return nil, types.NewValidationError("data source is not M365", nil)
	}

	conn, err := s.registry.GetConnector(ds.Type)
	if err != nil {
		return nil, fmt.Errorf("get connector: %w", err)
	}

	// Check if connector supports listing users
	// We need to type assert to *connector.M365Connector specifically, OR define an interface.
	// Since M365Connector is in `connector` package (as `M365Connector`), we can assert.
	// NOTE: m365.go is in `connector` package.
	// So we import "github.com/complyark/datalens/internal/infrastructure/connector" (which is self if we were in connector)
	// But we are in `service`.
	// We imported `github.com/complyark/datalens/internal/infrastructure/connector` above.
	// But wait, `M365Connector` is defined in `connector` package?
	// Yes, `m365.go` says `package connector`.
	// So we can assertion using `connector.M365Connector`.

	m365Conn, ok := conn.(*connector.M365Connector)
	if !ok {
		return nil, fmt.Errorf("connector does not support user listing")
	}

	if err := m365Conn.Connect(ctx, ds); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return m365Conn.ListUsers(ctx)
}

// ListM365Sites retrieves a list of sites from an M365 data source.
func (s *DataSourceService) ListM365Sites(ctx context.Context, id types.ID) ([]m365.Site, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if ds.Type != types.DataSourceMicrosoft365 && ds.Type != types.DataSourceOneDrive {
		return nil, types.NewValidationError("data source is not M365/OneDrive", nil)
	}

	conn, err := s.registry.GetConnector(ds.Type)
	if err != nil {
		return nil, fmt.Errorf("get connector: %w", err)
	}

	m365Conn, ok := conn.(*connector.M365Connector)
	if !ok {
		return nil, fmt.Errorf("connector does not support site listing")
	}

	if err := m365Conn.Connect(ctx, ds); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return m365Conn.ListSites(ctx)
}
