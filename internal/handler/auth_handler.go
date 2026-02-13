package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// AuthHandler handles authentication REST endpoints.
type AuthHandler struct {
	authSvc   *service.AuthService
	tenantSvc *service.TenantService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc *service.AuthService, tenantSvc *service.TenantService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, tenantSvc: tenantSvc}
}

// Routes returns a chi.Router with auth routes.
// Note: these routes are public (no auth middleware).
func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.RefreshToken)
	return r
}

// ProtectedRoutes returns routes that require authentication.
func (h *AuthHandler) ProtectedRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/me", h.Me)
	return r
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantName string `json:"tenant_name"`
		Domain     string `json:"domain"`
		Industry   string `json:"industry"`
		Country    string `json:"country"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		Password   string `json:"password"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	result, err := h.tenantSvc.Onboard(r.Context(), service.OnboardInput{
		TenantName: req.TenantName,
		Domain:     req.Domain,
		Industry:   req.Industry,
		Country:    req.Country,
		AdminEmail: req.Email,
		AdminName:  req.Name,
		AdminPass:  req.Password,
	})
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain   string `json:"domain"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	// If domain is provided, resolve tenant. If not, pass empty tenant ID to service/repo.
	var tenantID types.ID
	if req.Domain != "" {
		tenant, err := h.tenantSvc.GetByDomain(r.Context(), req.Domain)
		if err != nil {
			httputil.ErrorResponse(w, http.StatusNotFound, "TENANT_NOT_FOUND", "no tenant found for the given domain")
			return
		}
		tenantID = tenant.ID
	}

	pair, err := h.authSvc.Login(r.Context(), service.LoginInput{
		TenantID: tenantID,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, pair)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	pair, err := h.authSvc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, pair)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	user, err := h.authSvc.GetCurrentUser(r.Context(), userID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, user)
}
