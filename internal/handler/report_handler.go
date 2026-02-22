package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
)

// ReportHandler handles compliance reporting endpoints.
type ReportHandler struct {
	svc *service.ReportService
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// Routes returns a chi.Router with report-related routes.
func (h *ReportHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/compliance-snapshot", h.ComplianceSnapshot)
	r.Get("/export/{entity}", h.ExportEntity)
	return r
}

// ComplianceSnapshot handles GET /reports/compliance-snapshot?from=&to=
func (h *ReportHandler) ComplianceSnapshot(w http.ResponseWriter, r *http.Request) {
	from, to := parseDateRange(r)

	snapshot, err := h.svc.GenerateComplianceSnapshot(r.Context(), from, to)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, snapshot)
}

// ExportEntity handles GET /reports/export/{entity}?format=csv|json
func (h *ReportHandler) ExportEntity(w http.ResponseWriter, r *http.Request) {
	entity := chi.URLParam(r, "entity")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	data, filename, contentType, err := h.svc.ExportEntity(r.Context(), entity, format)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// parseDateRange extracts from/to query params, defaulting to last 30 days.
func parseDateRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now().UTC()
	to := now
	from := now.AddDate(0, 0, -30)

	if s := r.URL.Query().Get("from"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			from = t
		}
	}
	if s := r.URL.Query().Get("to"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			to = t
		}
	}

	return from, to
}
