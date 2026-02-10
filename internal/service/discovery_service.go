package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector"
	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// DiscoveryService orchestrates the scanning and PII detection process.
type DiscoveryService struct {
	dsRepo        discovery.DataSourceRepository
	inventoryRepo discovery.DataInventoryRepository
	entityRepo    discovery.DataEntityRepository
	fieldRepo     discovery.DataFieldRepository
	piiRepo       discovery.PIIClassificationRepository
	scanRunRepo   discovery.ScanRunRepository

	registry *connector.ConnectorRegistry
	detector *detection.ComposableDetector
	eventBus eventbus.EventBus
	logger   *slog.Logger
}

// NewDiscoveryService creates a new DiscoveryService.
func NewDiscoveryService(
	dsRepo discovery.DataSourceRepository,
	inventoryRepo discovery.DataInventoryRepository,
	entityRepo discovery.DataEntityRepository,
	fieldRepo discovery.DataFieldRepository,
	piiRepo discovery.PIIClassificationRepository,
	scanRunRepo discovery.ScanRunRepository,
	registry *connector.ConnectorRegistry,
	detector *detection.ComposableDetector,
	eb eventbus.EventBus,
	logger *slog.Logger,
) *DiscoveryService {
	return &DiscoveryService{
		dsRepo:        dsRepo,
		inventoryRepo: inventoryRepo,
		entityRepo:    entityRepo,
		fieldRepo:     fieldRepo,
		piiRepo:       piiRepo,
		scanRunRepo:   scanRunRepo,
		registry:      registry,
		detector:      detector,
		eventBus:      eb,
		logger:        logger.With("service", "discovery"),
	}
}

// ScanDataSource initiates a full scan of a data source.
// It detects schema changes and scans for PII.
func (s *DiscoveryService) ScanDataSource(ctx context.Context, dataSourceID types.ID) error {
	start := time.Now()

	// 1. Fetch Data Source
	ds, err := s.dsRepo.GetByID(ctx, dataSourceID)
	if err != nil {
		return fmt.Errorf("fetch data source: %w", err)
	}

	s.logger.InfoContext(ctx, "starting scan", "data_source_id", ds.ID, "name", ds.Name)

	// 2. Resolve Connector via Registry
	conn, err := s.registry.GetConnector(ds.Type)
	if err != nil {
		return fmt.Errorf("resolve connector: %w", err)
	}

	// 3. Connect
	if err := conn.Connect(ctx, ds); err != nil {
		s.logError(ctx, ds.ID, "connection failed", err)
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	// 4. Determine Scan Mode (Full vs Incremental)
	var discoveryInput discovery.DiscoveryInput

	// Check for previous successful scan
	// We need a method to get the last successful scan for this DS.
	// Since GetByDataSource returns all, we might want a specific method or filter here.
	// For now, let's assume we fetch recent scans and find the last success.
	// Improving ScanRunRepo to have GetLastSuccessful(ctx, dsID) would be better, but let's work with what we have.
	scanRuns, err := s.scanRunRepo.GetByDataSource(ctx, ds.ID)
	if err == nil {
		// Find latest COMPLETED scan
		var lastSuccess *discovery.ScanRun
		for i := range scanRuns {
			run := &scanRuns[i]
			if run.Status == discovery.ScanStatusCompleted && run.CompletedAt != nil {
				if lastSuccess == nil || run.CompletedAt.After(*lastSuccess.CompletedAt) {
					lastSuccess = run
				}
			}
		}

		if lastSuccess != nil {
			discoveryInput.ChangedSince = *lastSuccess.CompletedAt
			s.logger.InfoContext(ctx, "performing incremental scan",
				"data_source_id", ds.ID,
				"changed_since", discoveryInput.ChangedSince)
		} else {
			s.logger.InfoContext(ctx, "performing full scan (no previous success)", "data_source_id", ds.ID)
		}
	} else {
		s.logger.WarnContext(ctx, "failed to fetch scan history", "error", err)
	}

	// 5. Discover Schema (Inventory + Entities)
	inventory, entities, err := conn.DiscoverSchema(ctx, discoveryInput)
	if err != nil {
		s.logError(ctx, ds.ID, "schema discovery failed", err)
		return fmt.Errorf("discover schema: %w", err)
	}

	// 5. Sync Inventory
	// Check if inventory exists
	existingInv, err := s.inventoryRepo.GetByDataSource(ctx, ds.ID)
	if err != nil && !types.IsNotFoundError(err) {
		return err
	}

	if existingInv == nil {
		inventory.DataSourceID = ds.ID
		inventory.LastScannedAt = time.Now()
		if err := s.inventoryRepo.Create(ctx, inventory); err != nil {
			return fmt.Errorf("create inventory: %w", err)
		}
	} else {
		existingInv.TotalEntities = inventory.TotalEntities
		existingInv.LastScannedAt = time.Now()
		if err := s.inventoryRepo.Update(ctx, existingInv); err != nil {
			return fmt.Errorf("update inventory: %w", err)
		}
		inventory = existingInv
	}

	// 6. Process Entities
	piiCount := 0
	for _, entity := range entities {
		entity.InventoryID = inventory.ID

		// Create/Update Entity (Simplified: always create if not exists, skipping update logic for brevity)
		// Real implementation should check for existing by Name + InventoryID
		// Assuming we can list entities by inventory and find match.
		// For MVP, let's just create if not exists (implementation detail omitted or assumed Repo handles upsert?)
		// Repos are standard Postgres, so we need to check.

		// Let's assume we proceed to processing fields immediately.
		// We need the ID.
		// We'll list existing entities to match names.
		existingEntities, _ := s.entityRepo.GetByInventory(ctx, inventory.ID)
		var entityID types.ID
		var exists bool
		for _, e := range existingEntities {
			if e.Name == entity.Name {
				entityID = e.ID
				exists = true
				break
			}
		}

		if !exists {
			if err := s.entityRepo.Create(ctx, &entity); err != nil {
				return err
			}
			entityID = entity.ID
		}

		// 7. Get Fields from Connector
		fields, err := conn.GetFields(ctx, entity.Name)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to get fields", "entity", entity.Name, "error", err)
			continue
		}

		existingFields, _ := s.fieldRepo.GetByEntity(ctx, entityID)

		for _, field := range fields {
			field.EntityID = entityID

			// Check if field exists
			var fieldID types.ID
			var fExists bool
			for _, ef := range existingFields {
				if ef.Name == field.Name {
					fieldID = ef.ID
					fExists = true
					break
				}
			}

			if !fExists {
				if err := s.fieldRepo.Create(ctx, &field); err != nil {
					return err
				}
				fieldID = field.ID
			}

			// 8. Sample & Detect PII
			samples, err := conn.SampleData(ctx, entity.Name, field.Name, 10) // Limit 10 samples
			if err != nil {
				s.logger.WarnContext(ctx, "failed to sample data", "field", field.Name, "error", err)
			}

			// Only run detection if we have samples or use column name heuristic
			detectionInput := detection.Input{
				TableName:  entity.Name,
				ColumnName: field.Name,
				DataType:   field.DataType,
				Samples:    samples,
				// AdjacentColumns: ... (could gather all field names first)
			}

			report, err := s.detector.Detect(ctx, detectionInput)
			if err != nil {
				s.logger.WarnContext(ctx, "detection failed", "field", field.Name, "error", err)
				continue
			}

			if report.IsPII && report.TopMatch != nil {
				piiCount++

				// Create Classification
				cl := discovery.PIIClassification{
					FieldID:         fieldID,
					DataSourceID:    ds.ID,
					EntityName:      entity.Name,
					FieldName:       field.Name,
					Category:        report.TopMatch.Category,
					Type:            report.TopMatch.Type,
					Sensitivity:     report.TopMatch.Sensitivity,
					Confidence:      report.TopMatch.FinalConfidence,
					DetectionMethod: report.TopMatch.Methods[0], // Primary method
					Status:          types.VerificationPending,
					Reasoning:       report.TopMatch.Reasoning,
				}

				if err := s.piiRepo.Create(ctx, &cl); err != nil {
					s.logger.ErrorContext(ctx, "failed to save classification", "error", err)
				}
			}
		}
	}

	// Update inventory stats
	inventory.PIIFieldsCount = piiCount
	s.inventoryRepo.Update(ctx, inventory)

	duration := time.Since(start)
	s.logger.InfoContext(ctx, "scan completed", "duration", duration, "pii_count", piiCount)

	return nil
}

func (s *DiscoveryService) logError(ctx context.Context, dsID types.ID, msg string, err error) {
	s.logger.ErrorContext(ctx, msg, "data_source_id", dsID, "error", err)
}

// GetClassifications returns a paginated list of PII classifications with filters.
func (s *DiscoveryService) GetClassifications(ctx context.Context, tenantID types.ID, filter discovery.ClassificationFilter) (*types.PaginatedResult[discovery.PIIClassification], error) {
	return s.piiRepo.GetClassifications(ctx, tenantID, filter)
}

// TestConnection tests connectivity to a data source.
// It resolves the connector and calls its Connect/Close methods.
func (s *DiscoveryService) TestConnection(ctx context.Context, dataSourceID types.ID) error {
	// 1. Fetch Data Source
	ds, err := s.dsRepo.GetByID(ctx, dataSourceID)
	if err != nil {
		return fmt.Errorf("fetch data source: %w", err)
	}

	// 2. Resolve Connector
	conn, err := s.registry.GetConnector(ds.Type)
	if err != nil {
		return fmt.Errorf("resolve connector: %w", err)
	}

	// 3. Test Connection
	if err := conn.Connect(ctx, ds); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	// Close immediately as we just wanted to test connectivity
	if err := conn.Close(); err != nil {
		s.logger.WarnContext(ctx, "failed to close connection during test", "data_source_id", ds.ID, "error", err)
	}

	return nil
}
