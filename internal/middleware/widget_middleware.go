package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// Private context key to avoid collisions
type widgetContextKey string

const contextKeyWidget widgetContextKey = "consent_widget"

// WidgetAuthMiddleware creates a middleware that authenticates consent widgets.
func WidgetAuthMiddleware(widgetRepo consent.ConsentWidgetRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API Key
			apiKey := r.Header.Get("X-Widget-Key")
			if apiKey == "" {
				httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing widget api key")
				return
			}

			// Validate API Key
			widget, err := widgetRepo.GetByAPIKey(r.Context(), apiKey)
			if err != nil {
				if types.IsNotFoundError(err) {
					httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid widget api key")
					return
				}
				httputil.ErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to authenticate widget")
				return
			}

			// Check status
			if widget.Status != consent.WidgetStatusActive {
				httputil.ErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "widget is not active")
				return
			}

			// Inject context
			ctx := context.WithValue(r.Context(), types.ContextKeyTenantID, widget.TenantID)
			ctx = context.WithValue(ctx, types.ContextKeyWidgetID, widget.ID)

			// Store widget for CORS check
			ctx = context.WithValue(ctx, contextKeyWidget, widget)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WidgetCORSMiddleware validates the Origin header against the widget's allowed origins.
// Must be placed AFTER WidgetAuthMiddleware.
func WidgetCORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Retrieve widget from context (set by auth middleware)
			widget, ok := r.Context().Value(contextKeyWidget).(*consent.ConsentWidget)
			if !ok {
				// Should not happen if wired correctly
				httputil.ErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "widget context missing")
				return
			}

			origin := r.Header.Get("Origin")

			// If no origin (e.g. server-to-server or curl), we might block or allow depending on security policy.
			// For public widgets, browsers always send Origin.
			// If missing, we assume it's a direct API call (e.g. Postman) which is allowed if API key is valid.
			// But for strict security, we might enforce Origin for browser-based endpoints.
			// Let's allow empty Origin but don't set CORS headers.
			if origin != "" {
				if !isOriginAllowed(origin, widget.AllowedOrigins) {
					httputil.ErrorResponse(w, http.StatusForbidden, "CORS_ERROR", "origin not allowed")
					return
				}

				// Set CORS headers
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Widget-Key")
			}

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if the origin matches any of the allowed patterns.
// Supports wildcards (e.g., "*.example.com").
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, pattern := range allowedOrigins {
		if pattern == "*" {
			return true
		}
		if pattern == origin {
			return true
		}
		if strings.HasPrefix(pattern, "*.") {
			suffix := pattern[1:] // e.g., ".example.com"
			if strings.HasSuffix(origin, suffix) {
				return true
			}
		}
	}
	return false
}
