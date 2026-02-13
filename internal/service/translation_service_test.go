package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslationService_TranslateNotice_Success(t *testing.T) {
	// Setup Mock Server for HuggingFace
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)

		// Basic check of payload
		assert.Equal(t, "This is a privacy notice.", body["inputs"])

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		// IndicTrans2 format: [{"translation_text": "..."}]
		config := []map[string]string{
			{"translation_text": "Mock Translated Text"},
		}
		json.NewEncoder(w).Encode(config)
	}))
	defer ts.Close()

	// Setup Service
	transRepo := newMockTranslationRepo()
	noticeRepo := newMockNoticeRepo()
	eventBus := newMockEventBus()

	svc := NewTranslationService(transRepo, noticeRepo, eventBus, "test-api-key", ts.URL)
	// Lower timeout for tests
	svc.requestTimeout = 1 * time.Second

	// Setup Data
	ctx := context.Background()
	tenantID := types.NewID()

	// Create and Publish a Notice
	notice := &consent.ConsentNotice{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{ID: types.NewID()},
			TenantID:   tenantID,
		},
		Title:   "Privacy Policy",
		Content: "This is a privacy notice.",
		Status:  consent.NoticeStatusPublished, // Must be published
		Version: 1,
	}
	require.NoError(t, noticeRepo.Create(ctx, notice))

	// Execute
	translations, err := svc.TranslateNotice(ctx, notice.ID)
	require.NoError(t, err)

	// Assert
	// Only 22 languages? Check if we got results.
	// 22 Scheduled languages (excluding English)
	assert.NotEmpty(t, translations)
	assert.Len(t, translations, 22)

	// Check content of first translation
	first := translations[0]
	assert.Equal(t, notice.ID, first.NoticeID)
	assert.Equal(t, 1, first.NoticeVersion)
	assert.Equal(t, "Mock Translated Text", first.TranslatedText)
	assert.Equal(t, "INDICTRANS2", first.TranslationSource)

	// Verify Event Published
	assert.Len(t, eventBus.Events, 1)
	assert.Equal(t, "consent.notice_translated", eventBus.Events[0].Type)
}

func TestTranslationService_TranslateNotice_DraftError(t *testing.T) {
	svc := NewTranslationService(newMockTranslationRepo(), newMockNoticeRepo(), newMockEventBus(), "key", "url")
	ctx := context.Background()

	// Create Draft Notice
	notice := &consent.ConsentNotice{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: types.NewID()}},
		Status:       consent.NoticeStatusDraft,
	}
	repo := svc.noticeRepo.(*mockNoticeRepo) // Access mock underlying
	repo.Create(ctx, notice)

	_, err := svc.TranslateNotice(ctx, notice.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only published notices")
}

func TestTranslationService_OverrideTranslation(t *testing.T) {
	transRepo := newMockTranslationRepo()
	noticeRepo := newMockNoticeRepo()
	svc := NewTranslationService(transRepo, noticeRepo, newMockEventBus(), "key", "url")
	ctx := context.Background()

	notice := &consent.ConsentNotice{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: types.NewID()}},
		Status:       consent.NoticeStatusPublished,
		Version:      1,
	}
	noticeRepo.Create(ctx, notice)

	// 1. Create initial auto-translation (simulated)
	autoTrans := &consent.ConsentNoticeTranslation{
		BaseEntity:        types.BaseEntity{ID: types.NewID()},
		NoticeID:          notice.ID,
		NoticeVersion:     1,
		LanguageCode:      "hi",
		TranslatedText:    "Bad Auto Trans",
		TranslationSource: "INDICTRANS2",
	}
	transRepo.SaveTranslation(ctx, autoTrans)

	// 2. Override
	err := svc.OverrideTranslation(ctx, notice.ID, "hi", "Good Manual Trans")
	require.NoError(t, err)

	// 3. Verify
	stored, err := transRepo.GetByNoticeAndLang(ctx, notice.ID, 1, "hi")
	require.NoError(t, err)
	assert.Equal(t, "Good Manual Trans", stored.TranslatedText)
	assert.Equal(t, "MANUAL", stored.TranslationSource)
}

func TestTranslationService_GetTranslations(t *testing.T) {
	transRepo := newMockTranslationRepo()
	noticeRepo := newMockNoticeRepo()
	svc := NewTranslationService(transRepo, noticeRepo, newMockEventBus(), "key", "url")
	ctx := context.Background()

	notice := &consent.ConsentNotice{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: types.NewID()}},
		Status:       consent.NoticeStatusPublished,
		Version:      5,
	}
	noticeRepo.Create(ctx, notice)

	// Add translations for v5
	t1 := &consent.ConsentNoticeTranslation{NoticeID: notice.ID, NoticeVersion: 5, LanguageCode: "hi"}
	t2 := &consent.ConsentNoticeTranslation{NoticeID: notice.ID, NoticeVersion: 5, LanguageCode: "ta"}
	// Add old v4
	t3 := &consent.ConsentNoticeTranslation{NoticeID: notice.ID, NoticeVersion: 4, LanguageCode: "hi"}

	transRepo.SaveTranslation(ctx, t1)
	transRepo.SaveTranslation(ctx, t2)
	transRepo.SaveTranslation(ctx, t3)

	results, err := svc.GetTranslations(ctx, notice.ID)
	require.NoError(t, err)
	assert.Len(t, results, 2) // only v5
}
