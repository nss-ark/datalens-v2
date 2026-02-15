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

// AppealDPR creates an appeal for a rejected or completed DPR.
// DPDPA S18: Right to appeal to the Board (simulated here as internal DPO review first).
func (s *DataPrincipalService) AppealDPR(ctx context.Context, principalID, originalDPRID types.ID, reason string) (*consent.DPRRequest, error) {
	// 1. Get original DPR
	original, err := s.GetDPR(ctx, principalID, originalDPRID)
	if err != nil {
		return nil, err
	}

	// 2. Validate status
	// Can only appeal if Rejected or Completed (e.g. partial fulfillment)
	if original.Status != consent.DPRStatusRejected && original.Status != consent.DPRStatusCompleted {
		return nil, types.NewValidationError("cannot appeal request in current status", map[string]any{"status": original.Status})
	}

	// Check if already appealed
	existing, _ := s.GetAppeal(ctx, principalID, originalDPRID)
	if existing != nil {
		return nil, types.NewConflictError("DPR Appeal", originalDPRID.String(), "already exists")
	}

	// 3. Create Appeal DPR
	appeal := &consent.DPRRequest{
		BaseEntity: types.BaseEntity{
			ID:        types.NewID(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID:     original.TenantID,
		ProfileID:    principalID,
		Type:         original.Type, // Keep original type to preserve context
		Status:       consent.DPRStatusAppealed,
		SubmittedAt:  time.Now().UTC(),
		IsEscalated:  true,
		AppealOf:     &original.ID,
		AppealReason: &reason,
	}

	if err := s.dprRepo.Create(ctx, appeal); err != nil {
		return nil, fmt.Errorf("create appeal dpr: %w", err)
	}

	// 4. Create internal DSR for tracking the appeal
	dsr := &compliance.DSR{
		ID:          types.NewID(),
		TenantID:    appeal.TenantID,
		RequestType: compliance.RequestTypeAppeal,
		Status:      compliance.DSRStatusPending, // Appeals start as Pending review
		SubjectName: "Appeal for Request " + original.ID.String(),
		Priority:    "HIGH",
		Notes:       fmt.Sprintf("Appeal Reason: %s\nOriginal Request ID: %s", reason, original.ID),
		SLADeadline: time.Now().AddDate(0, 0, 30), // Standard SLA
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Metadata:    types.Metadata{"original_dpr_id": original.ID.String(), "appeal_dpr_id": appeal.ID.String()},
	}

	// We need subject details from profile
	profile, _ := s.profileRepo.GetByID(ctx, principalID)
	if profile != nil {
		dsr.SubjectName = profile.Email // Profile doesn't have Name, use Email
		dsr.SubjectEmail = profile.Email
		// Identifiers...
	}

	if err := s.dsrRepo.Create(ctx, dsr); err != nil {
		s.logger.Error("failed to create appeal DSR", "error", err)
	} else {
		// Link DPR to DSR
		appeal.DSRID = &dsr.ID
		s.dprRepo.Update(ctx, appeal)
	}

	return appeal, nil
}

// GetAppeal retrieves the appeal for a given original DPR, if it exists.
func (s *DataPrincipalService) GetAppeal(ctx context.Context, principalID, originalDPRID types.ID) (*consent.DPRRequest, error) {
	// Iterate to find appeal (naive but functional given low volume per user)
	dprs, err := s.dprRepo.GetByProfile(ctx, principalID)
	if err != nil {
		return nil, err
	}

	for _, d := range dprs {
		if d.AppealOf != nil && *d.AppealOf == originalDPRID {
			return &d, nil
		}
	}
	return nil, nil // Not found
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

// DPRDownloadResult is the response for downloading the result of a completed ACCESS DPR.
type DPRDownloadResult struct {
	DPRRequestID types.ID    `json:"dpr_request_id"`
	RequestType  string      `json:"request_type"`
	CompletedAt  *time.Time  `json:"completed_at"`
	Summary      string      `json:"summary,omitempty"`
	PersonalData interface{} `json:"personal_data"` // Aggregated task results from the DSR execution
}

// DownloadDPRData retrieves the completed result of an ACCESS-type DPR request.
// DPDPA S11(1): Data Principal has the right to obtain a summary of personal data.
func (s *DataPrincipalService) DownloadDPRData(ctx context.Context, principalID, dprID types.ID) (*DPRDownloadResult, error) {
	// GetDPR already validates ownership (ProfileID == principalID) and returns 403 on mismatch
	dpr, err := s.GetDPR(ctx, principalID, dprID)
	if err != nil {
		return nil, err
	}

	// Only allow download for completed requests
	if dpr.Status != consent.DPRStatusCompleted {
		return nil, types.NewForbiddenError("request is not yet completed — current status: " + string(dpr.Status))
	}

	result := &DPRDownloadResult{
		DPRRequestID: dpr.ID,
		RequestType:  dpr.Type,
		CompletedAt:  dpr.CompletedAt,
	}
	if dpr.ResponseSummary != nil {
		result.Summary = *dpr.ResponseSummary
	}

	// Compile personal data from the linked DSR's task results
	if dpr.DSRID != nil {
		tasks, err := s.dsrRepo.GetTasksByDSR(ctx, *dpr.DSRID)
		if err != nil {
			return nil, fmt.Errorf("fetch dsr tasks: %w", err)
		}

		taskResults := make([]map[string]interface{}, 0, len(tasks))
		for _, task := range tasks {
			if task.Result != nil {
				taskResults = append(taskResults, map[string]interface{}{
					"task_id":        task.ID,
					"data_source_id": task.DataSourceID,
					"status":         task.Status,
					"result":         task.Result,
					"completed_at":   task.CompletedAt,
				})
			}
		}
		result.PersonalData = taskResults
	} else {
		// No linked DSR — return empty data with note
		result.PersonalData = map[string]string{
			"note": "No linked data subject request found. The response summary contains all available information.",
		}
	}

	return result, nil
}
