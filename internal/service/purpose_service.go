package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// PurposeService handles data processing purpose operations.
type PurposeService struct {
	repo     governance.PurposeRepository
	eventBus eventbus.EventBus
	logger   *slog.Logger
}

// NewPurposeService creates a new PurposeService.
func NewPurposeService(repo governance.PurposeRepository, eb eventbus.EventBus, logger *slog.Logger) *PurposeService {
	return &PurposeService{
		repo:     repo,
		eventBus: eb,
		logger:   logger.With("service", "purpose"),
	}
}

// CreatePurposeInput holds fields for creating a purpose.
type CreatePurposeInput struct {
	TenantID        types.ID
	Code            string
	Name            string
	Description     string
	LegalBasis      types.LegalBasis
	RetentionDays   int
	RequiresConsent bool
}

// Create validates and persists a new purpose.
func (s *PurposeService) Create(ctx context.Context, in CreatePurposeInput) (*governance.Purpose, error) {
	if in.Code == "" {
		return nil, types.NewValidationError("code is required", nil)
	}
	if in.Name == "" {
		return nil, types.NewValidationError("name is required", nil)
	}
	if in.LegalBasis == "" {
		return nil, types.NewValidationError("legal_basis is required", nil)
	}
	if in.RetentionDays <= 0 {
		in.RetentionDays = 365
	}

	// Verify code uniqueness within tenant
	existing, err := s.repo.GetByCode(ctx, in.TenantID, in.Code)
	if err == nil && existing != nil {
		return nil, types.NewConflictError("Purpose", "code", in.Code)
	}

	p := &governance.Purpose{
		Code:            in.Code,
		Name:            in.Name,
		Description:     in.Description,
		LegalBasis:      in.LegalBasis,
		RetentionDays:   in.RetentionDays,
		IsActive:        true,
		RequiresConsent: in.RequiresConsent,
	}
	p.TenantID = in.TenantID

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create purpose: %w", err)
	}

	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(
		eventbus.EventPolicyCreated, "governance", in.TenantID,
		map[string]any{"id": p.ID, "code": p.Code, "name": p.Name},
	))

	s.logger.InfoContext(ctx, "purpose created", "id", p.ID, "code", p.Code)
	return p, nil
}

// GetByID retrieves a purpose by ID.
func (s *PurposeService) GetByID(ctx context.Context, id types.ID) (*governance.Purpose, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByTenant retrieves all purposes for a tenant.
func (s *PurposeService) ListByTenant(ctx context.Context, tenantID types.ID) ([]governance.Purpose, error) {
	return s.repo.GetByTenant(ctx, tenantID)
}

// UpdatePurposeInput holds updatable fields.
type UpdatePurposeInput struct {
	ID              types.ID
	Name            string
	Description     string
	LegalBasis      types.LegalBasis
	RetentionDays   int
	IsActive        *bool
	RequiresConsent *bool
}

// Update modifies an existing purpose.
func (s *PurposeService) Update(ctx context.Context, in UpdatePurposeInput) (*governance.Purpose, error) {
	p, err := s.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if in.Name != "" {
		p.Name = in.Name
	}
	if in.Description != "" {
		p.Description = in.Description
	}
	if in.LegalBasis != "" {
		p.LegalBasis = in.LegalBasis
	}
	if in.RetentionDays > 0 {
		p.RetentionDays = in.RetentionDays
	}
	if in.IsActive != nil {
		p.IsActive = *in.IsActive
	}
	if in.RequiresConsent != nil {
		p.RequiresConsent = *in.RequiresConsent
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(
		eventbus.EventPolicyUpdated, "governance", p.TenantID,
		map[string]any{"id": p.ID, "name": p.Name},
	))

	return p, nil
}

// Delete removes a purpose.
func (s *PurposeService) Delete(ctx context.Context, id types.ID) error {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.eventBus.Publish(ctx, eventbus.NewEvent(
		eventbus.EventPolicyDeleted, "governance", p.TenantID,
		map[string]any{"id": id, "code": p.Code},
	))

	s.logger.InfoContext(ctx, "purpose deleted", "id", id, "code", p.Code)
	return nil
}
