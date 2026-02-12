package service

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Suite Setup
// =============================================================================

func newTestNoticeService() (*NoticeService, *mockConsentNoticeRepo, *mockConsentWidgetRepo, *mockEventBus) {
	noticeRepo := newMockConsentNoticeRepo()
	widgetRepo := newMockConsentWidgetRepo()
	eventBus := newMockEventBus()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	svc := NewNoticeService(noticeRepo, widgetRepo, eventBus, logger)
	return svc, noticeRepo, widgetRepo, eventBus
}

// =============================================================================
// Tests
// =============================================================================

func TestNoticeService_Create(t *testing.T) {
	svc, _, _, eventBus := newTestNoticeService()
	ctx := context.Background()
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	t.Run("success", func(t *testing.T) {
		req := CreateNoticeRequest{
			Title:      "Privacy Policy",
			Content:    "We respect your privacy...",
			Regulation: "GDPR",
		}

		notice, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, notice.ID)
		assert.Equal(t, tenantID, notice.TenantID)
		assert.Equal(t, "Privacy Policy", notice.Title)
		assert.Equal(t, consent.NoticeStatusDraft, notice.Status)
		assert.Equal(t, 1, notice.Version)
		assert.NotEmpty(t, notice.SeriesID)

		// Verify event
		assert.Len(t, eventBus.Events, 1) // Create emits notice_created
		assert.Equal(t, "consent.notice_created", eventBus.Events[0].Type)
	})

	t.Run("missing title", func(t *testing.T) {
		_, err := svc.Create(ctx, CreateNoticeRequest{Content: "foo"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})
}

func TestNoticeService_Publish(t *testing.T) {
	svc, noticeRepo, _, eventBus := newTestNoticeService()
	ctx := context.Background()
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	// Create a draft notice directly in repo
	noticeID := types.NewID()
	notice := &consent.ConsentNotice{
		TenantEntity: types.TenantEntity{BaseEntity: types.BaseEntity{ID: noticeID}, TenantID: tenantID},
		Title:        "Draft Notice",
		Status:       consent.NoticeStatusDraft,
		Version:      1,
		SeriesID:     types.NewID(),
	}
	noticeRepo.Create(ctx, notice)

	t.Run("success", func(t *testing.T) {
		published, err := svc.Publish(ctx, noticeID)
		require.NoError(t, err)
		assert.Equal(t, noticeID, published.ID)
		assert.Equal(t, consent.NoticeStatusPublished, published.Status)
		assert.Equal(t, 2, published.Version) // Mock implementation logic: increments version
		assert.NotNil(t, published.PublishedAt)

		// Verify event
		assert.Len(t, eventBus.Events, 1)
		assert.Equal(t, "consent.notice_published", eventBus.Events[0].Type)
	})

	t.Run("already published", func(t *testing.T) {
		// Mock repo doesn't enforce transition rules strictly unless we add logic,
		// but service might check?
		// Service logic:
		// n, err := s.repo.GetByID...
		// if n.Status != Draft { error }

		// Set status to Published
		notice.Status = consent.NoticeStatusPublished
		noticeRepo.Update(ctx, notice)

		_, err := svc.Publish(ctx, noticeID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only draft notices can be published")
	})
}

// =============================================================================
// Mock Consent Notice Repository (Local Definition)
// =============================================================================

type mockConsentNoticeRepo struct {
	mu           sync.Mutex
	notices      map[types.ID]*consent.ConsentNotice
	translations map[types.ID][]consent.ConsentNoticeTranslation
}

func newMockConsentNoticeRepo() *mockConsentNoticeRepo {
	return &mockConsentNoticeRepo{
		notices:      make(map[types.ID]*consent.ConsentNotice),
		translations: make(map[types.ID][]consent.ConsentNoticeTranslation),
	}
}

func (r *mockConsentNoticeRepo) Create(_ context.Context, n *consent.ConsentNotice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n.ID == (types.ID{}) {
		n.ID = types.NewID()
	}
	r.notices[n.ID] = n
	return nil
}
func (r *mockConsentNoticeRepo) GetByID(_ context.Context, id types.ID) (*consent.ConsentNotice, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notices[id]
	if !ok {
		return nil, types.NewNotFoundError("notice", id)
	}
	return n, nil
}
func (r *mockConsentNoticeRepo) GetByTenant(_ context.Context, tenantID types.ID) ([]consent.ConsentNotice, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentNotice
	for _, n := range r.notices {
		if n.TenantID == tenantID {
			result = append(result, *n)
		}
	}
	return result, nil
}
func (r *mockConsentNoticeRepo) Update(_ context.Context, n *consent.ConsentNotice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notices[n.ID] = n
	return nil
}
func (r *mockConsentNoticeRepo) Publish(_ context.Context, id types.ID) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notices[id]
	if !ok {
		return 0, types.NewNotFoundError("notice", id)
	}
	n.Status = consent.NoticeStatusPublished
	n.Version++
	now := time.Now().UTC()
	n.PublishedAt = &now
	return n.Version, nil
}
func (r *mockConsentNoticeRepo) Archive(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notices[id]
	if !ok {
		return types.NewNotFoundError("notice", id)
	}
	n.Status = consent.NoticeStatusArchived
	return nil
}
func (r *mockConsentNoticeRepo) BindToWidgets(_ context.Context, noticeID types.ID, widgetIDs []types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notices[noticeID]
	if !ok {
		return types.NewNotFoundError("notice", noticeID)
	}
	n.WidgetIDs = widgetIDs
	return nil
}
func (r *mockConsentNoticeRepo) AddTranslation(_ context.Context, t *consent.ConsentNoticeTranslation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t.ID == (types.ID{}) {
		t.ID = types.NewID()
	}
	r.translations[t.NoticeID] = append(r.translations[t.NoticeID], *t)
	return nil
}
func (r *mockConsentNoticeRepo) GetTranslations(_ context.Context, noticeID types.ID) ([]consent.ConsentNoticeTranslation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.translations[noticeID], nil
}
func (r *mockConsentNoticeRepo) GetLatestVersion(_ context.Context, seriesID types.ID) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	maxVer := 0
	for _, n := range r.notices {
		if n.SeriesID == seriesID && n.Version > maxVer {
			maxVer = n.Version
		}
	}
	return maxVer, nil
}
