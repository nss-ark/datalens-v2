package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/eventbus"
	"github.com/complyark/datalens/pkg/types"
)

// DataPrincipalService handles portal-side data subject operations.
type DataPrincipalService struct {
	profileRepo consent.DataPrincipalProfileRepository
	dprRepo     consent.DPRRequestRepository
	dsrRepo     compliance.DSRRepository
	historyRepo consent.ConsentHistoryRepository
	eventBus    eventbus.EventBus
	logger      *slog.Logger
}

// NewDataPrincipalService creates a new DataPrincipalService.
func NewDataPrincipalService(
	profileRepo consent.DataPrincipalProfileRepository,
	dprRepo consent.DPRRequestRepository,
	dsrRepo compliance.DSRRepository,
	historyRepo consent.ConsentHistoryRepository,
	eventBus eventbus.EventBus,
	logger *slog.Logger,
) *DataPrincipalService {
	return &DataPrincipalService{
		profileRepo: profileRepo,
		dprRepo:     dprRepo,
		dsrRepo:     dsrRepo,
		historyRepo: historyRepo,
		eventBus:    eventBus,
		logger:      logger.With("service", "data_principal"),
	}
}

// GetProfile retrieves the profile for the authenticated principal.
func (s *DataPrincipalService) GetProfile(ctx context.Context, id types.ID) (*consent.DataPrincipalProfile, error) {
	return s.profileRepo.GetByID(ctx, id)
}

// GetConsentHistory retrieves the consent history for the authenticated principal.
func (s *DataPrincipalService) GetConsentHistory(ctx context.Context, principalID types.ID, pagination types.Pagination) (*types.PaginatedResult[consent.ConsentHistoryEntry], error) {
	profile, err := s.profileRepo.GetByID(ctx, principalID)
	if err != nil {
		return nil, err
	}

	// If profile has no linked subject ID yet, return empty
	if profile.SubjectID == nil {
		return &types.PaginatedResult[consent.ConsentHistoryEntry]{
			Items:      []consent.ConsentHistoryEntry{},
			Total:      0,
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalPages: 0,
		}, nil
	}

	return s.historyRepo.GetBySubject(ctx, profile.TenantID, *profile.SubjectID, pagination)
}

// CreateDPRRequestInput holds fields for creating a DPR.
type CreateDPRRequestInput struct {
	Type        string `json:"type"`        // ACCESS, ERASURE, etc.
	Description string `json:"description"` // Optional details
}

// SubmitDPR creates a new DPRRequest and logs it.
// CRITICAL: It also creates an internal DSR linked to this request.
func (s *DataPrincipalService) SubmitDPR(ctx context.Context, principalID types.ID, input CreateDPRRequestInput) (*consent.DPRRequest, error) {
	profile, err := s.profileRepo.GetByID(ctx, principalID)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	now := time.Now().UTC()

	// 1. Create Portal DPR implementation
	dpr := &consent.DPRRequest{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		TenantID:    profile.TenantID,
		ProfileID:   profile.ID,
		Type:        input.Type,
		Description: input.Description,
		Status:      consent.DPRStatusSubmitted, // Start as submitted
		SubmittedAt: now,
		VerifiedAt:  profile.VerifiedAt, // Profile is already verified
	}

	// 2. Create Internal Compliance DSR
	// If profile has a SubjectID, use it. If not, we might need resolution logic.
	// For now, if SubjectID is nil, we create DSR with empty SubjectID but with metadata to resolve it.
	dsr := &compliance.DSR{
		ID:          types.NewID(),
		CreatedAt:   now,
		UpdatedAt:   now,
		TenantID:    profile.TenantID,
		RequestType: compliance.DSRRequestType(input.Type),
		Status:      compliance.DSRStatusPending,
		// Source:      compliance.DSRSourcePortal, // Source field missing in DSR struct, using Notes or context
		// Deadline: not in DSR struct based on file view
		// RequestDetails not in DSR struct
		SubjectEmail: profile.Email,
		Notes:        fmt.Sprintf("Portal Request ID: %s. Description: %s", dpr.ID, input.Description),
	}

	// Link DSR to DPR
	dpr.DSRID = &dsr.ID

	// Transactional save (ideally) - executing sequentially for now
	if err := s.dsrRepo.Create(ctx, dsr); err != nil {
		return nil, fmt.Errorf("create internal dsr: %w", err)
	}

	if err := s.dprRepo.Create(ctx, dpr); err != nil {
		// Compensating action: delete DSR? Or just fail.
		return nil, fmt.Errorf("create portal dpr: %w", err)
	}

	// Publish Event
	event := eventbus.NewEvent("dpr.submitted", "portal", profile.TenantID, dpr)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.ErrorContext(ctx, "failed to publish dpr event", "error", err)
	}

	// If DSR was created, publish DSR event too
	if dsr != nil {
		dsrEvent := eventbus.NewEvent("dsr.created", "portal", profile.TenantID, dsr)
		if err := s.eventBus.Publish(ctx, dsrEvent); err != nil {
			s.logger.ErrorContext(ctx, "failed to publish dsr event", "error", err)
		}
	}

	return dpr, nil
}

// ListDPRs retrieves all requests for the principal.
func (s *DataPrincipalService) ListDPRs(ctx context.Context, principalID types.ID) ([]consent.DPRRequest, error) {
	return s.dprRepo.GetByProfile(ctx, principalID)
}

// GetDPR retrieves a specific request.
func (s *DataPrincipalService) GetDPR(ctx context.Context, principalID, dprID types.ID) (*consent.DPRRequest, error) {
	dpr, err := s.dprRepo.GetByID(ctx, dprID)
	if err != nil {
		return nil, err
	}

	if dpr.ProfileID != principalID {
		return nil, types.NewForbiddenError("access denied to this request")
	}

	return dpr, nil
}
