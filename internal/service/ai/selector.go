package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// =============================================================================
// Provider Selector — Fallback chain with preference routing
// =============================================================================
//
// The Selector wraps a Registry and adds intelligent provider selection:
//   - Fallback chain: try providers in configured order until one works
//   - Use-case routing: specific providers for specific tasks
//   - Automatic failover: skip unavailable providers
//
// Example fallback chain: ["openai", "anthropic", "ollama-local"]
//   - Try OpenAI first (best accuracy for PII detection)
//   - If OpenAI fails/rate-limited, fall back to Anthropic
//   - If Anthropic fails, use local Ollama (always available)

// Selector picks the best available provider for a request.
type Selector struct {
	registry      *Registry
	fallbackChain []string // Provider names in priority order

	mu         sync.RWMutex
	useCaseMap map[string]string // UseCase → preferred provider name

	logger *slog.Logger
}

// NewSelector creates a provider selector with a fallback chain.
func NewSelector(registry *Registry, fallbackChain []string, logger *slog.Logger) *Selector {
	if logger == nil {
		logger = slog.Default()
	}
	return &Selector{
		registry:      registry,
		fallbackChain: fallbackChain,
		useCaseMap:    make(map[string]string),
		logger:        logger,
	}
}

// SetUseCasePreference maps a use case to a preferred provider.
// e.g., SetUseCasePreference("pii_detection", "openai") means
// PII detection tasks prefer OpenAI over other providers.
// Safe for concurrent use.
func (s *Selector) SetUseCasePreference(useCase, providerName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.useCaseMap[useCase] = providerName
}

// CompleteWithFallback tries providers in the fallback chain until one succeeds.
// If a use-case preference is set and that provider is available, it's tried first.
func (s *Selector) CompleteWithFallback(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error) {
	// Build the ordered list of providers to try
	chain := s.buildChain(opts.UseCase)

	var lastErr error
	for _, name := range chain {
		provider := s.registry.Get(name)
		if provider == nil {
			continue
		}

		if !provider.IsAvailable(ctx) {
			s.logger.Warn("provider unavailable, skipping",
				"provider", name,
				"use_case", opts.UseCase,
			)
			continue
		}

		result, err := provider.Complete(ctx, prompt, opts)
		if err != nil {
			lastErr = err
			s.logger.Error("provider failed, trying next",
				"provider", name,
				"error", err,
				"use_case", opts.UseCase,
			)
			continue
		}

		// Check if result is empty — some providers return empty on edge cases
		if result.Response == "" {
			lastErr = fmt.Errorf("%s: empty response", name)
			s.logger.Warn("provider returned empty response, trying next",
				"provider", name,
			)
			continue
		}

		return result, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
	}
	return nil, fmt.Errorf("no providers available in fallback chain: %v", chain)
}

// buildChain constructs the ordered provider list for a request.
// Use-case preference goes first, followed by the fallback chain (deduplicated).
func (s *Selector) buildChain(useCase string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var chain []string
	seen := make(map[string]bool)

	// 1. Use-case preference first
	if preferred, ok := s.useCaseMap[useCase]; ok {
		chain = append(chain, preferred)
		seen[preferred] = true
	}

	// 2. Rest of fallback chain
	for _, name := range s.fallbackChain {
		if !seen[name] {
			chain = append(chain, name)
			seen[name] = true
		}
	}

	return chain
}

// =============================================================================
// Default Gateway Implementation
// =============================================================================

// DefaultGateway implements the Gateway interface using the Selector
// for provider management. This is the main entry point for application code.
type DefaultGateway struct {
	selector  *Selector
	sanitizer *Sanitizer
	logger    *slog.Logger
}

// NewDefaultGateway creates the main AI Gateway.
func NewDefaultGateway(selector *Selector, logger *slog.Logger) *DefaultGateway {
	if logger == nil {
		logger = slog.Default()
	}
	return &DefaultGateway{
		selector:  selector,
		sanitizer: NewSanitizer(),
		logger:    logger,
	}
}

// DetectPII analyzes a column for PII using the configured AI providers.
// Raw samples are sanitized before being sent to any provider.
func (g *DefaultGateway) DetectPII(ctx context.Context, input PIIDetectionInput) (*PIIDetectionResult, error) {
	start := time.Now()

	// Sanitize samples — NEVER send raw PII to cloud providers
	sanitizedSamples := input.SanitizedSamples
	if len(sanitizedSamples) > 0 {
		sanitizedSamples = g.sanitizer.SanitizeSamples(sanitizedSamples)
	}

	prompt := fmt.Sprintf(
		"Analyze this database column for PII and respond with JSON:\n"+
			"```json\n{\"is_pii\": bool, \"category\": \"string\", \"type\": \"string\","+
			" \"sensitivity\": \"LOW|MEDIUM|HIGH|CRITICAL\", \"confidence\": 0.0-1.0,"+
			" \"reasoning\": \"string\"}\n```\n\n"+
			"Table: %s\nColumn: %s\nData Type: %s\n"+
			"Sanitized Samples: %v\nAdjacent Columns: %v\nIndustry: %s",
		input.TableName, input.ColumnName, input.DataType,
		sanitizedSamples, input.AdjacentColumns, input.Industry,
	)

	opts := CompletionOptions{
		UseCase:      "pii_detection",
		Priority:     "accuracy",
		MaxTokens:    512,
		Temperature:  0.1,
		SystemPrompt: "You are a data privacy analysis assistant specializing in PII detection. Always respond with valid JSON only. No markdown fences.",
	}

	result, err := g.selector.CompleteWithFallback(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("pii detection: %w", err)
	}

	// Parse structured response from LLM
	piiResult, parseErr := parsePIIResponse(result)
	if parseErr != nil {
		g.logger.Warn("failed to parse structured PII response, returning raw",
			"error", parseErr,
			"provider", result.Provider,
		)
		// Return raw response as reasoning — detection service will handle
		return &PIIDetectionResult{
			Reasoning:  result.Response,
			Provider:   result.Provider,
			TokensUsed: result.TokensUsed,
			Duration:   time.Since(start),
		}, nil
	}

	piiResult.Provider = result.Provider
	piiResult.TokensUsed = result.TokensUsed
	piiResult.Duration = time.Since(start)
	return piiResult, nil
}

// parsePIIResponse attempts to parse the LLM JSON response into a PIIDetectionResult.
func parsePIIResponse(result *CompletionResult) (*PIIDetectionResult, error) {
	var parsed PIIDetectionResult
	if err := json.Unmarshal([]byte(result.Response), &parsed); err != nil {
		return nil, fmt.Errorf("json parse: %w", err)
	}
	return &parsed, nil
}

// SuggestPurposes recommends data processing purposes.
func (g *DefaultGateway) SuggestPurposes(ctx context.Context, input PurposeSuggestionInput) ([]PurposeSuggestion, error) {
	prompt := fmt.Sprintf(
		"Suggest data processing purposes. Respond with JSON array:\n"+
			"```json\n[{\"purpose_code\": \"string\", \"confidence\": 0.0-1.0,"+
			" \"reasoning\": \"string\", \"legal_basis\": \"CONSENT|CONTRACT|...\","+
			" \"requires_explicit_consent\": bool}]\n```\n\n"+
			"Data Source Type: %s\nEntity: %s\nColumn: %s\n"+
			"PII Type: %s\nIndustry: %s",
		input.DataSourceType, input.EntityName, input.ColumnName,
		input.PIIType, input.Industry,
	)

	opts := CompletionOptions{
		UseCase:      "purpose_suggestion",
		Priority:     "accuracy",
		MaxTokens:    512,
		Temperature:  0.2,
		SystemPrompt: "You are a data privacy compliance assistant. Suggest processing purposes based on GDPR, DPDPA, and other regulations. Respond with valid JSON only.",
	}

	result, err := g.selector.CompleteWithFallback(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("purpose suggestion: %w", err)
	}

	// Parse JSON array response
	var suggestions []PurposeSuggestion
	if err := json.Unmarshal([]byte(result.Response), &suggestions); err != nil {
		g.logger.Warn("failed to parse purpose suggestions",
			"error", err,
			"provider", result.Provider,
		)
		return nil, nil
	}

	return suggestions, nil
}

// Complete performs a generic completion.
func (g *DefaultGateway) Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error) {
	return g.selector.CompleteWithFallback(ctx, prompt, opts)
}
