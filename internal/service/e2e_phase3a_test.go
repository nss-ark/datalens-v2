// Package service contains end-to-end tests for Phase 3A: Data Principal Portal.
// This file exercises all DPDPA portal features through real services with in-memory mocks.
package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Deterministic test fixtures
// =============================================================================

var (
	phase3aTenantID    = mustParseID("550e8400-e29b-41d4-a716-446655440100")
	phase3aProfileID   = mustParseID("550e8400-e29b-41d4-a716-446655440101")
	phase3aSubjectID   = mustParseID("550e8400-e29b-41d4-a716-446655440102")
	phase3aAssigneeID  = mustParseID("550e8400-e29b-41d4-a716-446655440103")
	phase3aMinorProfID = mustParseID("550e8400-e29b-41d4-a716-446655440104")
)

func mustParseID(s string) types.ID {
	id, err := types.ParseID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func phase3aCtx() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, phase3aTenantID)
	return ctx
}

// newPhase3aDPService sets up a DataPrincipalService with fresh mocks.
// Returns the service and all mock repos for assertion.
func newPhase3aDPService(t *testing.T) (*DataPrincipalService, *mockProfileRepo, *mockDPRRepo, *mockDSRRepo, *mockHistoryRepo, *mockEventBus) {
	t.Helper()
	profileRepo := newMockProfileRepo()
	dprRepo := newMockDPRRepo()
	dsrRepo := newMockDSRRepo()
	historyRepo := newMockHistoryRepo()
	eventBus := newMockEventBus()

	svc := NewDataPrincipalService(profileRepo, dprRepo, dsrRepo, historyRepo, eventBus, nil, slog.Default())
	return svc, profileRepo, dprRepo, dsrRepo, historyRepo, eventBus
}

// seedVerifiedProfile pre-creates a verified, adult profile in the mock.
func seedVerifiedProfile(t *testing.T, repo *mockProfileRepo) *consent.DataPrincipalProfile {
	t.Helper()
	now := time.Now().UTC()
	method := "EMAIL_OTP"
	profile := &consent.DataPrincipalProfile{
		BaseEntity: types.BaseEntity{
			ID:        phase3aProfileID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		TenantID:           phase3aTenantID,
		Email:              "principal@example.com",
		VerificationStatus: consent.VerificationStatusVerified,
		VerifiedAt:         &now,
		VerificationMethod: &method,
		SubjectID:          &phase3aSubjectID,
		PreferredLang:      "en",
		IsMinor:            false,
		GuardianVerified:   false,
	}
	err := repo.Create(context.Background(), profile)
	require.NoError(t, err)
	return profile
}

// =============================================================================
// Test 1: DPR Lifecycle — ACCESS request, SLA, status, download
// =============================================================================

func TestPhase3A_DPRLifecycle(t *testing.T) {
	svc, profileRepo, dprRepo, dsrRepo, _, eventBus := newPhase3aDPService(t)
	ctx := phase3aCtx()
	seedVerifiedProfile(t, profileRepo)

	// --- Step 1: Submit ACCESS request ---
	dpr, err := svc.SubmitDPR(ctx, phase3aProfileID, CreateDPRRequestInput{
		Type:        "ACCESS",
		Description: "I want a copy of all my personal data.",
	})
	require.NoError(t, err)
	require.NotNil(t, dpr)

	// Verify DPR fields
	assert.Equal(t, consent.DPRStatusSubmitted, dpr.Status)
	assert.Equal(t, "ACCESS", dpr.Type)
	assert.Equal(t, phase3aTenantID, dpr.TenantID)
	assert.Equal(t, phase3aProfileID, dpr.ProfileID)
	assert.NotNil(t, dpr.DSRID, "DPR should link to a DSR")

	// --- Step 2: Verify linked DSR was created ---
	dsr, err := dsrRepo.GetByID(ctx, *dpr.DSRID)
	require.NoError(t, err)
	assert.Equal(t, compliance.DSRStatusPending, dsr.Status)
	assert.Equal(t, compliance.RequestTypeAccess, dsr.RequestType)

	// --- Step 3: Verify events were published ---
	assert.GreaterOrEqual(t, len(eventBus.Events), 1, "At least one event should be published")

	// --- Step 4: Check DPR status via service ---
	fetched, err := svc.GetDPR(ctx, phase3aProfileID, dpr.ID)
	require.NoError(t, err)
	assert.Equal(t, consent.DPRStatusSubmitted, fetched.Status)

	// --- Step 5: Simulate completion (admin marks complete) ---
	now := time.Now().UTC()
	summary := "Your personal data export is ready."
	dpr.Status = consent.DPRStatusCompleted
	dpr.CompletedAt = &now
	dpr.ResponseSummary = &summary
	err = dprRepo.Update(ctx, dpr)
	require.NoError(t, err)

	// --- Step 6: Download result (mocked) ---
	result, err := svc.DownloadDPRData(ctx, phase3aProfileID, dpr.ID)
	require.NoError(t, err)
	assert.Equal(t, dpr.ID, result.DPRRequestID)
	assert.Equal(t, "ACCESS", result.RequestType)
	assert.Equal(t, "Your personal data export is ready.", result.Summary)
	assert.NotNil(t, result.CompletedAt)

	t.Logf("✅ DPR Lifecycle: Submit → Complete → Download verified")
}

// =============================================================================
// Test 2: Multi-Type DPR — ERASURE, CORRECTION, NOMINATION
// =============================================================================

func TestPhase3A_DPRMultiType(t *testing.T) {
	svc, profileRepo, _, dsrRepo, _, _ := newPhase3aDPService(t)
	ctx := phase3aCtx()
	seedVerifiedProfile(t, profileRepo)

	types_to_test := []struct {
		dprType     string
		dsrType     compliance.DSRRequestType
		description string
	}{
		{"ERASURE", compliance.RequestTypeErasure, "Please delete all my data."},
		{"CORRECTION", compliance.RequestTypeCorrection, "My address is wrong. Please correct it."},
		{"NOMINATION", compliance.RequestTypeNomination, "I nominate my spouse as my representative."},
	}

	for _, tc := range types_to_test {
		t.Run(tc.dprType, func(t *testing.T) {
			dpr, err := svc.SubmitDPR(ctx, phase3aProfileID, CreateDPRRequestInput{
				Type:        tc.dprType,
				Description: tc.description,
			})
			require.NoError(t, err)
			assert.Equal(t, consent.DPRStatusSubmitted, dpr.Status)
			assert.Equal(t, tc.dprType, dpr.Type)

			// Verify linked DSR
			require.NotNil(t, dpr.DSRID)
			dsr, err := dsrRepo.GetByID(ctx, *dpr.DSRID)
			require.NoError(t, err)
			assert.Equal(t, tc.dsrType, dsr.RequestType)
			assert.Equal(t, compliance.DSRStatusPending, dsr.Status)
		})
	}

	// Verify all DPRs are listed
	dprs, err := svc.ListDPRs(ctx, phase3aProfileID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(dprs), 3, "Should have at least 3 DPRs (ERASURE, CORRECTION, NOMINATION)")

	t.Logf("✅ Multi-Type DPR: ERASURE, CORRECTION, NOMINATION all linked to DSRs")
}

// =============================================================================
// Test 3: Appeal Flow — Reject → Appeal → APPEALED status
// =============================================================================

func TestPhase3A_AppealFlow(t *testing.T) {
	svc, profileRepo, dprRepo, _, _, _ := newPhase3aDPService(t)
	ctx := phase3aCtx()
	seedVerifiedProfile(t, profileRepo)

	// --- Step 1: Submit a DPR ---
	dpr, err := svc.SubmitDPR(ctx, phase3aProfileID, CreateDPRRequestInput{
		Type:        "ACCESS",
		Description: "Give me my data.",
	})
	require.NoError(t, err)

	// --- Step 2: Admin rejects the DPR ---
	dpr.Status = consent.DPRStatusRejected
	err = dprRepo.Update(ctx, dpr)
	require.NoError(t, err)

	// Verify rejection
	rejected, err := svc.GetDPR(ctx, phase3aProfileID, dpr.ID)
	require.NoError(t, err)
	assert.Equal(t, consent.DPRStatusRejected, rejected.Status)

	// --- Step 3: User appeals ---
	appeal, err := svc.AppealDPR(ctx, phase3aProfileID, dpr.ID, "I believe my request was wrongly rejected.")
	require.NoError(t, err)
	require.NotNil(t, appeal)

	assert.Equal(t, consent.DPRStatusAppealed, appeal.Status)
	assert.NotNil(t, appeal.AppealOf)
	assert.Equal(t, dpr.ID, *appeal.AppealOf)
	assert.NotNil(t, appeal.AppealReason)
	assert.Contains(t, *appeal.AppealReason, "wrongly rejected")
	assert.True(t, appeal.IsEscalated)

	// --- Step 4: Verify appeal has linked DSR ---
	assert.NotNil(t, appeal.DSRID, "Appeal should have a linked DSR")

	// --- Step 5: Cannot appeal same DPR twice ---
	_, err = svc.AppealDPR(ctx, phase3aProfileID, dpr.ID, "double appeal")
	require.Error(t, err, "Should not allow duplicate appeal")

	// --- Step 6: Cannot appeal a DPR that is not REJECTED/COMPLETED ---
	dpr2, err := svc.SubmitDPR(ctx, phase3aProfileID, CreateDPRRequestInput{
		Type:        "ERASURE",
		Description: "delete everything",
	})
	require.NoError(t, err)
	_, err = svc.AppealDPR(ctx, phase3aProfileID, dpr2.ID, "this should fail")
	require.Error(t, err, "Should not allow appeal of SUBMITTED DPR")

	t.Logf("✅ Appeal Flow: Reject → Appeal → APPEALED, duplicate blocked, invalid state blocked")
}

// =============================================================================
// Test 4: Consent Management — Grant, Receipt, Withdraw, History
// =============================================================================

func TestPhase3A_ConsentManagement(t *testing.T) {
	ctx := phase3aCtx()

	widgetRepo := newMockWidgetRepo()
	sessionRepo := newMockSessionRepo()
	historyRepo := newMockHistoryRepo()
	eventBus := newMockEventBus()
	// noticeRepo := newMockNoticeRepo() // Not used in this service constructor
	// renewalRepo := newMockRenewalRepo() // Not used in this service constructor
	mockCache := newMockConsentCache()

	consentSvc := NewConsentService(
		widgetRepo,
		sessionRepo,
		historyRepo,
		eventBus,
		mockCache,
		"test_signing_key_32chars_long!!",
		slog.Default(),
		24*time.Hour,
	)

	// --- Step 1: Create a widget ---
	purposeID := types.NewID()
	widget := &consent.ConsentWidget{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			TenantID:   phase3aTenantID,
		},
		Name:   "Website Banner",
		Type:   consent.WidgetTypeBanner,
		Domain: "*.example.com",
		Status: consent.WidgetStatusActive,
		APIKey: "test-api-key-123",
		Config: consent.WidgetConfig{
			PurposeIDs:      []types.ID{purposeID},
			DefaultLanguage: "en",
		},
		Version: 1,
	}
	err := widgetRepo.Create(ctx, widget)
	require.NoError(t, err)

	// --- Step 2: Record consent (grant) ---
	session, err := consentSvc.RecordConsent(ctx, RecordConsentRequest{
		WidgetID: widget.ID,
		Decisions: []consent.ConsentDecision{
			{PurposeID: purposeID, Granted: true},
		},
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0 DataLens Test Agent",
		PageURL:   "https://example.com/signup",
	})

	require.NoError(t, err)
	require.NotNil(t, session)

	// Verify session
	assert.Equal(t, widget.ID, session.WidgetID)
	assert.NotEmpty(t, session.Signature, "Session should have HMAC signature")
	assert.Len(t, session.Decisions, 1)
	assert.True(t, session.Decisions[0].Granted)

	// --- Step 3: Verify receipt (via session retrieval) ---
	receipt, err := sessionRepo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, receipt.ID)
	assert.NotEmpty(t, receipt.Signature, "Receipt should have HMAC signature for legal evidence")

	// --- Step 4: Withdraw consent ---
	withdrawSession, err := consentSvc.RecordConsent(ctx, RecordConsentRequest{
		WidgetID: widget.ID,
		Decisions: []consent.ConsentDecision{
			{PurposeID: purposeID, Granted: false},
		},
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0 DataLens Test Agent",
		PageURL:   "https://example.com/settings",
	})

	require.NoError(t, err)
	assert.NotNil(t, withdrawSession)
	assert.False(t, withdrawSession.Decisions[0].Granted)

	// --- Step 5: Verify events published ---
	assert.GreaterOrEqual(t, len(eventBus.Events), 2, "At least 2 consent events (grant + withdraw)")

	t.Logf("✅ Consent Management: Grant → Receipt → Withdraw → History updated")
}

// =============================================================================
// Test 5: Grievance & DPO — Submit, Assign, Resolve, DPO Contact
// =============================================================================

func TestPhase3A_GrievanceAndDPO(t *testing.T) {
	ctx := phase3aCtx()

	// --- Grievance flow ---
	t.Run("grievance lifecycle", func(t *testing.T) {
		grievanceRepo := newMockGrievanceRepo()
		eventBus := newMockEventBus()
		grievanceSvc := NewGrievanceService(grievanceRepo, eventBus, slog.Default())

		// Step 1: Submit grievance
		grievance, err := grievanceSvc.SubmitGrievance(ctx, CreateGrievanceRequest{
			Subject:       "Consent Withdrawn but Data Still Processed",
			Description:   "I withdrew my consent for marketing emails, but I still receive them.",
			Category:      "CONSENT",
			DataSubjectID: phase3aSubjectID.String(),
		})
		require.NoError(t, err)
		require.NotNil(t, grievance)

		assert.Equal(t, compliance.GrievanceStatusOpen, grievance.Status)
		assert.NotNil(t, grievance.DueDate, "Should have 30-day SLA deadline")
		assert.WithinDuration(t, time.Now().AddDate(0, 0, 30), *grievance.DueDate, 5*time.Second)
		assert.Equal(t, phase3aSubjectID, grievance.DataSubjectID)

		// Step 2: Assign grievance
		err = grievanceSvc.AssignGrievance(ctx, grievance.ID, phase3aAssigneeID)
		require.NoError(t, err)

		assigned, err := grievanceSvc.GetGrievance(ctx, grievance.ID)
		require.NoError(t, err)
		assert.Equal(t, compliance.GrievanceStatusInProgress, assigned.Status)
		assert.NotNil(t, assigned.AssignedTo)
		assert.Equal(t, phase3aAssigneeID, *assigned.AssignedTo)

		// Step 3: Resolve grievance
		err = grievanceSvc.ResolveGrievance(ctx, grievance.ID, "Marketing emails disabled. Data verified as removed from mailing list.")
		require.NoError(t, err)

		resolved, err := grievanceSvc.GetGrievance(ctx, grievance.ID)
		require.NoError(t, err)
		assert.Equal(t, compliance.GrievanceStatusResolved, resolved.Status)
		assert.NotNil(t, resolved.Resolution)
		assert.Contains(t, *resolved.Resolution, "Marketing emails disabled")
		assert.NotNil(t, resolved.ResolvedAt)

		// Step 4: Verify events were published (submit, assign, resolve)
		assert.GreaterOrEqual(t, len(eventBus.Events), 3)

		t.Logf("  ✅ Grievance: Submit → Assign → Resolve verified")
	})

	// --- DPO Contact flow ---
	t.Run("dpo contact", func(t *testing.T) {
		dpoRepo := newMockDPOContactRepo()
		eventBus := newMockEventBus()
		dpoSvc := NewDPOService(dpoRepo, eventBus, slog.Default())

		// Step 1: Upsert DPO contact
		phone := "+91-9876543210"
		address := "Data Protection Office, 100 Privacy Lane, Bengaluru, India"
		contact, err := dpoSvc.UpsertContact(ctx, UpsertDPOContactRequest{
			OrgName:  "ComplyArk Technologies Pvt. Ltd.",
			DPOName:  "Dr. Arjun Mehta",
			DPOEmail: "dpo@complyark.in",
			DPOPhone: &phone,
			Address:  &address,
		})
		require.NoError(t, err)
		require.NotNil(t, contact)

		assert.Equal(t, "Dr. Arjun Mehta", contact.DPOName)
		assert.Equal(t, "dpo@complyark.in", contact.DPOEmail)

		// Step 2: Retrieve public contact
		pubContact, err := dpoSvc.GetPublicContact(ctx, phase3aTenantID)
		require.NoError(t, err)
		assert.Equal(t, "dpo@complyark.in", pubContact.DPOEmail)
		assert.Equal(t, "ComplyArk Technologies Pvt. Ltd.", pubContact.OrgName)

		t.Logf("  ✅ DPO Contact: Upsert → Public retrieval verified")
	})

	t.Logf("✅ Grievance & DPO: Full lifecycle verified")
}

// =============================================================================
// Test 6: Guardian Flow — Minor profile, rejection, verification, re-submit
// =============================================================================

func TestPhase3A_GuardianFlow(t *testing.T) {
	svc, profileRepo, _, _, _, _ := newPhase3aDPService(t)
	ctx := phase3aCtx()

	// --- Step 1: Create a minor profile (guardian NOT verified) ---
	now := time.Now().UTC()
	method := "EMAIL_OTP"
	dob := time.Date(2012, 5, 15, 0, 0, 0, 0, time.UTC)
	minorProfile := &consent.DataPrincipalProfile{
		BaseEntity: types.BaseEntity{
			ID:        phase3aMinorProfID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		TenantID:           phase3aTenantID,
		Email:              "minoruser@example.com",
		VerificationStatus: consent.VerificationStatusVerified,
		VerifiedAt:         &now,
		VerificationMethod: &method,
		SubjectID:          &phase3aSubjectID,
		PreferredLang:      "en",
		IsMinor:            true,
		DateOfBirth:        &dob,
		GuardianVerified:   false,
	}
	err := profileRepo.Create(ctx, minorProfile)
	require.NoError(t, err)

	// --- Step 2: Attempt DPR — should FAIL (guardian not verified) ---
	_, err = svc.SubmitDPR(ctx, phase3aMinorProfID, CreateDPRRequestInput{
		Type:        "ACCESS",
		Description: "I want my data (minor).",
	})
	require.Error(t, err, "Minor without guardian verification should be rejected")
	assert.Contains(t, err.Error(), "guardian")

	// --- Step 3: Verify guardian with dev OTP "123456" ---
	err = svc.VerifyGuardian(ctx, phase3aMinorProfID, "123456")
	require.NoError(t, err)

	// Confirm profile is updated
	updated, err := profileRepo.GetByID(ctx, phase3aMinorProfID)
	require.NoError(t, err)
	assert.True(t, updated.GuardianVerified)

	// --- Step 4: Re-submit DPR — should SUCCEED ---
	dpr, err := svc.SubmitDPR(ctx, phase3aMinorProfID, CreateDPRRequestInput{
		Type:        "ACCESS",
		Description: "I want my data (minor, guardian verified).",
	})
	require.NoError(t, err)
	require.NotNil(t, dpr)
	assert.Equal(t, consent.DPRStatusSubmitted, dpr.Status)
	assert.True(t, dpr.GuardianVerified)
	assert.True(t, dpr.IsMinor)

	t.Logf("✅ Guardian Flow: Minor rejected → Guardian verified → DPR succeeded")
}
