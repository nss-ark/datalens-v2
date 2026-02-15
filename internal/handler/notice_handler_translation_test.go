package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/handler"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repos
type mockTranslationRepoForTest struct {
	mock.Mock
}

func (m *mockTranslationRepoForTest) SaveTranslation(ctx context.Context, t *consent.ConsentNoticeTranslation) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *mockTranslationRepoForTest) GetByNoticeAndVersion(ctx context.Context, noticeID types.ID, version int) ([]consent.ConsentNoticeTranslation, error) {
	args := m.Called(ctx, noticeID, version)
	return args.Get(0).([]consent.ConsentNoticeTranslation), args.Error(1)
}

func (m *mockTranslationRepoForTest) GetByNoticeAndLang(ctx context.Context, noticeID types.ID, version int, lang string) (*consent.ConsentNoticeTranslation, error) {
	args := m.Called(ctx, noticeID, version, lang)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*consent.ConsentNoticeTranslation), args.Error(1)
}

func (m *mockTranslationRepoForTest) Upsert(ctx context.Context, t *consent.ConsentNoticeTranslation) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

// NoticeRepo mock (partial)
type mockNoticeRepoForTest struct {
	mock.Mock
}

func (m *mockNoticeRepoForTest) Create(ctx context.Context, n *consent.ConsentNotice) error {
	return m.Called(ctx, n).Error(0)

}
func (m *mockNoticeRepoForTest) GetByID(ctx context.Context, id types.ID) (*consent.ConsentNotice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*consent.ConsentNotice), args.Error(1)
}
func (m *mockNoticeRepoForTest) GetByTenant(ctx context.Context, tenantID types.ID) ([]consent.ConsentNotice, error) {
	return nil, nil // Not needed
}
func (m *mockNoticeRepoForTest) Update(ctx context.Context, n *consent.ConsentNotice) error {
	return nil
}
func (m *mockNoticeRepoForTest) Publish(ctx context.Context, id types.ID) (int, error) {
	return 0, nil
}
func (m *mockNoticeRepoForTest) Archive(ctx context.Context, id types.ID) error {
	return nil
}
func (m *mockNoticeRepoForTest) BindToWidgets(ctx context.Context, noticeID types.ID, widgetIDs []types.ID) error {
	return nil
}
func (m *mockNoticeRepoForTest) GetLatestVersion(ctx context.Context, seriesID types.ID) (int, error) {
	return 0, nil
}

// EventBus mock
type mockTranslationEventBus struct {
	mock.Mock
}

func (m *mockTranslationEventBus) Publish(ctx context.Context, event eventbus.Event) error {
	return m.Called(ctx, event).Error(0)
}

func (m *mockTranslationEventBus) Subscribe(ctx context.Context, pattern string, handler eventbus.EventHandler) (eventbus.Subscription, error) {
	return nil, nil
}

func (m *mockTranslationEventBus) Close() error {
	return nil
}

func TestNoticeHandler_GetTranslation(t *testing.T) {
	// Setup
	mockTransRepo := new(mockTranslationRepoForTest)
	mockNoticeRepo := new(mockNoticeRepoForTest)
	mockBus := new(mockTranslationEventBus)

	// Create service with mocks
	svc := service.NewTranslationService(mockTransRepo, mockNoticeRepo, mockBus, "", "")

	// Create handler
	h := handler.NewNoticeHandler(nil, svc) // We don't need NoticeService for this test

	// Router
	r := chi.NewRouter()
	r.Get("/notices/{id}/translations/{lang}", h.GetTranslation)

	// Data
	noticeID := types.NewID()
	lang := "hi"
	version := 1

	notice := &consent.ConsentNotice{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: noticeID}},
		Version:      version,
	}
	translation := &consent.ConsentNoticeTranslation{
		NoticeID:       noticeID,
		LanguageCode:   lang,
		TranslatedText: "नमस्ते",
	}

	// Mocks
	mockNoticeRepo.On("GetByID", mock.Anything, noticeID).Return(notice, nil)
	mockTransRepo.On("GetByNoticeAndLang", mock.Anything, noticeID, version, lang).Return(translation, nil)

	// User Request
	req := httptest.NewRequest("GET", "/notices/"+noticeID.String()+"/translations/"+lang, nil)
	w := httptest.NewRecorder()

	// Execute
	r.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var result consent.ConsentNoticeTranslation
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "नमस्ते", result.TranslatedText)
}

func TestNoticeHandler_GetTranslation_NotFound(t *testing.T) {
	// Setup
	mockTransRepo := new(mockTranslationRepoForTest)
	mockNoticeRepo := new(mockNoticeRepoForTest)
	mockBus := new(mockTranslationEventBus)

	svc := service.NewTranslationService(mockTransRepo, mockNoticeRepo, mockBus, "", "")
	h := handler.NewNoticeHandler(nil, svc)

	r := chi.NewRouter()
	r.Get("/notices/{id}/translations/{lang}", h.GetTranslation)

	noticeID := types.NewID()
	lang := "hi"
	version := 1

	notice := &consent.ConsentNotice{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: noticeID}},
		Version:      version,
	}

	// Mocks
	mockNoticeRepo.On("GetByID", mock.Anything, noticeID).Return(notice, nil)
	mockTransRepo.On("GetByNoticeAndLang", mock.Anything, noticeID, version, lang).Return(nil, nil) // Return nil, nil for not found

	// Request
	req := httptest.NewRequest("GET", "/notices/"+noticeID.String()+"/translations/"+lang, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
