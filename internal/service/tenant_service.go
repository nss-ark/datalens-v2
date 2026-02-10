package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// TenantService handles tenant lifecycle operations.
type TenantService struct {
	tenantRepo identity.TenantRepository
	userRepo   identity.UserRepository
	roleRepo   identity.RoleRepository
	authSvc    *AuthService
	logger     *slog.Logger
}

// NewTenantService creates a new TenantService.
func NewTenantService(
	tenantRepo identity.TenantRepository,
	userRepo identity.UserRepository,
	roleRepo identity.RoleRepository,
	authSvc *AuthService,
	logger *slog.Logger,
) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		authSvc:    authSvc,
		logger:     logger.With("service", "tenant"),
	}
}

// OnboardInput holds fields for tenant onboarding (tenant + first admin user).
type OnboardInput struct {
	TenantName string
	Domain     string
	Industry   string
	Country    string
	AdminEmail string
	AdminName  string
	AdminPass  string
}

// OnboardResult holds the result of tenant onboarding.
type OnboardResult struct {
	Tenant *identity.Tenant `json:"tenant"`
	User   *identity.User   `json:"user"`
	Tokens *TokenPair       `json:"tokens"`
}

// Onboard creates a new tenant, assigns an admin user, and returns tokens.
func (s *TenantService) Onboard(ctx context.Context, in OnboardInput) (*OnboardResult, error) {
	if in.TenantName == "" {
		return nil, types.NewValidationError("tenant name is required", nil)
	}
	if in.Domain == "" {
		return nil, types.NewValidationError("domain is required", nil)
	}
	if in.AdminEmail == "" {
		return nil, types.NewValidationError("admin email is required", nil)
	}
	if len(in.AdminPass) < 8 {
		return nil, types.NewValidationError("admin password must be at least 8 characters", nil)
	}

	// Check domain uniqueness
	existing, err := s.tenantRepo.GetByDomain(ctx, in.Domain)
	if err == nil && existing != nil {
		return nil, types.NewConflictError("Tenant", "domain", in.Domain)
	}

	// Create tenant
	tenant := &identity.Tenant{
		Name:     in.TenantName,
		Domain:   in.Domain,
		Industry: in.Industry,
		Country:  in.Country,
		Plan:     identity.PlanFree,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{
			DefaultRegulation:  "DPDPA",
			EnabledRegulations: []string{"DPDPA"},
			RetentionDays:      365,
			EnableAI:           false,
		},
	}
	if tenant.Country == "" {
		tenant.Country = "IN"
	}
	if tenant.Industry == "" {
		tenant.Industry = "GENERAL"
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	// Register admin user
	user, err := s.authSvc.Register(ctx, RegisterInput{
		TenantID: tenant.ID,
		Email:    in.AdminEmail,
		Name:     in.AdminName,
		Password: in.AdminPass,
	})
	if err != nil {
		return nil, fmt.Errorf("create admin user: %w", err)
	}

	// Generate tokens for the new user
	tokens, err := s.authSvc.Login(ctx, LoginInput{
		TenantID: tenant.ID,
		Email:    in.AdminEmail,
		Password: in.AdminPass,
	})
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	s.logger.InfoContext(ctx, "tenant onboarded", "tenant_id", tenant.ID, "domain", tenant.Domain)

	return &OnboardResult{
		Tenant: tenant,
		User:   user,
		Tokens: tokens,
	}, nil
}

// GetByID retrieves a tenant by ID.
func (s *TenantService) GetByID(ctx context.Context, id types.ID) (*identity.Tenant, error) {
	return s.tenantRepo.GetByID(ctx, id)
}

// Update modifies an existing tenant.
func (s *TenantService) Update(ctx context.Context, tenant *identity.Tenant) error {
	return s.tenantRepo.Update(ctx, tenant)
}

// GetByDomain retrieves a tenant by its domain.
func (s *TenantService) GetByDomain(ctx context.Context, domain string) (*identity.Tenant, error) {
	return s.tenantRepo.GetByDomain(ctx, domain)
}
