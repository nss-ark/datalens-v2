package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/service/analytics"
	"github.com/complyark/datalens/pkg/httputil"
)

// AnalyticsHandler exposes analytics endpoints.
type AnalyticsHandler struct {
	consentService *analytics.ConsentAnalyticsService
}

// NewAnalyticsHandler creates a new handler.
func NewAnalyticsHandler(consentService *analytics.ConsentAnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{consentService: consentService}
}

// Routes returns the analytics router.
func (h *AnalyticsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/consent/conversion", h.GetConversionStats)
	r.Get("/consent/purpose", h.GetPurposeStats)
	return r
}

// GetConversionStats handles GET /analytics/consent/conversion
func (h *AnalyticsHandler) GetConversionStats(w http.ResponseWriter, r *http.Request) {
	from, _ := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	to, _ := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	interval := r.URL.Query().Get("interval")

	// Defaults
	if to.IsZero() {
		to = time.Now()
	}
	if from.IsZero() {
		from = to.AddDate(0, 0, -30)
	}

	stats, err := h.consentService.GetConversionStats(r.Context(), from, to, interval)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, stats)
}

// GetPurposeStats handles GET /analytics/consent/purpose
func (h *AnalyticsHandler) GetPurposeStats(w http.ResponseWriter, r *http.Request) {
	from, _ := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	to, _ := time.Parse("2006-01-02", r.URL.Query().Get("to"))

	// Defaults
	if to.IsZero() {
		to = time.Now()
	}
	if from.IsZero() {
		from = to.AddDate(0, 0, -30)
	}

	stats, err := h.consentService.GetPurposeStats(r.Context(), from, to)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, stats)
}
