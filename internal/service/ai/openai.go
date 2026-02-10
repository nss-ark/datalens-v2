package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// =============================================================================
// OpenAI-Compatible Provider
// =============================================================================
//
// This single provider covers ANY API that speaks the OpenAI chat completions
// format. This includes (but is not limited to):
//
//	Provider        │ Endpoint                          │ Notes
//	────────────────┼───────────────────────────────────┼──────────
//	OpenAI          │ https://api.openai.com/v1         │ GPT-4o, GPT-4o-mini
//	Azure OpenAI    │ https://{}.openai.azure.com       │ Deployed models
//	Ollama          │ http://localhost:11434/v1          │ Local LLMs
//	vLLM            │ http://your-server:8000/v1        │ Self-hosted
//	Together.ai     │ https://api.together.xyz/v1       │ OSS models
//	Groq            │ https://api.groq.com/openai/v1    │ Fast inference
//	Mistral         │ https://api.mistral.ai/v1         │ Mistral/Mixtral
//	DeepSeek        │ https://api.deepseek.com/v1       │ DeepSeek
//	Perplexity      │ https://api.perplexity.ai         │ Sonar models
//	Anyscale        │ https://api.endpoints.anyscale.com│ OSS models
//	LiteLLM Proxy   │ http://your-proxy:4000/v1         │ Any model via proxy
//	LM Studio       │ http://localhost:1234/v1          │ Local GUI
//	OpenRouter      │ https://openrouter.ai/api/v1      │ Multi-provider
//
// Configuration example:
//
//	ProviderConfig{
//	    Name:         "ollama-local",
//	    Type:         ProviderTypeOpenAICompatible,
//	    Endpoint:     "http://localhost:11434/v1",
//	    DefaultModel: "llama3.2:latest",
//	    FastModel:    "llama3.2:1b",
//	    Timeout:      30 * time.Second,
//	}

// OpenAICompatProvider implements Provider for any OpenAI-compatible API.
type OpenAICompatProvider struct {
	config ProviderConfig
	client *http.Client
}

// NewOpenAICompatProvider creates a provider for any OpenAI-compatible endpoint.
func NewOpenAICompatProvider(cfg ProviderConfig) *OpenAICompatProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &OpenAICompatProvider{
		config: cfg,
		client: &http.Client{Timeout: timeout},
	}
}

func (p *OpenAICompatProvider) Name() string { return p.config.Name }

func (p *OpenAICompatProvider) IsAvailable(ctx context.Context) bool {
	return p.config.APIKey != "" || isLocalEndpoint(p.config.Endpoint)
}

// Complete sends a prompt to the OpenAI-compatible API and returns the response.
func (p *OpenAICompatProvider) Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error) {
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
		temp = 0.1 // Low temperature for structured JSON output
	}

	systemPrompt := opts.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = defaultSystemPrompt
	}

	reqBody := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   maxTokens,
		Temperature: temp,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: marshal request: %w", p.config.Name, err)
	}

	endpoint := strings.TrimRight(p.config.Endpoint, "/")
	completionPath := p.config.CompletionPath
	if completionPath == "" {
		completionPath = "/chat/completions"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+completionPath, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", p.config.Name, err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set auth header — Bearer token is the standard for OpenAI-compatible APIs
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	// Set any custom headers (e.g., Azure x-api-key, OpenRouter HTTP-Referer)
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

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response: %w", p.config.Name, err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("%s: no choices in response", p.config.Name)
	}

	return &CompletionResult{
		Response:   result.Choices[0].Message.Content,
		Provider:   p.config.Name,
		Model:      model,
		TokensUsed: result.Usage.TotalTokens,
		Duration:   time.Since(start),
	}, nil
}

// isLocalEndpoint checks if the endpoint is a local/self-hosted server.
func isLocalEndpoint(endpoint string) bool {
	return strings.Contains(endpoint, "localhost") ||
		strings.Contains(endpoint, "127.0.0.1") ||
		strings.Contains(endpoint, "0.0.0.0") ||
		strings.Contains(endpoint, "host.docker.internal")
}

// defaultSystemPrompt is used when CompletionOptions.SystemPrompt is empty.
const defaultSystemPrompt = "You are a data privacy analysis assistant. Always respond with valid JSON only."

// maxResponseBytes limits response body reads to 1MB to prevent OOM.
const maxResponseBytes = 1 << 20

// --- OpenAI wire format (used by all OpenAI-compatible APIs) ---

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
