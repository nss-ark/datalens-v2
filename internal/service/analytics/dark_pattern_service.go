package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/complyark/datalens/internal/service/ai"
)

// DarkPatternService analyzes content for dark patterns.
type DarkPatternService struct {
	aiGateway ai.Gateway
}

// NewDarkPatternService creates a new instance of DarkPatternService.
func NewDarkPatternService(gateway ai.Gateway) *DarkPatternService {
	return &DarkPatternService{
		aiGateway: gateway,
	}
}

// DarkPatternAnalysisResult holds the result of the analysis.
type DarkPatternAnalysisResult struct {
	DetectedPatterns []DarkPatternType `json:"detected_patterns"`
	Confidence       float64           `json:"confidence"`
	Explanation      string            `json:"explanation"`
	CitedClause      string            `json:"cited_clause"`
}

// DarkPatternType represents the specific dark pattern found.
type DarkPatternType string

const (
	DarkPatternFalseUrgency           DarkPatternType = "FALSE_URGENCY"
	DarkPatternBasketSneaking         DarkPatternType = "BASKET_SNEAKING"
	DarkPatternConfirmShaming         DarkPatternType = "CONFIRM_SHAMING"
	DarkPatternForcedAction           DarkPatternType = "FORCED_ACTION"
	DarkPatternSubscriptionTrap       DarkPatternType = "SUBSCRIPTION_TRAP"
	DarkPatternInterfaceInterference  DarkPatternType = "INTERFACE_INTERFERENCE"
	DarkPatternBaitAndSwitch          DarkPatternType = "BAIT_AND_SWITCH"
	DarkPatternDripPricing            DarkPatternType = "DRIP_PRICING"
	DarkPatternDisguisedAdvertisement DarkPatternType = "DISGUISED_ADVERTISEMENT"
	DarkPatternNagging                DarkPatternType = "NAGGING"
	DarkPatternTrickQuestion          DarkPatternType = "TRICK_QUESTION"
	DarkPatternSaaSBilling            DarkPatternType = "SAAS_BILLING"
	DarkPatternRogueMalwares          DarkPatternType = "ROGUE_MALWARES"
)

// AnalyzeContent analyzes the given content for dark patterns.
func (s *DarkPatternService) AnalyzeContent(ctx context.Context, contentType string, content string) (*DarkPatternAnalysisResult, error) {
	// Construct the prompt using the template
	// We use the gateway's Complete method which takes a raw prompt string.
	// So we need to format the template ourselves or let the gateway handle it if it supported templates.
	// The current Gateway.Complete takes a string.

	// Use text/template used in prompts.go manually here?
	// Or just do simple string replacement since it's simple enough.
	// prompts.go has Go templates, so I should process it.

	// Let's create a helper to process the template first.
	// Actually, for simplicity and to avoid importing text/template every time, I'll specific format here
	// But sticking to the pattern, lets match how others might use it.
	// The other services "SuggestPurposes" likely does this internally in the Gateway or Service.
	// I'll do it here.

	// Format the prompt
	prompt := ai.DarkPatternPrompt
	prompt = strings.ReplaceAll(prompt, "{{.ContentType}}", contentType)
	prompt = strings.ReplaceAll(prompt, "{{.Content}}", content)

	resp, err := s.aiGateway.Complete(ctx, prompt, ai.CompletionOptions{
		UseCase:     "dark_pattern_detection",
		Priority:    "accuracy", // Sensitivity implies we want accuracy over speed? Or maybe speed for UI?
		Temperature: 0.1,        // Low temperature for deterministic classification
		MaxTokens:   1000,
	})
	if err != nil {
		return nil, fmt.Errorf("ai gateway error: %w", err)
	}

	// Parse the response
	var result DarkPatternAnalysisResult
	if err := json.Unmarshal([]byte(resp.Response), &result); err != nil {
		// Clean the response if it contains markdown code blocks
		cleaned := strings.TrimPrefix(resp.Response, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
			return nil, fmt.Errorf("failed to parse AI response: %w. Response: %s", err, resp.Response)
		}
	}

	return &result, nil
}
