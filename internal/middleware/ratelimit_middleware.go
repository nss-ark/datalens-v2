package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// RateLimiter provides per-tenant rate limiting using an in-memory
// token bucket. For production with multiple instances, swap to Redis.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[types.ID]*bucket
	rate     int           // tokens per interval
	interval time.Duration // refill interval
	burst    int           // maximum bucket size
}

type bucket struct {
	tokens   int
	lastFill time.Time
}

// NewRateLimiter creates a new in-memory rate limiter.
// rate: requests allowed per interval. burst: max burst size.
func NewRateLimiter(rate int, interval time.Duration, burst int) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[types.ID]*bucket),
		rate:     rate,
		interval: interval,
		burst:    burst,
	}
}

// Middleware returns an HTTP middleware that enforces per-tenant rate limits.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, ok := TenantIDFromContext(r.Context())
			if !ok {
				// No tenant context = skip rate limiting (public routes)
				next.ServeHTTP(w, r)
				return
			}

			if !rl.allow(tenantID) {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.interval.Seconds())))
				httputil.ErrorResponse(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, please try again later")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) allow(tenantID types.ID) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[tenantID]
	now := time.Now()

	if !exists {
		rl.buckets[tenantID] = &bucket{tokens: rl.burst - 1, lastFill: now}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(b.lastFill)
	tokensToAdd := int(elapsed/rl.interval) * rl.rate
	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > rl.burst {
			b.tokens = rl.burst
		}
		b.lastFill = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}
