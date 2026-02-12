package service

import (
	"context"
	"testing"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoticeLifecycle_CreatePublishArchive(t *testing.T) {
	// Setup
	noticeRepo := newMockNoticeRepo()
	widgetRepo := newMockWidgetRepo()
	eventBus := newMockEventBus()
	logger := newTestLogger()

	svc := NewNoticeService(noticeRepo, widgetRepo, eventBus, logger)

	ctx := context.Background()
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	// 1. Create Draft
	req := CreateNoticeRequest{
		Title:      "Web Privacy Policy",
		Content:    "We track everything.",
		Purposes:   []types.ID{types.NewID()},
		Regulation: "GDPR",
	}
	notice, err := svc.Create(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, consent.NoticeStatusDraft, notice.Status)
	assert.Equal(t, 1, notice.Version)

	// 2. Update Draft
	err = svc.Update(ctx, UpdateNoticeRequest{
		ID:         notice.ID,
		Title:      "Web Privacy Policy V1",
		Content:    "We track mostly everything.",
		Purposes:   notice.Purposes,
		Regulation: "GDPR",
	})
	require.NoError(t, err)

	updated, err := svc.Get(ctx, notice.ID)
	require.NoError(t, err)
	assert.Equal(t, "Web Privacy Policy V1", updated.Title)

	// 3. Publish
	published, err := svc.Publish(ctx, notice.ID)
	require.NoError(t, err)
	assert.Equal(t, consent.NoticeStatusPublished, published.Status)
	assert.Equal(t, 2, published.Version) // Version increments on publish?
	// Wait, notice_service.go: Publish -> repo.Publish -> n.Version++; returns new version.
	// In Create, version is 1.
	// In Publish, version becomes 2.
	// Let's verify mock implementation of Publish: n.Version++.

	// 4. Archive
	err = svc.Archive(ctx, notice.ID)
	require.NoError(t, err)

	archived, err := svc.Get(ctx, notice.ID)
	require.NoError(t, err)
	assert.Equal(t, consent.NoticeStatusArchived, archived.Status)
}

func TestNoticePublish_CannotEditPublished(t *testing.T) {
	svc := NewNoticeService(newMockNoticeRepo(), newMockWidgetRepo(), newMockEventBus(), newTestLogger())
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, types.NewID())

	// Create and Publish
	notice, err := svc.Create(ctx, CreateNoticeRequest{Title: "T", Content: "C", Regulation: "R"})
	require.NoError(t, err)
	_, err = svc.Publish(ctx, notice.ID)
	require.NoError(t, err)

	// Attempt Update
	err = svc.Update(ctx, UpdateNoticeRequest{
		ID:      notice.ID,
		Title:   "New Title",
		Content: "New Content",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot update published")
}

func TestNoticeBinding_WidgetNotice(t *testing.T) {
	noticeRepo := newMockNoticeRepo()
	widgetRepo := newMockWidgetRepo()
	svc := NewNoticeService(noticeRepo, widgetRepo, newMockEventBus(), newTestLogger())

	ctx := context.Background()
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	// Create Notice
	notice, err := svc.Create(ctx, CreateNoticeRequest{Title: "T", Content: "C"})
	require.NoError(t, err)

	// Create Widget
	consentSvc := NewConsentService(widgetRepo, newMockSessionRepo(), newMockHistoryRepo(), newMockEventBus(), "key", newTestLogger())
	widget, err := consentSvc.CreateWidget(ctx, CreateWidgetRequest{Name: "W", Type: "BANNER"})
	require.NoError(t, err)

	// Bind
	err = svc.Bind(ctx, notice.ID, []types.ID{widget.ID})
	require.NoError(t, err)

	// Verify
	updated, err := svc.Get(ctx, notice.ID)
	require.NoError(t, err)
	assert.Contains(t, updated.WidgetIDs, widget.ID)
}
