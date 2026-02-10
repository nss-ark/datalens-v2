// Package middleware provides HTTP middleware for authentication,
// tenant isolation, rate limiting, and other cross-cutting concerns.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/complyark/datalens/internal/domain/identity"
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
	ContextKeyRoles    contextKey = "roles"
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

// RolesFromContext extracts the user's roles from the request context.
func RolesFromContext(ctx context.Context) []identity.Role {
	roles, _ := ctx.Value(ContextKeyRoles).([]identity.Role)
	return roles
}

// Auth returns middleware that validates JWT tokens or API keys and sets
// user/tenant context including roles for downstream permission checks.
func Auth(authSvc *service.AuthService, apiKeySvc *service.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try API key first (X-API-Key header)
			if apiKey := r.Header.Get("X-API-Key"); apiKey != "" && apiKeySvc != nil {
				tenantID, perms, err := apiKeySvc.ValidateKey(r.Context(), apiKey)
				if err != nil {
					httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired api key")
					return
				}

				ctx := r.Context()
				ctx = context.WithValue(ctx, ContextKeyTenantID, tenantID)
				// No UserID for API key auth — agents are not users

				// Convert permissions into a synthetic role for RequirePermission
				agentRole := identity.Role{
					Name:        "API_KEY_AGENT",
					Permissions: perms,
				}
				ctx = context.WithValue(ctx, ContextKeyRoles, []identity.Role{agentRole})

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Fall back to JWT Bearer token
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

			// Load user roles for RBAC (best-effort; missing roles = no permissions)
			roles, err := authSvc.GetUserRoles(ctx, claims.UserID)
			if err != nil {
				slog.Warn("failed to load user roles", "user_id", claims.UserID, "error", err)
				roles = nil
			}
			ctx = context.WithValue(ctx, ContextKeyRoles, roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission returns middleware that checks if the user has the
// required permission (resource + action) based on their assigned roles.
func RequirePermission(resource string, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles := RolesFromContext(r.Context())
			if !service.HasPermission(roles, resource, action) {
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
