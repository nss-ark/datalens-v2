package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTenantRepoExtended extends the standard mock with Admin-specific methods.
type mockTenantRepoExtended struct {
	*mockTenantRepo
	searchFunc   func(ctx context.Context, filter identity.TenantFilter) ([]identity.Tenant, int, error)
	getStatsFunc func(ctx context.Context) (*identity.TenantStats, error)
}

func (m *mockTenantRepoExtended) Search(ctx context.Context, filter identity.TenantFilter) ([]identity.Tenant, int, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (m *mockTenantRepoExtended) GetStats(ctx context.Context) (*identity.TenantStats, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx)
	}
	return &identity.TenantStats{}, nil
}

// mockUserRepoExtended extends the standard mock with Admin-specific methods.
type mockUserRepoExtended struct {
	*mockUserRepo
	countGlobalFunc      func(ctx context.Context) (int64, error)
	getByEmailGlobalFunc func(ctx context.Context, email string) (*identity.User, error)
	searchGlobalFunc     func(ctx context.Context, filter identity.UserFilter) ([]identity.User, int, error)
	updateStatusFunc     func(ctx context.Context, id types.ID, status identity.UserStatus) error
	assignRolesFunc      func(ctx context.Context, userID types.ID, roleIDs []types.ID) error
}

func (m *mockUserRepoExtended) SearchGlobal(ctx context.Context, filter identity.UserFilter) ([]identity.User, int, error) {
	if m.searchGlobalFunc != nil {
		return m.searchGlobalFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (m *mockUserRepoExtended) UpdateStatus(ctx context.Context, id types.ID, status identity.UserStatus) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, id, status)
	}
	// Default mock behavior: update in memory map if exists
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[id]; ok {
		u.Status = status
		return nil
	}
	return types.NewNotFoundError("user", id)
}

func (m *mockUserRepoExtended) AssignRoles(ctx context.Context, userID types.ID, roleIDs []types.ID) error {
	if m.assignRolesFunc != nil {
		return m.assignRolesFunc(ctx, userID, roleIDs)
	}
	// Default mock behavior
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[userID]; ok {
		u.RoleIDs = roleIDs
		return nil
	}
	return types.NewNotFoundError("user", userID)
}

func (m *mockUserRepoExtended) CountGlobal(ctx context.Context) (int64, error) {
	if m.countGlobalFunc != nil {
		return m.countGlobalFunc(ctx)
	}
	return 0, nil
}

func (m *mockUserRepoExtended) GetByEmailGlobal(ctx context.Context, email string) (*identity.User, error) {
	if m.getByEmailGlobalFunc != nil {
		return m.getByEmailGlobalFunc(ctx, email)
	}
	return nil, nil // Return nil if not found
}

func newTestAdminService() (*AdminService, *mockTenantRepoExtended, *mockUserRepoExtended) {
	baseTenantRepo := newMockTenantRepo()
	tenantRepo := &mockTenantRepoExtended{mockTenantRepo: baseTenantRepo}

	baseUserRepo := newMockUserRepo()
	userRepo := &mockUserRepoExtended{mockUserRepo: baseUserRepo}

	roleRepo := newMockRoleRepo()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	auditRepo := newMockAuditRepo()
	auditSvc := NewAuditService(auditRepo, logger)

	authSvc := NewAuthService(userRepo, roleRepo, "test-secret-key-32chars!!", 15*time.Minute, 7*24*time.Hour, logger, auditSvc)
	tenantSvc := NewTenantService(tenantRepo, userRepo, roleRepo, authSvc, logger)

	svc := NewAdminService(tenantRepo, userRepo, roleRepo, nil, nil, tenantSvc, logger)
	return svc, tenantRepo, userRepo
}

func TestAdminService_ListTenants(t *testing.T) {
	svc, tenantRepo, _ := newTestAdminService()
	ctx := context.Background()

	expectedTenants := []identity.Tenant{
		{Name: "Tenant A", Domain: "a.local", Status: identity.TenantActive},
		{Name: "Tenant B", Domain: "b.local", Status: identity.TenantSuspended},
	}

	tenantRepo.searchFunc = func(ctx context.Context, filter identity.TenantFilter) ([]identity.Tenant, int, error) {
		return expectedTenants, 2, nil
	}

	filter := identity.TenantFilter{Limit: 10, Offset: 0}
	tenants, total, err := svc.ListTenants(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, tenants, 2)
	assert.Equal(t, "Tenant A", tenants[0].Name)
}

func TestAdminService_GetStats(t *testing.T) {
	svc, tenantRepo, userRepo := newTestAdminService()
	ctx := context.Background()

	tenantRepo.getStatsFunc = func(ctx context.Context) (*identity.TenantStats, error) {
		return &identity.TenantStats{
			TotalTenants:  10,
			ActiveTenants: 8,
		}, nil
	}

	userRepo.countGlobalFunc = func(ctx context.Context) (int64, error) {
		return 50, nil
	}

	stats, err := svc.GetStats(ctx)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(10), stats.TotalTenants)
	assert.Equal(t, int64(8), stats.ActiveTenants)
	assert.Equal(t, int64(50), stats.TotalUsers)
}

func TestAdminService_OnboardTenant(t *testing.T) {
	svc, tenantRepo, userRepo := newTestAdminService()
	ctx := context.Background()

	// Implement required mocks for checking duplicates
	tenantRepo.searchFunc = func(ctx context.Context, filter identity.TenantFilter) ([]identity.Tenant, int, error) {
		return nil, 0, nil
	}
	userRepo.getByEmailGlobalFunc = func(ctx context.Context, email string) (*identity.User, error) {
		return nil, types.NewNotFoundError("user", "email")
	}

	input := OnboardInput{
		TenantName: "New Corp",
		Domain:     "new.corp",
		AdminEmail: "admin@new.corp",
		AdminName:  "Admin",
		AdminPass:  "securePass123!",
	}

	result, err := svc.OnboardTenant(ctx, input)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Corp", result.Tenant.Name)
	assert.Equal(t, "new.corp", result.Tenant.Domain)
	assert.Equal(t, "admin@new.corp", result.User.Email)
}
