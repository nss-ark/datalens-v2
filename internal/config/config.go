// Package config loads and validates application configuration
// from environment variables and configuration files.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the complete application configuration.
type Config struct {
	App       AppConfig
	DB        DatabaseConfig
	Redis     RedisConfig
	NATS      NATSConfig
	AI        AIConfig
	JWT       JWTConfig
	Agent     AgentConfig
	Consent   ConsentConfig
	Portal    PortalConfig
	Microsoft MicrosoftConfig
	Google    GoogleConfig
	Identity  IdentityConfig
}

// IdentityConfig holds identity provider settings.
type IdentityConfig struct {
	DigiLocker DigiLockerConfig
}

// DigiLockerConfig holds DigiLocker integration settings.
type DigiLockerConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// GoogleConfig holds Google Workspace integration settings.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// MicrosoftConfig holds Microsoft 365 integration settings.
type MicrosoftConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	TenantID     string // Common or specific tenant ID
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Env       string // development, staging, production
	Port      int
	LogLevel  string
	SecretKey string
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DSN returns the PostgreSQL data source name.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Addr returns the Redis address string.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// NATSConfig holds NATS connection settings.
type NATSConfig struct {
	URL       string
	ClusterID string
}

// AIConfig holds AI provider settings.
type AIConfig struct {
	DefaultProvider string
	OpenAI          OpenAIConfig
	Anthropic       AnthropicConfig
	LocalLLM        LocalLLMConfig
}

// OpenAIConfig holds OpenAI-specific settings.
type OpenAIConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// AnthropicConfig holds Anthropic-specific settings.
type AnthropicConfig struct {
	APIKey string
	Model  string
}

// LocalLLMConfig holds local LLM (Ollama) settings.
type LocalLLMConfig struct {
	Endpoint string
	Model    string
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// AgentConfig holds on-premise agent settings.
type AgentConfig struct {
	ID                    string
	APIKey                string
	ControlCentreEndpoint string
}

// ConsentConfig holds consent module settings.
type ConsentConfig struct {
	SigningKey string        // HMAC-SHA256 key for signing consent records
	CacheTTL   time.Duration // TTL for consent cache (default 300s)
}

// PortalConfig holds settings for the Data Principal Portal.
type PortalConfig struct {
	JWTSecret string
	JWTExpiry time.Duration
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Env:       getEnv("APP_ENV", "development"),
			Port:      getEnvInt("APP_PORT", 8080),
			LogLevel:  getEnv("APP_LOG_LEVEL", "debug"),
			SecretKey: getEnv("APP_SECRET_KEY", "change-me-in-prod"),
		},
		DB: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			Name:            getEnv("DB_NAME", "datalens"),
			User:            getEnv("DB_USER", "datalens"),
			Password:        getEnv("DB_PASSWORD", "datalens_dev"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		NATS: NATSConfig{
			URL:       getEnv("NATS_URL", "nats://localhost:4222"),
			ClusterID: getEnv("NATS_CLUSTER_ID", "datalens"),
		},
		AI: AIConfig{
			DefaultProvider: getEnv("AI_DEFAULT_PROVIDER", "local"),
			OpenAI: OpenAIConfig{
				APIKey:    getEnv("OPENAI_API_KEY", ""),
				Model:     getEnv("OPENAI_MODEL", "gpt-4o"),
				MaxTokens: getEnvInt("OPENAI_MAX_TOKENS", 4096),
			},
			Anthropic: AnthropicConfig{
				APIKey: getEnv("ANTHROPIC_API_KEY", ""),
				Model:  getEnv("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022"),
			},
			LocalLLM: LocalLLMConfig{
				Endpoint: getEnv("LOCAL_LLM_ENDPOINT", "http://localhost:11434"),
				Model:    getEnv("LOCAL_LLM_MODEL", "llama3.2"),
			},
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", getEnv("APP_SECRET_KEY", "change-me-in-prod")),
			AccessTokenExpiry:  getEnvDuration("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvDuration("JWT_REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		},
		Agent: AgentConfig{
			ID:                    getEnv("AGENT_ID", ""),
			APIKey:                getEnv("AGENT_API_KEY", ""),
			ControlCentreEndpoint: getEnv("CONTROL_CENTRE_ENDPOINT", "http://localhost:8080"),
		},
		Consent: ConsentConfig{
			SigningKey: getEnv("CONSENT_SIGNING_KEY", "dev-consent-signing-key-change-me"),
			CacheTTL:   getEnvDuration("CONSENT_CACHE_TTL_SECONDS", 300*time.Second),
		},
		Portal: PortalConfig{
			JWTSecret: getEnv("PORTAL_JWT_SECRET", "portal-secret-key-change-me-in-prod-32chars"),
			JWTExpiry: getEnvDuration("PORTAL_JWT_EXPIRY", 15*time.Minute),
		},
		Microsoft: MicrosoftConfig{
			ClientID:     getEnv("MICROSOFT_CLIENT_ID", ""),
			ClientSecret: getEnv("MICROSOFT_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("MICROSOFT_REDIRECT_URL", "http://localhost:8080/api/v2/auth/m365/callback"),
			TenantID:     getEnv("MICROSOFT_TENANT_ID", "common"),
		},
		Google: GoogleConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v2/auth/google/callback"),
		},
		Identity: IdentityConfig{
			DigiLocker: DigiLockerConfig{
				ClientID:     getEnv("DIGILOCKER_CLIENT_ID", ""),
				ClientSecret: getEnv("DIGILOCKER_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("DIGILOCKER_REDIRECT_URI", "http://localhost:8080/api/v2/identity/digilocker/callback"),
			},
		},
	}

	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	if c.App.Env == "production" && c.App.SecretKey == "change-me-in-prod" {
		return fmt.Errorf("APP_SECRET_KEY must be set in production")
	}
	return nil
}

// =============================================================================
// Helpers
// =============================================================================

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
