package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// AdminService handles platform administration tasks.
type AdminService struct {
	tenantRepo       identity.TenantRepository
	userRepo         identity.UserRepository
	roleRepo         identity.RoleRepository
	dsrRepo          compliance.DSRRepository
	retentionRepo    compliance.RetentionPolicyRepository
	subscriptionRepo identity.SubscriptionRepository
	moduleAccessRepo identity.ModuleAccessRepository
	settingsRepo     identity.PlatformSettingsRepository

	tenantSvc *TenantService
	logger    *slog.Logger
}

// NewAdminService creates a new AdminService.
func NewAdminService(
	tenantRepo identity.TenantRepository,
	userRepo identity.UserRepository,
	roleRepo identity.RoleRepository,
	dsrRepo compliance.DSRRepository,
	retentionRepo compliance.RetentionPolicyRepository,
	subscriptionRepo identity.SubscriptionRepository,
	moduleAccessRepo identity.ModuleAccessRepository,
	settingsRepo identity.PlatformSettingsRepository,
	tenantSvc *TenantService,
	logger *slog.Logger,
) *AdminService {
	return &AdminService{
		tenantRepo:       tenantRepo,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		dsrRepo:          dsrRepo,
		retentionRepo:    retentionRepo,
		subscriptionRepo: subscriptionRepo,
		moduleAccessRepo: moduleAccessRepo,
		settingsRepo:     settingsRepo,
		tenantSvc:        tenantSvc,
		logger:           logger.With("service", "admin"),
	}
}

// ListTenants retrieves a paginated list of tenants.
func (s *AdminService) ListTenants(ctx context.Context, filter identity.TenantFilter) ([]identity.Tenant, int, error) {
	return s.tenantRepo.Search(ctx, filter)
}

// OnboardTenant handles the creation of a new tenant and its administrator.
func (s *AdminService) OnboardTenant(ctx context.Context, input OnboardInput) (*OnboardResult, error) {
	// Re-use existing tenant onboarding logic which handles validation,
	// tenant creation, user creation, and role assignment.
	return s.tenantSvc.Onboard(ctx, input)
}

// GetTenant retrieves a single tenant by ID.
func (s *AdminService) GetTenant(ctx context.Context, id types.ID) (*identity.Tenant, error) {
	return s.tenantSvc.GetByID(ctx, id)
}

// UpdateTenant updates a tenant's details.
func (s *AdminService) UpdateTenant(ctx context.Context, id types.ID, update identity.Tenant) (*identity.Tenant, error) {
	// 1. Get existing
	existing, err := s.tenantSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Apply updates (only allow specific fields for now)
	if update.Name != "" {
		existing.Name = update.Name
	}
	if update.Industry != "" {
		existing.Industry = update.Industry
	}
	if update.Country != "" {
		existing.Country = update.Country
	}
	if update.Plan != "" {
		existing.Plan = update.Plan
	}
	if update.Status != "" {
		existing.Status = update.Status
	}
	// Settings updates
	if update.Settings.RetentionDays > 0 {
		existing.Settings.RetentionDays = update.Settings.RetentionDays
	}
	// EnableAI logic
	existing.Settings.EnableAI = update.Settings.EnableAI

	// 3. Save
	if err := s.tenantSvc.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update tenant: %w", err)
	}

	return existing, nil
}

// GlobalStats returns aggregate platform statistics.
type GlobalStats struct {
	TotalTenants  int64 `json:"total_tenants"`
	ActiveTenants int64 `json:"active_tenants"`
	TotalUsers    int64 `json:"total_users"`
}

// GetStats gathers platform-wide statistics.
func (s *AdminService) GetStats(ctx context.Context) (*GlobalStats, error) {
	tenantStats, err := s.tenantRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tenant stats: %w", err)
	}

	userCount, err := s.userRepo.CountGlobal(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	return &GlobalStats{
		TotalTenants:  tenantStats.TotalTenants,
		ActiveTenants: tenantStats.ActiveTenants,
		TotalUsers:    userCount,
	}, nil
}

// ListUsers retrieves a paginated list of users across all tenants.
func (s *AdminService) ListUsers(ctx context.Context, filter identity.UserFilter) ([]identity.User, int, error) {
	return s.userRepo.SearchGlobal(ctx, filter)
}

// GetUser retrieves a user by ID.
func (s *AdminService) GetUser(ctx context.Context, id types.ID) (*identity.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// SuspendUser suspends a user account.
func (s *AdminService) SuspendUser(ctx context.Context, id types.ID) error {
	return s.userRepo.UpdateStatus(ctx, id, identity.UserSuspended)
}

// ActivateUser activates a user account.
func (s *AdminService) ActivateUser(ctx context.Context, id types.ID) error {
	return s.userRepo.UpdateStatus(ctx, id, identity.UserActive)
}

// AssignRoles assigns system roles to a user.
func (s *AdminService) AssignRoles(ctx context.Context, userID types.ID, roleIDs []types.ID) error {
	// TODO: Validate that roles exist and are system roles or tenant roles appropriate for the user
	return s.userRepo.AssignRoles(ctx, userID, roleIDs)
}

// ListRoles retrieves all system roles.
// Note: For now we only expose system roles to platform admins.
// Tenant-specific roles are managed by tenant admins.
func (s *AdminService) ListRoles(ctx context.Context) ([]identity.Role, error) {
	return s.roleRepo.GetSystemRoles(ctx)
}

// GetAllDSRs retrieves all DSRs across all tenants.
func (s *AdminService) GetAllDSRs(ctx context.Context, pagination types.Pagination, status *compliance.DSRStatus, reqType *compliance.DSRRequestType) (*types.PaginatedResult[compliance.DSR], error) {
	return s.dsrRepo.GetAll(ctx, pagination, status, reqType)
}

// GetDSR retrieves a specific DSR by ID (cross-tenant).
func (s *AdminService) GetDSR(ctx context.Context, id types.ID) (*compliance.DSR, error) {
	return s.dsrRepo.GetByID(ctx, id)
}

// -------------------------------------------------------------------------
// Retention Policies (DPDP R8)
// -------------------------------------------------------------------------

// CreateRetentionPolicy creates a new data retention policy.
func (s *AdminService) CreateRetentionPolicy(ctx context.Context, req compliance.RetentionPolicy) (*compliance.RetentionPolicy, error) {
	// Validate
	if req.PurposeID == (types.ID{}) {
		return nil, types.NewValidationError("purpose_id is required", nil)
	}
	if req.MaxRetentionDays < 1 {
		return nil, types.NewValidationError("max_retention_days must be greater than 0", nil)
	}

	policy := &compliance.RetentionPolicy{
		ID:               types.NewID(),
		TenantID:         req.TenantID,
		PurposeID:        req.PurposeID,
		MaxRetentionDays: req.MaxRetentionDays,
		DataCategories:   req.DataCategories,
		Status:           compliance.RetentionPolicyActive,
		AutoErase:        req.AutoErase,
		Description:      req.Description,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := s.retentionRepo.Create(ctx, policy); err != nil {
		return nil, fmt.Errorf("create retention policy: %w", err)
	}

	return policy, nil
}

// GetRetentionPolicy retrieves a retention policy by ID.
func (s *AdminService) GetRetentionPolicy(ctx context.Context, id types.ID) (*compliance.RetentionPolicy, error) {
	return s.retentionRepo.GetByID(ctx, id)
}

// ListRetentionPolicies retrieves all retention policies for a tenant.
func (s *AdminService) ListRetentionPolicies(ctx context.Context, tenantID types.ID) ([]compliance.RetentionPolicy, error) {
	return s.retentionRepo.GetByTenant(ctx, tenantID)
}

// UpdateRetentionPolicy updates an existing retention policy.
func (s *AdminService) UpdateRetentionPolicy(ctx context.Context, id types.ID, update compliance.RetentionPolicy) (*compliance.RetentionPolicy, error) {
	policy, err := s.retentionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update allowed fields
	if update.MaxRetentionDays > 0 {
		policy.MaxRetentionDays = update.MaxRetentionDays
	}
	if len(update.DataCategories) > 0 {
		policy.DataCategories = update.DataCategories
	}
	if update.Description != "" {
		policy.Description = update.Description
	}
	if update.Status != "" {
		policy.Status = update.Status
	}
	// Toggle boolean? Need explicit pointer or dedicated method for booleans usually, but sticking to simple struct replacement for design task
	policy.AutoErase = update.AutoErase

	policy.UpdatedAt = time.Now().UTC()

	if err := s.retentionRepo.Update(ctx, policy); err != nil {
		return nil, fmt.Errorf("update retention policy: %w", err)
	}

	return policy, nil
}

// -------------------------------------------------------------------------
// Subscription Management
// -------------------------------------------------------------------------

// GetSubscription retrieves the subscription for a tenant.
// If none exists, a default FREE subscription is created.
func (s *AdminService) GetSubscription(ctx context.Context, tenantID types.ID) (*identity.Subscription, error) {
	sub, err := s.subscriptionRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		// If not found, create a default subscription
		if types.IsNotFoundError(err) {
			sub = &identity.Subscription{
				TenantID:   tenantID,
				Plan:       identity.PlanFree,
				AutoRevoke: true,
				Status:     identity.SubscriptionActive,
			}
			if createErr := s.subscriptionRepo.Create(ctx, sub); createErr != nil {
				return nil, fmt.Errorf("create default subscription: %w", createErr)
			}
			// Also seed default modules for FREE plan
			if seedErr := s.ApplyPlanDefaults(ctx, tenantID, identity.PlanFree); seedErr != nil {
				s.logger.WarnContext(ctx, "failed to seed default modules", "error", seedErr)
			}
			return sub, nil
		}
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return sub, nil
}

// UpdateSubscription updates a tenant's subscription.
func (s *AdminService) UpdateSubscription(ctx context.Context, tenantID types.ID, update identity.Subscription) (*identity.Subscription, error) {
	existing, err := s.GetSubscription(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	oldPlan := existing.Plan

	if update.Plan != "" {
		existing.Plan = update.Plan
	}
	if update.BillingStart != nil {
		existing.BillingStart = update.BillingStart
	}
	if update.BillingEnd != nil {
		existing.BillingEnd = update.BillingEnd
	}
	existing.AutoRevoke = update.AutoRevoke
	if update.Status != "" {
		existing.Status = update.Status
	}

	if err := s.subscriptionRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update subscription: %w", err)
	}

	// If plan changed, also update the tenant.plan column and seed modules
	if oldPlan != existing.Plan {
		tenant, err := s.tenantSvc.GetByID(ctx, tenantID)
		if err == nil {
			tenant.Plan = existing.Plan
			_ = s.tenantSvc.Update(ctx, tenant)
		}
		if seedErr := s.ApplyPlanDefaults(ctx, tenantID, existing.Plan); seedErr != nil {
			s.logger.WarnContext(ctx, "failed to apply plan defaults", "error", seedErr)
		}
	}

	return existing, nil
}

// -------------------------------------------------------------------------
// Module Access
// -------------------------------------------------------------------------

// GetModuleAccess retrieves module access for a tenant.
// If none exist, seeds from PlanModuleDefaults based on current plan.
func (s *AdminService) GetModuleAccess(ctx context.Context, tenantID types.ID) ([]identity.ModuleAccess, error) {
	modules, err := s.moduleAccessRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get module access: %w", err)
	}

	// If empty, seed from defaults
	if len(modules) == 0 {
		tenant, tErr := s.tenantSvc.GetByID(ctx, tenantID)
		if tErr != nil {
			return nil, fmt.Errorf("get tenant for module defaults: %w", tErr)
		}
		if seedErr := s.ApplyPlanDefaults(ctx, tenantID, tenant.Plan); seedErr != nil {
			return nil, fmt.Errorf("seed module defaults: %w", seedErr)
		}
		return s.moduleAccessRepo.GetByTenantID(ctx, tenantID)
	}

	return modules, nil
}

// ModuleAccessInput is the request body for updating module access.
type ModuleAccessInput struct {
	ModuleName identity.ModuleName `json:"module_name"`
	Enabled    bool                `json:"enabled"`
}

// SetModuleAccess replaces all module access for a tenant.
func (s *AdminService) SetModuleAccess(ctx context.Context, tenantID types.ID, inputs []ModuleAccessInput) ([]identity.ModuleAccess, error) {
	modules := make([]identity.ModuleAccess, len(inputs))
	for i, in := range inputs {
		modules[i] = identity.ModuleAccess{
			TenantID:   tenantID,
			ModuleName: in.ModuleName,
			Enabled:    in.Enabled,
		}
	}

	if err := s.moduleAccessRepo.SetModules(ctx, tenantID, modules); err != nil {
		return nil, fmt.Errorf("set module access: %w", err)
	}

	return s.moduleAccessRepo.GetByTenantID(ctx, tenantID)
}

// ApplyPlanDefaults seeds module_access rows from PlanModuleDefaults.
func (s *AdminService) ApplyPlanDefaults(ctx context.Context, tenantID types.ID, plan identity.PlanType) error {
	enabled := identity.PlanModuleDefaults[plan]
	enabledSet := make(map[identity.ModuleName]bool, len(enabled))
	for _, m := range enabled {
		enabledSet[m] = true
	}

	// Build full module list — all modules present, enabled flag set per plan
	var modules []identity.ModuleAccess
	for _, m := range identity.AllModules {
		modules = append(modules, identity.ModuleAccess{
			TenantID:   tenantID,
			ModuleName: m,
			Enabled:    enabledSet[m],
		})
	}

	return s.moduleAccessRepo.SetModules(ctx, tenantID, modules)
}

// GetPlatformSettings retrieves all platform configuration.
func (s *AdminService) GetPlatformSettings(ctx context.Context) (map[string]any, error) {
	settings, err := s.settingsRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}

	result := make(map[string]any)
	for _, setting := range settings {
		var val any
		if err := json.Unmarshal(setting.Value, &val); err != nil {
			s.logger.WarnContext(ctx, "failed to unmarshal setting value", "key", setting.Key, "error", err)
			continue
		}
		result[setting.Key] = val
	}
	return result, nil
}

// UpdatePlatformSetting updates a single platform setting key.
func (s *AdminService) UpdatePlatformSetting(ctx context.Context, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal setting value: %w", err)
	}

	setting := &identity.PlatformSetting{
		Key:   key,
		Value: raw,
	}
	return s.settingsRepo.Set(ctx, setting)
}

// CheckSubscriptionExpiry checks for expired subscriptions and suspends tenants.
func (s *AdminService) CheckSubscriptionExpiry(ctx context.Context) error {
	subs, err := s.subscriptionRepo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("get active subscriptions: %w", err)
	}

	now := time.Now().UTC()
	for i := range subs {
		sub := &subs[i]

		// Skip if no billing end date
		if sub.BillingEnd == nil {
			continue
		}

		// Check if expired
		if sub.BillingEnd.Before(now) {
			s.logger.InfoContext(ctx, "subscription expired", "tenant_id", sub.TenantID, "billing_end", sub.BillingEnd)

			// 1. Mark subscription as expired
			sub.Status = identity.SubscriptionExpired
			if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
				s.logger.ErrorContext(ctx, "failed to update subscription status", "error", err, "tenant_id", sub.TenantID)
				continue
			}

			// 2. If auto-revoke is enabled, suspend the tenant
			if sub.AutoRevoke {
				s.logger.InfoContext(ctx, "auto-revoking tenant due to expiry", "tenant_id", sub.TenantID)

				// Suspend the tenant
				_, err := s.UpdateTenant(ctx, sub.TenantID, identity.Tenant{Status: identity.TenantSuspended})
				if err != nil {
					s.logger.ErrorContext(ctx, "failed to suspend tenant", "error", err, "tenant_id", sub.TenantID)
				}
			}
		} else {
			// Check for warning (e.g. 7 days left)
			daysLeft := sub.BillingEnd.Sub(now).Hours() / 24
			if daysLeft <= 7 && daysLeft > 0 {
				s.logger.InfoContext(ctx, "subscription expiring soon", "tenant_id", sub.TenantID, "days_left", daysLeft)
				// TODO: Send email notification
			}
		}
	}
	return nil
}
