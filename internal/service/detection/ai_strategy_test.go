package detection_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/service/ai"
	"github.com/complyark/datalens/internal/service/detection"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Mock Gateway
// =============================================================================

type MockGateway struct {
	mock.Mock
}

func (m *MockGateway) DetectPII(ctx context.Context, input ai.PIIDetectionInput) (*ai.PIIDetectionResult, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ai.PIIDetectionResult), args.Error(1)
}

func (m *MockGateway) SuggestPurposes(ctx context.Context, input ai.PurposeSuggestionInput) ([]ai.PurposeSuggestion, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]ai.PurposeSuggestion), args.Error(1)
}

func (m *MockGateway) Complete(ctx context.Context, prompt string, opts ai.CompletionOptions) (*ai.CompletionResult, error) {
	args := m.Called(ctx, prompt, opts)
	return args.Get(0).(*ai.CompletionResult), args.Error(1)
}

// =============================================================================
// Tests
// =============================================================================

func TestAIStrategy_Detect(t *testing.T) {
	ctx := context.Background()

	t.Run("success_pii_found", func(t *testing.T) {
		mockGateway := new(MockGateway)
		strategy := detection.NewAIStrategy(mockGateway, 0.8)

		input := detection.Input{
			TableName:  "users",
			ColumnName: "full_name",
			Samples:    []string{"John Doe", "Jane Smith"},
		}

		// Mock expectations
		mockGateway.On("DetectPII", ctx, mock.MatchedBy(func(arg ai.PIIDetectionInput) bool {
			return arg.ColumnName == "full_name" && len(arg.SanitizedSamples) == 2
		})).Return(&ai.PIIDetectionResult{
			IsPII:       true,
			Category:    types.PIICategoryIdentity,
			Type:        types.PIITypeName,
			Sensitivity: types.SensitivityMedium,
			Confidence:  0.95,
			Reasoning:   "Names identified",
		}, nil)

		results, err := strategy.Detect(ctx, input)
		require.NoError(t, err)
		require.Len(t, results, 1)

		assert.Equal(t, types.PIICategoryIdentity, results[0].Category)
		assert.Equal(t, types.PIITypeName, results[0].Type)
		assert.Equal(t, 0.95, results[0].Confidence)
		assert.Equal(t, types.DetectionMethodAI, results[0].Method)

		mockGateway.AssertExpectations(t)
	})

	t.Run("success_no_pii", func(t *testing.T) {
		mockGateway := new(MockGateway)
		strategy := detection.NewAIStrategy(mockGateway, 0.8)

		input := detection.Input{
			TableName:  "products",
			ColumnName: "sku",
			Samples:    []string{"SKU-123", "SKU-456"},
		}

		mockGateway.On("DetectPII", ctx, mock.Anything).Return(&ai.PIIDetectionResult{
			IsPII: false,
		}, nil)

		results, err := strategy.Detect(ctx, input)
		require.NoError(t, err)
		assert.Empty(t, results)

		mockGateway.AssertExpectations(t)
	})

	t.Run("gateway_error", func(t *testing.T) {
		mockGateway := new(MockGateway)
		strategy := detection.NewAIStrategy(mockGateway, 0.8)

		input := detection.Input{ColumnName: "error_col"}

		mockGateway.On("DetectPII", ctx, mock.Anything).Return(nil, errors.New("api outage"))

		results, err := strategy.Detect(ctx, input)
		assert.Error(t, err)
		assert.Empty(t, results)
		assert.Contains(t, err.Error(), "api outage")
	})

	t.Run("inference_fallback", func(t *testing.T) {
		mockGateway := new(MockGateway)
		strategy := detection.NewAIStrategy(mockGateway, 0.8)

		input := detection.Input{ColumnName: "email"}

		// Gateway returns Type but missing Category/Sensitivity
		mockGateway.On("DetectPII", ctx, mock.Anything).Return(&ai.PIIDetectionResult{
			IsPII:      true,
			Type:       types.PIITypeEmail,
			Confidence: 0.99,
		}, nil)

		results, err := strategy.Detect(ctx, input)
		require.NoError(t, err)
		require.Len(t, results, 1)

		// Strategy should infer these
		assert.Equal(t, types.PIICategoryContact, results[0].Category)
		assert.Equal(t, types.SensitivityMedium, results[0].Sensitivity)
	})
	t.Run("empty_input", func(t *testing.T) {
		mockGateway := new(MockGateway)
		strategy := detection.NewAIStrategy(mockGateway, 0.8)

		input := detection.Input{
			TableName:  "empty_table",
			ColumnName: "empty_col",
			Samples:    []string{},
		}

		mockGateway.On("DetectPII", ctx, mock.Anything).Return(&ai.PIIDetectionResult{
			IsPII: false,
		}, nil)

		results, err := strategy.Detect(ctx, input)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("context_cancellation", func(t *testing.T) {
		mockGateway := new(MockGateway)
		strategy := detection.NewAIStrategy(mockGateway, 0.8)

		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		mockGateway.On("DetectPII", cancelCtx, mock.Anything).Return(nil, context.Canceled)

		_, err := strategy.Detect(cancelCtx, detection.Input{ColumnName: "test"})
		assert.ErrorIs(t, err, context.Canceled)
	})
}
