package handler

import (
	"net/http"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
	"github.com/go-chi/chi/v5"
)

// DPOHandler handles HTTP requests for DPO contact operations.
type DPOHandler struct {
	service *service.DPOService
}

// NewDPOHandler creates a new DPOHandler.
func NewDPOHandler(service *service.DPOService) *DPOHandler {
	return &DPOHandler{service: service}
}

// Routes returns the private API routes for DPO contact management.
// Mounted at /api/v2/compliance/dpo (or similar).
func (h *DPOHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Get)
	r.Put("/", h.Upsert)
	return r
}

// PublicRoutes returns the public API routes for DPO contact display.
// Mounted at /api/public/compliance/dpo.
func (h *DPOHandler) PublicRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetPublic)
	return r
}

// Get returns the DPO contact for the authenticated tenant.
func (h *DPOHandler) Get(w http.ResponseWriter, r *http.Request) {
	contact, err := h.service.GetContact(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, contact)
}

// Upsert creates or updates the DPO contact for the authenticated tenant.
func (h *DPOHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req service.UpsertDPOContactRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	contact, err := h.service.UpsertContact(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, contact)
}

// GetPublic returns the DPO contact for a specific tenant (public).
// Requires ?tenant_id=... query parameter.
func (h *DPOHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.URL.Query().Get("tenant_id")
	if tenantIDStr == "" {
		httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "tenant_id is required")
		return
	}

	tenantID, err := types.ParseID(tenantIDStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid tenant_id format")
		return
	}

	contact, err := h.service.GetPublicContact(r.Context(), tenantID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, contact)
}
