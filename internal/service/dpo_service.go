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

// DPOService handles business logic for Data Protection Officer contacts.
type DPOService struct {
	repo     compliance.DPOContactRepository
	eventBus eventbus.EventBus
	logger   *slog.Logger
}

// NewDPOService creates a new DPOService.
func NewDPOService(repo compliance.DPOContactRepository, eventBus eventbus.EventBus, logger *slog.Logger) *DPOService {
	return &DPOService{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// UpsertDPOContactRequest defines the payload for creating/updating a DPO contact.
type UpsertDPOContactRequest struct {
	OrgName    string  `json:"org_name"`
	DPOName    string  `json:"dpo_name"`
	DPOEmail   string  `json:"dpo_email"`
	DPOPhone   *string `json:"dpo_phone"`
	Address    *string `json:"address"`
	WebsiteURL *string `json:"website_url"`
}

// Validate checks required fields.
func (r UpsertDPOContactRequest) Validate() error {
	if r.OrgName == "" {
		return types.NewValidationError("organization name is required", map[string]any{"field": "org_name"})
	}
	if r.DPOName == "" {
		return types.NewValidationError("DPO name is required", map[string]any{"field": "dpo_name"})
	}
	if r.DPOEmail == "" {
		return types.NewValidationError("DPO email is required", map[string]any{"field": "dpo_email"})
	}
	return nil
}

// UpsertContact creates or updates the DPO contact for the authenticated tenant.
func (s *DPOService) UpsertContact(ctx context.Context, req UpsertDPOContactRequest) (*compliance.DPOContact, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	contact := &compliance.DPOContact{
		TenantEntity: types.TenantEntity{
			TenantID: tenantID,
		},
		OrgName:    req.OrgName,
		DPOName:    req.DPOName,
		DPOEmail:   req.DPOEmail,
		DPOPhone:   req.DPOPhone,
		Address:    req.Address,
		WebsiteURL: req.WebsiteURL,
		UpdatedAt:  time.Now().UTC(),
	}

	if err := s.repo.Upsert(ctx, contact); err != nil {
		return nil, fmt.Errorf("upsert contact: %w", err)
	}

	// Publish event
	event := eventbus.NewEvent("compliance.dpo_contact_updated", "compliance", tenantID, contact)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("failed to publish dpo_contact_updated event", "error", err)
		// Don't fail the request if event publishing fails
	}

	return contact, nil
}

// GetContact retrieves the DPO contact for the authenticated tenant.
func (s *DPOService) GetContact(ctx context.Context) (*compliance.DPOContact, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.Get(ctx, tenantID)
}

// GetPublicContact retrieves the DPO contact for a specific tenant (public access).
func (s *DPOService) GetPublicContact(ctx context.Context, tenantID types.ID) (*compliance.DPOContact, error) {
	return s.repo.Get(ctx, tenantID)
}
