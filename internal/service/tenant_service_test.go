package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTenantService() (*TenantService, *mockTenantRepo, *mockUserRepo) {
	tenantRepo := newMockTenantRepo()
	userRepo := newMockUserRepo()
	roleRepo := newMockRoleRepo()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	auditRepo := newMockAuditRepo()
	auditSvc := NewAuditService(auditRepo, logger)

	authSvc := NewAuthService(userRepo, roleRepo, "test-secret-key-32chars!!", 15*time.Minute, 7*24*time.Hour, logger, auditSvc)
	svc := NewTenantService(tenantRepo, userRepo, roleRepo, authSvc, logger)
	return svc, tenantRepo, userRepo
}

func TestTenantService_Onboard_Success(t *testing.T) {
	svc, _, _ := newTestTenantService()
	ctx := context.Background()

	result, err := svc.Onboard(ctx, OnboardInput{
		TenantName: "Acme Corp",
		Domain:     "acme.local",
		Industry:   "FINTECH",
		Country:    "IN",
		AdminEmail: "admin@acme.local",
		AdminName:  "Acme Admin",
		AdminPass:  "securepassword123",
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Tenant)
	assert.NotNil(t, result.User)
	assert.NotNil(t, result.Tokens)
	assert.Equal(t, "Acme Corp", result.Tenant.Name)
	assert.Equal(t, "acme.local", result.Tenant.Domain)
	assert.Equal(t, "admin@acme.local", result.User.Email)
	assert.NotEmpty(t, result.Tokens.AccessToken)
}

func TestTenantService_Onboard_MissingTenantName(t *testing.T) {
	svc, _, _ := newTestTenantService()
	ctx := context.Background()

	_, err := svc.Onboard(ctx, OnboardInput{
		TenantName: "",
		Domain:     "acme.local",
		AdminEmail: "admin@acme.local",
		AdminName:  "Admin",
		AdminPass:  "securepassword123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant name")
}

func TestTenantService_Onboard_MissingDomain(t *testing.T) {
	svc, _, _ := newTestTenantService()
	ctx := context.Background()

	_, err := svc.Onboard(ctx, OnboardInput{
		TenantName: "Acme",
		Domain:     "",
		AdminEmail: "admin@acme.local",
		AdminName:  "Admin",
		AdminPass:  "securepassword123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "domain")
}

func TestTenantService_Onboard_ShortPassword(t *testing.T) {
	svc, _, _ := newTestTenantService()
	ctx := context.Background()

	_, err := svc.Onboard(ctx, OnboardInput{
		TenantName: "Acme",
		Domain:     "acme.local",
		AdminEmail: "admin@acme.local",
		AdminName:  "Admin",
		AdminPass:  "short",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "password")
}

func TestTenantService_Onboard_DuplicateDomain(t *testing.T) {
	svc, _, _ := newTestTenantService()
	ctx := context.Background()

	_, err := svc.Onboard(ctx, OnboardInput{
		TenantName: "Acme Corp",
		Domain:     "acme.local",
		AdminEmail: "admin@acme.local",
		AdminName:  "Admin",
		AdminPass:  "securepassword123",
	})
	require.NoError(t, err)

	_, err = svc.Onboard(ctx, OnboardInput{
		TenantName: "Acme Corp 2",
		Domain:     "acme.local",
		AdminEmail: "admin2@acme.local",
		AdminName:  "Admin 2",
		AdminPass:  "anotherpassword123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflict")
}

func TestTenantService_Onboard_DefaultCountryIndustry(t *testing.T) {
	svc, _, _ := newTestTenantService()
	ctx := context.Background()

	result, err := svc.Onboard(ctx, OnboardInput{
		TenantName: "Acme Corp",
		Domain:     "acme2.local",
		AdminEmail: "admin@acme2.local",
		AdminName:  "Admin",
		AdminPass:  "securepassword123",
		// Country and Industry deliberately omitted
	})

	require.NoError(t, err)
	assert.Equal(t, "IN", result.Tenant.Country)
	assert.Equal(t, "GENERAL", result.Tenant.Industry)
}

func TestTenantService_GetByID(t *testing.T) {
	svc, _, _ := newTestTenantService()
	ctx := context.Background()

	result, _ := svc.Onboard(ctx, OnboardInput{
		TenantName: "Acme", Domain: "acme.local",
		AdminEmail: "a@acme.local", AdminName: "A", AdminPass: "password12345",
	})

	fetched, err := svc.GetByID(ctx, result.Tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, "Acme", fetched.Name)
}

func TestTenantService_GetByDomain(t *testing.T) {
	svc, _, _ := newTestTenantService()
	ctx := context.Background()

	_, _ = svc.Onboard(ctx, OnboardInput{
		TenantName: "Acme", Domain: "acme.local",
		AdminEmail: "a@acme.local", AdminName: "A", AdminPass: "password12345",
	})

	fetched, err := svc.GetByDomain(ctx, "acme.local")
	require.NoError(t, err)
	assert.Equal(t, "Acme", fetched.Name)
}

func TestTenantService_GetByDomain_NotFound(t *testing.T) {
	svc, _, _ := newTestTenantService()
	ctx := context.Background()

	_, err := svc.GetByDomain(ctx, "nonexistent.local")
	require.Error(t, err)
}

func TestTenantService_Update(t *testing.T) {
	svc, _, _ := newTestTenantService()
	ctx := context.Background()

	result, _ := svc.Onboard(ctx, OnboardInput{
		TenantName: "Acme", Domain: "acme.local",
		AdminEmail: "a@acme.local", AdminName: "A", AdminPass: "password12345",
	})

	result.Tenant.Name = "Acme Updated"
	err := svc.Update(ctx, result.Tenant)
	require.NoError(t, err)

	fetched, _ := svc.GetByID(ctx, result.Tenant.ID)
	assert.Equal(t, "Acme Updated", fetched.Name)
}
