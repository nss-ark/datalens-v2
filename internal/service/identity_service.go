package service

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

type IdentityService struct {
	profileRepo identity.IdentityProfileRepository
	providers   map[string]identity.IdentityProvider
	logger      *slog.Logger
}

func NewIdentityService(profileRepo identity.IdentityProfileRepository, providers []identity.IdentityProvider, logger *slog.Logger) *IdentityService {
	providerMap := make(map[string]identity.IdentityProvider)
	for _, p := range providers {
		providerMap[p.Name()] = p
	}

	return &IdentityService{
		profileRepo: profileRepo,
		providers:   providerMap,
		logger:      logger,
	}
}

// GetStatus returns the current IAL and verification status for a subject
func (s *IdentityService) GetStatus(ctx context.Context, subjectID types.ID) (*identity.IdentityProfile, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	profile, err := s.profileRepo.GetBySubject(ctx, tenantID, subjectID)
	if err != nil {
		if types.IsNotFoundError(err) {
			// Return a default empty profile if not found
			return &identity.IdentityProfile{
				TenantEntity: types.TenantEntity{
					BaseEntity: types.BaseEntity{ID: types.NewID()},
					TenantID:   tenantID,
				},
				SubjectID:          subjectID,
				AssuranceLevel:     identity.AssuranceLevelNone,
				VerificationStatus: identity.VerificationStatusPending,
			}, nil
		}
		return nil, fmt.Errorf("get identity profile: %w", err)
	}

	return profile, nil
}

// LinkProvider handles the linking of an external identity provider
func (s *IdentityService) LinkProvider(ctx context.Context, subjectID types.ID, providerName string, authCode string) (*identity.IdentityProfile, error) {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil, types.NewForbiddenError("tenant context required")
	}

	provider, ok := s.providers[providerName]
	if !ok {
		return nil, types.NewNotFoundError("provider not found", map[string]any{"provider": providerName})
	}

	// 1. Exchange auth code for token
	tokenResp, err := provider.ExchangeToken(ctx, authCode)
	if err != nil {
		s.logger.Error("failed to exchange token", "error", err, "provider", providerName)
		return nil, types.NewValidationError("failed to authenticate with provider", nil)
	}

	// 2. Fetch user profile and documents
	userProfile, err := provider.GetUserProfile(ctx, tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("failed to fetch user profile", "error", err, "provider", providerName)
		return nil, fmt.Errorf("fetch user profile: %w", err)
	}
	_ = userProfile // Ignore for now, used only for debug/future

	documents, err := provider.FetchDocuments(ctx, tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("failed to fetch documents", "error", err, "provider", providerName)
		// We might proceed without documents if profile is enough, but for now let's log error
	}

	// 3. Get or create IdentityProfile
	profile, err := s.profileRepo.GetBySubject(ctx, tenantID, subjectID)
	if err != nil {
		if !types.IsNotFoundError(err) {
			return nil, fmt.Errorf("get identity profile: %w", err)
		}
		// Create new profile
		profile = &identity.IdentityProfile{
			TenantEntity: types.TenantEntity{
				BaseEntity: types.BaseEntity{ID: types.NewID()},
				TenantID:   tenantID,
			},
			SubjectID: subjectID,
		}
	}

	// 4. Update Profile with new information
	// Determine Assurance Level based on provider and documents
	newLevel := identity.AssuranceLevelBasic
	if providerName == "DigiLocker" && len(documents) > 0 {
		newLevel = identity.AssuranceLevelSubstantial
	}

	// Update only if level is increasing
	if isHigherAssurance(newLevel, profile.AssuranceLevel) {
		profile.AssuranceLevel = newLevel
	}

	profile.VerificationStatus = identity.VerificationStatusVerified
	now := time.Now()
	profile.LastVerifiedAt = &now

	// Merge documents
	profile.Documents = append(profile.Documents, documents...)

	// Save
	if profile.CreatedAt.IsZero() {
		if err := s.profileRepo.Create(ctx, profile); err != nil {
			return nil, fmt.Errorf("create identity profile: %w", err)
		}
	} else {
		if err := s.profileRepo.Update(ctx, profile); err != nil {
			return nil, fmt.Errorf("update identity profile: %w", err)
		}
	}

	s.logger.Info("identity provider linked", "subject_id", subjectID, "provider", providerName, "assurance_level", profile.AssuranceLevel)

	return profile, nil
}

func isHigherAssurance(newLevel, oldLevel identity.AssuranceLevel) bool {
	levels := map[identity.AssuranceLevel]int{
		identity.AssuranceLevelNone:        0,
		identity.AssuranceLevelBasic:       1,
		identity.AssuranceLevelSubstantial: 2,
		identity.AssuranceLevelHigh:        3,
	}
	return levels[newLevel] > levels[oldLevel]
}
