package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole_Allowed(t *testing.T) {
	// Setup request with PLATFORM_ADMIN role
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	roles := []identity.Role{
		{Name: "PLATFORM_ADMIN"},
	}
	ctx := context.WithValue(req.Context(), types.ContextKeyRoles, roles)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Handler that should be reached
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	mw := middleware.RequireRole("PLATFORM_ADMIN")
	handler := mw(nextHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireRole_Denied(t *testing.T) {
	// Setup request with only USER role
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	roles := []identity.Role{
		{Name: "USER"},
	}
	ctx := context.WithValue(req.Context(), types.ContextKeyRoles, roles)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.RequireRole("PLATFORM_ADMIN")
	handler := mw(nextHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireRole_NoRoles(t *testing.T) {
	// Setup request with NO roles
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	// No context value for roles

	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.RequireRole("PLATFORM_ADMIN")
	handler := mw(nextHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}
