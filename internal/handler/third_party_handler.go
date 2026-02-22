package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
)

// ThirdPartyHandler handles HTTP requests for third-party CRUD.
type ThirdPartyHandler struct {
	service *service.ThirdPartyService
}

// NewThirdPartyHandler creates a new ThirdPartyHandler.
func NewThirdPartyHandler(s *service.ThirdPartyService) *ThirdPartyHandler {
	return &ThirdPartyHandler{service: s}
}

// Routes returns a chi.Router with third-party routes.
// Mounted at /api/v2/third-parties.
func (h *ThirdPartyHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// Create handles POST /api/v2/third-parties.
func (h *ThirdPartyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateThirdPartyRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	tp, err := h.service.Create(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, tp)
}

// List handles GET /api/v2/third-parties.
func (h *ThirdPartyHandler) List(w http.ResponseWriter, r *http.Request) {
	thirdParties, err := h.service.List(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, thirdParties)
}

// GetByID handles GET /api/v2/third-parties/{id}.
func (h *ThirdPartyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	tp, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, tp)
}

// Update handles PUT /api/v2/third-parties/{id}.
func (h *ThirdPartyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req service.UpdateThirdPartyRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	tp, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, tp)
}

// Delete handles DELETE /api/v2/third-parties/{id}.
func (h *ThirdPartyHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
