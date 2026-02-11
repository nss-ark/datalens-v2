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

// PolicyService manages governance policies and their enforcement.
type PolicyService struct {
	policyRepo    governance.PolicyRepository
	violationRepo governance.ViolationRepository
	mappingRepo   governance.DataMappingRepository
	dsRepo        discovery.DataSourceRepository
	piiRepo       discovery.PIIClassificationRepository
	eventBus      eventbus.EventBus
	logger        *slog.Logger
}

// NewPolicyService creates a new PolicyService.
func NewPolicyService(
	policyRepo governance.PolicyRepository,
	violationRepo governance.ViolationRepository,
	mappingRepo governance.DataMappingRepository,
	dsRepo discovery.DataSourceRepository,
	piiRepo discovery.PIIClassificationRepository,
	eventBus eventbus.EventBus,
	logger *slog.Logger,
) *PolicyService {
	return &PolicyService{
		policyRepo:    policyRepo,
		violationRepo: violationRepo,
		mappingRepo:   mappingRepo,
		dsRepo:        dsRepo,
		piiRepo:       piiRepo,
		eventBus:      eventBus,
		logger:        logger.With("service", "policy_service"),
	}
}

// CreatePolicy creates a new governance policy.
func (s *PolicyService) CreatePolicy(ctx context.Context, p *governance.Policy) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}

	p.ID = types.NewID()
	p.TenantID = tenantID
	p.IsActive = true
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	if err := s.policyRepo.Create(ctx, p); err != nil {
		return err
	}

	event := eventbus.NewEvent("governance.policy_created", "governance", tenantID, p)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("failed to publish policy created event", "error", err)
		// We don't fail the request if event publishing fails, just log it
	}
	return nil
}

// GetPolicies retrieves all active policies for the tenant.
func (s *PolicyService) GetPolicies(ctx context.Context) ([]governance.Policy, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.policyRepo.GetActive(ctx, tenantID)
}

// GetViolations retrieves violations, optionally filtering by status.
func (s *PolicyService) GetViolations(ctx context.Context, status *governance.ViolationStatus) ([]governance.Violation, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.violationRepo.GetByTenant(ctx, tenantID, status)
}

// EvaluatePolicies runs the policy engine against the tenant's data inventory.
func (s *PolicyService) EvaluatePolicies(ctx context.Context, tenantID types.ID) error {
	s.logger.Info("Starting policy evaluation", "tenant_id", tenantID)

	policies, err := s.policyRepo.GetActive(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("fetch policies: %w", err)
	}

	for _, policy := range policies {
		if err := s.evaluatePolicy(ctx, policy); err != nil {
			s.logger.Error("Policy evaluation failed", "policy_id", policy.ID, "error", err)
			// Continue with other policies
		}
	}

	return nil
}

func (s *PolicyService) evaluatePolicy(ctx context.Context, policy governance.Policy) error {
	// Stub implementation to pass compilation
	return nil
}
