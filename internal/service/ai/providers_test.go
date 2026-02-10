package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestOpenAICompatProvider_Complete tests the OpenAI-compatible provider
// against a mock HTTP server that speaks the OpenAI wire format.
func TestOpenAICompatProvider_Complete(t *testing.T) {
	// Mock server returning OpenAI-format response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correct headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type: got %q, want application/json", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Authorization: got %q, want 'Bearer test-key'", r.Header.Get("Authorization"))
		}

		// Verify request body
		var req openAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "gpt-4o-mini" {
			t.Errorf("model: got %q, want gpt-4o-mini", req.Model)
		}

		// Return mock response
		resp := openAIResponse{
			Choices: []openAIChoice{
				{Message: openAIMessage{Role: "assistant", Content: `{"is_pii": true}`}},
			},
			Usage: openAIUsage{PromptTokens: 50, CompletionTokens: 10, TotalTokens: 60},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAICompatProvider(ProviderConfig{
		Name:         "test-openai",
		Type:         ProviderTypeOpenAICompatible,
		APIKey:       "test-key",
		Endpoint:     server.URL,
		DefaultModel: "gpt-4o",
		FastModel:    "gpt-4o-mini",
		Timeout:      5 * time.Second,
	})

	// Test with "speed" priority — should use FastModel
	result, err := provider.Complete(context.Background(), "test prompt", CompletionOptions{
		Priority: "speed",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response != `{"is_pii": true}` {
		t.Errorf("response: got %q", result.Response)
	}
	if result.TokensUsed != 60 {
		t.Errorf("tokens: got %d, want 60", result.TokensUsed)
	}
	if result.Provider != "test-openai" {
		t.Errorf("provider: got %q, want test-openai", result.Provider)
	}
}

func TestOpenAICompatProvider_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limit exceeded"}`))
	}))
	defer server.Close()

	provider := NewOpenAICompatProvider(ProviderConfig{
		Name:         "test-openai",
		Endpoint:     server.URL,
		APIKey:       "key",
		DefaultModel: "gpt-4o",
	})

	_, err := provider.Complete(context.Background(), "test", CompletionOptions{})
	if err == nil {
		t.Fatal("expected error on 429, got nil")
	}
}

func TestOpenAICompatProvider_IsAvailable(t *testing.T) {
	// With API key
	p1 := NewOpenAICompatProvider(ProviderConfig{APIKey: "key", Endpoint: "https://api.openai.com/v1"})
	if !p1.IsAvailable(context.Background()) {
		t.Error("should be available with API key")
	}

	// Local endpoint without key
	p2 := NewOpenAICompatProvider(ProviderConfig{Endpoint: "http://localhost:11434/v1"})
	if !p2.IsAvailable(context.Background()) {
		t.Error("should be available for localhost")
	}

	// Cloud endpoint without key
	p3 := NewOpenAICompatProvider(ProviderConfig{Endpoint: "https://api.openai.com/v1"})
	if p3.IsAvailable(context.Background()) {
		t.Error("should NOT be available for cloud without API key")
	}
}

// TestAnthropicProvider_Complete tests the Anthropic provider against a mock server.
func TestAnthropicProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Anthropic-specific headers
		if r.Header.Get("x-api-key") != "sk-ant-test" {
			t.Errorf("x-api-key: got %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("anthropic-version: got %q", r.Header.Get("anthropic-version"))
		}

		resp := anthropicResponse{
			Content: []anthropicContentBlock{
				{Type: "text", Text: `{"is_pii": false}`},
			},
			Usage: anthropicUsage{InputTokens: 40, OutputTokens: 8},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewAnthropicProvider(ProviderConfig{
		Name:         "test-anthropic",
		Type:         ProviderTypeAnthropic,
		APIKey:       "sk-ant-test",
		Endpoint:     server.URL,
		DefaultModel: "claude-3-5-sonnet",
	})

	result, err := provider.Complete(context.Background(), "test prompt", CompletionOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response != `{"is_pii": false}` {
		t.Errorf("response: got %q", result.Response)
	}
	if result.TokensUsed != 48 {
		t.Errorf("tokens: got %d, want 48", result.TokensUsed)
	}
}

// TestGenericHTTPProvider_Complete tests the generic provider with template + JSONPath.
func TestGenericHTTPProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"result": map[string]any{
				"text": "detected PII: email",
			},
			"metadata": map[string]any{
				"tokens": 25,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewGenericHTTPProvider(ProviderConfig{
		Name:                "test-generic",
		Type:                ProviderTypeGenericHTTP,
		Endpoint:            server.URL,
		DefaultModel:        "custom-v1",
		RequestBodyTemplate: `{"text": "{{.Prompt}}", "max_length": {{.MaxTokens}}}`,
		ResponseContentPath: "result.text",
		ResponseTokensPath:  "metadata.tokens",
	})

	result, err := provider.Complete(context.Background(), "test prompt", CompletionOptions{MaxTokens: 256})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response != "detected PII: email" {
		t.Errorf("response: got %q", result.Response)
	}
	if result.TokensUsed != 25 {
		t.Errorf("tokens: got %d, want 25", result.TokensUsed)
	}
}

func TestExtractJSONPath(t *testing.T) {
	data := map[string]any{
		"choices": []any{
			map[string]any{
				"message": map[string]any{
					"content": "hello world",
				},
			},
		},
		"usage": map[string]any{
			"total_tokens": float64(42),
		},
	}

	tests := []struct {
		path string
		want string
	}{
		{"choices.0.message.content", "hello world"},
		{"usage.total_tokens", "42"},
	}

	for _, tt := range tests {
		got := extractJSONPath(data, tt.path)
		if got != tt.want {
			t.Errorf("extractJSONPath(%q): got %q, want %q", tt.path, got, tt.want)
		}
	}
}

// TestRegistry tests provider registration and retrieval.
func TestRegistry(t *testing.T) {
	reg := NewRegistry()

	cfg := ProviderConfig{Name: "test", DefaultModel: "m1"}
	provider := NewOpenAICompatProvider(cfg)
	reg.Register("test", provider, cfg)

	// Get
	got := reg.Get("test")
	if got == nil {
		t.Fatal("expected provider, got nil")
	}
	if got.Name() != "test" {
		t.Errorf("name: got %q", got.Name())
	}

	// GetConfig
	gotCfg, ok := reg.GetConfig("test")
	if !ok {
		t.Fatal("expected config found")
	}
	if gotCfg.DefaultModel != "m1" {
		t.Errorf("model: got %q", gotCfg.DefaultModel)
	}

	// List
	names := reg.List()
	if len(names) != 1 || names[0] != "test" {
		t.Errorf("list: got %v", names)
	}

	// Remove
	reg.Remove("test")
	if reg.Get("test") != nil {
		t.Error("expected nil after remove")
	}
}

func TestBuildFromConfig(t *testing.T) {
	configs := []ProviderConfig{
		{Name: "oai", Type: ProviderTypeOpenAICompatible, DefaultModel: "gpt-4o"},
		{Name: "ant", Type: ProviderTypeAnthropic, DefaultModel: "claude-3"},
		{Name: "gen", Type: ProviderTypeGenericHTTP, DefaultModel: "custom"},
	}

	for _, cfg := range configs {
		p, err := BuildFromConfig(cfg)
		if err != nil {
			t.Fatalf("BuildFromConfig(%s): %v", cfg.Name, err)
		}
		if p.Name() != cfg.Name {
			t.Errorf("name: got %q, want %q", p.Name(), cfg.Name)
		}
	}

	// Unknown type should error
	_, err := BuildFromConfig(ProviderConfig{Type: "unknown"})
	if err == nil {
		t.Error("expected error for unknown type")
	}
}
