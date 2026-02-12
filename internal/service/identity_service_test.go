package service

import (
	"context"
	"testing"

	"log/slog"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider
type MockIdentityProvider struct {
	mock.Mock
}

func (m *MockIdentityProvider) Name() string {
	return "DigiLocker"
}

func (m *MockIdentityProvider) GetAuthorizationURL(state string) string {
	return "http://mock-auth-url"
}

func (m *MockIdentityProvider) ExchangeToken(ctx context.Context, code string) (*identity.TokenResponse, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.TokenResponse), args.Error(1)
}

func (m *MockIdentityProvider) GetUserProfile(ctx context.Context, token string) (*identity.UserProfile, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.UserProfile), args.Error(1)
}

func (m *MockIdentityProvider) FetchDocuments(ctx context.Context, token string) ([]identity.IdentityDocument, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]identity.IdentityDocument), args.Error(1)
}

// MockRepo
type MockIdentityProfileRepo struct {
	mock.Mock
}

func (m *MockIdentityProfileRepo) Create(ctx context.Context, profile *identity.IdentityProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockIdentityProfileRepo) GetBySubject(ctx context.Context, tenantID, subjectID types.ID) (*identity.IdentityProfile, error) {
	args := m.Called(ctx, tenantID, subjectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.IdentityProfile), args.Error(1)
}

func (m *MockIdentityProfileRepo) Update(ctx context.Context, profile *identity.IdentityProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func TestIdentityService_LinkProvider_UpgradeIAL(t *testing.T) {
	// Setup
	mockRepo := new(MockIdentityProfileRepo)
	mockProvider := new(MockIdentityProvider)
	logger := slog.Default()

	svc := NewIdentityService(mockRepo, []identity.IdentityProvider{mockProvider}, logger)

	ctx := context.Background()
	tenantID := types.NewID()
	subjectID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	// Scenarios
	mockProvider.On("ExchangeToken", ctx, "auth-code").Return(&identity.TokenResponse{AccessToken: "token"}, nil)
	mockProvider.On("GetUserProfile", ctx, "token").Return(&identity.UserProfile{Name: "Test User"}, nil)

	// Mock returning documents (Aadhaar)
	docs := []identity.IdentityDocument{
		{Type: identity.DocumentTypeAadhaar, ReferenceID: "1234"},
	}
	mockProvider.On("FetchDocuments", ctx, "token").Return(docs, nil)

	// Mock existing profile (Not Found -> New)
	mockRepo.On("GetBySubject", ctx, tenantID, subjectID).Return(nil, types.NewNotFoundError("profile", subjectID))

	// Expect Create with Substantial Assurance
	mockRepo.On("Create", ctx, mock.MatchedBy(func(p *identity.IdentityProfile) bool {
		return p.AssuranceLevel == identity.AssuranceLevelSubstantial &&
			p.SubjectID == subjectID &&
			len(p.Documents) == 1
	})).Return(nil)

	// Execute
	profile, err := svc.LinkProvider(ctx, subjectID, "DigiLocker", "auth-code")

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, identity.AssuranceLevelSubstantial, profile.AssuranceLevel)
	mockRepo.AssertExpectations(t)
	mockProvider.AssertExpectations(t)
}
