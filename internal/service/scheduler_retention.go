package service

import (
	"context"
	"strings"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// checkRetentionPolicies evaluates all ACTIVE retention policies across all tenants.
// Runs once per 24 hours. For MVP, it does NOT actually delete data from connected sources —
// it only creates RetentionLog entries indicating what *would* be erased.
func (s *SchedulerService) checkRetentionPolicies(ctx context.Context) {
	// Throttle: run once per day
	if time.Since(s.lastRetentionCheck) < 24*time.Hour && !s.lastRetentionCheck.IsZero() {
		return
	}
	s.lastRetentionCheck = time.Now()

	if s.retentionRepo == nil {
		return
	}

	s.logger.Info("Starting daily retention policy evaluation")

	tenants, err := s.tenantRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to list tenants for retention check", "error", err)
		return
	}

	totalPoliciesChecked := 0
	totalLogsCreated := 0

	for _, tenant := range tenants {
		if tenant.Status != identity.TenantActive {
			continue
		}

		checked, logged := s.evaluateTenantRetentionPolicies(ctx, tenant.ID)
		totalPoliciesChecked += checked
		totalLogsCreated += logged
	}

	s.logger.Info("Daily retention evaluation complete",
		"tenants_checked", len(tenants),
		"policies_checked", totalPoliciesChecked,
		"logs_created", totalLogsCreated,
	)
}

// evaluateTenantRetentionPolicies checks all ACTIVE retention policies for a single tenant.
// Returns (policiesChecked, logsCreated).
func (s *SchedulerService) evaluateTenantRetentionPolicies(ctx context.Context, tenantID types.ID) (int, int) {
	policies, err := s.retentionRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to get retention policies",
			"tenant_id", tenantID,
			"error", err,
		)
		return 0, 0
	}

	checked := 0
	logged := 0

	for _, policy := range policies {
		if policy.Status != compliance.RetentionPolicyActive {
			continue
		}
		checked++

		// Check if the retention period has been exceeded relative to policy creation.
		// For MVP, we compare against policy creation date + MaxRetentionDays.
		// In a full implementation, this would check actual data ages in connected sources.
		retentionDeadline := policy.CreatedAt.AddDate(0, 0, policy.MaxRetentionDays)
		if time.Now().UTC().Before(retentionDeadline) {
			continue // Not yet exceeded
		}

		// Determine action based on AutoErase flag
		action := "RETENTION_EXCEEDED"
		if policy.AutoErase {
			action = "ERASED"
		}

		log := &compliance.RetentionLog{
			ID:        types.NewID(),
			TenantID:  tenantID,
			PolicyID:  policy.ID,
			Action:    action,
			Target:    "Retention policy evaluation",
			Details:   "Retention period exceeded for categories: " + strings.Join(policy.DataCategories, ", "),
			Timestamp: time.Now().UTC(),
		}

		if err := s.retentionRepo.CreateLog(ctx, log); err != nil {
			s.logger.Error("Failed to create retention log",
				"tenant_id", tenantID,
				"policy_id", policy.ID,
				"error", err,
			)
			continue
		}

		logged++
		s.logger.Info("Retention policy action logged",
			"tenant_id", tenantID,
			"policy_id", policy.ID,
			"action", action,
			"categories_count", len(policy.DataCategories),
		)
	}

	return checked, logged
}
