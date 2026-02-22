package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/types"
)

// ThirdPartyService provides business logic for third-party management.
type ThirdPartyService struct {
	repo     governance.ThirdPartyRepository
	auditSvc *AuditService
	logger   *slog.Logger
}

// NewThirdPartyService creates a new ThirdPartyService.
func NewThirdPartyService(repo governance.ThirdPartyRepository, auditSvc *AuditService, logger *slog.Logger) *ThirdPartyService {
	return &ThirdPartyService{
		repo:     repo,
		auditSvc: auditSvc,
		logger:   logger.With("service", "third_party"),
	}
}

// CreateThirdPartyRequest holds input for creating a third party.
type CreateThirdPartyRequest struct {
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Country      string     `json:"country"`
	DPADocPath   *string    `json:"dpa_doc_path"`
	PurposeIDs   []types.ID `json:"purpose_ids"`
	DPAStatus    string     `json:"dpa_status"`
	DPASignedAt  *time.Time `json:"dpa_signed_at"`
	DPAExpiresAt *time.Time `json:"dpa_expires_at"`
	DPANotes     string     `json:"dpa_notes"`
	ContactName  string     `json:"contact_name"`
	ContactEmail string     `json:"contact_email"`
}

// UpdateThirdPartyRequest holds input for updating a third party.
type UpdateThirdPartyRequest struct {
	Name         *string    `json:"name,omitempty"`
	Type         *string    `json:"type,omitempty"`
	Country      *string    `json:"country,omitempty"`
	DPADocPath   *string    `json:"dpa_doc_path,omitempty"`
	PurposeIDs   []types.ID `json:"purpose_ids,omitempty"`
	DPAStatus    *string    `json:"dpa_status,omitempty"`
	DPASignedAt  *time.Time `json:"dpa_signed_at,omitempty"`
	DPAExpiresAt *time.Time `json:"dpa_expires_at,omitempty"`
	DPANotes     *string    `json:"dpa_notes,omitempty"`
	ContactName  *string    `json:"contact_name,omitempty"`
	ContactEmail *string    `json:"contact_email,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

// Create creates a new third party for the tenant.
func (s *ThirdPartyService) Create(ctx context.Context, req CreateThirdPartyRequest) (*governance.ThirdParty, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	if req.Name == "" {
		return nil, types.NewValidationError("name is required", nil)
	}

	dpaStatus := req.DPAStatus
	if dpaStatus == "" {
		dpaStatus = governance.DPAStatusNone
	}

	tp := &governance.ThirdParty{
		TenantEntity: types.TenantEntity{
			BaseEntity: types.BaseEntity{},
			TenantID:   tenantID,
		},
		Name:         req.Name,
		Type:         governance.ThirdPartyType(req.Type),
		Country:      req.Country,
		DPADocPath:   req.DPADocPath,
		IsActive:     true,
		PurposeIDs:   req.PurposeIDs,
		DPAStatus:    dpaStatus,
		DPASignedAt:  req.DPASignedAt,
		DPAExpiresAt: req.DPAExpiresAt,
		DPANotes:     req.DPANotes,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
	}

	if err := s.repo.Create(ctx, tp); err != nil {
		return nil, fmt.Errorf("create third party: %w", err)
	}

	// Audit log
	userID, _ := types.UserIDFromContext(ctx)
	s.auditSvc.Log(ctx, userID, "THIRD_PARTY_CREATE", "THIRD_PARTY", tp.ID, nil,
		map[string]any{"name": tp.Name, "type": string(tp.Type)}, tenantID)

	s.logger.Info("third party created",
		slog.String("tenant_id", tenantID.String()),
		slog.String("third_party_id", tp.ID.String()),
	)

	return tp, nil
}

// GetByID retrieves a third party by ID.
func (s *ThirdPartyService) GetByID(ctx context.Context, id types.ID) (*governance.ThirdParty, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all third parties for the tenant.
func (s *ThirdPartyService) List(ctx context.Context) ([]governance.ThirdParty, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}
	return s.repo.GetByTenant(ctx, tenantID)
}

// Update updates an existing third party.
func (s *ThirdPartyService) Update(ctx context.Context, id types.ID, req UpdateThirdPartyRequest) (*governance.ThirdParty, error) {
	tp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldName := tp.Name

	if req.Name != nil {
		tp.Name = *req.Name
	}
	if req.Type != nil {
		tp.Type = governance.ThirdPartyType(*req.Type)
	}
	if req.Country != nil {
		tp.Country = *req.Country
	}
	if req.DPADocPath != nil {
		tp.DPADocPath = req.DPADocPath
	}
	if req.PurposeIDs != nil {
		tp.PurposeIDs = req.PurposeIDs
	}
	if req.DPAStatus != nil {
		tp.DPAStatus = *req.DPAStatus
	}
	if req.DPASignedAt != nil {
		tp.DPASignedAt = req.DPASignedAt
	}
	if req.DPAExpiresAt != nil {
		tp.DPAExpiresAt = req.DPAExpiresAt
	}
	if req.DPANotes != nil {
		tp.DPANotes = *req.DPANotes
	}
	if req.ContactName != nil {
		tp.ContactName = *req.ContactName
	}
	if req.ContactEmail != nil {
		tp.ContactEmail = *req.ContactEmail
	}
	if req.IsActive != nil {
		tp.IsActive = *req.IsActive
	}

	tp.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, tp); err != nil {
		return nil, fmt.Errorf("update third party: %w", err)
	}

	// Audit log
	userID, _ := types.UserIDFromContext(ctx)
	tenantID, _ := types.TenantIDFromContext(ctx)
	s.auditSvc.Log(ctx, userID, "THIRD_PARTY_UPDATE", "THIRD_PARTY", tp.ID,
		map[string]any{"name": oldName},
		map[string]any{"name": tp.Name}, tenantID)

	s.logger.Info("third party updated",
		slog.String("third_party_id", id.String()),
	)

	return tp, nil
}

// Delete removes a third party by ID.
func (s *ThirdPartyService) Delete(ctx context.Context, id types.ID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	userID, _ := types.UserIDFromContext(ctx)
	tenantID, _ := types.TenantIDFromContext(ctx)
	s.auditSvc.Log(ctx, userID, "THIRD_PARTY_DELETE", "THIRD_PARTY", id, nil, nil, tenantID)

	s.logger.Info("third party deleted",
		slog.String("third_party_id", id.String()),
	)

	return nil
}
