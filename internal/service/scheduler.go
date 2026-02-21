package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// SchedulerService manages automated scan scheduling based on cron expressions.
type SchedulerService struct {
	dsRepo             discovery.DataSourceRepository
	tenantRepo         identity.TenantRepository
	policySvc          *PolicyService
	scanService        ScanOrchestrator
	expirySvc          *ConsentExpiryService
	retentionRepo      compliance.RetentionPolicyRepository
	logger             *slog.Logger
	parser             cron.Parser
	ticker             *time.Ticker
	stopChan           chan struct{}
	lastPolicyEval     time.Time
	lastRetentionCheck time.Time
}

// NewSchedulerService creates a new SchedulerService.
func NewSchedulerService(
	dsRepo discovery.DataSourceRepository,
	tenantRepo identity.TenantRepository,
	policySvc *PolicyService,
	scanService ScanOrchestrator,
	expirySvc *ConsentExpiryService,
	retentionRepo compliance.RetentionPolicyRepository,
	logger *slog.Logger,
) *SchedulerService {
	return &SchedulerService{
		dsRepo:        dsRepo,
		tenantRepo:    tenantRepo,
		policySvc:     policySvc,
		scanService:   scanService,
		expirySvc:     expirySvc,
		retentionRepo: retentionRepo,
		logger:        logger.With("service", "scheduler"),
		parser:        cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		ticker:        time.NewTicker(60 * time.Second),
		stopChan:      make(chan struct{}),
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
				s.checkConsentExpiries(ctx)
				s.checkRetentionPolicies(ctx)
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
	// 1. Iterate over all active tenants
	tenants, err := s.tenantRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to list tenants for scheduling", "error", err)
		return
	}

	for _, tenant := range tenants {
		if tenant.Status != identity.TenantActive {
			continue
		}

		func(t identity.Tenant) {
			// 2. Get data sources for tenant
			dataSources, err := s.dsRepo.GetByTenant(ctx, t.ID)
			if err != nil {
				s.logger.Error("failed to list data sources", "tenant_id", t.ID, "error", err)
				return
			}

			// 3. Check schedules
			for _, ds := range dataSources {
				if ds.ScanSchedule == nil || *ds.ScanSchedule == "" {
					continue
				}

				// Only schedule connected data sources
				if ds.Status != discovery.ConnectionStatusConnected {
					continue
				}

				// Check if due
				isDue, err := s.IsDue(*ds.ScanSchedule, ds.LastSyncAt)
				if err != nil {
					s.logger.Error("invalid cron schedule",
						"tenant_id", t.ID,
						"ds_id", ds.ID,
						"schedule", *ds.ScanSchedule,
						"error", err)
					continue
				}

				if isDue {
					s.logger.Info("triggering scheduled scan",
						"tenant_id", t.ID,
						"ds_id", ds.ID,
						"schedule", *ds.ScanSchedule)

					// Trigger scan
					if _, err := s.scanService.EnqueueScan(ctx, ds.ID, t.ID, discovery.ScanTypeFull); err != nil {
						s.logger.Error("failed to enqueue scheduled scan",
							"tenant_id", t.ID,
							"ds_id", ds.ID,
							"error", err)
					}
				}
			}
		}(tenant)
	}
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
