package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
)

// DepartmentHandler handles HTTP requests for department CRUD and notifications.
type DepartmentHandler struct {
	service *service.DepartmentService
}

// NewDepartmentHandler creates a new DepartmentHandler.
func NewDepartmentHandler(s *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: s}
}

// Routes returns a chi.Router with department routes.
// Mounted at /api/v2/departments.
func (h *DepartmentHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/notify", h.Notify)
	return r
}

// Create handles POST /api/v2/departments.
func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateDepartmentRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	dept, err := h.service.Create(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, dept)
}

// List handles GET /api/v2/departments.
func (h *DepartmentHandler) List(w http.ResponseWriter, r *http.Request) {
	departments, err := h.service.List(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, departments)
}

// GetByID handles GET /api/v2/departments/{id}.
func (h *DepartmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	dept, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, dept)
}

// Update handles PUT /api/v2/departments/{id}.
func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req service.UpdateDepartmentRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	dept, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, dept)
}

// Delete handles DELETE /api/v2/departments/{id}.
func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

// Notify handles POST /api/v2/departments/{id}/notify.
func (h *DepartmentHandler) Notify(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req service.NotifyDepartmentRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if err := h.service.Notify(r.Context(), id, req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"message": "notification sent"})
}
