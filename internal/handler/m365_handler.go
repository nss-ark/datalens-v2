package handler

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
	"github.com/go-chi/chi/v5"
)

// M365Handler handles Microsoft 365 authentication requests.
type M365Handler struct {
	service *service.M365AuthService
}

// NewM365Handler creates a new M365Handler.
func NewM365Handler(service *service.M365AuthService) *M365Handler {
	return &M365Handler{service: service}
}

// Routes returns the router for M365 auth.
func (h *M365Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/connect", h.Connect)
	r.Get("/callback", h.Callback)
	return r
}

// Connect initiates the OAuth2 flow.
func (h *M365Handler) Connect(w http.ResponseWriter, r *http.Request) {
	// Generate a state parameter to prevent CSRF.
	// In a real app, store this in a cookie or Redis/Session with a short expiry.
	// For now, we'll base64 encode a random string and maybe the tenant ID?
	// To be stateless but secure, we could sign it, but for this task we'll keep it simple:
	// Just random bytes.
	// NOTE: If we need to know WHICH tenant initiated this, we should pass it or rely on the user being logged in
	// when they hit /callback. The callback handler needs to know the TenantID to create the datasource.
	// IF the callback comes from the browser, the Auth cookie (JWT) should still be there.
	// So we can extract TenantID from context in Callback.

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// We might want to store state in a secure cookie to verify later.
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/", // Should be specific path but for simplicity /
		HttpOnly: true,
		Secure:   r.TLS != nil, // Check if TLS
		Expires:  time.Now().Add(10 * time.Minute),
	})

	url := h.service.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Callback handles the OAuth2 callback.
func (h *M365Handler) Callback(w http.ResponseWriter, r *http.Request) {
	// Verify state
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_STATE", "missing state cookie")
		return
	}

	if r.URL.Query().Get("state") != cookie.Value {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_STATE", "state mismatch")
		return
	}

	// We need the TenantID to associate the data source.
	// The user SHOULD be authenticated.
	tenantID, ok := types.TenantIDFromContext(r.Context())
	if !ok {
		// If the user isn't authenticated (e.g. cookie expired), we can't accept this.
		// NOTE: The /callback endpoint must be protected by Auth Middleware for this to work.
		// If it's public, we can't get TenantID.
		// So we must ensure this route is mounted under Auth middleware.
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		httputil.ErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "authorization code missing")
		return
	}

	ds, err := h.service.ExchangeAndConnect(r.Context(), code, tenantID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	// success, redirect to frontend data sources page
	// Or return JSON if handled by a popup.
	// Usually for a full redirect flow, we redirect back to the app.
	// Let's redirect to /data-sources with a success query param.
	// Check config for frontend URL or assume it.
	// For now, simple JSON response or redirect?
	// Task says: "Upon successful auth... Store refresh_token".
	// The verified step says "Verify credentials column".
	// Let's return JSON for now so we can see the result, OR redirect.
	// A JSON response is easier to debug for the "Backend Task", but a redirect is better for UX.
	// Let's do a redirect to a success page or just return JSON if the user is calling this via API directly (unlikely for OAuth).
	// Actually, standard OAuth pattern: The frontend opens a window to /connect, which redirects to MS, which redirects to /callback.
	// The /callback returns a script to close the window and reload parent, OR redirects.

	// Let's just return JSON with the created DataSource for now, as it's an API task.
	httputil.JSON(w, http.StatusCreated, ds)
}
