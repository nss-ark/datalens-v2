package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/complyark/datalens/pkg/types"
)

// ConsentCache defines the interface for caching consent decisions.
type ConsentCache interface {
	// GetConsentStatus retrieves the cached consent status.
	// Returns:
	// - true: Consent granted
	// - false: Consent not granted (withdrawn or never granted)
	// - nil: Cache miss (entry not found)
	// - error: Redis error
	GetConsentStatus(ctx context.Context, tenantID, subjectID, purposeID types.ID) (*bool, error)

	// SetConsentStatus caches the consent status.
	SetConsentStatus(ctx context.Context, tenantID, subjectID, purposeID types.ID, granted bool, ttl time.Duration) error

	// InvalidateSubject removes all consent entries for a subject (wildcard delete).
	// Note: Redis SCAN is slow for many keys; use with caution or consider Sets.
	// For now, we'll implement a specific key invalidation or discuss strategy.
	// Actually, the requirement asks for "InvalidateSubject". valid keys are specific to purpose.
	// We might need to store a set of purposes for a subject to invalidate efficiently.
	// IMPROVEMENT: Use a Set to track purposes for a subject, or just SCAN if volume is low.
	// Given <50ms requirement for *reads*, specific invalidation is better.
	// For this iteration, we'll assume we can construct the key or use a pattern.
	// Pattern: consent:{tenant}:{subject}:*
	InvalidateSubject(ctx context.Context, tenantID, subjectID types.ID) error

	// InvalidateAll flushes the cache for a tenant (dangerous but needed for key rotation etc).
	InvalidateAll(ctx context.Context, tenantID types.ID) error
}

// RedisConsentCache implements ConsentCache using Redis.
type RedisConsentCache struct {
	client *redis.Client
}

// NewRedisConsentCache creates a new RedisConsentCache.
func NewRedisConsentCache(client *redis.Client) *RedisConsentCache {
	return &RedisConsentCache{client: client}
}

func (c *RedisConsentCache) key(tenantID, subjectID, purposeID types.ID) string {
	return fmt.Sprintf("consent:%s:%s:%s", tenantID, subjectID, purposeID)
}

func (c *RedisConsentCache) GetConsentStatus(ctx context.Context, tenantID, subjectID, purposeID types.ID) (*bool, error) {
	val, err := c.client.Get(ctx, c.key(tenantID, subjectID, purposeID)).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	granted := val == "1"
	return &granted, nil
}

func (c *RedisConsentCache) SetConsentStatus(ctx context.Context, tenantID, subjectID, purposeID types.ID, granted bool, ttl time.Duration) error {
	val := "0"
	if granted {
		val = "1"
	}
	return c.client.Set(ctx, c.key(tenantID, subjectID, purposeID), val, ttl).Err()
}

func (c *RedisConsentCache) InvalidateSubject(ctx context.Context, tenantID, subjectID types.ID) error {
	// Pattern scan for consent:{tenant}:{subject}:*
	pattern := fmt.Sprintf("consent:%s:%s:*", tenantID, subjectID)
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (c *RedisConsentCache) InvalidateAll(ctx context.Context, tenantID types.ID) error {
	pattern := fmt.Sprintf("consent:%s:*", tenantID)
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
