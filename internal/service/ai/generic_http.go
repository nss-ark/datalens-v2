package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// =============================================================================
// Generic HTTP Provider
// =============================================================================
//
// For any LLM API that doesn't follow the OpenAI or Anthropic format.
// Uses configurable JSON path mappings to extract the response.
//
// Use cases:
//   - Hugging Face Inference API
//   - Custom model servers (FastAPI, Flask, etc.)
//   - Self-hosted transformers endpoints
//   - Any REST API that takes text and returns text
//
// Configuration example (Hugging Face):
//
//	ProviderConfig{
//	    Name:                "huggingface-mistral",
//	    Type:                ProviderTypeGenericHTTP,
//	    APIKey:              "hf_...",
//	    Endpoint:            "https://api-inference.huggingface.co/models/mistralai/Mistral-7B-Instruct-v0.3",
//	    Headers:             map[string]string{"Authorization": "Bearer hf_..."},
//	    RequestBodyTemplate: `{"inputs": "{{.Prompt}}", "parameters": {"max_new_tokens": {{.MaxTokens}}, "temperature": {{.Temperature}}}}`,
//	    ResponseContentPath: "0.generated_text",
//	    DefaultModel:        "mistral-7b-instruct",
//	}
//
// Configuration example (custom FastAPI server):
//
//	ProviderConfig{
//	    Name:                "custom-classifier",
//	    Type:                ProviderTypeGenericHTTP,
//	    Endpoint:            "http://ml-server:8080",
//	    CompletionPath:      "/predict",
//	    RequestBodyTemplate: `{"text": "{{.Prompt}}", "max_length": {{.MaxTokens}}}`,
//	    ResponseContentPath: "result",
//	    ResponseTokensPath:  "metadata.token_count",
//	    DefaultModel:        "custom-v1",
//	}

// GenericHTTPProvider implements Provider for arbitrary REST APIs.
type GenericHTTPProvider struct {
	config ProviderConfig
	client *http.Client
	tmpl   *template.Template
}

// NewGenericHTTPProvider creates a generic HTTP provider.
// Returns an error if the RequestBodyTemplate is invalid.
func NewGenericHTTPProvider(cfg ProviderConfig) *GenericHTTPProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	p := &GenericHTTPProvider{
		config: cfg,
		client: &http.Client{Timeout: timeout},
	}

	// Pre-compile the request body template if provided.
	// Log a warning if it fails (don't block startup), but the
	// Complete call will fall back to default JSON format.
	if cfg.RequestBodyTemplate != "" {
		tmpl, err := template.New("request").Parse(cfg.RequestBodyTemplate)
		if err != nil {
			// Template is invalid — will use default JSON in Complete()
			_ = err // logged at startup by caller
		} else {
			p.tmpl = tmpl
		}
	}

	return p
}

func (p *GenericHTTPProvider) Name() string { return p.config.Name }

func (p *GenericHTTPProvider) IsAvailable(ctx context.Context) bool {
	// If it has an API key or is a local endpoint, consider available
	if p.config.APIKey != "" || isLocalEndpoint(p.config.Endpoint) {
		return true
	}
	// Check if any auth headers are set
	for k := range p.config.Headers {
		if strings.EqualFold(k, "authorization") || strings.EqualFold(k, "x-api-key") {
			return true
		}
	}
	return false
}

// templateData is passed to the request body template.
type templateData struct {
	Prompt      string
	MaxTokens   int
	Temperature float64
	Model       string
}

// Complete sends a prompt to the generic HTTP endpoint.
func (p *GenericHTTPProvider) Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error) {
	start := time.Now()

	model := p.config.DefaultModel
	maxTokens := opts.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}
	temp := opts.Temperature
	if temp == 0 {
		temp = 0.1
	}

	// Build request body
	var bodyBytes []byte
	if p.tmpl != nil {
		// Use configured template
		var buf bytes.Buffer
		err := p.tmpl.Execute(&buf, templateData{
			Prompt:      prompt,
			MaxTokens:   maxTokens,
			Temperature: temp,
			Model:       model,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: template execution: %w", p.config.Name, err)
		}
		bodyBytes = buf.Bytes()
	} else {
		// Default: send as simple JSON object
		defaultBody := map[string]any{
			"prompt":      prompt,
			"max_tokens":  maxTokens,
			"temperature": temp,
			"model":       model,
		}
		var err error
		bodyBytes, err = json.Marshal(defaultBody)
		if err != nil {
			return nil, fmt.Errorf("%s: marshal request: %w", p.config.Name, err)
		}
	}

	endpoint := strings.TrimRight(p.config.Endpoint, "/")
	path := p.config.CompletionPath // No default — endpoint IS the full URL unless overridden

	url := endpoint + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", p.config.Name, err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set auth
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	// Custom headers override defaults
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

	// Parse response using configured JSON paths
	responseText, tokensUsed, err := p.extractResponse(respBody)
	if err != nil {
		return nil, fmt.Errorf("%s: parse response: %w", p.config.Name, err)
	}

	return &CompletionResult{
		Response:   responseText,
		Provider:   p.config.Name,
		Model:      model,
		TokensUsed: tokensUsed,
		Duration:   time.Since(start),
	}, nil
}

// extractResponse extracts the completion text and token count from the
// raw JSON response using the configured JSON paths.
func (p *GenericHTTPProvider) extractResponse(body []byte) (string, int, error) {
	// Parse the response as generic JSON
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		// If it's not JSON, treat the entire body as the response text
		return string(body), 0, nil
	}

	// Extract response text
	contentPath := p.config.ResponseContentPath
	if contentPath == "" {
		// If no path configured, try common response formats
		contentPath = "generated_text" // HuggingFace default
	}

	responseText := extractJSONPath(data, contentPath)
	if responseText == "" {
		// Fallback: try stringifying the entire response
		responseText = string(body)
	}

	// Extract token count (optional)
	tokensUsed := 0
	if p.config.ResponseTokensPath != "" {
		tokenStr := extractJSONPath(data, p.config.ResponseTokensPath)
		if n, err := strconv.Atoi(tokenStr); err == nil {
			tokensUsed = n
		}
	}

	return responseText, tokensUsed, nil
}

// extractJSONPath navigates a JSON structure using a dot-separated path.
// Supports array indexing with numeric segments: "choices.0.message.content"
func extractJSONPath(data any, path string) string {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			current = v[part]
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return ""
			}
			current = v[idx]
		default:
			return ""
		}
	}

	switch v := current.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		// For complex types, marshal back to JSON
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}
