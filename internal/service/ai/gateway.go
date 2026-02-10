// Package ai implements the AI Gateway — a unified abstraction layer
// over multiple LLM providers (OpenAI, Anthropic, Ollama) with
// fallback orchestration, rate limiting, caching, and cost tracking.
package ai

import (
	"context"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// Core Gateway Interface
// =============================================================================

// Gateway is the unified AI abstraction. Application code calls Gateway
// methods; the implementation handles provider selection, fallback,
// caching, and cost tracking transparently.
type Gateway interface {
	// DetectPII analyzes column metadata and optional samples to determine
	// whether a data field contains PII. This is the most critical AI use case.
	DetectPII(ctx context.Context, input PIIDetectionInput) (*PIIDetectionResult, error)

	// SuggestPurposes recommends data processing purposes based on the
	// data source context (table name, column name, PII type, industry).
	SuggestPurposes(ctx context.Context, input PurposeSuggestionInput) ([]PurposeSuggestion, error)

	// Complete performs a generic LLM completion. Used for ad-hoc tasks
	// like breach impact assessment, DSR identity verification, etc.
	Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error)
}

// =============================================================================
// Provider Interface — implemented by each LLM backend
// =============================================================================

// Provider abstracts a single LLM provider (OpenAI, Anthropic, Ollama, etc.).
type Provider interface {
	// Name returns the provider identifier (e.g., "openai", "anthropic", "ollama").
	Name() string

	// Complete sends a prompt and returns the raw completion.
	Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error)

	// IsAvailable returns whether this provider is currently reachable
	// and within rate/budget limits.
	IsAvailable(ctx context.Context) bool
}

// =============================================================================
// PII Detection Types
// =============================================================================

// PIIDetectionInput contains all context needed for AI-based PII detection.
// IMPORTANT: SanitizedSamples must never contain real PII — only patterns
// like "[EMAIL: format valid]" or "[NAME: 2 words]".
type PIIDetectionInput struct {
	TableName        string   `json:"table_name"`
	ColumnName       string   `json:"column_name"`
	DataType         string   `json:"data_type"`
	SanitizedSamples []string `json:"sanitized_samples"`
	AdjacentColumns  []string `json:"adjacent_columns"`
	Industry         string   `json:"industry,omitempty"`
}

// PIIDetectionResult holds the AI's analysis of a single column.
type PIIDetectionResult struct {
	IsPII          bool                   `json:"is_pii"`
	Category       types.PIICategory      `json:"category,omitempty"`
	Type           types.PIIType          `json:"type,omitempty"`
	Sensitivity    types.SensitivityLevel `json:"sensitivity,omitempty"`
	Confidence     float64                `json:"confidence"`
	Reasoning      string                 `json:"reasoning"`
	RequiresReview bool                   `json:"requires_review"`
	Provider       string                 `json:"provider"`
	TokensUsed     int                    `json:"tokens_used"`
	Duration       time.Duration          `json:"duration"`
}

// =============================================================================
// Purpose Suggestion Types
// =============================================================================

// PurposeSuggestionInput provides context for purpose inference.
type PurposeSuggestionInput struct {
	DataSourceType string `json:"data_source_type"`
	EntityName     string `json:"entity_name"`
	ColumnName     string `json:"column_name"`
	PIIType        string `json:"pii_type"`
	Industry       string `json:"industry"`
}

// PurposeSuggestion is a single suggested purpose with confidence.
type PurposeSuggestion struct {
	PurposeCode             string           `json:"purpose_code"`
	Confidence              float64          `json:"confidence"`
	Reasoning               string           `json:"reasoning"`
	LegalBasis              types.LegalBasis `json:"legal_basis"`
	RequiresExplicitConsent bool             `json:"requires_explicit_consent"`
}

// =============================================================================
// Generic Completion Types
// =============================================================================

// CompletionOptions configures an LLM request.
type CompletionOptions struct {
	UseCase      string        `json:"use_case"`           // e.g., "pii_detection", "purpose_suggestion"
	Priority     string        `json:"priority,omitempty"` // "accuracy" (slower model) or "speed" (faster model)
	MaxTokens    int           `json:"max_tokens,omitempty"`
	Temperature  float64       `json:"temperature,omitempty"`
	CacheTTL     time.Duration `json:"cache_ttl,omitempty"`
	SystemPrompt string        `json:"system_prompt,omitempty"` // Overrides default system prompt
}

// CompletionResult holds the raw LLM response with metadata.
type CompletionResult struct {
	Response   string        `json:"response"`
	Provider   string        `json:"provider"`
	Model      string        `json:"model"`
	TokensUsed int           `json:"tokens_used"`
	Duration   time.Duration `json:"duration"`
	Cached     bool          `json:"cached"`
}

// =============================================================================
// Provider Configuration
// =============================================================================

// ProviderType defines the type of LLM provider.
type ProviderType string

const (
	ProviderTypeOpenAICompatible ProviderType = "openai_compatible"
	ProviderTypeAnthropic        ProviderType = "anthropic"
	ProviderTypeGenericHTTP      ProviderType = "generic_http"
)

// ProviderConfig holds the configuration for a single AI provider.
type ProviderConfig struct {
	// Core identity
	Name string       `json:"name"`
	Type ProviderType `json:"type"` // "openai_compatible", "anthropic", "generic_http"

	// Connection
	APIKey   string `json:"api_key,omitempty"`
	Endpoint string `json:"endpoint,omitempty"` // Base URL (e.g., "https://api.openai.com/v1", "http://localhost:11434/v1")

	// Model selection
	DefaultModel string `json:"default_model"`        // Model for accuracy tasks (e.g., "gpt-4o", "claude-3-5-sonnet")
	FastModel    string `json:"fast_model,omitempty"` // Model for speed tasks (e.g., "gpt-4o-mini", "claude-3-haiku")

	// Rate limits
	RequestsPerMinute int `json:"requests_per_minute"`
	TokensPerMinute   int `json:"tokens_per_minute"`

	// Cost
	CostPer1KTokens float64 `json:"cost_per_1k_tokens"`

	// Timeouts
	Timeout time.Duration `json:"timeout"`

	// Headers — additional HTTP headers (e.g., {"x-api-key": "..."})
	// Useful for custom auth schemes or Hugging Face tokens.
	Headers map[string]string `json:"headers,omitempty"`

	// GenericHTTP-specific — JSON path mappings for custom APIs
	// Only used when Type = "generic_http"
	RequestBodyTemplate string `json:"request_body_template,omitempty"` // Go template for request body
	ResponseContentPath string `json:"response_content_path,omitempty"` // JSONPath to response text (e.g., "choices.0.message.content")
	ResponseTokensPath  string `json:"response_tokens_path,omitempty"`  // JSONPath to token count
	CompletionPath      string `json:"completion_path,omitempty"`       // API path (e.g., "/v1/chat/completions")

	// Anthropic-specific
	AnthropicVersion string `json:"anthropic_version,omitempty"` // e.g., "2023-06-01"
}

// GatewayConfig holds the overall AI Gateway configuration.
type GatewayConfig struct {
	Providers      []ProviderConfig `json:"providers"`      // Slice fed to BuildRegistryFromConfig
	FallbackChain  []string         `json:"fallback_chain"` // e.g., ["openai", "anthropic", "ollama"]
	DefaultTimeout time.Duration    `json:"default_timeout"`
	CacheEnabled   bool             `json:"cache_enabled"`
}
