package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// AdminService handles platform administration tasks.
type AdminService struct {
	tenantRepo    identity.TenantRepository
	userRepo      identity.UserRepository
	roleRepo      identity.RoleRepository
	dsrRepo       compliance.DSRRepository
	retentionRepo compliance.RetentionPolicyRepository
	tenantSvc     *TenantService
	logger        *slog.Logger
}

// NewAdminService creates a new AdminService.
func NewAdminService(
	tenantRepo identity.TenantRepository,
	userRepo identity.UserRepository,
	roleRepo identity.RoleRepository,
	dsrRepo compliance.DSRRepository,
	retentionRepo compliance.RetentionPolicyRepository,
	tenantSvc *TenantService,
	logger *slog.Logger,
) *AdminService {
	return &AdminService{
		tenantRepo:    tenantRepo,
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		dsrRepo:       dsrRepo,
		retentionRepo: retentionRepo,
		tenantSvc:     tenantSvc,
		logger:        logger.With("service", "admin"),
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
