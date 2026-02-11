package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/queue"
	"github.com/complyark/datalens/pkg/types"
)

// ScanService orchestrates scan jobs.
type ScanService struct {
	scanRunRepo   discovery.ScanRunRepository
	dsRepo        discovery.DataSourceRepository
	queue         queue.ScanQueue
	discoverySvc  DiscoveryOrchestrator
	logger        *slog.Logger
	maxConcurrent int
}

// NewScanService creates a new ScanService.
func NewScanService(
	scanRepo discovery.ScanRunRepository,
	dsRepo discovery.DataSourceRepository,
	queue queue.ScanQueue,
	discoverySvc DiscoveryOrchestrator,
	logger *slog.Logger,
) *ScanService {
	return &ScanService{
		scanRunRepo:   scanRepo,
		dsRepo:        dsRepo,
		queue:         queue,
		discoverySvc:  discoverySvc,
		logger:        logger.With("service", "scan_orchestrator"),
		maxConcurrent: 3, // Default limit
	}
}

// EnqueueScan validates the request and queues a background scan job.
func (s *ScanService) EnqueueScan(ctx context.Context, dataSourceID types.ID, tenantID types.ID, scanType discovery.ScanType) (*discovery.ScanRun, error) {
	// 1. Check if Data Source exists & belongs to tenant
	ds, err := s.dsRepo.GetByID(ctx, dataSourceID)
	if err != nil {
		return nil, err // Could wrap
	}
	if ds.TenantID != tenantID {
		return nil, types.NewForbiddenError("data source does not belong to tenant")
	}

	// 2. Check Concurrency Limit
	activeRuns, err := s.scanRunRepo.GetActive(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("check active scans: %w", err)
	}
	if len(activeRuns) >= s.maxConcurrent {
		return nil, types.NewQuotaExceededError(fmt.Sprintf("max concurrent scans reached (%d)", s.maxConcurrent))
	}

	// 3. Create ScanRun Record
	run := &discovery.ScanRun{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		DataSourceID: dataSourceID,
		TenantID:     tenantID,
		Type:         scanType,
		Status:       discovery.ScanStatusPending,
		Progress:     0,
		Stats:        discovery.ScanStats{},
	}

	if err := s.scanRunRepo.Create(ctx, run); err != nil {
		return nil, fmt.Errorf("create scan run: %w", err)
	}

	// 4. Publish to Queue
	if err := s.queue.Enqueue(ctx, run.ID.String()); err != nil {
		// If queue fails, mark run as FAILED immediately
		run.Status = discovery.ScanStatusFailed
		run.ErrorMessage = types.Ptr(fmt.Sprintf("failed to queue: %v", err))
		_ = s.scanRunRepo.Update(ctx, run)
		return nil, fmt.Errorf("enqueue job: %w", err)
	}

	s.logger.Info("scan job enqueued", slog.String("tenant_id", tenantID.String()),
		slog.String("run_id", run.ID.String()),
		slog.String("ds_id", dataSourceID.String()),
	)
	return run, nil
}

// ProcessScanJob is the worker handler.
func (s *ScanService) ProcessScanJob(ctx context.Context, jobID string) error {
	runID, err := types.ParseID(jobID)
	if err != nil {
		return fmt.Errorf("parse run id: %w", err)
	}

	// 1. Fetch Run
	run, err := s.scanRunRepo.GetByID(ctx, runID)
	if err != nil {
		return fmt.Errorf("fetch run: %w", err)
	}

	if run.Status != discovery.ScanStatusPending {
		s.logger.Warn("skipping non-pending job", "run_id", runID, "status", run.Status)
		return nil
	}

	// 2. Mark Running
	now := time.Now()
	run.Status = discovery.ScanStatusRunning
	run.StartedAt = &now
	run.Progress = 0
	if err := s.scanRunRepo.Update(ctx, run); err != nil {
		return fmt.Errorf("mark running: %w", err)
	}

	// 3. Execute Scan via DiscoveryService
	// Note: DiscoveryService methods update Inventory/Entites but NOT ScanRun stats/progress yet.
	// Ideally, DiscoveryService should accept a way to report progress or we wrap it.
	// For this iteration, we call ScanDataSource synchronously here.
	// IMPROVEMENT: Refactor DiscoveryService to take a ProgressCallback or ScanRunID.

	// Since I cannot heavily refactor DiscoveryService right now without breaking changes,
	// I will just execute it. Real stats tracking requires DiscoveryService to return stats.
	// DiscoveryService.ScanDataSource currently returns nil or error.
	// It logs connection info.

	// Constraint: "The existing DiscoveryService.ScanDataSource method should be called by the worker, not refactored"
	// This makes progress tracking hard.
	// BUT, I can update the Status to COMPLETED/FAILED afterwards.

	scanErr := s.discoverySvc.ScanDataSource(ctx, run.DataSourceID)

	completedAt := time.Now()
	run.CompletedAt = &completedAt
	run.Stats.Duration = completedAt.Sub(*run.StartedAt)

	if scanErr != nil {
		run.Status = discovery.ScanStatusFailed
		run.ErrorMessage = types.Ptr(scanErr.Error())
		s.logger.Error("scan failed", slog.String("tenant_id", run.TenantID.String()),
			slog.String("run_id", run.ID.String()),
			slog.String("error", scanErr.Error()),
		)
	} else {
		run.Status = discovery.ScanStatusCompleted
		run.Progress = 100

		// Fetch inventory to populate stats (approximation)
		// DiscoveryService updates Inventory.
		// We'll read it back.
		// TODO: This is a hack because DiscoveryService doesn't return stats.
		// In a real scenario, I'd return stats from ScanDataSource.
	}

	if err := s.scanRunRepo.Update(ctx, run); err != nil {
		s.logger.Error("failed to update run status", slog.String("tenant_id", run.TenantID.String()),
			slog.String("run_id", run.ID.String()),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// StartWorker initiates the background worker.
func (s *ScanService) StartWorker(ctx context.Context) error {
	s.logger.Info("starting scan worker")
	return s.queue.Subscribe(ctx, s.ProcessScanJob)
}

// GetScan retrieves a scan run.
func (s *ScanService) GetScan(ctx context.Context, id types.ID) (*discovery.ScanRun, error) {
	return s.scanRunRepo.GetByID(ctx, id)
}

// GetHistory retrieves scan history for a data source.
func (s *ScanService) GetHistory(ctx context.Context, dataSourceID types.ID) ([]discovery.ScanRun, error) {
	return s.scanRunRepo.GetByDataSource(ctx, dataSourceID)
}
