package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// GoogleHandler handles Google Workspace authentication endpoints.
type GoogleHandler struct {
	authSvc *service.GoogleAuthService
}

// NewGoogleHandler creates a new GoogleHandler.
func NewGoogleHandler(authSvc *service.GoogleAuthService) *GoogleHandler {
	return &GoogleHandler{authSvc: authSvc}
}

// Routes returns a chi.Router with Google auth routes.
func (h *GoogleHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/connect", h.Connect)
	r.Get("/callback", h.Callback)
	return r
}

// Connect initiates the OAuth2 flow.
func (h *GoogleHandler) Connect(w http.ResponseWriter, r *http.Request) {
	// Generate random state or use tenant ID if needed for correlation
	// Ideally state should be cryptographically random and stored in session/cookie to prevent CSRF.
	// For now, we'll use a simple "state" or tenant ID if context available.
	// But /connect might be called from frontend which should provide state or we generate it.

	state := "state-token" // TODO: Implement proper state handling
	url := h.authSvc.GetAuthURL(state)

	httputil.JSON(w, http.StatusOK, map[string]string{"url": url})
}

// Callback handles the OAuth2 callback.
func (h *GoogleHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "code is required")
		return
	}

	tenantID, ok := types.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "tenant context required")
		return
	}

	ds, err := h.authSvc.ExchangeAndConnect(r.Context(), code, tenantID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, ds)
}
