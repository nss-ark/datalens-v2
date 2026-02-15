package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

type NoticeService struct {
	repo       consent.ConsentNoticeRepository
	widgetRepo consent.ConsentWidgetRepository
	eventBus   eventbus.EventBus
	logger     *slog.Logger
}

func NewNoticeService(repo consent.ConsentNoticeRepository, widgetRepo consent.ConsentWidgetRepository, eventBus eventbus.EventBus, logger *slog.Logger) *NoticeService {
	return &NoticeService{
		repo:       repo,
		widgetRepo: widgetRepo,
		eventBus:   eventBus,
		logger:     logger,
	}
}

type CreateNoticeRequest struct {
	Title      string                     `json:"title"`
	Content    string                     `json:"content"`
	Purposes   []types.ID                 `json:"purposes"`
	Regulation string                     `json:"regulation"`
	Schema     consent.NoticeSchemaFields `json:"schema"`
	SeriesID   *types.ID                  `json:"series_id"` // Optional: if creating a new version of existing series
}

func (s *NoticeService) Create(ctx context.Context, req CreateNoticeRequest) (*consent.ConsentNotice, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	if req.Title == "" {
		return nil, types.NewValidationError("title is required", nil)
	}

	seriesID := types.NewID()
	version := 1

	// If SeriesID is provided, we are creating a new version in that series
	if req.SeriesID != nil {
		seriesID = *req.SeriesID
		// Version management could be enhanced here to check latest version
	}

	notice := &consent.ConsentNotice{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{
				ID:        types.NewID(),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			TenantID: tenantID,
		},
		SeriesID:   seriesID,
		Title:      req.Title,
		Content:    req.Content,
		Version:    version,
		Status:     consent.NoticeStatusDraft,
		Purposes:   req.Purposes,
		WidgetIDs:  []types.ID{},
		Regulation: req.Regulation,
		Schema:     req.Schema,
	}

	if err := s.repo.Create(ctx, notice); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "consent.notice_created", tenantID, notice)
	return notice, nil
}

func (s *NoticeService) List(ctx context.Context) ([]consent.ConsentNotice, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.GetByTenant(ctx, tenantID)
}

func (s *NoticeService) Get(ctx context.Context, id types.ID) (*consent.ConsentNotice, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Verify tenant access
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok || n.TenantID != tenantID {
		return nil, types.NewNotFoundError("notice not found", map[string]any{"id": id})
	}
	return n, nil
}

// GetPublic fetches a notice by ID without tenant context check, but ensures it is PUBLISHED.
func (s *NoticeService) GetPublic(ctx context.Context, id types.ID) (*consent.ConsentNotice, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if n.Status != consent.NoticeStatusPublished {
		return nil, types.NewNotFoundError("notice not found or not published", map[string]any{"id": id})
	}
	return n, nil
}

type UpdateNoticeRequest struct {
	ID         types.ID                   `json:"id"`
	Title      string                     `json:"title"`
	Content    string                     `json:"content"`
	Purposes   []types.ID                 `json:"purposes"`
	Regulation string                     `json:"regulation"`
	Schema     consent.NoticeSchemaFields `json:"schema"`
}

func (s *NoticeService) Update(ctx context.Context, req UpdateNoticeRequest) error {
	n, err := s.Get(ctx, req.ID)
	if err != nil {
		return err
	}

	if n.Status == consent.NoticeStatusPublished || n.Status == consent.NoticeStatusArchived {
		return types.NewDomainError("CONFLICT", "cannot update published or archived notice; create a new version instead")
	}

	n.Title = req.Title
	n.Content = req.Content
	n.Purposes = req.Purposes
	n.Regulation = req.Regulation
	n.Schema = req.Schema
	n.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, n); err != nil {
		return err
	}

	s.publishEvent(ctx, "consent.notice_updated", n.TenantID, n)
	return nil
}

func (s *NoticeService) Publish(ctx context.Context, id types.ID) (*consent.ConsentNotice, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if n.Status != consent.NoticeStatusDraft {
		return nil, types.NewDomainError("CONFLICT", "only draft notices can be published")
	}

	// Validate Schema (DPDP Rule 3(1) Schedule I)
	if missing := s.ValidateSchema(n); len(missing) > 0 {
		return nil, types.NewValidationError(fmt.Sprintf("schema validation failed: missing required fields: %v", missing), map[string]any{"missing_fields": missing})
	}

	newVersion, err := s.repo.Publish(ctx, id)
	if err != nil {
		return nil, err
	}

	n.Status = consent.NoticeStatusPublished
	n.Version = newVersion
	now := time.Now().UTC()
	n.PublishedAt = &now

	s.publishEvent(ctx, "consent.notice_published", n.TenantID, map[string]any{
		"id":        id,
		"series_id": n.SeriesID,
		"version":   n.Version,
	})
	return n, nil
}

func (s *NoticeService) Archive(ctx context.Context, id types.ID) error {
	n, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Archive(ctx, id); err != nil {
		return err
	}

	s.publishEvent(ctx, "consent.notice_archived", n.TenantID, map[string]any{"id": id})
	return nil
}

func (s *NoticeService) Bind(ctx context.Context, noticeID types.ID, widgetIDs []types.ID) error {
	n, err := s.Get(ctx, noticeID)
	if err != nil {
		return err
	}

	// Verify widgets belong to tenant
	for _, wid := range widgetIDs {
		w, err := s.widgetRepo.GetByID(ctx, wid)
		if err != nil {
			return fmt.Errorf("verify widget %s: %w", wid, err)
		}
		if w.TenantID != n.TenantID {
			return types.NewForbiddenError(fmt.Sprintf("widget %s belongs to different tenant", wid))
		}
	}

	if err := s.repo.BindToWidgets(ctx, noticeID, widgetIDs); err != nil {
		return err
	}

	s.publishEvent(ctx, "consent.notice_bound", n.TenantID, map[string]any{
		"notice_id":  noticeID,
		"series_id":  n.SeriesID,
		"widget_ids": widgetIDs,
	})
	return nil
}

// ValidateSchema checks if the notice contains all required fields per Schedule I.
func (s *NoticeService) ValidateSchema(n *consent.ConsentNotice) []string {
	var missing []string

	// R3(1)(a)
	if len(n.Schema.DataTypesCollected) == 0 {
		missing = append(missing, "data_types_collected")
	}
	if len(n.Schema.Purposes) == 0 {
		missing = append(missing, "purposes")
	}

	// R3(1)(b)
	if n.Schema.FiduciaryName == "" {
		missing = append(missing, "fiduciary_name")
	}
	if n.Schema.FiduciaryContact == "" {
		missing = append(missing, "fiduciary_contact")
	}

	// R3(1)(d) - Rights
	// Checking if boolean flags are explicitly false is tricky as false is default.
	// However, usually "false" means "not included in notice", which is a violation if the right exists.
	// But simply checking bools might be too strict if we assume default false means "no".
	// The requirement is that the notice MUST contain info about these rights.
	// So we assume the UI sets these to true when the user confirms they added the text.
	// If they are false, it means the user hasn't confirmed inclusion.
	if !n.Schema.RightsWithdraw {
		missing = append(missing, "rights_withdraw")
	}
	if !n.Schema.RightsAccess {
		missing = append(missing, "rights_access")
	}
	if !n.Schema.RightsCorrection {
		missing = append(missing, "rights_correction")
	}
	if !n.Schema.RightsGrievance {
		missing = append(missing, "rights_grievance")
	}
	if !n.Schema.RightsNomination {
		missing = append(missing, "rights_nomination")
	}

	// R3(1)(e)
	if n.Schema.ComplaintMethod == "" {
		missing = append(missing, "complaint_method")
	}
	if n.Schema.BoardComplaintMethod == "" {
		missing = append(missing, "board_complaint_method")
	}

	// R3(1)(f)
	// SharingCategories might be empty if no sharing, but good to enforce explicit "None" string if so?
	// For now, let's assume empty slice is allowed IF the user explicitly says "No sharing".
	// But our schema is just fields. Let's enforce it for now as "must declare sharing or lack thereof".
	// Actually, strict schema validation usually requires something.
	// Let's rely on the fact that the specific field is populated.
	// If the user intends "None", they should probably select "None".
	if len(n.Schema.SharingCategories) == 0 {
		missing = append(missing, "sharing_categories")
	}

	// R3(1)(g)
	if n.Schema.CrossBorderTransfer == "" {
		missing = append(missing, "cross_border_transfer")
	}

	// R3(1)(h)
	if n.Schema.RetentionPeriod == "" {
		missing = append(missing, "retention_period")
	}

	// Grievance Officer / DPO
	if n.Schema.DPOName == "" {
		missing = append(missing, "dpo_name")
	}
	if n.Schema.DPOContact == "" {
		missing = append(missing, "dpo_contact")
	}

	return missing
}

// CheckCompliance runs validation validation without publishing.
func (s *NoticeService) CheckCompliance(ctx context.Context, id types.ID) (map[string]any, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	missing := s.ValidateSchema(n)
	valid := len(missing) == 0

	return map[string]any{
		"valid":          valid,
		"missing_fields": missing,
		"schema":         n.Schema,
	}, nil
}

// publishEvent publishes a domain event (best-effort)
func (s *NoticeService) publishEvent(ctx context.Context, eventType string, tenantID types.ID, data any) {
	event := eventbus.NewEvent(eventType, "consent", tenantID, data)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("failed to publish event",
			slog.String("event_type", eventType),
			slog.String("error", err.Error()),
		)
	}
}
