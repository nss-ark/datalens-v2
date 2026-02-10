// Package middleware provides HTTP middleware for authentication,
// tenant isolation, rate limiting, and other cross-cutting concerns.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// Context keys for authenticated request data.
type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyTenantID contextKey = "tenant_id"
	ContextKeyEmail    contextKey = "email"
	ContextKeyName     contextKey = "name"
)

// UserIDFromContext extracts the user ID from the request context.
func UserIDFromContext(ctx context.Context) (types.ID, bool) {
	id, ok := ctx.Value(ContextKeyUserID).(types.ID)
	return id, ok
}

// TenantIDFromContext extracts the tenant ID from the request context.
func TenantIDFromContext(ctx context.Context) (types.ID, bool) {
	id, ok := ctx.Value(ContextKeyTenantID).(types.ID)
	return id, ok
}

// Auth returns middleware that validates JWT tokens and sets user context.
func Auth(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header")
				return
			}

			claims, err := authSvc.ValidateToken(token)
			if err != nil {
				httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
				return
			}

			// Set authenticated user data on context
			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyTenantID, claims.TenantID)
			ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
			ctx = context.WithValue(ctx, ContextKeyName, claims.Name)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission returns middleware that checks if the user has the required permission.
// This is a placeholder that checks context; full RBAC can be wired later.
func RequirePermission(resource string, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For now, all authenticated users pass permission checks.
			// TODO: Load user roles and check permissions against resource+action.
			_, ok := UserIDFromContext(r.Context())
			if !ok {
				httputil.ErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
