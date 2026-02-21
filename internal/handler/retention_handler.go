package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
)

// RetentionHandler handles HTTP requests for retention policy CRUD.
type RetentionHandler struct {
	service *service.RetentionService
}

// NewRetentionHandler creates a new RetentionHandler.
func NewRetentionHandler(s *service.RetentionService) *RetentionHandler {
	return &RetentionHandler{service: s}
}

// Routes returns a chi.Router with retention policy routes.
// Mounted at /api/v2/retention.
func (h *RetentionHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Get("/{id}/logs", h.GetLogs)
	return r
}

// Create handles POST /api/v2/retention.
func (h *RetentionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateRetentionPolicyRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	policy, err := h.service.Create(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, policy)
}

// List handles GET /api/v2/retention.
func (h *RetentionHandler) List(w http.ResponseWriter, r *http.Request) {
	policies, err := h.service.GetByTenant(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, policies)
}

// GetByID handles GET /api/v2/retention/{id}.
func (h *RetentionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	policy, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, policy)
}

// Update handles PUT /api/v2/retention/{id}.
func (h *RetentionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req service.UpdateRetentionPolicyRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	policy, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, policy)
}

// Delete handles DELETE /api/v2/retention/{id}.
func (h *RetentionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetLogs handles GET /api/v2/retention/{id}/logs.
func (h *RetentionHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	pagination := httputil.ParsePagination(r)

	result, err := h.service.GetLogs(r.Context(), &id, pagination)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSONWithPagination(w, result.Items, pagination.Page, pagination.PageSize, result.Total)
}
