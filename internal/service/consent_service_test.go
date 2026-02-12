package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Suite Setup
// =============================================================================

func newTestConsentService() (*ConsentService, *mockConsentWidgetRepo, *mockConsentSessionRepo, *mockConsentHistoryRepo, *mockEventBus) {
	widgetRepo := newMockConsentWidgetRepo()
	sessionRepo := newMockConsentSessionRepo()
	historyRepo := newMockConsentHistoryRepo()
	eventBus := newMockEventBus()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	svc := NewConsentService(widgetRepo, sessionRepo, historyRepo, eventBus, "test-secret-key", logger)
	return svc, widgetRepo, sessionRepo, historyRepo, eventBus
}

// =============================================================================
// Tests
// =============================================================================

func TestConsentService_CreateWidget(t *testing.T) {
	svc, _, _, _, eventBus := newTestConsentService()
	ctx := context.Background()
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	t.Run("success", func(t *testing.T) {
		req := CreateWidgetRequest{
			Name:   "Cookie Banner",
			Type:   "BANNER",
			Domain: "example.com",
			Config: consent.WidgetConfig{
				RegulationRef: "DPDPA",
			},
			AllowedOrigins: []string{"https://example.com"},
		}

		widget, err := svc.CreateWidget(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, widget.ID)
		assert.Equal(t, tenantID, widget.TenantID)
		assert.Equal(t, "Cookie Banner", widget.Name)
		assert.Equal(t, consent.WidgetStatusDraft, widget.Status)
		assert.Equal(t, 1, widget.Version)
		assert.NotEmpty(t, widget.APIKey)
		assert.Contains(t, widget.EmbedCode, widget.APIKey)

		// Verify event
		assert.Len(t, eventBus.Events, 1)
		assert.Equal(t, eventbus.EventConsentWidgetCreated, eventBus.Events[0].Type)
	})

	t.Run("missing tenant context", func(t *testing.T) {
		_, err := svc.CreateWidget(context.Background(), CreateWidgetRequest{Name: "Fail"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tenant context required")
	})

	t.Run("validation error", func(t *testing.T) {
		_, err := svc.CreateWidget(ctx, CreateWidgetRequest{Name: ""}) // Missing name
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})
}

func TestConsentService_RecordConsent(t *testing.T) {
	svc, widgetRepo, sessionRepo, historyRepo, _ := newTestConsentService()
	ctx := context.Background()
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	// Create a widget first
	widget := &consent.ConsentWidget{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{
				ID:        types.NewID(),
				CreatedAt: time.Now(),
			},
			TenantID: tenantID,
		},
		Name:    "Test Widget",
		Version: 5,
		Status:  consent.WidgetStatusActive,
	}
	widgetRepo.Create(ctx, widget)

	t.Run("success", func(t *testing.T) {
		purposeID := types.NewID()
		subjectID := types.NewID()

		req := RecordConsentRequest{
			WidgetID:  widget.ID,
			SubjectID: &subjectID,
			Decisions: []consent.ConsentDecision{
				{PurposeID: purposeID, Granted: true},
			},
			IPAddress:     "127.0.0.1",
			UserAgent:     "Mozilla/5.0",
			PageURL:       "https://example.com",
			NoticeVersion: "v1.0",
		}

		session, err := svc.RecordConsent(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, widget.ID, session.WidgetID)
		assert.Equal(t, 5, session.WidgetVersion)
		assert.NotEmpty(t, session.Signature)

		// Verify HMAC signature
		// Re-construct canonical data to verify
		canonical := struct {
			Decisions []consent.ConsentDecision `json:"decisions"`
			Timestamp string                    `json:"timestamp"`
		}{
			Decisions: req.Decisions,
			Timestamp: session.CreatedAt.Format(time.RFC3339Nano),
		}
		data, _ := json.Marshal(canonical)
		mac := hmac.New(sha256.New, []byte("test-secret-key"))
		mac.Write(data)
		expectedSig := "sha256:" + hex.EncodeToString(mac.Sum(nil))
		assert.Equal(t, expectedSig, session.Signature)

		// Verify session persisted
		persistedSession, err := sessionRepo.GetBySubject(ctx, tenantID, subjectID)
		require.NoError(t, err)
		assert.Len(t, persistedSession, 1)

		// Verify history created
		history, err := historyRepo.GetBySubject(ctx, tenantID, subjectID, types.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, history.Total)
		assert.Equal(t, "GRANTED", history.Items[0].NewStatus)
		assert.NotEmpty(t, history.Items[0].Signature)
	})

	t.Run("missing widget", func(t *testing.T) {
		req := RecordConsentRequest{
			WidgetID: types.NewID(), // Random ID
			Decisions: []consent.ConsentDecision{
				{PurposeID: types.NewID(), Granted: true},
			},
		}
		_, err := svc.RecordConsent(ctx, req)
		require.Error(t, err)
	})
}

func TestConsentService_CheckConsent(t *testing.T) {
	svc, _, _, historyRepo, _ := newTestConsentService()
	ctx := context.Background()
	tenantID := types.NewID()
	subjectID := types.NewID()
	purposeID := types.NewID()

	// Initially no consent
	granted, err := svc.CheckConsent(ctx, tenantID, subjectID, purposeID)
	require.NoError(t, err)
	assert.False(t, granted)

	// Add granted history (old)
	historyRepo.Create(ctx, &consent.ConsentHistoryEntry{
		BaseEntity: types.BaseEntity{ID: types.NewID(), CreatedAt: time.Now().Add(-2 * time.Hour)},
		TenantID:   tenantID,
		SubjectID:  subjectID,
		PurposeID:  purposeID,
		NewStatus:  "GRANTED",
	})

	granted, err = svc.CheckConsent(ctx, tenantID, subjectID, purposeID)
	require.NoError(t, err)
	assert.True(t, granted)

	// Add withdrawn history (newer)
	historyRepo.Create(ctx, &consent.ConsentHistoryEntry{
		BaseEntity: types.BaseEntity{ID: types.NewID(), CreatedAt: time.Now().Add(-1 * time.Hour)},
		TenantID:   tenantID,
		SubjectID:  subjectID,
		PurposeID:  purposeID,
		NewStatus:  "WITHDRAWN",
	})

	granted, err = svc.CheckConsent(ctx, tenantID, subjectID, purposeID)
	require.NoError(t, err)
	assert.False(t, granted)
}

// =============================================================================
// Mocks (Local Implementation)
// =============================================================================

// Mock ConsentWidgetRepository
type mockConsentWidgetRepo struct {
	mu      sync.Mutex
	widgets map[types.ID]*consent.ConsentWidget
	byKey   map[string]*consent.ConsentWidget
}

func newMockConsentWidgetRepo() *mockConsentWidgetRepo {
	return &mockConsentWidgetRepo{
		widgets: make(map[types.ID]*consent.ConsentWidget),
		byKey:   make(map[string]*consent.ConsentWidget),
	}
}

func (r *mockConsentWidgetRepo) Create(_ context.Context, w *consent.ConsentWidget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if w.ID == (types.ID{}) {
		w.ID = types.NewID()
	}
	r.widgets[w.ID] = w
	if w.APIKey != "" {
		r.byKey[w.APIKey] = w
	}
	return nil
}

func (r *mockConsentWidgetRepo) GetByID(_ context.Context, id types.ID) (*consent.ConsentWidget, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.widgets[id]
	if !ok {
		return nil, types.NewNotFoundError("consent widget", id)
	}
	return w, nil
}

func (r *mockConsentWidgetRepo) GetByTenant(_ context.Context, tenantID types.ID) ([]consent.ConsentWidget, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentWidget
	for _, w := range r.widgets {
		if w.TenantID == tenantID {
			result = append(result, *w)
		}
	}
	return result, nil
}

func (r *mockConsentWidgetRepo) GetByAPIKey(_ context.Context, apiKey string) (*consent.ConsentWidget, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.byKey[apiKey]
	if !ok {
		return nil, types.NewNotFoundError("consent widget", nil)
	}
	return w, nil
}

func (r *mockConsentWidgetRepo) Update(_ context.Context, w *consent.ConsentWidget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.widgets[w.ID] = w
	if w.APIKey != "" {
		r.byKey[w.APIKey] = w
	}
	return nil
}

func (r *mockConsentWidgetRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.widgets[id]
	if ok {
		delete(r.widgets, id)
		delete(r.byKey, w.APIKey)
	}
	return nil
}

// Mock ConsentSessionRepository
type mockConsentSessionRepo struct {
	mu       sync.Mutex
	sessions []consent.ConsentSession
}

func newMockConsentSessionRepo() *mockConsentSessionRepo {
	return &mockConsentSessionRepo{}
}

func (r *mockConsentSessionRepo) Create(_ context.Context, s *consent.ConsentSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == (types.ID{}) {
		s.ID = types.NewID()
	}
	r.sessions = append(r.sessions, *s)
	return nil
}

func (r *mockConsentSessionRepo) GetBySubject(_ context.Context, tenantID, subjectID types.ID) ([]consent.ConsentSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentSession
	for _, s := range r.sessions {
		if s.TenantID == tenantID && s.SubjectID != nil && *s.SubjectID == subjectID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *mockConsentSessionRepo) GetConversionStats(_ context.Context, tenantID types.ID, from, to time.Time, interval string) ([]consent.ConversionStat, error) {
	return nil, nil
}

func (r *mockConsentSessionRepo) GetPurposeStats(_ context.Context, tenantID types.ID, from, to time.Time) ([]consent.PurposeStat, error) {
	return nil, nil
}

// Mock ConsentHistoryRepository
type mockConsentHistoryRepo struct {
	mu      sync.Mutex
	history []consent.ConsentHistoryEntry
}

func newMockConsentHistoryRepo() *mockConsentHistoryRepo {
	return &mockConsentHistoryRepo{}
}

func (r *mockConsentHistoryRepo) Create(_ context.Context, entry *consent.ConsentHistoryEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry.ID == (types.ID{}) {
		entry.ID = types.NewID()
	}
	r.history = append(r.history, *entry)
	return nil
}

func (r *mockConsentHistoryRepo) GetBySubject(_ context.Context, tenantID, subjectID types.ID, p types.Pagination) (*types.PaginatedResult[consent.ConsentHistoryEntry], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var items []consent.ConsentHistoryEntry
	for _, h := range r.history {
		if h.TenantID == tenantID && h.SubjectID == subjectID {
			items = append(items, h)
		}
	}
	return &types.PaginatedResult[consent.ConsentHistoryEntry]{Items: items, Total: len(items)}, nil
}

func (r *mockConsentHistoryRepo) GetByPurpose(_ context.Context, tenantID, purposeID types.ID) ([]consent.ConsentHistoryEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentHistoryEntry
	for _, h := range r.history {
		if h.TenantID == tenantID && h.PurposeID == purposeID {
			result = append(result, h)
		}
	}
	return result, nil
}

func (r *mockConsentHistoryRepo) GetLatestState(_ context.Context, tenantID, subjectID, purposeID types.ID) (*consent.ConsentHistoryEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var latest *consent.ConsentHistoryEntry
	for i := range r.history {
		h := &r.history[i]
		if h.TenantID == tenantID && h.SubjectID == subjectID && h.PurposeID == purposeID {
			if latest == nil || h.CreatedAt.After(latest.CreatedAt) {
				latest = h
			}
		}
	}
	return latest, nil
}
