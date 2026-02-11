package middleware

import (
	"context"
	"net/http"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

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
			// We use a private context key or simply make it accessible, but keeping it simple:
			// Just use the retrieved widget for CORS logic in the NEXT middleware if needed,
			// or handle CORS right here? The plan separates them. Let's stick to the plan request to verify origin.
			// Actually, let's inject the whole widget or just the allowed origins if separate middleware needs it.
			// For now, let's create a specific context key for the widget itself if needed, but ID + TenantID is usually enough.
			// Since CORS middleware needs AllowedOrigins, let's add a context key or fetch it again (bad for perf).
			// Efficient way: context key for the widget object.
			ctx = context.WithValue(ctx, contextKeyWidget, widget)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Private context key to avoid collisions
type widgetContextKey string

const contextKeyWidget widgetContextKey = "consent_widget"

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

			// If no origin (e.g. server-to-server or curl), we might allow or block.
			// For browser widgets, Origin is expected.
			// If allowed_origins is empty/loose, maybe allow all?
			// Spec says "AllowedOrigins TEXT[]". If empty, maybe deny all or allow all?
			// Let's assume strict: if defined, must match. If empty, maybe allow none (secure default).
			// But for a widget to work, it must have allowed origins.

			if origin != "" {
				allowed := false
				for _, o := range widget.AllowedOrigins {
					if o == "*" || o == origin {
						allowed = true
						break
					}
				}

				if !allowed {
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
