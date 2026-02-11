package governance

import (
	"context"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/internal/domain/governance/templates"
	"github.com/complyark/datalens/internal/service/ai"
)

// ContextEngine orchestrates the purpose suggestion logic.
type ContextEngine struct {
	templateLoader *templates.Loader
	aiGateway      ai.Gateway
	logger         *slog.Logger
}

// NewContextEngine creates a new ContextEngine.
func NewContextEngine(
	templateLoader *templates.Loader,
	aiGateway ai.Gateway,
	logger *slog.Logger,
) *ContextEngine {
	return &ContextEngine{
		templateLoader: templateLoader,
		aiGateway:      aiGateway,
		logger:         logger,
	}
}

// SuggestPurposes generates purpose suggestions for a list of table/column pairs.
// It tries strategies in order: Pattern -> AI (if enabled/requested).
func (e *ContextEngine) SuggestPurposes(
	ctx context.Context,
	items []PurposeSuggestionItem,
	useAI bool,
) ([]governance.PurposeSuggestion, error) {
	var suggestions []governance.PurposeSuggestion

	for _, item := range items {
		// 1. Pattern Strategy (Fast, deterministic)
		if matched, suggestion := e.runPatternStrategy(item); matched {
			suggestions = append(suggestions, suggestion)
			continue // Skip AI if pattern matches (cost saving)
		}

		// 2. AI Strategy (Slower, smarter)
		if useAI {
			aiSuggestions, err := e.runAIStrategy(ctx, item)
			if err != nil {
				e.logger.Error("AI strategy failed", "error", err, "table", item.TableName, "column", item.ColumnName)
			} else if len(aiSuggestions) > 0 {
				suggestions = append(suggestions, aiSuggestions...)
				continue
			}
		}

		// 3. Fallback / No match
	}

	return suggestions, nil
}

// PurposeSuggestionItem represents a single data element to analyze.
type PurposeSuggestionItem struct {
	TableName  string
	ColumnName string
	DataType   string
	Industry   string // e.g. "E-Commerce", "Healthcare"
}

func (e *ContextEngine) runPatternStrategy(item PurposeSuggestionItem) (bool, governance.PurposeSuggestion) {
	if e.templateLoader == nil {
		return false, governance.PurposeSuggestion{}
	}

	code, confidence, reason, found := e.templateLoader.FindMatch(item.TableName, item.ColumnName)
	if !found {
		return false, governance.PurposeSuggestion{}
	}

	return true, governance.PurposeSuggestion{
		Table:                item.TableName,
		Column:               item.ColumnName,
		SuggestedPurposeCode: code,
		Confidence:           confidence,
		Reason:               reason,
		Strategy:             "PATTERN",
	}
}

func (e *ContextEngine) runAIStrategy(ctx context.Context, item PurposeSuggestionItem) ([]governance.PurposeSuggestion, error) {
	if e.aiGateway == nil {
		return nil, nil
	}

	input := ai.PurposeSuggestionInput{
		EntityName: item.TableName,
		ColumnName: item.ColumnName,
		// Industry: item.Industry, // Gateway input has Industry
	}
	// TODO: Pass industry if available in item

	aiResults, err := e.aiGateway.SuggestPurposes(ctx, input)
	if err != nil {
		return nil, err
	}

	var suggestions []governance.PurposeSuggestion
	for _, res := range aiResults {
		suggestions = append(suggestions, governance.PurposeSuggestion{
			Table:                item.TableName,
			Column:               item.ColumnName,
			SuggestedPurposeCode: res.PurposeCode,
			Confidence:           res.Confidence,
			Reason:               res.Reasoning,
			Strategy:             "AI",
		})
	}

	return suggestions, nil
}
