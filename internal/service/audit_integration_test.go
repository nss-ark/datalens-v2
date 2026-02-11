//go:build integration

package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

func TestAudit_Login(t *testing.T) {
	pool := setupPostgres(t)
	ctx := context.Background()

	// 1. Setup Dependencies
	auditRepo := repository.NewPostgresAuditRepository(pool)
	userRepo := repository.NewUserRepo(pool)
	roleRepo := repository.NewRoleRepo(pool)
	tenantRepo := repository.NewTenantRepo(pool)

	logger := slog.Default()
	auditService := NewAuditService(auditRepo, logger)

	authService := NewAuthService(
		userRepo,
		roleRepo,
		"secret-key",
		15*time.Minute,
		24*time.Hour,
		logger,
		auditService,
	)

	// 2. Seed Data
	// Create Tenant
	tenant := &identity.Tenant{
		BaseEntity: types.BaseEntity{
			ID: types.NewID(),
		},
		Name:   "Audit Test Tenant",
		Domain: "audit.test.com",
		Status: identity.TenantActive,
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	// Create User
	// We need to hash password first if we insert directly via repo, OR use Register.
	// We will use Register below.

	regInput := RegisterInput{
		TenantID: tenant.ID,
		Email:    "audit_test_user@test.com",
		Name:     "Audit Test User",
		Password: "password123", // Service hashes this
	}

	registeredUser, err := authService.Register(ctx, regInput)
	require.NoError(t, err)

	// 3. Perform Action: Login
	loginInput := LoginInput{
		TenantID: tenant.ID,
		Email:    regInput.Email,
		Password: regInput.Password,
	}

	_, err = authService.Login(ctx, loginInput)
	require.NoError(t, err)

	// Wait a bit for async audit log?
	// AuditService.Log emits a goroutine.
	time.Sleep(100 * time.Millisecond)

	// 4. Verify Audit Log
	// There should be a LOGIN log in audit_logs
	// We can query manually via pgxpool or use auditRepo if it has Get method.
	// auditRepo has GetByTenant.

	logs, err := auditRepo.GetByTenant(ctx, tenant.ID, 10)
	require.NoError(t, err)

	// Should have at least 1 log. Could have more if other tests ran on same tenant (they shouldn't).
	// But we just created the tenant, so it should be clean.
	// Actually, Register might not log audit? Let's check AuthService.Register ... it logs via logger, but not AuditService?
	// Checking AuthService.go ... Register does NOT call auditService.Log.
	// Login calls `auditService.Log(..., "LOGIN", ...)`

	var foundLogin bool
	for _, l := range logs {
		if l.Action == "LOGIN" && l.ActorID == registeredUser.ID {
			foundLogin = true
			break
		}
	}
	assert.True(t, foundLogin, "Expected audit log for LOGIN action")
}

type MockEventBus struct{}

func (m *MockEventBus) Publish(ctx context.Context, event eventbus.Event) error { return nil }
func (m *MockEventBus) Subscribe(ctx context.Context, pattern string, handler eventbus.EventHandler) (eventbus.Subscription, error) {
	return nil, nil
}
func (m *MockEventBus) Close() error { return nil }

func TestAudit_PolicyCreate(t *testing.T) {
	pool := setupPostgres(t)
	ctx := context.Background()

	// 1. Setup Dependencies
	auditRepo := repository.NewPostgresAuditRepository(pool)
	policyRepo := repository.NewPostgresPolicyRepository(pool)
	tenantRepo := repository.NewTenantRepo(pool)

	// Creates necessary but unused repos for PolicyService
	violationRepo := repository.NewPostgresViolationRepository(pool)
	mappingRepo := repository.NewPostgresDataMappingRepository(pool)
	dsRepo := repository.NewDataSourceRepo(pool)
	piiRepo := repository.NewPIIClassificationRepo(pool)

	logger := slog.Default()
	auditService := NewAuditService(auditRepo, logger)
	eventBus := &MockEventBus{}

	policyService := NewPolicyService(
		policyRepo,
		violationRepo,
		mappingRepo,
		dsRepo,
		piiRepo,
		eventBus,
		auditService,
		logger,
	)

	// 2. Seed Data
	tenant := &identity.Tenant{
		BaseEntity: types.BaseEntity{
			ID: types.NewID(),
		},
		Name:   "Policy Audit Tenant",
		Domain: "policy.audit.test.com",
		Status: identity.TenantActive,
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	// Setup Context for Tenant
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenant.ID)

	// 3. Perform Action: Create Policy
	policy := &governance.Policy{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{
				ID: types.NewID(),
			},
			TenantID: tenant.ID,
		},
		Name:        "Test Policy",
		Description: "Test Description",
		Type:        governance.PolicyTypeRetention,
		Rules:       []governance.PolicyRule{},
		Actions:     []governance.PolicyAction{},
		Severity:    types.SeverityCritical,
	}

	err := policyService.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	// Wait for async
	time.Sleep(100 * time.Millisecond)

	// 4. Verify Audit Log
	logs, err := auditRepo.GetByTenant(ctx, tenant.ID, 10)
	require.NoError(t, err)

	var foundPolicy bool
	for _, l := range logs {
		if l.Action == "POLICY_CREATE" && l.ResourceID == policy.ID {
			foundPolicy = true
			break
		}
	}
	assert.True(t, foundPolicy, "Expected audit log for POLICY_CREATE action")
}
