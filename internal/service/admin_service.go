package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// AdminService handles platform administration tasks.
type AdminService struct {
	tenantRepo identity.TenantRepository
	userRepo   identity.UserRepository
	roleRepo   identity.RoleRepository
	tenantSvc  *TenantService
	logger     *slog.Logger
}

// NewAdminService creates a new AdminService.
func NewAdminService(
	tenantRepo identity.TenantRepository,
	userRepo identity.UserRepository,
	roleRepo identity.RoleRepository,
	tenantSvc *TenantService,
	logger *slog.Logger,
) *AdminService {
	return &AdminService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		tenantSvc:  tenantSvc,
		logger:     logger.With("service", "admin"),
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
