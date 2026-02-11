//go:build integration

package service

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/pkg/types"
)

// TestE2E_GovernanceFlow simulates the Governance workflow:
// 1. Setup: Tenant, DataSource, Inventory, Entity, Field (simulating scan results)
// 2. Context Engine: Suggest Purposes (using Pattern strategy)
// 3. User Action: Apply Purposes (Create DataMapping)
// 4. Policy Engine: Create a Policy (e.g. "Alert on High Sensitivity")
// 5. Policy Engine: Evaluate Policies
// 6. Verification: Check for Violations
func TestE2E_GovernanceFlow(t *testing.T) {
	pool := setupPostgres(t) // Reuse shared testcontainers setup
	// Do not close pool here as it is shared

	ctx := context.Background()

	// =========================================================================
	// Setup Repositories
	// =========================================================================
	tenantRepo := repository.NewTenantRepo(pool)
	dsRepo := repository.NewDataSourceRepo(pool)
	invRepo := repository.NewDataInventoryRepo(pool)
	entityRepo := repository.NewDataEntityRepo(pool)
	fieldRepo := repository.NewDataFieldRepo(pool)
	piiRepo := repository.NewPIIClassificationRepo(pool)
	purposeRepo := repository.NewPurposeRepo(pool)
	policyRepo := repository.NewPostgresPolicyRepository(pool)
	violationRepo := repository.NewPostgresViolationRepository(pool)
	mappingRepo := repository.NewPostgresDataMappingRepository(pool)

	// Setup Services
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	eb := newMockEventBus()

	// We need ContextEngine for suggestions. For E2E, we'll use a mocked AI Gateway or just rely on Pattern Strategy.
	// Let's use Pattern Strategy via TemplateLoader (mocked or real).
	// Actually, ContextEngine expects a template loader.
	// Let's construct it manually or mock the ContextEngine logic if it's complex to set up templates here.
	// Simplified: We will manually create the Purpose to simulate "Applying" a suggestion.

	auditRepo := repository.NewPostgresAuditRepository(pool)
	auditSvc := NewAuditService(auditRepo, logger)

	policySvc := NewPolicyService(policyRepo, violationRepo, mappingRepo, dsRepo, piiRepo, eb, auditSvc, logger)

	// Create Tenant
	tenant := &identity.Tenant{
		Name:     "GovTestCo",
		Domain:   "gov-" + types.NewID().String()[:8] + ".com",
		Plan:     identity.PlanEnterprise,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{DefaultRegulation: "DPDPA"},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenant.ID)

	// Create Data Source
	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{TenantID: tenant.ID},
		Name:         "HR Database",
		Type:         types.DataSourcePostgreSQL,
		Status:       discovery.ConnectionStatusConnected,
	}
	ds.ID = types.NewID()
	require.NoError(t, dsRepo.Create(ctx, ds))

	// Create Inventory -> Entity -> Field
	inv := &discovery.DataInventory{DataSourceID: ds.ID, SchemaVersion: "v1"}
	require.NoError(t, invRepo.Create(ctx, inv))

	entity := &discovery.DataEntity{InventoryID: inv.ID, Name: "employees", Type: discovery.EntityTypeTable}
	require.NoError(t, entityRepo.Create(ctx, entity))

	field := &discovery.DataField{EntityID: entity.ID, Name: "salary", DataType: "numeric"}
	require.NoError(t, fieldRepo.Create(ctx, field))

	// Simulate PII Detection (Verified)
	pii := &discovery.PIIClassification{
		BaseEntity:   types.BaseEntity{ID: types.NewID()},
		FieldID:      field.ID,
		DataSourceID: ds.ID,
		EntityName:   "employees",
		FieldName:    "salary",
		Category:     types.PIICategoryFinancial,
		Type:         types.PIITypeCreditCard, // Mismatched but serves for test
		Sensitivity:  types.SensitivityCritical,
		Confidence:   0.99,
		Status:       types.VerificationVerified,
	}
	require.NoError(t, piiRepo.Create(ctx, pii))

	// =========================================================================
	// Step 1: Create Purpose (Simulate Applying Suggestion)
	// =========================================================================
	purpose := &governance.Purpose{
		TenantEntity:  types.TenantEntity{TenantID: tenant.ID},
		Code:          "PAYROLL",
		Name:          "Payroll Processing",
		LegalBasis:    types.LegalBasisContract,
		RetentionDays: 3650,
		IsActive:      true,
	}
	require.NoError(t, purposeRepo.Create(ctx, purpose))

	// Create User for MappedBy
	userRepo := repository.NewUserRepo(pool)
	user := &identity.User{
		TenantEntity: types.TenantEntity{TenantID: tenant.ID},
		Email:        "admin@govtest.com",
		Name:         "Gov Admin",
		Password:     "securepass",
		Status:       identity.UserActive,
		MFAEnabled:   false,
	}
	require.NoError(t, userRepo.Create(ctx, user))

	// Link Purpose to PII (DataMapping)
	mapping := &governance.DataMapping{
		TenantEntity:     types.TenantEntity{TenantID: tenant.ID},
		ClassificationID: pii.ID,
		PurposeIDs:       []types.ID{purpose.ID},
		ThirdPartyIDs:    []types.ID{}, // Initialize to avoid NULL
		MappedBy:         user.ID,      // Use valid User ID
		MappedAt:         time.Now(),
	}
	require.NoError(t, mappingRepo.Create(ctx, mapping))

	// =========================================================================
	// Step 2: Create Policy
	// =========================================================================
	// Rule: If Sensitivity == CRITICAL, trigger violation
	policy := &governance.Policy{
		TenantEntity: types.TenantEntity{TenantID: tenant.ID},
		Name:         "High Sensitivity Alert",
		Type:         governance.PolicyTypeAlert,
		Severity:     types.SeverityCritical,
		IsActive:     true,
		Actions:      []governance.PolicyAction{}, // Initialize to avoid NULL
		Rules: []governance.PolicyRule{
			{
				Field:    "sensitivity",
				Operator: "EQ",
				Value:    "CRITICAL",
			},
		},
	}
	require.NoError(t, policySvc.CreatePolicy(ctx, policy))

	// =========================================================================
	// Step 3: Evaluate Policies
	// =========================================================================
	err := policySvc.EvaluatePolicies(ctx, tenant.ID)
	require.NoError(t, err)

	// =========================================================================
	// Step 4: Verify Violations
	// =========================================================================
	violations, err := policySvc.GetViolations(ctx, nil)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(violations), 1, "Should have at least 1 violation")

	found := false
	for _, v := range violations {
		if v.PolicyID == policy.ID && v.FieldName == "salary" {
			found = true
			assert.Equal(t, governance.ViolationStatusOpen, v.Status)
			assert.Equal(t, types.SeverityCritical, v.Severity)
			break
		}
	}
	assert.True(t, found, "Expected violation for salary field not found")

	t.Log("✅ Governance E2E Flow verified: Mapping -> Policy -> Evaluation -> Violation")
}
