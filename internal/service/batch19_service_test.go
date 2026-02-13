package service

import (
	"context"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/breach"
	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBatch19_GuardianFlow validates the minor/guardian verification logic from Batch 18.
func TestBatch19_GuardianFlow(t *testing.T) {
	// Setup Dependencies
	profileRepo := newMockProfileRepo()
	dprRepo := newMockDPRRepo()
	dsrRepo := newMockDSRRepo() // DataPrincipalService creates internal DSR too
	historyRepo := newMockHistoryRepo()
	eventBus := newMockEventBus()
	logger := newTestLogger()

	// Using nil Redis means service will fallback to DEV mode (log only) and "123456" OTP
	svc := NewDataPrincipalService(profileRepo, dprRepo, dsrRepo, historyRepo, eventBus, nil, logger)

	// Context with Tenant
	tenantID := types.NewID()
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, tenantID)

	// Create a Minor Profile
	minorID := types.NewID()
	minorProfile := &consent.DataPrincipalProfile{
		BaseEntity:       types.BaseEntity{ID: minorID},
		TenantID:         tenantID,
		Email:            "minor@example.com",
		IsMinor:          true,
		GuardianVerified: false,
	}
	require.NoError(t, profileRepo.Create(ctx, minorProfile))

	t.Run("Submitting DPR without guardian verification should FAIL", func(t *testing.T) {
		_, err := svc.SubmitDPR(ctx, minorID, CreateDPRRequestInput{
			Type:        "ACCESS",
			Description: "I want my data",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "guardian verification required")
	})

	t.Run("Initiate Guardian Verification (OTP Generation)", func(t *testing.T) {
		err := svc.InitiateGuardianVerification(ctx, minorID, "parent@example.com")
		require.NoError(t, err)
		// Logs would show OTP "123456" in dev mode
	})

	t.Run("Verify Guardian with Wrong OTP", func(t *testing.T) {
		err := svc.VerifyGuardian(ctx, minorID, "999999")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid OTP")
	})

	t.Run("Verify Guardian with Correct OTP", func(t *testing.T) {
		err := svc.VerifyGuardian(ctx, minorID, "123456")
		require.NoError(t, err)

		// Check profile updated
		p, err := profileRepo.GetByID(ctx, minorID)
		require.NoError(t, err)
		assert.True(t, p.GuardianVerified)
	})

	t.Run("Submitting DPR AFTER verification should SUCCEED", func(t *testing.T) {
		dpr, err := svc.SubmitDPR(ctx, minorID, CreateDPRRequestInput{
			Type:        "ERASURE",
			Description: "Delete my data",
		})
		require.NoError(t, err)
		assert.NotNil(t, dpr)
		assert.Equal(t, consent.DPRStatusSubmitted, dpr.Status)
		assert.True(t, dpr.GuardianVerified)

		// Check internal DSR created
		require.NotNil(t, dpr.DSRID)
		dsr, err := dsrRepo.GetByID(ctx, *dpr.DSRID)
		require.NoError(t, err)
		assert.Equal(t, compliance.DSRStatusPending, dsr.Status)
		assert.Equal(t, "minor@example.com", dsr.SubjectEmail)
	})
}

// TestBatch19_AdminDSR validates the cross-tenant DSR listing from Batch 18.1.
// Service Only Test
func TestBatch19_AdminDSR_Service(t *testing.T) {
	// Setup Dependencies
	dsrRepo := newMockDSRRepo()
	tenantRepo := newMockTenantRepo() // AdminService needs these
	userRepo := newMockUserRepo()
	roleRepo := newMockRoleRepo()
	logger := newTestLogger()

	// Needed for AdminService
	tenantSvc := NewTenantService(tenantRepo, userRepo, roleRepo, nil, logger)
	adminSvc := NewAdminService(tenantRepo, userRepo, roleRepo, dsrRepo, tenantSvc, logger)

	// Create DSRs in Different Tenants
	tenantA := types.NewID()
	tenantB := types.NewID()

	dsr1 := &compliance.DSR{ID: types.NewID(), TenantID: tenantA, Status: compliance.DSRStatusPending, RequestType: compliance.RequestTypeAccess}
	dsr2 := &compliance.DSR{ID: types.NewID(), TenantID: tenantB, Status: compliance.DSRStatusApproved, RequestType: compliance.RequestTypeErasure}

	require.NoError(t, dsrRepo.Create(context.Background(), dsr1))
	require.NoError(t, dsrRepo.Create(context.Background(), dsr2))

	t.Run("Service: GetAllDSRs returns all items", func(t *testing.T) {
		result, err := adminSvc.GetAllDSRs(context.Background(), types.Pagination{PageSize: 10, Page: 1}, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Total)
		assert.Len(t, result.Items, 2)
	})
}

// TestBatch19_BreachNotification validates that High Severity breaches trigger notifications.
func TestBatch19_BreachNotification(t *testing.T) {
	// Setup Dependencies
	breachRepo := newMockBreachRepo()
	auditRepo := newMockAuditRepo()
	auditSvc := NewAuditService(auditRepo, newTestLogger())
	eventBus := newMockEventBus()
	logger := newTestLogger()

	svc := NewBreachService(breachRepo, newMockProfileRepo(), nil, auditSvc, eventBus, logger)

	// Context
	tenantID := types.NewID()
	userID := types.NewID()
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, tenantID)
	ctx = context.WithValue(ctx, types.ContextKeyUserID, userID)

	t.Run("Create HIGH severity breach publishes event", func(t *testing.T) {
		// Clear events
		eventBus.Events = nil

		req := CreateIncidentRequest{
			Title:      "Data leak",
			DetectedAt: time.Now(),
			Severity:   breach.SeverityHigh,
		}

		incident, err := svc.CreateIncident(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, incident)
		assert.True(t, incident.IsReportableToCertIn)

		// Check Event Bus
		require.Len(t, eventBus.Events, 1)
		evt := eventBus.Events[0]
		assert.Equal(t, "breach.incident_created", evt.Type)
		assert.Equal(t, incident.ID, evt.Data.(*breach.BreachIncident).ID)
	})

	t.Run("Create LOW severity breach publishes event too (but workflow might handle differently)", func(t *testing.T) {
		// Clear events
		eventBus.Events = nil

		req := CreateIncidentRequest{
			Title:      "Minor glitch",
			DetectedAt: time.Now(),
			Severity:   breach.SeverityLow,
		}

		incident, err := svc.CreateIncident(ctx, req)
		require.NoError(t, err)
		assert.False(t, incident.IsReportableToCertIn)

		// It still triggers "incident_created"
		require.Len(t, eventBus.Events, 1)
	})
}
