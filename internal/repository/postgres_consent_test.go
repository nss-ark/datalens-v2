package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/repository"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// ConsentWidgetRepo Tests
// =============================================================================

func TestConsentWidgetRepo_GetByAPIKey(t *testing.T) {
	widgetRepo := repository.NewConsentWidgetRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	// Setup Tenant
	tenant := &identity.Tenant{
		Name:     "WidgetTestCo",
		Domain:   "widgettest-" + types.NewID().String()[:8] + ".com",
		Plan:     identity.PlanFree,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{DefaultRegulation: "DPDPA"},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	// Create Widget
	apiKey := "test-api-key-" + types.NewID().String()
	widget := &consent.ConsentWidget{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{
				ID:        types.NewID(),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			TenantID: tenant.ID,
		},
		Name:    "Test Widget",
		Type:    consent.WidgetType(consent.WidgetTypeBanner),
		Domain:  "example.com",
		Status:  consent.WidgetStatusActive,
		APIKey:  apiKey,
		Version: 1,
		Config:  consent.WidgetConfig{RegulationRef: "DPDPA"},
	}
	require.NoError(t, widgetRepo.Create(ctx, widget))

	// Test GetByAPIKey
	t.Run("found", func(t *testing.T) {
		got, err := widgetRepo.GetByAPIKey(ctx, apiKey)
		require.NoError(t, err)
		assert.Equal(t, widget.ID, got.ID)
		assert.Equal(t, "DPDPA", got.Config.RegulationRef)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := widgetRepo.GetByAPIKey(ctx, "invalid-key")
		require.Error(t, err)
		// Should be types.ErrNotFound, checking if error string contains "not found"
		// or if types.NewNotFoundError returns something checkable.
		// Usually we check assert.True(t, errors.Is(err, types.ErrNotFound)) if exported
		// or verify error message.
	})
}

// =============================================================================
// ConsentSessionRepo Tests
// =============================================================================

func TestConsentSessionRepo_Create(t *testing.T) {
	sessionRepo := repository.NewConsentSessionRepo(testPool)
	widgetRepo := repository.NewConsentWidgetRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	// Setup Tenant & Widget
	tenant := &identity.Tenant{
		Name:     "SessionTestCo",
		Domain:   "sessiontest-" + types.NewID().String()[:8] + ".com",
		Plan:     identity.PlanFree,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{DefaultRegulation: "DPDPA"},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	widget := &consent.ConsentWidget{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{
				ID:        types.NewID(),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			TenantID: tenant.ID,
		},
		Name:    "Session Widget",
		Status:  consent.WidgetStatusActive,
		APIKey:  "session-key-" + types.NewID().String(),
		Version: 1,
	}
	require.NoError(t, widgetRepo.Create(ctx, widget))

	// Create Session
	subjectID := types.NewID()
	session := &consent.ConsentSession{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: time.Now().UTC(),
		},
		TenantID:  tenant.ID,
		WidgetID:  widget.ID,
		SubjectID: &subjectID,
		Decisions: []consent.ConsentDecision{
			{PurposeID: types.NewID(), Granted: true},
			{PurposeID: types.NewID(), Granted: false},
		},
		IPAddress:     "192.168.1.1",
		UserAgent:     "GoTest",
		PageURL:       "http://localhost",
		WidgetVersion: 1,
		NoticeVersion: "v1",
		Signature:     "sha256:signature",
	}

	err := sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	// Verify Retrieval
	sessions, err := sessionRepo.GetBySubject(ctx, tenant.ID, subjectID)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, session.ID, sessions[0].ID)
	assert.Len(t, sessions[0].Decisions, 2)
	assert.Equal(t, "sha256:signature", sessions[0].Signature)
}

// =============================================================================
// ConsentHistoryRepo Tests
// =============================================================================

func TestConsentHistoryRepo_LatestState(t *testing.T) {
	historyRepo := repository.NewConsentHistoryRepo(testPool)
	tenantRepo := repository.NewTenantRepo(testPool)
	ctx := context.Background()

	// Setup Tenant
	tenant := &identity.Tenant{
		Name:     "HistoryTestCo",
		Domain:   "historytest-" + types.NewID().String()[:8] + ".com",
		Plan:     identity.PlanFree,
		Status:   identity.TenantActive,
		Settings: identity.TenantSettings{DefaultRegulation: "DPDPA"},
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	subjectID := types.NewID()
	purposeID := types.NewID()

	// 1. Create older entry (GRANTED)
	t1 := time.Now().UTC().Add(-2 * time.Hour)
	e1 := &consent.ConsentHistoryEntry{
		BaseEntity:    types.BaseEntity{ID: types.NewID(), CreatedAt: t1},
		TenantID:      tenant.ID,
		SubjectID:     subjectID,
		PurposeID:     purposeID,
		PurposeName:   "Marketing",
		NewStatus:     "GRANTED",
		Source:        "BANNER",
		NoticeVersion: "v1",
		Signature:     "sig1",
	}
	require.NoError(t, historyRepo.Create(ctx, e1))

	// 2. Create newer entry (WITHDRAWN)
	t2 := time.Now().UTC().Add(-1 * time.Hour)
	e2 := &consent.ConsentHistoryEntry{
		BaseEntity:     types.BaseEntity{ID: types.NewID(), CreatedAt: t2},
		TenantID:       tenant.ID,
		SubjectID:      subjectID,
		PurposeID:      purposeID,
		PurposeName:    "Marketing",
		PreviousStatus: types.Ptr("GRANTED"),
		NewStatus:      "WITHDRAWN",
		Source:         "PORTAL",
		NoticeVersion:  "v1",
		Signature:      "sig2",
	}
	require.NoError(t, historyRepo.Create(ctx, e2))

	// GetLatestState should return e2
	latest, err := historyRepo.GetLatestState(ctx, tenant.ID, subjectID, purposeID)
	require.NoError(t, err)
	assert.NotNil(t, latest)
	assert.Equal(t, "WITHDRAWN", latest.NewStatus)
	assert.Equal(t, "PORTAL", latest.Source)
}
