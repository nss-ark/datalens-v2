// Package service contains unit tests for domain services.
// This file provides in-memory mock implementations for Batch 16 repositories.
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Mock Translation Repository
// =============================================================================

type mockTranslationRepo struct {
	mu           sync.Mutex
	translations map[types.ID]*consent.ConsentNoticeTranslation
}

func newMockTranslationRepo() *mockTranslationRepo {
	return &mockTranslationRepo{
		translations: make(map[types.ID]*consent.ConsentNoticeTranslation),
	}
}

func (r *mockTranslationRepo) SaveTranslation(_ context.Context, t *consent.ConsentNoticeTranslation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t.ID == (types.ID{}) {
		t.ID = types.NewID()
	}
	r.translations[t.ID] = t
	return nil
}

func (r *mockTranslationRepo) GetByNoticeAndVersion(_ context.Context, noticeID types.ID, version int) ([]consent.ConsentNoticeTranslation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.ConsentNoticeTranslation
	for _, t := range r.translations {
		if t.NoticeID == noticeID && t.NoticeVersion == version {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (r *mockTranslationRepo) GetByNoticeAndLang(_ context.Context, noticeID types.ID, version int, lang string) (*consent.ConsentNoticeTranslation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.translations {
		if t.NoticeID == noticeID && t.NoticeVersion == version && t.LanguageCode == lang {
			return t, nil
		}
	}
	// Return nil, nil if not found (as per typical repo pattern for "check existence")
	// Or error if strict. The service code checks `if existing != nil`.
	// Let's return nil, nil to indicate "not found but no error".
	return nil, nil
}

func (r *mockTranslationRepo) Upsert(_ context.Context, t *consent.ConsentNoticeTranslation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for existing to update
	for i, existing := range r.translations {
		if existing.NoticeID == t.NoticeID && existing.NoticeVersion == t.NoticeVersion && existing.LanguageCode == t.LanguageCode {
			// Update in place (preserving ID if needed, but manual override usually takes precedence)
			t.ID = existing.ID // Keep original ID
			r.translations[i] = t
			return nil
		}
	}

	// Create new
	if t.ID == (types.ID{}) {
		t.ID = types.NewID()
	}
	r.translations[t.ID] = t
	return nil
}

// =============================================================================
// Mock Notification Repository
// =============================================================================

type mockNotificationRepo struct {
	mu            sync.Mutex
	notifications map[types.ID]*consent.ConsentNotification
}

func newMockNotificationRepo() *mockNotificationRepo {
	return &mockNotificationRepo{
		notifications: make(map[types.ID]*consent.ConsentNotification),
	}
}

func (r *mockNotificationRepo) Create(_ context.Context, n *consent.ConsentNotification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n.ID == (types.ID{}) {
		n.ID = types.NewID()
	}
	r.notifications[n.ID] = n
	return nil
}

func (r *mockNotificationRepo) GetByID(_ context.Context, id types.ID) (*consent.ConsentNotification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notifications[id]
	if !ok {
		return nil, fmt.Errorf("notification not found")
	}
	return n, nil
}

func (r *mockNotificationRepo) ListByTenant(_ context.Context, tenantID types.ID, filter consent.NotificationFilter, pagination types.Pagination) (*types.PaginatedResult[consent.ConsentNotification], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var items []consent.ConsentNotification
	for _, n := range r.notifications {
		if n.TenantID != tenantID {
			continue
		}
		if filter.RecipientID != nil && n.RecipientID != *filter.RecipientID {
			continue
		}
		if filter.EventType != nil && n.EventType != *filter.EventType {
			continue
		}
		if filter.Channel != nil && n.Channel != *filter.Channel {
			continue
		}
		if filter.Status != nil && n.Status != *filter.Status {
			continue
		}
		items = append(items, *n)
	}
	return &types.PaginatedResult[consent.ConsentNotification]{Items: items, Total: len(items)}, nil
}

func (r *mockNotificationRepo) UpdateStatus(_ context.Context, id types.ID, status string, sentAt *time.Time, errorMessage *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notifications[id]
	if !ok {
		return fmt.Errorf("notification not found")
	}
	n.Status = status
	if sentAt != nil {
		n.SentAt = sentAt
	}
	if errorMessage != nil {
		n.ErrorMessage = errorMessage
	}
	return nil
}

// =============================================================================
// Mock Template Repository
// =============================================================================

type mockTemplateRepo struct {
	mu        sync.Mutex
	templates map[types.ID]*consent.NotificationTemplate
}

func newMockTemplateRepo() *mockTemplateRepo {
	return &mockTemplateRepo{
		templates: make(map[types.ID]*consent.NotificationTemplate),
	}
}

func (r *mockTemplateRepo) Create(_ context.Context, t *consent.NotificationTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t.ID == (types.ID{}) {
		t.ID = types.NewID()
	}
	r.templates[t.ID] = t
	return nil
}

func (r *mockTemplateRepo) Update(_ context.Context, t *consent.NotificationTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.templates[t.ID] = t
	return nil
}

func (r *mockTemplateRepo) GetByID(_ context.Context, id types.ID) (*consent.NotificationTemplate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.templates[id]
	if !ok {
		return nil, fmt.Errorf("template not found")
	}
	return t, nil
}

func (r *mockTemplateRepo) ListByTenant(_ context.Context, tenantID types.ID) ([]consent.NotificationTemplate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []consent.NotificationTemplate
	for _, t := range r.templates {
		if t.TenantID == tenantID {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (r *mockTemplateRepo) GetByEventAndChannel(_ context.Context, tenantID types.ID, eventType, channel string) (*consent.NotificationTemplate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.templates {
		if t.TenantID == tenantID && t.EventType == eventType && t.Channel == channel {
			return t, nil
		}
	}
	return nil, types.NewNotFoundError("template", map[string]any{"event": eventType, "channel": channel})
}

func (r *mockTemplateRepo) Delete(_ context.Context, id types.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.templates, id)
	return nil
}

// =============================================================================
// Mock Grievance Repository
// =============================================================================

type mockGrievanceRepo struct {
	mu         sync.Mutex
	grievances map[types.ID]*compliance.Grievance
}

func newMockGrievanceRepo() *mockGrievanceRepo {
	return &mockGrievanceRepo{
		grievances: make(map[types.ID]*compliance.Grievance),
	}
}

func (r *mockGrievanceRepo) Create(_ context.Context, g *compliance.Grievance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g.ID == (types.ID{}) {
		g.ID = types.NewID()
	}
	r.grievances[g.ID] = g
	return nil
}

func (r *mockGrievanceRepo) GetByID(_ context.Context, id types.ID) (*compliance.Grievance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.grievances[id]
	if !ok {
		return nil, types.NewNotFoundError("grievance", map[string]any{"id": id})
	}
	return g, nil
}

func (r *mockGrievanceRepo) ListByTenant(_ context.Context, tenantID types.ID, filters map[string]any, pagination types.Pagination) (*types.PaginatedResult[compliance.Grievance], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var items []compliance.Grievance
	for _, g := range r.grievances {
		if g.TenantID != tenantID {
			continue
		}
		// Apply basic filters if needed for tests
		if subjectID, ok := filters["data_subject_id"].(string); ok && subjectID != "" {
			if g.DataSubjectID.String() != subjectID {
				continue
			}
		}
		items = append(items, *g)
	}
	return &types.PaginatedResult[compliance.Grievance]{Items: items, Total: len(items)}, nil
}

func (r *mockGrievanceRepo) ListBySubject(_ context.Context, tenantID, subjectID types.ID) ([]compliance.Grievance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []compliance.Grievance
	for _, g := range r.grievances {
		if g.TenantID == tenantID && g.DataSubjectID == subjectID {
			result = append(result, *g)
		}
	}
	return result, nil
}

func (r *mockGrievanceRepo) Update(_ context.Context, g *compliance.Grievance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.grievances[g.ID] = g
	return nil
}
