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

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/pkg/types"
)

// TestE2E_PortalFlow simulates the Data Principal Portal experience:
// 1. Principal requests OTP (InitiateLogin)
// 2. Principal submits OTP (VerifyLogin) -> gets Token + Profile
// 3. Principal views Profile
// 4. Principal submits a DSR (Subject Access Request)
// 5. Verify DSR is created in the backend
func TestE2E_PortalFlow(t *testing.T) {
	pool := setupPostgres(t)
	// Do not close pool here as it is shared

	ctx := context.Background()

	// =========================================================================
	// Setup: Tenant & Repos
	// =========================================================================
	tenantRepo := repository.NewTenantRepo(pool)
	profileRepo := repository.NewDataPrincipalProfileRepo(pool)
	dsrRepo := repository.NewDSRRepo(pool)
	dsRepo := repository.NewDataSourceRepo(pool)

	// Create Tenant
	tenant := &identity.Tenant{
		Name:     "PortalTestCo",
		Domain:   "portal-" + types.NewID().String()[:8] + ".com",
		Plan:     identity.PlanEnterprise,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{DefaultRegulation: "DPDPA"},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	// Create a Data Source (required for DSR task decomposition later)
	ds := &discovery.DataSource{
		TenantEntity: types.TenantEntity{
			TenantID: tenant.ID,
		},
		Name:   "Production DB",
		Type:   types.DataSourcePostgreSQL,
		Status: discovery.ConnectionStatusConnected,
	}
	ds.ID = types.NewID() // Ensure ID is set
	require.NoError(t, dsRepo.Create(ctx, ds))

	// Setup Services
	// Note: We pass nil for Redis to force PortalAuthService into dev mode (accepts "123456")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	authSvc := NewPortalAuthService(profileRepo, nil, "super-secret-key", 1*time.Hour, logger)

	eb := newMockEventBus() // Use mock event bus to capture events
	dsrQueue := newMockDSRQueue()
	dsrSvc := NewDSRService(dsrRepo, dsRepo, dsrQueue, eb, logger)

	// =========================================================================
	// Step 1: Initiate Login (Request OTP)
	// =========================================================================
	email := "alice@example.com"
	err := authSvc.InitiateLogin(ctx, tenant.ID, email, "")
	require.NoError(t, err)

	// =========================================================================
	// Step 2: Verify Login (Submit OTP)
	// =========================================================================
	// In dev mode (nil redis), VerifyLogin accepts any code if it warns, OR specific logic.
	// Looking at code: "if redis == nil ... if code != '123456' ... return error"
	// So we must use "123456"
	tokenResp, profile, err := authSvc.VerifyLogin(ctx, tenant.ID, email, "", "123456")
	require.NoError(t, err)
	require.NotNil(t, tokenResp)
	require.NotNil(t, profile)

	assert.Equal(t, email, profile.Email)
	assert.Equal(t, consent.VerificationStatusVerified, profile.VerificationStatus)
	assert.NotEmpty(t, tokenResp.AccessToken)

	// =========================================================================
	// Step 3: View Profile (Verify Persistence)
	// =========================================================================
	fetchedProfile, err := profileRepo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.Equal(t, profile.ID, fetchedProfile.ID)
	assert.Equal(t, email, fetchedProfile.Email)

	// =========================================================================
	// Step 4: Submit DSR
	// =========================================================================
	// Context must have TenantID for DSR creation
	ctxWithTenant := context.WithValue(ctx, types.ContextKeyTenantID, tenant.ID)

	req := CreateDSRRequest{
		RequestType:        compliance.RequestTypeAccess,
		SubjectName:        "Alice Wonderland",
		SubjectEmail:       email,
		SubjectIdentifiers: map[string]string{"email": email},
		Priority:           "MEDIUM",
		Notes:              "I want to see my data",
	}

	dsr, err := dsrSvc.CreateDSR(ctxWithTenant, req)
	require.NoError(t, err)
	assert.NotNil(t, dsr)
	assert.Equal(t, compliance.DSRStatusPending, dsr.Status)
	assert.Equal(t, email, dsr.SubjectEmail)

	// =========================================================================
	// Step 5: Verify DSR Persistence
	// =========================================================================
	fetchedDSR, err := dsrRepo.GetByID(ctx, dsr.ID)
	require.NoError(t, err)
	assert.Equal(t, dsr.ID, fetchedDSR.ID)
	assert.Equal(t, "Alice Wonderland", fetchedDSR.SubjectName)

	t.Log("✅ Portal E2E Flow verified: OTP -> Login -> Profile -> DSR Creation")
}
