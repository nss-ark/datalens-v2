package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/identity"
)

// AdminService handles platform administration tasks.
type AdminService struct {
	tenantRepo identity.TenantRepository
	userRepo   identity.UserRepository
	tenantSvc  *TenantService
	logger     *slog.Logger
}

// NewAdminService creates a new AdminService.
func NewAdminService(
	tenantRepo identity.TenantRepository,
	userRepo identity.UserRepository,
	tenantSvc *TenantService,
	logger *slog.Logger,
) *AdminService {
	return &AdminService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
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
