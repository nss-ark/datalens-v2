package detection

import (
	"context"
	"fmt"
	"strings"

	"github.com/complyark/datalens/internal/service/ai"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// AI Detection Strategy
// =============================================================================
//
// AIStrategy wraps the AI Gateway (any LLM provider — OpenAI, Ollama,
// Anthropic, etc.) as a detection Strategy so it can be composed with
// regex and heuristic strategies in the ComposableDetector.
//
// Flow:
//
//	ComposableDetector
//	  ├── PatternStrategy   (regex, weight 0.6)
//	  ├── HeuristicStrategy (column names, weight 0.4)
//	  └── AIStrategy        ← this file (LLM, weight 0.8)
//	          │
//	          └── Gateway.DetectPII() → OpenAI / Anthropic / Ollama / ...
//
// The AI strategy has the highest weight because LLMs provide the best
// contextual understanding, but lower-weight strategies act as guardrails
// and boost confidence when they agree.

// AIStrategy implements the Strategy interface using the AI Gateway.
type AIStrategy struct {
	gateway ai.Gateway
	weight  float64
}

// NewAIStrategy creates an AI-based detection strategy.
// The weight determines how much influence AI results have in the
// merged confidence score (recommended: 0.8).
func NewAIStrategy(gateway ai.Gateway, weight float64) *AIStrategy {
	if weight <= 0 || weight > 1.0 {
		weight = 0.8
	}
	return &AIStrategy{
		gateway: gateway,
		weight:  weight,
	}
}

func (s *AIStrategy) Name() string                  { return "ai" }
func (s *AIStrategy) Method() types.DetectionMethod { return types.DetectionMethodAI }
func (s *AIStrategy) Weight() float64               { return s.weight }

// Detect sends the column metadata to the AI Gateway and converts the
// AI response into the detection Result format.
func (s *AIStrategy) Detect(ctx context.Context, input Input) ([]Result, error) {
	// Build adjacent column names as flat strings for the AI prompt
	adjacentNames := make([]string, len(input.AdjacentColumns))
	for i, col := range input.AdjacentColumns {
		adjacentNames[i] = col.Name
	}

	// Prepare samples — prefer pre-sanitized, fall back to raw
	// (the Gateway's DetectPII also sanitizes, so double-sanitizing is safe)
	samples := input.SanitizedSamples
	if len(samples) == 0 {
		samples = input.Samples
	}

	aiInput := ai.PIIDetectionInput{
		TableName:        input.TableName,
		ColumnName:       input.ColumnName,
		DataType:         input.DataType,
		SanitizedSamples: samples,
		AdjacentColumns:  adjacentNames,
		Industry:         input.Industry,
	}

	aiResult, err := s.gateway.DetectPII(ctx, aiInput)
	if err != nil {
		return nil, fmt.Errorf("ai strategy: %w", err)
	}

	// If the AI says it's not PII, return no results
	if !aiResult.IsPII {
		return nil, nil
	}

	// Convert AI result to detection Result
	result := Result{
		Category:    aiResult.Category,
		Type:        aiResult.Type,
		Sensitivity: aiResult.Sensitivity,
		Confidence:  aiResult.Confidence,
		Method:      types.DetectionMethodAI,
		Reasoning:   aiResult.Reasoning,
	}

	// Validate category/type — AI sometimes returns unexpected values
	if result.Category == "" {
		result.Category = inferCategory(result.Type)
	}
	if result.Sensitivity == "" {
		result.Sensitivity = inferSensitivity(result.Category)
	}

	return []Result{result}, nil
}

// inferCategory maps a PIIType to its most likely PIICategory
// when the AI response doesn't include category information.
func inferCategory(piiType types.PIIType) types.PIICategory {
	categoryMap := map[types.PIIType]types.PIICategory{
		types.PIITypeName:          types.PIICategoryIdentity,
		types.PIITypeDOB:           types.PIICategoryIdentity,
		types.PIITypeGender:        types.PIICategoryIdentity,
		types.PIITypeEmail:         types.PIICategoryContact,
		types.PIITypePhone:         types.PIICategoryContact,
		types.PIITypeAddress:       types.PIICategoryContact,
		types.PIITypeAadhaar:       types.PIICategoryGovernmentID,
		types.PIITypePAN:           types.PIICategoryGovernmentID,
		types.PIITypePassport:      types.PIICategoryGovernmentID,
		types.PIITypeSSN:           types.PIICategoryGovernmentID,
		types.PIITypeNationalID:    types.PIICategoryGovernmentID,
		types.PIITypeBankAccount:   types.PIICategoryFinancial,
		types.PIITypeCreditCard:    types.PIICategoryFinancial,
		types.PIITypeIPAddress:     types.PIICategoryBehavioral,
		types.PIITypeMACAddress:    types.PIICategoryBehavioral,
		types.PIITypeDeviceID:      types.PIICategoryBehavioral,
		types.PIITypeBiometric:     types.PIICategoryBiometric,
		types.PIITypeMedicalRecord: types.PIICategoryHealth,
		types.PIITypePhoto:         types.PIICategoryIdentity,
		types.PIITypeSignature:     types.PIICategoryIdentity,
	}

	if cat, ok := categoryMap[piiType]; ok {
		return cat
	}
	return types.PIICategoryIdentity // Safe default
}

// inferSensitivity maps a PIICategory to its default sensitivity level.
func inferSensitivity(category types.PIICategory) types.SensitivityLevel {
	switch category {
	case types.PIICategoryBiometric, types.PIICategoryGenetic, types.PIICategoryHealth, types.PIICategoryMinor:
		return types.SensitivityCritical
	case types.PIICategoryGovernmentID, types.PIICategoryFinancial:
		return types.SensitivityHigh
	case types.PIICategoryIdentity, types.PIICategoryContact, types.PIICategoryLocation:
		return types.SensitivityMedium
	case types.PIICategoryBehavioral, types.PIICategoryProfessional:
		return types.SensitivityLow
	default:
		return types.SensitivityMedium
	}
}

// =============================================================================
// Convenience Constructor
// =============================================================================

// NewDefaultDetector creates a ComposableDetector with the standard strategy
// stack: Pattern (regex), Heuristic (column names), and AI (LLM).
// Pass nil for gateway to skip the AI strategy (offline mode).
func NewDefaultDetector(gateway ai.Gateway) *ComposableDetector {
	strategies := []Strategy{
		NewPatternStrategy(),
		NewHeuristicStrategy(),
		NewIndustryStrategy(),
	}

	if gateway != nil {
		strategies = append(strategies, NewAIStrategy(gateway, 0.8))
	}

	return NewComposableDetector(strategies...)
}

// NewOfflineDetector creates a detector with only pattern + heuristic
// strategies (no AI). Useful for testing or when no LLM is configured.
func NewOfflineDetector() *ComposableDetector {
	return NewDefaultDetector(nil)
}

// =============================================================================
// Helper — suppress unused import warning
// =============================================================================

var _ = strings.TrimSpace // used when we expand adjacentNames logic
