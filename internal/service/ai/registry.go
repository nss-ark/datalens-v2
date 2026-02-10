package ai

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// =============================================================================
// Provider Registry — Dynamic provider management
// =============================================================================

// Registry manages AI providers. Providers can be registered at startup
// from config, or added at runtime via the API. The registry is safe
// for concurrent access.
//
// Architecture:
//
//	┌──────────────────────────────────────────────────────────────────┐
//	│                     Provider Registry                            │
//	│                                                                  │
//	│  "openai"       → OpenAICompatProvider (api.openai.com)          │
//	│  "azure"        → OpenAICompatProvider (*.openai.azure.com)      │
//	│  "ollama"       → OpenAICompatProvider (localhost:11434)          │
//	│  "vllm"         → OpenAICompatProvider (your-vllm-server:8000)   │
//	│  "together"     → OpenAICompatProvider (api.together.xyz)        │
//	│  "groq"         → OpenAICompatProvider (api.groq.com)            │
//	│  "mistral"      → OpenAICompatProvider (api.mistral.ai)          │
//	│  "litellm"      → OpenAICompatProvider (your-litellm-proxy)      │
//	│  "anthropic"    → AnthropicProvider    (api.anthropic.com)        │
//	│  "huggingface"  → GenericHTTPProvider  (api-inference.hf.co)     │
//	│  "custom"       → GenericHTTPProvider  (your-custom-endpoint)    │
//	│  [any name]     → [any Provider impl]                            │
//	└──────────────────────────────────────────────────────────────────┘
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	configs   map[string]ProviderConfig
}

// NewRegistry creates an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		configs:   make(map[string]ProviderConfig),
	}
}

// Register adds a provider to the registry. If a provider with the same
// name already exists, it will be replaced.
func (r *Registry) Register(name string, provider Provider, config ProviderConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
	r.configs[name] = config
}

// Get returns a provider by name. Returns nil if not found.
func (r *Registry) Get(name string) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[name]
}

// GetConfig returns the config for a provider. Returns empty config if not found.
func (r *Registry) GetConfig(name string) (ProviderConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cfg, ok := r.configs[name]
	return cfg, ok
}

// List returns all registered provider names, sorted alphabetically.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListAvailable returns names of providers that are currently available
// (reachable, within budget, API key set, etc.).
func (r *Registry) ListAvailable() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var available []string
	for name, p := range r.providers {
		if p.IsAvailable(context.Background()) {
			available = append(available, name)
		}
	}
	sort.Strings(available)
	return available
}

// Remove removes a provider from the registry.
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, name)
	delete(r.configs, name)
}

// =============================================================================
// Registry Factory — Build providers from config
// =============================================================================

// BuildFromConfig creates a provider from a ProviderConfig.
// The config's Type field determines which implementation to use.
func BuildFromConfig(cfg ProviderConfig) (Provider, error) {
	switch cfg.Type {
	case ProviderTypeOpenAICompatible:
		return NewOpenAICompatProvider(cfg), nil
	case ProviderTypeAnthropic:
		return NewAnthropicProvider(cfg), nil
	case ProviderTypeGenericHTTP:
		return NewGenericHTTPProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %q (use openai_compatible, anthropic, or generic_http)", cfg.Type)
	}
}

// BuildRegistryFromConfig creates a registry populated from a slice of configs.
func BuildRegistryFromConfig(configs []ProviderConfig) (*Registry, error) {
	registry := NewRegistry()
	for _, cfg := range configs {
		provider, err := BuildFromConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("building provider %q: %w", cfg.Name, err)
		}
		registry.Register(cfg.Name, provider, cfg)
	}
	return registry, nil
}
