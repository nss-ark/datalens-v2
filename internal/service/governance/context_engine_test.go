package governance

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/complyark/datalens/internal/domain/governance/templates"
	"github.com/complyark/datalens/internal/service/ai"
)

// Mock AI Gateway
type mockAIGateway struct {
	suggestFunc func(ctx context.Context, input ai.PurposeSuggestionInput) ([]ai.PurposeSuggestion, error)
}

func (m *mockAIGateway) DetectPII(ctx context.Context, input ai.PIIDetectionInput) (*ai.PIIDetectionResult, error) {
	return nil, nil
}

func (m *mockAIGateway) SuggestPurposes(ctx context.Context, input ai.PurposeSuggestionInput) ([]ai.PurposeSuggestion, error) {
	if m.suggestFunc != nil {
		return m.suggestFunc(ctx, input)
	}
	return nil, nil
}

func (m *mockAIGateway) Complete(ctx context.Context, prompt string, opts ai.CompletionOptions) (*ai.CompletionResult, error) {
	return nil, nil
}

func TestContextEngine_SuggestPurposes(t *testing.T) {
	// Initialize real template loader (it uses embedded files, so it's safe for unit tests)
	loader, err := templates.NewLoader()
	if err != nil {
		t.Fatalf("Failed to create template loader: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("Pattern Strategy Matches", func(t *testing.T) {
		// Mock AI that should NOT be called
		mockAI := &mockAIGateway{
			suggestFunc: func(ctx context.Context, input ai.PurposeSuggestionInput) ([]ai.PurposeSuggestion, error) {
				t.Error("AI should not be called when pattern matches")
				return nil, nil
			},
		}

		engine := NewContextEngine(loader, mockAI, logger)

		items := []PurposeSuggestionItem{
			{TableName: "users", ColumnName: "id"},
		}

		suggestions, err := engine.SuggestPurposes(context.Background(), items, true)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(suggestions) != 1 {
			t.Fatalf("Expected 1 suggestion, got %d", len(suggestions))
		}

		if suggestions[0].SuggestedPurposeCode != "IDENTITY" {
			t.Errorf("Expected purpose IDENTITY, got %s", suggestions[0].SuggestedPurposeCode)
		}
		if suggestions[0].Strategy != "PATTERN" {
			t.Errorf("Expected strategy PATTERN, got %s", suggestions[0].Strategy)
		}
	})

	t.Run("AI Strategy Called When Pattern Fails", func(t *testing.T) {
		mockAI := &mockAIGateway{
			suggestFunc: func(ctx context.Context, input ai.PurposeSuggestionInput) ([]ai.PurposeSuggestion, error) {
				return []ai.PurposeSuggestion{
					{
						PurposeCode: "MARKETING",
						Confidence:  0.85,
						Reasoning:   "AI reasoning",
					},
				}, nil
			},
		}

		engine := NewContextEngine(loader, mockAI, logger)

		items := []PurposeSuggestionItem{
			{TableName: "unknown_table", ColumnName: "random_col"},
		}

		suggestions, err := engine.SuggestPurposes(context.Background(), items, true)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(suggestions) != 1 {
			t.Fatalf("Expected 1 suggestion, got %d", len(suggestions))
		}

		if suggestions[0].SuggestedPurposeCode != "MARKETING" {
			t.Errorf("Expected purpose MARKETING, got %s", suggestions[0].SuggestedPurposeCode)
		}
		if suggestions[0].Strategy != "AI" {
			t.Errorf("Expected strategy AI, got %s", suggestions[0].Strategy)
		}
	})

	t.Run("AI Skipped When Disabled", func(t *testing.T) {
		mockAI := &mockAIGateway{
			suggestFunc: func(ctx context.Context, input ai.PurposeSuggestionInput) ([]ai.PurposeSuggestion, error) {
				t.Error("AI should not be called when useAI is false")
				return nil, nil
			},
		}

		engine := NewContextEngine(loader, mockAI, logger)

		items := []PurposeSuggestionItem{
			{TableName: "unknown_table", ColumnName: "random_col"},
		}

		suggestions, err := engine.SuggestPurposes(context.Background(), items, false)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(suggestions) != 0 {
			t.Errorf("Expected 0 suggestions, got %d", len(suggestions))
		}
	})
}
