package service

import (
	"context"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpiryChecker_DetectsExpiringConsent(t *testing.T) {
	// Setup
	sessionRepo := newMockSessionRepo()
	renewalRepo := newMockRenewalRepo()
	historyRepo := newMockHistoryRepo()
	widgetRepo := newMockWidgetRepo()
	eventBus := newMockEventBus()
	logger := newTestLogger()

	consentSvc := NewConsentService(widgetRepo, sessionRepo, historyRepo, eventBus, "key", logger)
	expirySvc := NewConsentExpiryService(sessionRepo, renewalRepo, historyRepo, widgetRepo, eventBus, logger, consentSvc)

	ctx := context.Background()
	tenantID := types.NewID()
	subjectID := types.NewID()
	purposeID := types.NewID()

	// 1. Create Widget with 30-day expiry
	widget := &consent.ConsentWidget{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID()},
			TenantID:   tenantID,
		},
		Config: consent.WidgetConfig{ConsentExpiryDays: 30},
	}
	widgetRepo.Create(ctx, widget)

	// 2. Create Session created 25 days ago (Expires in 5 days)
	createdAt := time.Now().Add(-25 * 24 * time.Hour)
	session := consent.ConsentSession{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: createdAt,
		},
		TenantID:  tenantID,
		WidgetID:  widget.ID,
		SubjectID: &subjectID,
		Decisions: []consent.ConsentDecision{
			{PurposeID: purposeID, Granted: true},
		},
	}
	sessionRepo.Create(ctx, &session)

	// 3. Setup Initial History State (GRANTED)
	historyRepo.Create(ctx, &consent.ConsentHistoryEntry{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: createdAt,
		},
		TenantID:      tenantID,
		SubjectID:     subjectID,
		PurposeID:     purposeID,
		NewStatus:     "GRANTED",
		NoticeVersion: "1.0",
		WidgetID:      &widget.ID,
	})

	// 4. Run CheckExpiries
	err := expirySvc.CheckExpiries(ctx)
	require.NoError(t, err)

	// 5. Verify Reminder Event
	// 5 days remaining <= 7 days -> expect 7d reminder
	found := false
	for _, e := range eventBus.Events {
		if e.Type == "consent.expiry_reminder_7d" {
			data := e.Data.(map[string]any)
			if data["subject_id"] == subjectID.String() && data["purpose_id"] == purposeID.String() {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "expected 7d expiry reminder")
}

func TestExpiryChecker_MarksExpired(t *testing.T) {
	// Setup
	sessionRepo := newMockSessionRepo()
	renewalRepo := newMockRenewalRepo()
	historyRepo := newMockHistoryRepo()
	widgetRepo := newMockWidgetRepo()
	eventBus := newMockEventBus()
	logger := newTestLogger()

	consentSvc := NewConsentService(widgetRepo, sessionRepo, historyRepo, eventBus, "key", logger)
	expirySvc := NewConsentExpiryService(sessionRepo, renewalRepo, historyRepo, widgetRepo, eventBus, logger, consentSvc)

	ctx := context.Background()
	tenantID := types.NewID()
	subjectID := types.NewID()
	purposeID := types.NewID()

	// 1. Create Widget with 30-day expiry
	widget := &consent.ConsentWidget{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID()},
			TenantID:   tenantID,
		},
		Config: consent.WidgetConfig{ConsentExpiryDays: 30},
	}
	widgetRepo.Create(ctx, widget)

	// 2. Create Session created 31 days ago (Expired 1 day ago)
	createdAt := time.Now().Add(-31 * 24 * time.Hour)
	session := consent.ConsentSession{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: createdAt,
		},
		TenantID:  tenantID,
		WidgetID:  widget.ID,
		SubjectID: &subjectID,
		Decisions: []consent.ConsentDecision{
			{PurposeID: purposeID, Granted: true},
		},
	}
	sessionRepo.Create(ctx, &session)

	// 3. Setup Initial History State
	historyRepo.Create(ctx, &consent.ConsentHistoryEntry{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: createdAt,
		},
		TenantID:      tenantID,
		SubjectID:     subjectID,
		PurposeID:     purposeID,
		NewStatus:     "GRANTED",
		NoticeVersion: "1.0",
		WidgetID:      &widget.ID,
	})

	// 4. Run CheckExpiries
	err := expirySvc.CheckExpiries(ctx)
	require.NoError(t, err)

	// 5. Verify Expiry Event and History Update
	// Check history for "EXPIRED" entry
	latest, err := historyRepo.GetLatestState(ctx, tenantID, subjectID, purposeID)
	require.NoError(t, err)
	assert.Equal(t, "EXPIRED", latest.NewStatus)

	// Check event
	found := false
	for _, e := range eventBus.Events {
		if e.Type == "consent.expired" {
			data := e.Data.(map[string]any)
			if data["subject_id"] == subjectID.String() && data["purpose_id"] == purposeID.String() {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "expected expired event")
}
