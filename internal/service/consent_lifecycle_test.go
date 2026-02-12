package service

import (
	"context"
	"testing"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsentLifecycle_GrantCheckWithdrawCheck(t *testing.T) {
	// Setup
	widgetRepo := newMockWidgetRepo()
	sessionRepo := newMockSessionRepo()
	historyRepo := newMockHistoryRepo()
	eventBus := newMockEventBus()
	logger := newTestLogger()

	svc := NewConsentService(widgetRepo, sessionRepo, historyRepo, eventBus, "test-secret", logger)

	ctx := context.Background()
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	// Pre-requisite: Create a widget
	widget, err := svc.CreateWidget(ctx, CreateWidgetRequest{
		Name: "Test Widget",
		Type: "BANNER",
	})
	require.NoError(t, err)

	subjectID := types.NewID()
	purposeA := types.NewID()
	purposeB := types.NewID()
	purposeC := types.NewID()

	// 1. Grant consent for purposes A, B, C
	t.Run("Grant Consent", func(t *testing.T) {
		req := RecordConsentRequest{
			WidgetID:  widget.ID,
			SubjectID: &subjectID,
			Decisions: []consent.ConsentDecision{
				{PurposeID: purposeA, Granted: true},
				{PurposeID: purposeB, Granted: true},
				{PurposeID: purposeC, Granted: true},
			},
			IPAddress:     "127.0.0.1",
			UserAgent:     "TestAgent",
			NoticeVersion: "1.0",
		}
		session, err := svc.RecordConsent(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, 3, len(session.Decisions))
	})

	// 2. Check consent for A -> should be GRANTED
	t.Run("Check Consent A Granted", func(t *testing.T) {
		granted, err := svc.CheckConsent(ctx, tenantID, subjectID, purposeA)
		require.NoError(t, err)
		assert.True(t, granted)
	})

	// 3. Withdraw consent for A
	t.Run("Withdraw Consent A", func(t *testing.T) {
		err := svc.WithdrawConsent(ctx, WithdrawConsentRequest{
			SubjectID:     subjectID,
			PurposeID:     purposeA,
			Source:        "PREFERENCE_CENTER",
			IPAddress:     "127.0.0.1",
			NoticeVersion: "1.0",
		})
		require.NoError(t, err)
	})

	// 4. Check consent for A -> should be WITHDRAWN (false)
	t.Run("Check Consent A Withdrawn", func(t *testing.T) {
		granted, err := svc.CheckConsent(ctx, tenantID, subjectID, purposeA)
		require.NoError(t, err)
		assert.False(t, granted)
	})

	// 5. Check consent for B -> should still be GRANTED
	t.Run("Check Consent B Still Granted", func(t *testing.T) {
		granted, err := svc.CheckConsent(ctx, tenantID, subjectID, purposeB)
		require.NoError(t, err)
		assert.True(t, granted)
	})
}

func TestConsentWithdrawal_EventEmitted(t *testing.T) {
	// Setup
	widgetRepo := newMockWidgetRepo()
	sessionRepo := newMockSessionRepo()
	historyRepo := newMockHistoryRepo()
	eventBus := newMockEventBus()
	logger := newTestLogger()

	svc := NewConsentService(widgetRepo, sessionRepo, historyRepo, eventBus, "test-secret", logger)

	ctx := context.Background()
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	subjectID := types.NewID()
	purposeID := types.NewID()

	// Withdraw consent
	err := svc.WithdrawConsent(ctx, WithdrawConsentRequest{
		SubjectID:     subjectID,
		PurposeID:     purposeID,
		Source:        "API",
		IPAddress:     "1.1.1.1",
		NoticeVersion: "1.0",
	})
	require.NoError(t, err)

	// Verify event
	require.Len(t, eventBus.Events, 1)
	assert.Equal(t, "consent.withdrawn", eventBus.Events[0].Type)
	data := eventBus.Events[0].Data.(map[string]any)
	assert.Equal(t, subjectID.String(), data["subject_id"])
	assert.Equal(t, purposeID.String(), data["purpose_id"])
}
