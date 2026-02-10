package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/pkg/types"
)

// CachedGateway wraps an AI Gateway with Redis caching and token budgeting.
type CachedGateway struct {
	next   Gateway
	redis  *redis.Client
	logger *slog.Logger
	cfg    config.AIConfig
}

// NewCachedGateway creates a new CachedGateway decorator.
func NewCachedGateway(next Gateway, rdb *redis.Client, logger *slog.Logger, cfg config.AIConfig) *CachedGateway {
	return &CachedGateway{
		next:   next,
		redis:  rdb,
		logger: logger.With("component", "ai_cache"),
		cfg:    cfg,
	}
}

// =============================================================================
// Gateway Interface Implementation
// =============================================================================

func (g *CachedGateway) DetectPII(ctx context.Context, input PIIDetectionInput) (*PIIDetectionResult, error) {
	if err := g.checkBudget(ctx); err != nil {
		return nil, err
	}

	key := g.cacheKey("pii", input)
	if cached, err := g.getFromCache(ctx, key, &PIIDetectionResult{}); err == nil {
		// On cache hit, we don't track *new* token usage against the budget,
		// but we might want to log it or track "saved" tokens.
		// For now, we return the cached result immediately.
		cached.(*PIIDetectionResult).Duration = 0 // Cached response is instant
		cached.(*PIIDetectionResult).Provider = cached.(*PIIDetectionResult).Provider + " (cached)"
		return cached.(*PIIDetectionResult), nil
	}

	result, err := g.next.DetectPII(ctx, input)
	if err != nil {
		return nil, err
	}

	// Cache successful results (24h default)
	if err := g.setToCache(ctx, key, result, 24*time.Hour); err != nil {
		g.logger.WarnContext(ctx, "failed to cache pii result", "error", err)
	}

	// Track usage
	if err := g.trackUsage(ctx, result.TokensUsed); err != nil {
		g.logger.WarnContext(ctx, "failed to track token usage", "error", err)
	}

	return result, nil
}

func (g *CachedGateway) SuggestPurposes(ctx context.Context, input PurposeSuggestionInput) ([]PurposeSuggestion, error) {
	if err := g.checkBudget(ctx); err != nil {
		return nil, err
	}

	key := g.cacheKey("purposes", input)
	var cachedSuggestions []PurposeSuggestion
	if cached, err := g.getFromCache(ctx, key, &cachedSuggestions); err == nil {
		return *cached.(*[]PurposeSuggestion), nil
	}

	result, err := g.next.SuggestPurposes(ctx, input)
	if err != nil {
		return nil, err
	}

	// Cache successful results
	if len(result) > 0 {
		if err := g.setToCache(ctx, key, result, 48*time.Hour); err != nil {
			g.logger.WarnContext(ctx, "failed to cache purpose suggestions", "error", err)
		}
	}

	// Purpose suggestion token usage is harder to track precisely if not returned by next layer
	// Assuming a fixed cost or tracking if available in future interface updates.
	// For now, we skip explicit token tracking for purposes as it's usually low volume.

	return result, nil
}

func (g *CachedGateway) Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error) {
	if err := g.checkBudget(ctx); err != nil {
		return nil, err
	}

	// Calculate cache key based on prompt + options
	cacheInput := struct {
		Prompt string            `json:"prompt"`
		Opts   CompletionOptions `json:"opts"`
	}{Prompt: prompt, Opts: opts}
	key := g.cacheKey("completion", cacheInput)

	if opts.CacheTTL > 0 {
		if cached, err := g.getFromCache(ctx, key, &CompletionResult{}); err == nil {
			res := cached.(*CompletionResult)
			res.Cached = true
			return res, nil
		}
	}

	result, err := g.next.Complete(ctx, prompt, opts)
	if err != nil {
		return nil, err
	}

	if opts.CacheTTL > 0 {
		if err := g.setToCache(ctx, key, result, opts.CacheTTL); err != nil {
			g.logger.WarnContext(ctx, "failed to cache completion", "error", err)
		}
	}

	if err := g.trackUsage(ctx, result.TokensUsed); err != nil {
		g.logger.WarnContext(ctx, "failed to track token usage", "error", err)
	}

	return result, nil
}

// =============================================================================
// Internals: Caching
// =============================================================================

func (g *CachedGateway) cacheKey(prefix string, input any) string {
	hash := sha256.New()
	if err := json.NewEncoder(hash).Encode(input); err != nil {
		// Should rarely happen with basic structs; fallback to random/timestamp to avoid collision on error
		return fmt.Sprintf("ai:%s:err:%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("ai:%s:%s", prefix, hex.EncodeToString(hash.Sum(nil)))
}

func (g *CachedGateway) getFromCache(ctx context.Context, key string, target any) (any, error) {
	val, err := g.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(val), target); err != nil {
		return nil, err
	}
	return target, nil
}

func (g *CachedGateway) setToCache(ctx context.Context, key string, value any, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return g.redis.Set(ctx, key, bytes, ttl).Err()
}

// =============================================================================
// Internals: Budgeting
// =============================================================================

func (g *CachedGateway) trackUsage(ctx context.Context, tokens int) error {
	if tokens <= 0 {
		return nil
	}
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil // System call or no tenant context
	}

	today := time.Now().Format("20060102")
	key := fmt.Sprintf("tenant:%s:ai:tokens:%s", tenantID, today)

	// Increment daily usage (expire after 30 days for reporting)
	pipe := g.redis.Pipeline()
	pipe.IncrBy(ctx, key, int64(tokens))
	pipe.Expire(ctx, key, 30*24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

func (g *CachedGateway) checkBudget(ctx context.Context) error {
	tenantID, ok := types.TenantIDFromContext(ctx)
	if !ok {
		return nil
	}

	// 1. Check if tenant has a custom budget set
	budgetKey := fmt.Sprintf("tenant:%s:ai:budget", tenantID)
	budgetStr, err := g.redis.Get(ctx, budgetKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil // No specific budget set, allow (or fallback to global plan limit)
	}
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to check budget", "error", err)
		return nil // Fail open on Redis error
	}

	var budget int
	if _, err := fmt.Sscanf(budgetStr, "%d", &budget); err != nil {
		return nil
	}

	// 2. Check usage today (simplistic budget = daily limit for now)
	// Real implementation might aggregate monthly.
	today := time.Now().Format("20060102")
	usageKey := fmt.Sprintf("tenant:%s:ai:tokens:%s", tenantID, today)
	usage, err := g.redis.Get(ctx, usageKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		g.logger.ErrorContext(ctx, "failed to get usage", "error", err)
		return nil
	}

	if usage >= budget {
		return types.NewDomainError(
			types.ErrCodeQuotaExceeded,
			fmt.Sprintf("AI token budget exceeded (%d/%d)", usage, budget),
		)
	}

	return nil
}
