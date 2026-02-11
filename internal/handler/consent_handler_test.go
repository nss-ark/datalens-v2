package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/handler"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Test Setup
// =============================================================================

func newTestConsentHandler() (*handler.ConsentHandler, *mockConsentWidgetRepo) {
	widgetRepo := &mockConsentWidgetRepo{
		widgets: make(map[types.ID]*consent.ConsentWidget),
		byKey:   make(map[string]*consent.ConsentWidget),
	}
	sessionRepo := &mockConsentSessionRepo{}
	historyRepo := &mockConsentHistoryRepo{}
	eventBus := &mockEventBus{}
	logger := slog.Default()

	svc := service.NewConsentService(widgetRepo, sessionRepo, historyRepo, eventBus, "test-secret", logger)
	h := handler.NewConsentHandler(svc)
	return h, widgetRepo
}

// =============================================================================
// Tests
// =============================================================================

func TestConsentHandler_PublicRoutes_GetWidgetConfig(t *testing.T) {
	h, widgetRepo := newTestConsentHandler()
	r := chi.NewRouter()
	r.Mount("/api/public/consent", h.PublicRoutes())

	// Setup data
	widgetID := types.NewID()
	apiKey := "valid-api-key"
	widget := &consent.ConsentWidget{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: widgetID},
			TenantID:   types.NewID(),
		},
		Name:    "Test Widget",
		APIKey:  apiKey,
		Status:  consent.WidgetStatusActive,
		Config:  consent.WidgetConfig{RegulationRef: "DPDPA"},
		Version: 1,
	}
	widgetRepo.Create(context.Background(), widget)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/public/consent/widget/config", nil)
		req.Header.Set("X-Widget-Key", apiKey)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Success bool                 `json:"success"`
			Data    consent.WidgetConfig `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "DPDPA", resp.Data.RegulationRef)
	})

	t.Run("missing api key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/public/consent/widget/config", nil)
		// No Header

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		// Service returns error because lookup by empty key fails (or returns not found)
		// Expect 404 or 400. Service GetByAPIKey("") -> Not Found likely.
		// Checking implementation: repo.GetByAPIKey relies on SQL.
		// Mock implementation checks map key.
		assert.NotEqual(t, http.StatusOK, rec.Code)
	})

	t.Run("invalid api key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/public/consent/widget/config", nil)
		req.Header.Set("X-Widget-Key", "invalid")

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestConsentHandler_PublicRoutes_RecordSession(t *testing.T) {
	h, widgetRepo := newTestConsentHandler()
	r := chi.NewRouter()
	r.Mount("/api/public/consent", h.PublicRoutes())

	widgetID := types.NewID()
	apiKey := "valid-api-key"
	widget := &consent.ConsentWidget{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: widgetID},
			TenantID:   types.NewID(),
		},
		Name:    "Test Widget",
		APIKey:  apiKey,
		Status:  consent.WidgetStatusActive,
		Version: 1,
	}
	widgetRepo.Create(context.Background(), widget)

	t.Run("success", func(t *testing.T) {
		// Middleware usually injects context, but here we are testing handler logic which gets ID from context
		// HOWEVER, the handler `recordConsent` says:
		// ctxWidgetID, ok := r.Context().Value(types.ContextKeyWidgetID).(types.ID)
		// If testing without middleware, we must inject this context manually OR assume middleware is present.
		// Since we are mounting PublicRoutes(), which doesn't include the middleware (middleware is likely applied globally or in broader router),
		// we should inject it in the test request.

		payload := service.RecordConsentRequest{
			WidgetID: widgetID, // Still needed in payload by struct, but handler overrides/checks context
			Decisions: []consent.ConsentDecision{
				{PurposeID: types.NewID(), Granted: true},
			},
			IPAddress: "1.2.3.4",
			UserAgent: "TestAgent",
			PageURL:   "http://localhost",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/public/consent/sessions", bytes.NewReader(body))

		// Inject ContextKeyWidgetID and TenantID as middleware would
		ctx := req.Context()
		ctx = context.WithValue(ctx, types.ContextKeyWidgetID, widgetID)
		ctx = context.WithValue(ctx, types.ContextKeyTenantID, widget.TenantID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp struct {
			Success bool                   `json:"success"`
			Data    consent.ConsentSession `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.NotEmpty(t, resp.Data.Signature)
	})

	t.Run("missing context (bypass middleware)", func(t *testing.T) {
		payload := service.RecordConsentRequest{
			WidgetID: widgetID,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/public/consent/sessions", bytes.NewReader(body))
		// No context injection

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// =============================================================================
// Mocks
// =============================================================================

type mockConsentWidgetRepo struct {
	widgets map[types.ID]*consent.ConsentWidget
	byKey   map[string]*consent.ConsentWidget
}

func (m *mockConsentWidgetRepo) Create(ctx context.Context, w *consent.ConsentWidget) error {
	m.widgets[w.ID] = w
	m.byKey[w.APIKey] = w
	return nil
}
func (m *mockConsentWidgetRepo) GetByID(ctx context.Context, id types.ID) (*consent.ConsentWidget, error) {
	if w, ok := m.widgets[id]; ok {
		return w, nil
	}
	return nil, types.NewNotFoundError("widget", id)
}
func (m *mockConsentWidgetRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]consent.ConsentWidget, error) {
	return nil, nil
}
func (m *mockConsentWidgetRepo) GetByAPIKey(ctx context.Context, key string) (*consent.ConsentWidget, error) {
	if w, ok := m.byKey[key]; ok {
		return w, nil
	}
	return nil, types.NewNotFoundError("widget", nil)
}
func (m *mockConsentWidgetRepo) Update(ctx context.Context, w *consent.ConsentWidget) error {
	return nil
}
func (m *mockConsentWidgetRepo) Delete(ctx context.Context, id types.ID) error { return nil }

type mockConsentSessionRepo struct{}

func (m *mockConsentSessionRepo) Create(ctx context.Context, s *consent.ConsentSession) error {
	s.CreatedAt = time.Now()
	return nil
}
func (m *mockConsentSessionRepo) GetBySubject(ctx context.Context, t, s types.ID) ([]consent.ConsentSession, error) {
	return nil, nil
}

type mockConsentHistoryRepo struct{}

func (m *mockConsentHistoryRepo) Create(ctx context.Context, h *consent.ConsentHistoryEntry) error {
	return nil
}
func (m *mockConsentHistoryRepo) GetBySubject(ctx context.Context, t, s types.ID, p types.Pagination) (*types.PaginatedResult[consent.ConsentHistoryEntry], error) {
	return nil, nil
}
func (m *mockConsentHistoryRepo) GetByPurpose(ctx context.Context, t, p types.ID) ([]consent.ConsentHistoryEntry, error) {
	return nil, nil
}
func (m *mockConsentHistoryRepo) GetLatestState(ctx context.Context, t, s, p types.ID) (*consent.ConsentHistoryEntry, error) {
	return nil, nil
}

type mockEventBus struct {
	Events []eventbus.Event
}

func (m *mockEventBus) Publish(ctx context.Context, e eventbus.Event) error {
	m.Events = append(m.Events, e)
	return nil
}
func (m *mockEventBus) Subscribe(ctx context.Context, t string, h eventbus.EventHandler) (eventbus.Subscription, error) {
	return nil, nil
}
func (m *mockEventBus) Close() error { return nil }
