package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// CreateGrievanceRequest payload for submitting a grievance.
type CreateGrievanceRequest struct {
	Subject       string `json:"subject"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	DataSubjectID string `json:"data_subject_id"` // Provided by portal context or admin
}

// GrievanceService handles grievance lifecycle logic.
type GrievanceService struct {
	repo     compliance.GrievanceRepository
	eventBus eventbus.EventBus
	logger   *slog.Logger
}

// NewGrievanceService creates a new GrievanceService.
func NewGrievanceService(
	repo compliance.GrievanceRepository,
	eventBus eventbus.EventBus,
	logger *slog.Logger,
) *GrievanceService {
	return &GrievanceService{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger.With("service", "grievance"),
	}
}

// SubmitGrievance creates a new grievance with OPEN status and 30-day SLA.
func (s *GrievanceService) SubmitGrievance(ctx context.Context, req CreateGrievanceRequest) (*compliance.Grievance, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	if req.Subject == "" || req.Description == "" {
		return nil, types.NewValidationError("subject and description are required", nil)
	}

	// Parse DataSubjectID
	subjectID, err := types.ParseID(req.DataSubjectID)
	if err != nil {
		return nil, types.NewValidationError("invalid data_subject_id", map[string]any{"error": err.Error()})
	}

	// DPDPA allows 30 days for grievance redressal
	dueDate := time.Now().UTC().Add(30 * 24 * time.Hour)

	grievance := &compliance.Grievance{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{
				ID:        types.NewID(),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			TenantID: tenantID,
		},
		DataSubjectID: subjectID,
		Subject:       req.Subject,
		Description:   req.Description,
		Category:      req.Category,
		Status:        compliance.GrievanceStatusOpen,
		SubmittedAt:   time.Now().UTC(),
		DueDate:       &dueDate,
	}

	if err := s.repo.Create(ctx, grievance); err != nil {
		return nil, fmt.Errorf("create grievance: %w", err)
	}

	s.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventGrievanceSubmitted, "compliance", tenantID, grievance))
	return grievance, nil
}

// AssignGrievance assigns a grievance to a DPO/handler.
func (s *GrievanceService) AssignGrievance(ctx context.Context, id types.ID, assigneeID types.ID) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}

	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if g.TenantID != tenantID {
		return types.NewNotFoundError("grievance", map[string]any{"id": id})
	}

	g.AssignedTo = &assigneeID
	g.Status = compliance.GrievanceStatusInProgress
	g.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, g); err != nil {
		return fmt.Errorf("assign grievance: %w", err)
	}

	s.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventGrievanceAssigned, "compliance", tenantID, g))
	return nil
}

// ResolveGrievance marks a grievance as resolved with a resolution note.
func (s *GrievanceService) ResolveGrievance(ctx context.Context, id types.ID, resolution string) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}

	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if g.TenantID != tenantID {
		return types.NewNotFoundError("grievance", map[string]any{"id": id})
	}

	now := time.Now().UTC()
	g.Resolution = &resolution
	g.ResolvedAt = &now
	g.Status = compliance.GrievanceStatusResolved
	g.UpdatedAt = now

	if err := s.repo.Update(ctx, g); err != nil {
		return fmt.Errorf("resolve grievance: %w", err)
	}

	s.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventGrievanceResolved, "compliance", tenantID, g))
	return nil
}

// EscalateGrievance marks a grievance as escalated to a DPA.
func (s *GrievanceService) EscalateGrievance(ctx context.Context, id types.ID, authority string) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}

	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if g.TenantID != tenantID {
		return types.NewNotFoundError("grievance", map[string]any{"id": id})
	}

	g.EscalatedTo = &authority
	g.Status = compliance.GrievanceStatusEscalated
	g.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, g); err != nil {
		return fmt.Errorf("escalate grievance: %w", err)
	}

	s.eventBus.Publish(ctx, eventbus.NewEvent(eventbus.EventGrievanceEscalated, "compliance", tenantID, g))
	return nil
}

// SubmitFeedback captures data principal satisfaction feedback.
func (s *GrievanceService) SubmitFeedback(ctx context.Context, id types.ID, rating int, comment string) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}

	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if g.TenantID != tenantID {
		return types.NewNotFoundError("grievance", map[string]any{"id": id})
	}

	if g.Status != compliance.GrievanceStatusResolved {
		return types.NewValidationError("can only submit feedback for resolved grievances", nil)
	}

	g.FeedbackRating = &rating
	g.FeedbackComment = &comment
	g.Status = compliance.GrievanceStatusClosed
	g.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, g); err != nil {
		return fmt.Errorf("submit feedback: %w", err)
	}

	return nil
}

// GetGrievance retrieves a single grievance.
func (s *GrievanceService) GetGrievance(ctx context.Context, id types.ID) (*compliance.Grievance, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if g.TenantID != tenantID {
		return nil, types.NewNotFoundError("grievance", map[string]any{"id": id})
	}

	return g, nil
}

// ListByTenant lists grievances with filters.
func (s *GrievanceService) ListByTenant(ctx context.Context, filters map[string]any, pagination types.Pagination) (*types.PaginatedResult[compliance.Grievance], error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	return s.repo.ListByTenant(ctx, tenantID, filters, pagination)
}

// ListBySubject lists grievances for the authenticated portal user.
func (s *GrievanceService) ListBySubject(ctx context.Context, subjectID types.ID) ([]compliance.Grievance, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	return s.repo.ListBySubject(ctx, tenantID, subjectID)
}
