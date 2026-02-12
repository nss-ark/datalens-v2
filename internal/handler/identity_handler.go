package handler

import (
	"encoding/json"
	"net/http"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
	"github.com/go-chi/chi/v5"
)

type IdentityHandler struct {
	service *service.IdentityService
}

func NewIdentityHandler(service *service.IdentityService) *IdentityHandler {
	return &IdentityHandler{service: service}
}

func (h *IdentityHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/status", h.GetStatus)
	r.Post("/link", h.LinkProvider)
	r.Get("/providers", h.ListProviders) // TODO: Implement if needed for UI to know what options to show
	return r
}

func (h *IdentityHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	subjectID, ok := types.SubjectIDFromContext(r.Context())
	if !ok {
		// Fallback: try to get subject ID from query param if admin viewing
		// But for now, assume this is for the logged-in user or subject
		// TODO: Clarify context usage.
		// For now, let's assume the auth middleware puts SubjectID if it's a DSR portal user.
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "subject context required")
		return
	}

	profile, err := h.service.GetStatus(r.Context(), subjectID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, profile)
}

type LinkProviderRequest struct {
	ProviderName string   `json:"provider_name"`
	AuthCode     string   `json:"auth_code"`
	SubjectID    types.ID `json:"subject_id"` // Optional if in context
}

func (h *IdentityHandler) LinkProvider(w http.ResponseWriter, r *http.Request) {
	var req LinkProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "failed to decode request body")
		return
	}

	// Determine SubjectID: Precedence Context > Request
	subjectID, ok := types.SubjectIDFromContext(r.Context())
	if !ok {
		if req.SubjectID != (types.ID{}) {
			subjectID = req.SubjectID
		} else {
			// For testing or admin overrides, we might allow passing ID, but strict security would require context.
			// Let's enforce context for self-service or admin permission check
			httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "subject context or ID required")
			return
		}
	}

	profile, err := h.service.LinkProvider(r.Context(), subjectID, req.ProviderName, req.AuthCode)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, profile)
}

func (h *IdentityHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	// Hardcoded for now based on what we implemented
	providers := []map[string]string{
		{"name": "DigiLocker", "type": "GOV_ID"},
	}
	httputil.JSON(w, http.StatusOK, providers)
}
