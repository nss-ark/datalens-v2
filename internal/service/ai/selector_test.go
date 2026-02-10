package ai

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSelector_CompleteWithFallback_FirstSucceeds(t *testing.T) {
	// Mock server for the primary provider
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []openAIChoice{
				{Message: openAIMessage{Content: "response from primary"}},
			},
			Usage: openAIUsage{TotalTokens: 10},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	reg := NewRegistry()
	reg.Register("primary", NewOpenAICompatProvider(ProviderConfig{
		Name:         "primary",
		APIKey:       "key",
		Endpoint:     server.URL,
		DefaultModel: "test-model",
	}), ProviderConfig{})

	sel := NewSelector(reg, []string{"primary"}, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	result, err := sel.CompleteWithFallback(context.Background(), "test", CompletionOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response != "response from primary" {
		t.Errorf("response: got %q", result.Response)
	}
}

func TestSelector_CompleteWithFallback_FallsBack(t *testing.T) {
	// Primary returns error, fallback succeeds
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer badServer.Close()

	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []openAIChoice{
				{Message: openAIMessage{Content: "response from fallback"}},
			},
			Usage: openAIUsage{TotalTokens: 5},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer goodServer.Close()

	reg := NewRegistry()
	reg.Register("primary", NewOpenAICompatProvider(ProviderConfig{
		Name: "primary", APIKey: "key", Endpoint: badServer.URL, DefaultModel: "m",
	}), ProviderConfig{})
	reg.Register("fallback", NewOpenAICompatProvider(ProviderConfig{
		Name: "fallback", APIKey: "key", Endpoint: goodServer.URL, DefaultModel: "m",
	}), ProviderConfig{})

	sel := NewSelector(reg, []string{"primary", "fallback"}, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	result, err := sel.CompleteWithFallback(context.Background(), "test", CompletionOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response != "response from fallback" {
		t.Errorf("expected fallback response, got %q", result.Response)
	}
}

func TestSelector_CompleteWithFallback_AllFail(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer badServer.Close()

	reg := NewRegistry()
	reg.Register("only", NewOpenAICompatProvider(ProviderConfig{
		Name: "only", APIKey: "key", Endpoint: badServer.URL, DefaultModel: "m",
	}), ProviderConfig{})

	sel := NewSelector(reg, []string{"only"}, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	_, err := sel.CompleteWithFallback(context.Background(), "test", CompletionOptions{})
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestSelector_UseCasePreference(t *testing.T) {
	// Set up two providers, preferred one for "pii_detection"
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []openAIChoice{{Message: openAIMessage{Content: "from-preferred"}}},
			Usage:   openAIUsage{TotalTokens: 1},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []openAIChoice{{Message: openAIMessage{Content: "from-default"}}},
			Usage:   openAIUsage{TotalTokens: 1},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server2.Close()

	reg := NewRegistry()
	reg.Register("preferred", NewOpenAICompatProvider(ProviderConfig{
		Name: "preferred", APIKey: "k", Endpoint: server1.URL, DefaultModel: "m",
	}), ProviderConfig{})
	reg.Register("default", NewOpenAICompatProvider(ProviderConfig{
		Name: "default", APIKey: "k", Endpoint: server2.URL, DefaultModel: "m",
	}), ProviderConfig{})

	sel := NewSelector(reg, []string{"default", "preferred"}, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	sel.SetUseCasePreference("pii_detection", "preferred")

	// PII detection should go to preferred first
	result, err := sel.CompleteWithFallback(context.Background(), "test", CompletionOptions{
		UseCase: "pii_detection",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response != "from-preferred" {
		t.Errorf("expected preferred provider, got %q", result.Response)
	}
}

func TestSelector_EmptyChain(t *testing.T) {
	reg := NewRegistry()
	sel := NewSelector(reg, []string{}, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	_, err := sel.CompleteWithFallback(context.Background(), "test", CompletionOptions{})
	if err == nil {
		t.Fatal("expected error with empty chain")
	}
}
