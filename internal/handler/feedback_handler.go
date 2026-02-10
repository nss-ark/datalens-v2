package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// FeedbackHandler handles detection feedback REST endpoints.
type FeedbackHandler struct {
	svc *service.FeedbackService
}

// NewFeedbackHandler creates a new FeedbackHandler.
func NewFeedbackHandler(svc *service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{svc: svc}
}

// Routes returns a chi.Router with feedback routes mounted.
func (h *FeedbackHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// POST /api/v2/discovery/feedback — submit verify/correct/reject
	r.Post("/", h.SubmitFeedback)

	// GET /api/v2/discovery/feedback — list all feedback for tenant (paginated)
	r.Get("/", h.List)

	// GET /api/v2/discovery/feedback/classification/{classificationID} — feedback for a classification
	r.Get("/classification/{classificationID}", h.GetByClassification)

	// GET /api/v2/discovery/feedback/accuracy/{method} — accuracy stats for a detection method
	r.Get("/accuracy/{method}", h.GetAccuracyStats)

	return r
}

// SubmitFeedback handles POST /api/v2/discovery/feedback
func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusForbidden, "TENANT_REQUIRED", "tenant context is required")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusForbidden, "USER_REQUIRED", "user context is required")
		return
	}

	var req service.SubmitFeedbackInput
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	result, err := h.svc.SubmitFeedback(r.Context(), tenantID, userID, req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, result)
}

// List handles GET /api/v2/discovery/feedback
func (h *FeedbackHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusForbidden, "TENANT_REQUIRED", "tenant context is required")
		return
	}

	pagination := httputil.ParsePagination(r)

	result, err := h.svc.ListByTenant(r.Context(), tenantID, pagination)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSONWithPagination(w, result.Items, result.Page, result.PageSize, result.Total)
}

// GetByClassification handles GET /api/v2/discovery/feedback/classification/{classificationID}
func (h *FeedbackHandler) GetByClassification(w http.ResponseWriter, r *http.Request) {
	classificationID, err := httputil.ParseID(chi.URLParam(r, "classificationID"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	feedback, err := h.svc.GetByClassification(r.Context(), classificationID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, feedback)
}

// GetAccuracyStats handles GET /api/v2/discovery/feedback/accuracy/{method}
func (h *FeedbackHandler) GetAccuracyStats(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusForbidden, "TENANT_REQUIRED", "tenant context is required")
		return
	}

	method := types.DetectionMethod(chi.URLParam(r, "method"))

	stats, err := h.svc.GetAccuracyStats(r.Context(), tenantID, method)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, stats)
}
