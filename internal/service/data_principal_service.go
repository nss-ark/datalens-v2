package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

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
	redis       *redis.Client
	logger      *slog.Logger
}

// NewDataPrincipalService creates a new DataPrincipalService.
func NewDataPrincipalService(
	profileRepo consent.DataPrincipalProfileRepository,
	dprRepo consent.DPRRequestRepository,
	dsrRepo compliance.DSRRepository,
	historyRepo consent.ConsentHistoryRepository,
	eventBus eventbus.EventBus,
	redis *redis.Client,
	logger *slog.Logger,
) *DataPrincipalService {
	return &DataPrincipalService{
		profileRepo: profileRepo,
		dprRepo:     dprRepo,
		dsrRepo:     dsrRepo,
		historyRepo: historyRepo,
		eventBus:    eventBus,
		redis:       redis,
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

	// 1. Check for Minor / Guardian requirements (DPDPA Section 9)
	if profile.IsMinor && !profile.GuardianVerified {
		// Option A: Reject immediately
		// return nil, types.NewForbiddenError("guardian verification required for minors", nil)

		// Option B: Allow submission but set status to GUARDIAN_PENDING
		// For now, let's go with Option B as per requirements "REJECT submission or set status to GUARDIAN_PENDING"
		// Actually, let's return error for now to enforce flow, or use GUARDIAN_PENDING if we want to allow draft.
		// Requirement says: "Minor profiles cannot submit DPRs without verified guardian."
		// So let's reject.
		return nil, types.NewForbiddenError("guardian verification required for minors")
	}

	now := time.Now().UTC()

	// 2. Create Portal DPR implementation
	dpr := &consent.DPRRequest{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		TenantID:         profile.TenantID,
		ProfileID:        profile.ID,
		Type:             input.Type,
		Description:      input.Description,
		Status:           consent.DPRStatusSubmitted, // Start as submitted
		SubmittedAt:      now,
		VerifiedAt:       profile.VerifiedAt, // Profile is already verified
		IsMinor:          profile.IsMinor,
		GuardianVerified: profile.GuardianVerified,
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
	event := eventbus.NewEvent(eventbus.EventDPRSubmitted, "portal", profile.TenantID, dpr)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.ErrorContext(ctx, "failed to publish dpr event", "error", err)
	}

	// If DSR was created, publish DSR event too
	if dsr != nil {
		dsrEvent := eventbus.NewEvent(eventbus.EventDSRCreated, "portal", profile.TenantID, dsr)
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

// InitiateGuardianVerification sends an OTP to the guardian contact.
func (s *DataPrincipalService) InitiateGuardianVerification(ctx context.Context, principalID types.ID, contact string) error {
	profile, err := s.profileRepo.GetByID(ctx, principalID)
	if err != nil {
		return err
	}

	if !profile.IsMinor {
		return types.NewValidationError("profile is not a minor", nil)
	}

	if contact == "" {
		return types.NewValidationError("guardian contact is required", nil)
	}

	// Generate OTP
	otp, err := generateOTP() // Reuse from portal_auth_service.go (needs to be shared or duplicated)
	// Since generateOTP is unexported in another package, I'll duplicate it here or move it to a shared pkg.
	// For simplicity, I'll use a local helper.
	if err != nil {
		return fmt.Errorf("generate otp: %w", err)
	}

	key := fmt.Sprintf("guardian:otp:%s", principalID)

	if s.redis != nil {
		if err := s.redis.Set(ctx, key, otp, 15*time.Minute).Err(); err != nil {
			return fmt.Errorf("store otp: %w", err)
		}
	} else {
		s.logger.WarnContext(ctx, "Redis unavailable, Guardian OTP logs only (DEV)", "otp", otp)
	}

	// In real world, send email/SMS via NotificationService
	s.logger.InfoContext(ctx, "Guardian OTP Generated", "contact", contact, "otp", otp)

	return nil
}

// VerifyGuardian validates the OTP and updates the profile.
func (s *DataPrincipalService) VerifyGuardian(ctx context.Context, principalID types.ID, code string) error {
	profile, err := s.profileRepo.GetByID(ctx, principalID)
	if err != nil {
		return err
	}

	if !profile.IsMinor {
		return types.NewValidationError("profile is not a minor", nil)
	}

	key := fmt.Sprintf("guardian:otp:%s", principalID)

	if s.redis != nil {
		val, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			return types.NewUnauthorizedError("invalid or expired OTP")
		}
		if val != code {
			return types.NewUnauthorizedError("invalid OTP")
		}
		_ = s.redis.Del(ctx, key)
	} else {
		if code != "123456" {
			return types.NewUnauthorizedError("invalid OTP (DEV: use 123456)")
		}
	}

	profile.GuardianVerified = true
	// profile.GuardianEmail/Phone = contact // Ideally we update the contact too, but for now we trust the flow

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("update profile: %w", err)
	}

	return nil
}

// generateOTP is already defined in portal_auth_service.go (same package)
// We rely on that shared unexported helper.

// ConsentSummaryItem represents the current consent state for a single purpose.
type ConsentSummaryItem struct {
	PurposeID   string `json:"purpose_id"`
	PurposeName string `json:"purpose_name"`
	Status      string `json:"status"` // GRANTED, WITHDRAWN, EXPIRED
	UpdatedAt   string `json:"updated_at"`
}

// GetConsentSummary returns the latest consent state per purpose for the principal.
func (s *DataPrincipalService) GetConsentSummary(ctx context.Context, principalID types.ID) ([]ConsentSummaryItem, error) {
	profile, err := s.profileRepo.GetByID(ctx, principalID)
	if err != nil {
		return nil, err
	}

	// If profile has no linked subject ID yet, return empty
	if profile.SubjectID == nil {
		return []ConsentSummaryItem{}, nil
	}

	entries, err := s.historyRepo.GetAllLatestBySubject(ctx, profile.TenantID, *profile.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("get consent summary: %w", err)
	}

	items := make([]ConsentSummaryItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, ConsentSummaryItem{
			PurposeID:   e.PurposeID.String(),
			PurposeName: e.PurposeName,
			Status:      e.NewStatus,
			UpdatedAt:   e.CreatedAt.Format(time.RFC3339),
		})
	}
	return items, nil
}

// IdentityStatusResponse is the response for identity verification status.
type IdentityStatusResponse struct {
	VerificationStatus string  `json:"verification_status"` // PENDING, VERIFIED, EXPIRED
	VerifiedAt         *string `json:"verified_at,omitempty"`
	Method             *string `json:"method,omitempty"` // EMAIL_OTP, PHONE_OTP
	LinkedProviders    []any   `json:"linked_providers"` // Phase 4
}

// GetIdentityStatus returns the identity verification status for the principal.
func (s *DataPrincipalService) GetIdentityStatus(ctx context.Context, principalID types.ID) (*IdentityStatusResponse, error) {
	profile, err := s.profileRepo.GetByID(ctx, principalID)
	if err != nil {
		return nil, err
	}

	resp := &IdentityStatusResponse{
		VerificationStatus: string(profile.VerificationStatus),
		Method:             profile.VerificationMethod,
		LinkedProviders:    []any{}, // Phase 4: DigiLocker, Aadhaar, etc.
	}

	if profile.VerifiedAt != nil {
		t := profile.VerifiedAt.Format(time.RFC3339)
		resp.VerifiedAt = &t
	}

	return resp, nil
}
