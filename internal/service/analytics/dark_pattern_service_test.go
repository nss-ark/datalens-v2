package analytics

import (
	"context"
	"testing"

	"github.com/complyark/datalens/internal/service/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGateway is a mock implementation of ai.Gateway
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ai.PurposeSuggestion), args.Error(1)
}

func (m *MockGateway) Complete(ctx context.Context, prompt string, opts ai.CompletionOptions) (*ai.CompletionResult, error) {
	args := m.Called(ctx, prompt, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ai.CompletionResult), args.Error(1)
}

func TestDarkPatternService_AnalyzeContent(t *testing.T) {
	// Setup
	mockGateway := new(MockGateway)
	service := NewDarkPatternService(mockGateway)
	ctx := context.Background()

	// Test Case 1: Detect "False Urgency"
	t.Run("Detect False Urgency", func(t *testing.T) {
		input := "Only 2 left! Buy now!"
		expectedResponse := `{
			"detected_patterns": ["FALSE_URGENCY"],
			"confidence": 0.95,
			"explanation": "Create false sense of urgency",
			"cited_clause": "Annexure 1(1) False Urgency"
		}`

		mockGateway.On("Complete", ctx, mock.Anything, mock.MatchedBy(func(opts ai.CompletionOptions) bool {
			return opts.UseCase == "dark_pattern_detection"
		})).Return(&ai.CompletionResult{
			Response: expectedResponse,
		}, nil).Once()

		result, err := service.AnalyzeContent(ctx, "TEXT", input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result.DetectedPatterns, DarkPatternFalseUrgency)
		assert.Equal(t, 0.95, result.Confidence)
	})

	// Test Case 2: Detect "Confirm Shaming"
	t.Run("Detect Confirm Shaming", func(t *testing.T) {
		input := "No, I like paying full price."
		expectedResponse := `{
			"detected_patterns": ["CONFIRM_SHAMING"],
			"confidence": 0.90,
			"explanation": "Guilts user into compliance",
			"cited_clause": "Annexure 1(3) Confirm Shaming"
		}`

		mockGateway.On("Complete", ctx, mock.Anything, mock.MatchedBy(func(opts ai.CompletionOptions) bool {
			return opts.UseCase == "dark_pattern_detection"
		})).Return(&ai.CompletionResult{
			Response: expectedResponse,
		}, nil).Once()

		result, err := service.AnalyzeContent(ctx, "TEXT", input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result.DetectedPatterns, DarkPatternConfirmShaming)
	})

	// Test Case 3: No Patterns
	t.Run("No Patterns", func(t *testing.T) {
		input := "Click here to sign up."
		expectedResponse := `{
			"detected_patterns": [],
			"confidence": 0.10,
			"explanation": "No dark patterns detected",
			"cited_clause": ""
		}`

		mockGateway.On("Complete", ctx, mock.Anything, mock.MatchedBy(func(opts ai.CompletionOptions) bool {
			return opts.UseCase == "dark_pattern_detection"
		})).Return(&ai.CompletionResult{
			Response: expectedResponse,
		}, nil).Once()

		result, err := service.AnalyzeContent(ctx, "TEXT", input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.DetectedPatterns)
	})

	// Test Case 4: JSON Parsing Error correction (Markdown strip)
	t.Run("Handle Markdown JSON", func(t *testing.T) {
		input := "Test markdown"
		expectedResponse := "```json\n{\n  \"detected_patterns\": [\"NAGGING\"],\n  \"confidence\": 0.8\n}\n```"

		mockGateway.On("Complete", ctx, mock.Anything, mock.MatchedBy(func(opts ai.CompletionOptions) bool {
			return opts.UseCase == "dark_pattern_detection"
		})).Return(&ai.CompletionResult{
			Response: expectedResponse,
		}, nil).Once()

		result, err := service.AnalyzeContent(ctx, "TEXT", input)
		assert.NoError(t, err)
		assert.Contains(t, result.DetectedPatterns, DarkPatternNagging)
	})
}
