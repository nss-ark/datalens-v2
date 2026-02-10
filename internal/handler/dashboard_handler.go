package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	mw "github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
)

// DashboardHandler handles dashboard related endpoints.
type DashboardHandler struct {
	svc *service.DashboardService
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Routes returns a chi.Router with dashboard routes mounted.
func (h *DashboardHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/stats", h.GetStats)
	return r
}

// GetStats returns aggregated statistics for the tenant dashboard.
// GET /api/v2/dashboard/stats
func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mw.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "tenant context missing")
		return
	}

	stats, err := h.svc.GetStats(r.Context(), tenantID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, stats)
}
