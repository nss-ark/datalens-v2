package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// PortalAuthMiddleware handles JWT validation for portal requests.
type PortalAuthMiddleware struct {
	authService *service.PortalAuthService
	profileRepo consent.DataPrincipalProfileRepository
}

// NewPortalAuthMiddleware creates a new PortalAuthMiddleware.
func NewPortalAuthMiddleware(authService *service.PortalAuthService, profileRepo consent.DataPrincipalProfileRepository) *PortalAuthMiddleware {
	return &PortalAuthMiddleware{authService: authService, profileRepo: profileRepo}
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

		// Inject principal_id and tenant_id into context
		ctx := context.WithValue(r.Context(), types.ContextKeyPrincipalID, claims.PrincipalID)
		ctx = context.WithValue(ctx, types.ContextKeyTenantID, claims.TenantID)

		// Resolve subject_id from profile for downstream handlers (e.g. grievance listing)
		if m.profileRepo != nil {
			if profile, err := m.profileRepo.GetByID(ctx, claims.PrincipalID); err == nil && profile.SubjectID != nil {
				ctx = context.WithValue(ctx, types.ContextKeySubjectID, *profile.SubjectID)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
