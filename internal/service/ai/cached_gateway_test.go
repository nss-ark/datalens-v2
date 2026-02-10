package ai

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/pkg/types"
)

// MockGateway
type MockGateway struct {
	mock.Mock
}

func (m *MockGateway) DetectPII(ctx context.Context, input PIIDetectionInput) (*PIIDetectionResult, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PIIDetectionResult), args.Error(1)
}

func (m *MockGateway) SuggestPurposes(ctx context.Context, input PurposeSuggestionInput) ([]PurposeSuggestion, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PurposeSuggestion), args.Error(1)
}

func (m *MockGateway) Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error) {
	args := m.Called(ctx, prompt, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CompletionResult), args.Error(1)
}

// Name implements Provider interface (though Gateway doesn't have Name, Provider does. Gateway has Complete/Detect/Suggest)
// The MockGateway implements Gateway interface.
func (m *MockGateway) Name() string {
	return "mock"
}

func setupTest(t *testing.T) (*CachedGateway, *MockGateway, *miniredis.Miniredis) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	mockNext := new(MockGateway)
	logger := slog.Default() // Use default logger for tests

	// Dummy cfg
	cfg := config.AIConfig{}

	gateway := NewCachedGateway(mockNext, rdb, logger, cfg)
	return gateway, mockNext, s
}

func TestCachedGateway_DetectPII_CacheMiss(t *testing.T) {
	gateway, mockNext, s := setupTest(t)
	defer s.Close()

	ctx := context.Background()
	// Needs tenant context for usage tracking
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	input := PIIDetectionInput{
		SanitizedSamples: []string{"My name is John Doe"},
		TableName:        "users",
		ColumnName:       "bio",
		DataType:         "text",
	}
	expectedParams := PIIDetectionResult{
		IsPII:      true,
		TokensUsed: 10,
		Provider:   "openai",
	}

	// Expect call to next
	mockNext.On("DetectPII", ctx, input).Return(&expectedParams, nil)

	// 1. First Call (Miss)
	res, err := gateway.DetectPII(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, "openai", res.Provider)

	mockNext.AssertExpectations(t)

	// Verify it's in Redis
	// Key calculation is hidden, but we can verify usage key
	// Usage key: tenant:{id}:ai:tokens:{date}
	today := time.Now().Format("20060102")
	usageKey := "tenant:" + tenantID.String() + ":ai:tokens:" + today
	val, err := s.Get(usageKey)
	require.NoError(t, err)
	assert.Equal(t, "10", val)
}

func TestCachedGateway_DetectPII_CacheHit(t *testing.T) {
	gateway, mockNext, s := setupTest(t)
	defer s.Close()

	ctx := context.Background()
	input := PIIDetectionInput{
		SanitizedSamples: []string{"Cached Text"},
		TableName:        "users",
		ColumnName:       "bio_cached",
		DataType:         "text",
	}

	expected := &PIIDetectionResult{
		IsPII:      true,
		TokensUsed: 50,
		Provider:   "openai",
	}
	mockNext.On("DetectPII", mock.Anything, input).Return(expected, nil).Once()

	// 1. Prime Cache
	_, err := gateway.DetectPII(ctx, input)
	require.NoError(t, err)

	// 2. Second Call (Hit)
	res, err := gateway.DetectPII(ctx, input)
	require.NoError(t, err)

	// Should be cached status
	assert.Contains(t, res.Provider, "(cached)")
	assert.Equal(t, time.Duration(0), res.Duration)

	mockNext.AssertExpectations(t) // Should only be called once
}

func TestCachedGateway_BudgetExceeded(t *testing.T) {
	gateway, _, s := setupTest(t)
	defer s.Close()

	ctx := context.Background()
	tenantID := types.NewID()
	ctx = context.WithValue(ctx, types.ContextKeyTenantID, tenantID)

	// Set Budget: 100
	s.Set("tenant:"+tenantID.String()+":ai:budget", "100")

	// Set Usage: 101
	today := time.Now().Format("20060102")
	s.Set("tenant:"+tenantID.String()+":ai:tokens:"+today, "101")

	input := PIIDetectionInput{SanitizedSamples: []string{"Cheap request"}}

	_, err := gateway.DetectPII(ctx, input)
	assert.Error(t, err)

	// Check error type
	var domainErr *types.DomainError
	isDomainErr := errors.As(err, &domainErr)
	assert.True(t, isDomainErr)
	assert.Equal(t, types.ErrCodeQuotaExceeded, domainErr.Code)
}

func TestCachedGateway_RedisFailure_FailOpen(t *testing.T) {
	gateway, mockNext, s := setupTest(t)

	// Close Redis to force error
	s.Close()

	ctx := context.Background()
	input := PIIDetectionInput{SanitizedSamples: []string{"Redis is down"}}
	expected := &PIIDetectionResult{IsPII: false}

	// Should still call next
	mockNext.On("DetectPII", ctx, input).Return(expected, nil)

	res, err := gateway.DetectPII(ctx, input)

	// Expect NO error (fail open logic)
	require.NoError(t, err)
	assert.Equal(t, expected, res)
	mockNext.AssertExpectations(t)
}
