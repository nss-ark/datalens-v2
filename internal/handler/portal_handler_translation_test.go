package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/handler"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPortalHandler_GetNotice_WithTranslation(t *testing.T) {
	// Setup Mocks
	mockNoticeRepo := new(mockNoticeRepoForTest)
	mockTransRepo := new(mockTranslationRepoForTest)
	mockBus := new(mockTranslationEventBus)

	// Services
	noticeSvc := service.NewNoticeService(mockNoticeRepo, nil, mockBus, nil)
	transSvc := service.NewTranslationService(mockTransRepo, mockNoticeRepo, mockBus, "", "")

	// Handler (pass nils for unused dependencies)
	h := handler.NewPortalHandler(nil, nil, nil, nil, noticeSvc, transSvc, nil, nil)

	// Router
	router := h.Routes()
	// PortalHandler.Routes() mounts at root of its sub-router.
	// `r.Get("/notice/{id}", h.getNotice)` is inside `Routes()`.

	// Let's check `Routes()` implementation in `portal_handler.go`.
	// It returns a chi.Router.
	// `r.Get("/notice/{id}", h.getNotice)`

	// So if I use `h.Routes()`, the path is `/notice/{id}` relative to that router.

	// Data
	noticeID := types.NewID()
	lang := "hi"
	version := 1

	notice := &consent.ConsentNotice{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: noticeID}},
		Version:      version,
		Status:       consent.NoticeStatusPublished,
		Title:        "Privacy Policy",
		Content:      "Original English Content",
	}

	translation := &consent.ConsentNoticeTranslation{
		NoticeID:       noticeID,
		LanguageCode:   lang,
		TranslatedText: "Translated Hindi Content",
	}

	// Expectations
	mockNoticeRepo.On("GetByID", mock.Anything, noticeID).Return(notice, nil)
	mockTransRepo.On("GetByNoticeAndLang", mock.Anything, noticeID, version, lang).Return(translation, nil)

	// Request
	req := httptest.NewRequest("GET", "/notice/"+noticeID.String()+"?lang="+lang, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var result consent.ConsentNotice
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, "Translated Hindi Content", result.Content)
	assert.Equal(t, "Translated Hindi Content", result.Title) // Assuming we overlaid both?
	// In my impl: `notice.Title = translation.TranslatedText` AND `notice.Content = translation.TranslatedText`
	// Because translation only has one text field.
}
