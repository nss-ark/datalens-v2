package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
)

// RoPAHandler handles HTTP requests for RoPA (Record of Processing Activities).
type RoPAHandler struct {
	service *service.RoPAService
}

// NewRoPAHandler creates a new RoPAHandler.
func NewRoPAHandler(s *service.RoPAService) *RoPAHandler {
	return &RoPAHandler{service: s}
}

// Routes returns a chi.Router with RoPA routes.
// Mounted at /api/v2/ropa.
func (h *RoPAHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Generate)            // Auto-generate new version
	r.Get("/", h.GetLatest)            // Latest version
	r.Get("/versions", h.ListVersions) // Paginated history
	r.Get("/versions/{version}", h.GetByVersion)
	r.Put("/", h.SaveEdit)             // User edit → new minor version
	r.Post("/publish", h.Publish)      // Mark as PUBLISHED
	r.Post("/promote", h.PromoteMajor) // Major version bump
	return r
}

// Generate handles POST /api/v2/ropa — auto-generates a new RoPA version.
func (h *RoPAHandler) Generate(w http.ResponseWriter, r *http.Request) {
	version, err := h.service.Generate(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, version)
}

// GetLatest handles GET /api/v2/ropa — returns the latest RoPA version.
func (h *RoPAHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	version, err := h.service.GetLatest(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	// If no versions exist, return 200 with nil data — frontend handles empty state
	httputil.JSON(w, http.StatusOK, version)
}

// ListVersions handles GET /api/v2/ropa/versions — paginated history.
func (h *RoPAHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	pagination := httputil.ParsePagination(r)

	result, err := h.service.ListVersions(r.Context(), pagination)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSONWithPagination(w, result.Items, pagination.Page, pagination.PageSize, result.Total)
}

// GetByVersion handles GET /api/v2/ropa/versions/{version} — specific version.
func (h *RoPAHandler) GetByVersion(w http.ResponseWriter, r *http.Request) {
	versionStr := chi.URLParam(r, "version")
	if versionStr == "" {
		httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "version parameter is required")
		return
	}

	version, err := h.service.GetByVersion(r.Context(), versionStr)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, version)
}

// SaveEdit handles PUT /api/v2/ropa — user edit creates a new minor version.
func (h *RoPAHandler) SaveEdit(w http.ResponseWriter, r *http.Request) {
	var req service.SaveEditRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	version, err := h.service.SaveEdit(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, version)
}

// Publish handles POST /api/v2/ropa/publish — marks a version as PUBLISHED.
func (h *RoPAHandler) Publish(w http.ResponseWriter, r *http.Request) {
	var req service.PublishRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if err := h.service.Publish(r.Context(), req.ID); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"status": "published"})
}

// PromoteMajor handles POST /api/v2/ropa/promote — major version bump.
func (h *RoPAHandler) PromoteMajor(w http.ResponseWriter, r *http.Request) {
	version, err := h.service.PromoteMajor(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, version)
}
