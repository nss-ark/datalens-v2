package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConsentCache
type MockConsentCache struct {
	mock.Mock
}

func (m *MockConsentCache) GetConsentStatus(ctx context.Context, tenantID, subjectID, purposeID types.ID) (*bool, error) {
	args := m.Called(ctx, tenantID, subjectID, purposeID)
	// Check if nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Safely cast
	if val, ok := args.Get(0).(bool); ok {
		return &val, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockConsentCache) SetConsentStatus(ctx context.Context, tenantID, subjectID, purposeID types.ID, granted bool, ttl time.Duration) error {
	return m.Called(ctx, tenantID, subjectID, purposeID, granted, ttl).Error(0)
}

func (m *MockConsentCache) InvalidateSubject(ctx context.Context, tenantID, subjectID types.ID) error {
	return m.Called(ctx, tenantID, subjectID).Error(0)
}

func (m *MockConsentCache) InvalidateAll(ctx context.Context, tenantID types.ID) error {
	return m.Called(ctx, tenantID).Error(0)
}

func TestCheckConsent_CacheHit(t *testing.T) {
	// Setup
	svc, _, _, _, _ := newTestConsentService()
	mockCache := new(MockConsentCache)
	svc.cache = mockCache // Inject mock cache

	ctx := context.Background()
	tenantID := types.NewID()
	subjectID := types.NewID()
	purposeID := types.NewID()

	// Expectation: Cache returns true
	mockCache.On("GetConsentStatus", ctx, tenantID, subjectID, purposeID).Return(true, nil)

	// Execute
	granted, err := svc.CheckConsent(ctx, tenantID, subjectID, purposeID)

	// Verify
	assert.NoError(t, err)
	assert.True(t, granted)
	mockCache.AssertExpectations(t)
}

func TestCheckConsent_CacheMiss_DBHit(t *testing.T) {
	// Setup
	svc, _, _, historyRepo, _ := newTestConsentService()
	mockCache := new(MockConsentCache)
	svc.cache = mockCache

	ctx := context.Background()
	tenantID := types.NewID()
	subjectID := types.NewID()
	purposeID := types.NewID()

	// Seed DB with GRANTED status
	err := historyRepo.Create(ctx, &consent.ConsentHistoryEntry{
		BaseEntity: types.BaseEntity{ID: types.NewID(), CreatedAt: time.Now()},
		TenantID:   tenantID,
		SubjectID:  subjectID,
		PurposeID:  purposeID,
		NewStatus:  "GRANTED",
	})
	assert.NoError(t, err)

	// Expectation: Cache returns nil (miss)
	mockCache.On("GetConsentStatus", ctx, tenantID, subjectID, purposeID).Return(nil, nil)

	// Expectation: Cache Set is called with true
	mockCache.On("SetConsentStatus", ctx, tenantID, subjectID, purposeID, true, svc.cacheTTL).Return(nil)

	// Execute
	granted, err := svc.CheckConsent(ctx, tenantID, subjectID, purposeID)

	// Verify
	assert.NoError(t, err)
	assert.True(t, granted)
	mockCache.AssertExpectations(t)
}

func TestCheckConsent_CacheError_FallsThrough(t *testing.T) {
	// Setup
	svc, _, _, historyRepo, _ := newTestConsentService()
	mockCache := new(MockConsentCache)
	svc.cache = mockCache

	ctx := context.Background()
	tenantID := types.NewID()
	subjectID := types.NewID()
	purposeID := types.NewID()

	// Seed DB with WITHDRAWN status
	err := historyRepo.Create(ctx, &consent.ConsentHistoryEntry{
		BaseEntity: types.BaseEntity{ID: types.NewID(), CreatedAt: time.Now()},
		TenantID:   tenantID,
		SubjectID:  subjectID,
		PurposeID:  purposeID,
		NewStatus:  "WITHDRAWN",
	})
	assert.NoError(t, err)

	// Expectation: Cache returns error
	mockCache.On("GetConsentStatus", ctx, tenantID, subjectID, purposeID).Return(nil, errors.New("redis error"))

	// Expectation: Cache Set is called with false (we try to cache the DB result even if Get failed)
	mockCache.On("SetConsentStatus", ctx, tenantID, subjectID, purposeID, false, svc.cacheTTL).Return(nil)

	// Execute
	granted, err := svc.CheckConsent(ctx, tenantID, subjectID, purposeID)

	// Verify
	assert.NoError(t, err)
	assert.False(t, granted)
	mockCache.AssertExpectations(t)
}
