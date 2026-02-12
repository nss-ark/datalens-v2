package service_test

import (
	"context"
	"testing"

	"log/slog"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockIdentityProfileRepository for testing PolicyEnforcer
type MockIdentityProfileRepository struct {
	mock.Mock
}

func (m *MockIdentityProfileRepository) Create(ctx context.Context, profile *identity.IdentityProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockIdentityProfileRepository) GetBySubject(ctx context.Context, tenantID, subjectID types.ID) (*identity.IdentityProfile, error) {
	args := m.Called(ctx, tenantID, subjectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.IdentityProfile), args.Error(1)
}

func (m *MockIdentityProfileRepository) Update(ctx context.Context, profile *identity.IdentityProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func TestPolicyEnforcer_CheckAccess(t *testing.T) {
	repo := new(MockIdentityProfileRepository)
	logger := slog.Default()
	enforcer := service.NewPolicyEnforcer(repo, logger)

	tenantID := types.NewID()
	subjectID := types.NewID()
	ctx := context.WithValue(context.Background(), types.ContextKeyTenantID, tenantID)

	// Profiles
	basicProfile := &identity.IdentityProfile{
		TenantEntity:   types.TenantEntity{BaseEntity: types.BaseEntity{ID: types.NewID()}, TenantID: tenantID},
		SubjectID:      subjectID,
		AssuranceLevel: identity.AssuranceLevelBasic,
	}

	verifiedProfile := &identity.IdentityProfile{
		TenantEntity:   types.TenantEntity{BaseEntity: types.BaseEntity{ID: types.NewID()}, TenantID: tenantID},
		SubjectID:      subjectID,
		AssuranceLevel: identity.AssuranceLevelSubstantial,
	}

	// Scenarios

	t.Run("Scenario A: Policy=Strict (Substantial), User=Basic -> Error", func(t *testing.T) {
		// Mock repo to return Basic profile
		repo.On("GetBySubject", ctx, tenantID, subjectID).Return(basicProfile, nil).Once()

		err := enforcer.CheckAccess(ctx, subjectID, identity.AssuranceLevelSubstantial)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient identity assurance")
		assert.Contains(t, err.Error(), "got BASIC")
	})

	t.Run("Scenario B: Policy=Flexible (Basic), User=Basic -> Success", func(t *testing.T) {
		// Mock repo to return Basic profile
		repo.On("GetBySubject", ctx, tenantID, subjectID).Return(basicProfile, nil).Once()

		err := enforcer.CheckAccess(ctx, subjectID, identity.AssuranceLevelBasic)
		require.NoError(t, err)
	})

	t.Run("Scenario C: User=Verified (Substantial), Policy=Strict -> Success", func(t *testing.T) {
		// Mock repo to return Verified profile
		repo.On("GetBySubject", ctx, tenantID, subjectID).Return(verifiedProfile, nil).Once()

		err := enforcer.CheckAccess(ctx, subjectID, identity.AssuranceLevelSubstantial)
		require.NoError(t, err)
	})

	t.Run("Scenario D: No Profile, Policy=Basic -> Error", func(t *testing.T) {
		// Mock repo to return NotFound
		repo.On("GetBySubject", ctx, tenantID, subjectID).Return(nil, types.NewNotFoundError("profile not found", nil)).Once()

		err := enforcer.CheckAccess(ctx, subjectID, identity.AssuranceLevelBasic)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "need BASIC, got NONE")
	})

	t.Run("Scenario E: No Profile, Policy=None -> Success", func(t *testing.T) {
		// Mock repo to return NotFound
		repo.On("GetBySubject", ctx, tenantID, subjectID).Return(nil, types.NewNotFoundError("profile not found", nil)).Once()

		err := enforcer.CheckAccess(ctx, subjectID, identity.AssuranceLevelNone)
		require.NoError(t, err)
	})
}
