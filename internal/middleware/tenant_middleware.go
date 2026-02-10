package middleware

import (
	"net/http"

	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// TenantIsolation ensures every request has a valid tenant context.
// It reads the tenant ID set by Auth middleware and rejects requests
// without one on protected routes.
func TenantIsolation() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := types.TenantIDFromContext(r.Context())
			if !ok {
				httputil.ErrorResponse(w, http.StatusForbidden, "TENANT_REQUIRED", "tenant context is required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
