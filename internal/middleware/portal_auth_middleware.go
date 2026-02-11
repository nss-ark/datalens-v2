package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// PortalAuthMiddleware handles JWT validation for portal requests.
type PortalAuthMiddleware struct {
	authService *service.PortalAuthService
}

// NewPortalAuthMiddleware creates a new PortalAuthMiddleware.
func NewPortalAuthMiddleware(authService *service.PortalAuthService) *PortalAuthMiddleware {
	return &PortalAuthMiddleware{authService: authService}
}

// PortalJWTAuth middleware validates the Bearer token and injects claims into context.
func (m *PortalAuthMiddleware) PortalJWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "authorization header required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header format")
			return
		}

		tokenString := parts[1]
		claims, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			return
		}

		// Inject configuration into context
		ctx := context.WithValue(r.Context(), types.ContextKey("principal_id"), claims.PrincipalID)
		ctx = context.WithValue(ctx, types.ContextKeyTenantID, claims.TenantID)
		// We use types.ContextKeyTenantID because standard handlers might rely on it,
		// but typically portal handlers should be specific.
		// However, for TenantID, using the standard key allows reuse of multi-tenant logic if needed.

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
