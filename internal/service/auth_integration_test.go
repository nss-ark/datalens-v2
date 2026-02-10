package service_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Auth Integration Tests — Real Postgres via testcontainers
// =============================================================================

var integrationPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("datalens_auth_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(120*time.Second),
		),
	)
	if err != nil {
		panic("failed to start postgres: " + err.Error())
	}
	defer container.Terminate(ctx)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}

	integrationPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}
	defer integrationPool.Close()

	if err := applyMigrations(ctx, integrationPool); err != nil {
		panic("failed to apply migrations: " + err.Error())
	}

	os.Exit(m.Run())
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	files := []string{"001_initial_schema.sql", "002_api_keys.sql"}
	for _, f := range files {
		sql, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return err
		}
	}
	return nil
}

// =============================================================================
// Helpers
// =============================================================================

func newIntegrationAuthService(t *testing.T) (*service.AuthService, *service.TenantService) {
	t.Helper()
	userRepo := repository.NewUserRepo(integrationPool)
	tenantRepo := repository.NewTenantRepo(integrationPool)
	roleRepo := repository.NewRoleRepo(integrationPool)

	authSvc := service.NewAuthService(
		userRepo,
		roleRepo,
		"integration-test-secret-key-32chars!",
		15*time.Minute,
		24*time.Hour,
		slog.Default(),
	)
	tenantSvc := service.NewTenantService(tenantRepo, userRepo, roleRepo, authSvc, slog.Default())
	return authSvc, tenantSvc
}

func uniqueDomain(prefix string) string {
	return prefix + "-" + types.NewID().String()[:8] + ".com"
}

// =============================================================================
// End-to-End Auth Flow
// =============================================================================

func TestAuthIntegration_RegisterLoginRefresh(t *testing.T) {
	authSvc, tenantSvc := newIntegrationAuthService(t)
	ctx := context.Background()

	// 1. Onboard — creates tenant + admin user + tokens
	domain := uniqueDomain("authflow")
	result, err := tenantSvc.Onboard(ctx, service.OnboardInput{
		TenantName: "AuthTestCo",
		Domain:     domain,
		Industry:   "technology",
		Country:    "IN",
		AdminEmail: "admin@" + domain,
		AdminName:  "Admin User",
		AdminPass:  "SecureP@ss123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.Tenant.ID)
	assert.NotEmpty(t, result.Tokens.AccessToken)
	assert.NotEmpty(t, result.Tokens.RefreshToken)

	// 2. Login with the same credentials
	loginTokens, err := authSvc.Login(ctx, service.LoginInput{
		TenantID: result.Tenant.ID,
		Email:    "admin@" + domain,
		Password: "SecureP@ss123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, loginTokens.AccessToken)

	// 3. Validate the access token
	claims, err := authSvc.ValidateToken(loginTokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, result.Tenant.ID, claims.TenantID)

	// 4. Refresh the token
	newTokens, err := authSvc.RefreshToken(ctx, loginTokens.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEqual(t, loginTokens.AccessToken, newTokens.AccessToken)

	// 5. New token is still valid
	newClaims, err := authSvc.ValidateToken(newTokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, claims.UserID, newClaims.UserID)
}

func TestAuthIntegration_LoginWrongPassword(t *testing.T) {
	authSvc, tenantSvc := newIntegrationAuthService(t)
	ctx := context.Background()

	domain := uniqueDomain("wrongpw")
	result, err := tenantSvc.Onboard(ctx, service.OnboardInput{
		TenantName: "WrongPwCo",
		Domain:     domain,
		AdminEmail: "user@" + domain,
		AdminName:  "Test User",
		AdminPass:  "CorrectPass123",
	})
	require.NoError(t, err)

	_, err = authSvc.Login(ctx, service.LoginInput{
		TenantID: result.Tenant.ID,
		Email:    "user@" + domain,
		Password: "WrongPassword",
	})
	assert.Error(t, err, "login with wrong password should fail")
}

func TestAuthIntegration_LoginNonexistentUser(t *testing.T) {
	authSvc, tenantSvc := newIntegrationAuthService(t)
	ctx := context.Background()

	domain := uniqueDomain("nouser")
	result, err := tenantSvc.Onboard(ctx, service.OnboardInput{
		TenantName: "NoUserCo",
		Domain:     domain,
		AdminEmail: "admin@" + domain,
		AdminName:  "Admin",
		AdminPass:  "SecurePass123",
	})
	require.NoError(t, err)

	_, err = authSvc.Login(ctx, service.LoginInput{
		TenantID: result.Tenant.ID,
		Email:    "nonexistent@" + domain,
		Password: "AnyPassword",
	})
	assert.Error(t, err, "login with non-existent user should fail")
}

func TestAuthIntegration_InvalidToken(t *testing.T) {
	authSvc, _ := newIntegrationAuthService(t)

	_, err := authSvc.ValidateToken("completely.invalid.token")
	assert.Error(t, err, "invalid token should fail validation")
}

func TestAuthIntegration_TenantIsolation(t *testing.T) {
	authSvc, tenantSvc := newIntegrationAuthService(t)
	ctx := context.Background()

	// Create two tenants
	d1 := uniqueDomain("iso1")
	r1, err := tenantSvc.Onboard(ctx, service.OnboardInput{
		TenantName: "IsoTenant1",
		Domain:     d1,
		AdminEmail: "admin@" + d1,
		AdminName:  "Admin1",
		AdminPass:  "SecurePass123",
	})
	require.NoError(t, err)

	d2 := uniqueDomain("iso2")
	_, err = tenantSvc.Onboard(ctx, service.OnboardInput{
		TenantName: "IsoTenant2",
		Domain:     d2,
		AdminEmail: "admin@" + d2,
		AdminName:  "Admin2",
		AdminPass:  "SecurePass123",
	})
	require.NoError(t, err)

	// Tenant2's user should NOT be found via Tenant1's tenant_id
	_, err = authSvc.Login(ctx, service.LoginInput{
		TenantID: r1.Tenant.ID,
		Email:    "admin@" + d2,
		Password: "SecurePass123",
	})
	assert.Error(t, err, "cross-tenant login should fail")
}

func TestAuthIntegration_APIKeyLifecycle(t *testing.T) {
	_, tenantSvc := newIntegrationAuthService(t)
	ctx := context.Background()

	domain := uniqueDomain("apikey")
	result, err := tenantSvc.Onboard(ctx, service.OnboardInput{
		TenantName: "APIKeyCo",
		Domain:     domain,
		AdminEmail: "admin@" + domain,
		AdminName:  "Admin",
		AdminPass:  "SecurePass123",
	})
	require.NoError(t, err)

	// Create API key
	apiKeySvc := service.NewAPIKeyService(integrationPool, slog.Default())
	keyResult, err := apiKeySvc.CreateKey(ctx, result.Tenant.ID, "test-agent", []identity.Permission{
		{Resource: "DATA_SOURCE", Actions: []string{"READ"}},
	})
	require.NoError(t, err)
	assert.True(t, len(keyResult.RawKey) > 20, "raw key should be long enough")
	assert.Contains(t, keyResult.RawKey, "dlk_", "key should have dlk_ prefix")

	// Validate the key
	gotTenantID, perms, err := apiKeySvc.ValidateKey(ctx, keyResult.RawKey)
	require.NoError(t, err)
	assert.Equal(t, result.Tenant.ID, gotTenantID)
	assert.Len(t, perms, 1)

	// Revoke the key
	err = apiKeySvc.RevokeKey(ctx, keyResult.ID)
	require.NoError(t, err)

	// Revoked key should fail validation
	_, _, err = apiKeySvc.ValidateKey(ctx, keyResult.RawKey)
	assert.Error(t, err, "revoked key should fail validation")
}
