package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// PolicyEnforcer handles Identity Assurance Level (IAL) enforcement.
type PolicyEnforcer struct {
	profileRepo identity.IdentityProfileRepository
	logger      *slog.Logger
}

// NewPolicyEnforcer creates a new PolicyEnforcer.
func NewPolicyEnforcer(profileRepo identity.IdentityProfileRepository, logger *slog.Logger) *PolicyEnforcer {
	return &PolicyEnforcer{
		profileRepo: profileRepo,
		logger:      logger,
	}
}

// CheckAccess verifies if a subject meets the required assurance level.
func (e *PolicyEnforcer) CheckAccess(ctx context.Context, subjectID types.ID, requiredLevel identity.AssuranceLevel) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return types.NewForbiddenError("tenant context required")
	}

	profile, err := e.profileRepo.GetBySubject(ctx, tenantID, subjectID)
	if err != nil {
		if types.IsNotFoundError(err) {
			// Implicitly NONE if profile doesn't exist
			if requiredLevel == identity.AssuranceLevelNone {
				return nil
			}
			return types.NewForbiddenError(fmt.Sprintf("identity verification required: need %s, got NONE", requiredLevel))
		}
		return fmt.Errorf("get identity profile: %w", err)
	}

	if !e.isSufficient(profile.AssuranceLevel, requiredLevel) {
		return types.NewForbiddenError(fmt.Sprintf("insufficient identity assurance: need %s, got %s", requiredLevel, profile.AssuranceLevel))
	}

	return nil
}

// isSufficient checks if actual level >= required level
func (e *PolicyEnforcer) isSufficient(actual, required identity.AssuranceLevel) bool {
	levels := map[identity.AssuranceLevel]int{
		identity.AssuranceLevelNone:        0,
		identity.AssuranceLevelBasic:       1,
		identity.AssuranceLevelSubstantial: 2,
		identity.AssuranceLevelHigh:        3,
	}

	// Handle case where level might be invalid string by determining 0 default
	return levels[actual] >= levels[required]
}
