package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// =============================================================================
// Anthropic Provider (Messages API)
// =============================================================================
//
// Anthropic uses a different API format from OpenAI. This provider
// handles the Messages API used by Claude models.
//
// Configuration example:
//
//	ProviderConfig{
//	    Name:             "anthropic",
//	    Type:             ProviderTypeAnthropic,
//	    APIKey:           "sk-ant-...",
//	    Endpoint:         "https://api.anthropic.com",
//	    DefaultModel:     "claude-3-5-sonnet-20241022",
//	    FastModel:        "claude-3-haiku-20240307",
//	    AnthropicVersion: "2023-06-01",
//	}

// AnthropicProvider implements Provider for Anthropic's Messages API.
type AnthropicProvider struct {
	config ProviderConfig
	client *http.Client
}

// NewAnthropicProvider creates an Anthropic provider.
func NewAnthropicProvider(cfg ProviderConfig) *AnthropicProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	if cfg.AnthropicVersion == "" {
		cfg.AnthropicVersion = "2023-06-01"
	}
	return &AnthropicProvider{
		config: cfg,
		client: &http.Client{Timeout: timeout},
	}
}

func (p *AnthropicProvider) Name() string { return p.config.Name }

func (p *AnthropicProvider) IsAvailable(ctx context.Context) bool {
	return p.config.APIKey != ""
}

// Complete sends a prompt to the Anthropic Messages API.
func (p *AnthropicProvider) Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error) {
	start := time.Now()

	model := p.config.DefaultModel
	if opts.Priority == "speed" && p.config.FastModel != "" {
		model = p.config.FastModel
	}

	maxTokens := opts.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	temp := opts.Temperature
	if temp == 0 {
		temp = 0.1
	}

	systemPrompt := opts.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = defaultSystemPrompt
	}

	reqBody := anthropicRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: temp,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: marshal request: %w", p.config.Name, err)
	}

	endpoint := p.config.Endpoint
	if endpoint == "" {
		endpoint = "https://api.anthropic.com"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", p.config.Name, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.config.APIKey)
	req.Header.Set("anthropic-version", p.config.AnthropicVersion)

	// Custom headers
	for k, v := range p.config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", p.config.Name, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("%s: read response: %w", p.config.Name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: API error %d: %s", p.config.Name, resp.StatusCode, string(respBody))
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response: %w", p.config.Name, err)
	}

	// Extract text from content blocks
	var responseText string
	for _, block := range result.Content {
		if block.Type == "text" {
			responseText += block.Text
		}
	}

	return &CompletionResult{
		Response:   responseText,
		Provider:   p.config.Name,
		Model:      model,
		TokensUsed: result.Usage.InputTokens + result.Usage.OutputTokens,
		Duration:   time.Since(start),
	}, nil
}

// --- Anthropic wire format ---

type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	Temperature float64            `json:"temperature,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []anthropicContentBlock `json:"content"`
	Usage   anthropicUsage          `json:"usage"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
