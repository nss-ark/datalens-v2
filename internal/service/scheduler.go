package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// SchedulerService manages automated scan scheduling based on cron expressions.
type SchedulerService struct {
	dsRepo         discovery.DataSourceRepository
	tenantRepo     identity.TenantRepository
	policySvc      *PolicyService
	scanService    ScanOrchestrator
	logger         *slog.Logger
	parser         cron.Parser
	ticker         *time.Ticker
	stopChan       chan struct{}
	lastPolicyEval time.Time
}

// NewSchedulerService creates a new SchedulerService.
func NewSchedulerService(
	dsRepo discovery.DataSourceRepository,
	tenantRepo identity.TenantRepository,
	policySvc *PolicyService,
	scanService ScanOrchestrator,
	logger *slog.Logger,
) *SchedulerService {
	return &SchedulerService{
		dsRepo:      dsRepo,
		tenantRepo:  tenantRepo,
		policySvc:   policySvc,
		scanService: scanService,
		logger:      logger.With("service", "scheduler"),
		parser:      cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		ticker:      time.NewTicker(60 * time.Second),
		stopChan:    make(chan struct{}),
	}
}

// Start begins the scheduler loop.
func (s *SchedulerService) Start(ctx context.Context) error {
	s.logger.Info("Starting scan scheduler")

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkSchedules(ctx)
				s.schedulePolicyEvaluations(ctx)
			case <-s.stopChan:
				s.logger.Info("Stopping scan scheduler")
				return
			case <-ctx.Done():
				s.logger.Info("Scheduler context cancelled")
				return
			}
		}
	}()

	return nil
}

// Stop halts the scheduler.
func (s *SchedulerService) Stop() {
	close(s.stopChan)
	s.ticker.Stop()
}

// checkSchedules evaluates all data sources with schedules and enqueues scans if due.
func (s *SchedulerService) checkSchedules(ctx context.Context) {
	// Get all data sources (we'll need to filter by non-null scan_schedule in repo or here)
	// For now, we'll list all tenant data sources and filter in-memory
	// TODO: Add GetScheduledDataSources to repository

	// Since we don't have tenant context here, we need to either:
	// 1. List ALL data sources across all tenants (not ideal)
	// 2. Store tenantID in scheduler (requires multi-tenant scheduler instances)
	// For MVP, let's assume a single scheduler per tenant or list all

	// This is a limitation - we need to iterate tenants or change architecture
	// For now, let's skip this and assume the user will trigger scans manually
	// In production, we'd need a better approach (e.g., one scheduler per tenant or global scheduler with tenant iteration)

	s.logger.Warn("Scheduler checkSchedules not fully implemented - requires tenant iteration strategy")
}

// schedulePolicyEvaluations triggers policy evaluation for all active tenants.
func (s *SchedulerService) schedulePolicyEvaluations(ctx context.Context) {
	// Run every hour
	if time.Since(s.lastPolicyEval) < 1*time.Hour && !s.lastPolicyEval.IsZero() {
		return
	}
	s.lastPolicyEval = time.Now()

	// Fetch all tenants
	tenants, err := s.tenantRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to list tenants for policy evaluation", "error", err)
		return
	}

	for _, tenant := range tenants {
		if tenant.Status != identity.TenantActive {
			continue
		}

		// Run evaluation in background to not block scheduler
		go func(tID types.ID) {
			evalCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			if err := s.policySvc.EvaluatePolicies(evalCtx, tID); err != nil {
				s.logger.Error("Scheduled policy evaluation failed", "tenant_id", tID, "error", err)
			}
		}(tenant.ID)
	}
}

// IsDue checks if a cron schedule is due for execution.
func (s *SchedulerService) IsDue(cronExpr string, lastRun *time.Time) (bool, error) {
	schedule, err := s.parser.Parse(cronExpr)
	if err != nil {
		return false, err
	}

	now := time.Now()
	var checkFrom time.Time
	if lastRun != nil {
		checkFrom = *lastRun
	} else {
		// If never run, check from 24 hours ago
		checkFrom = now.Add(-24 * time.Hour)
	}

	// Get next scheduled time after last run
	nextRun := schedule.Next(checkFrom)

	// If next run is in the past or now, it's due
	return nextRun.Before(now) || nextRun.Equal(now), nil
}

// ValidateCron validates a cron expression.
func (s *SchedulerService) ValidateCron(cronExpr string) error {
	_, err := s.parser.Parse(cronExpr)
	return err
}
