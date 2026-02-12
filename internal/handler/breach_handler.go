package handler

import (
	"net/http"

	"github.com/complyark/datalens/internal/domain/breach"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/go-chi/chi/v5"
)

type BreachHandler struct {
	service *service.BreachService
}

func NewBreachHandler(service *service.BreachService) *BreachHandler {
	return &BreachHandler{service: service}
}

func (h *BreachHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Get("/{id}/report/cert-in", h.GetCertInReport)
	return r
}

func (h *BreachHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateIncidentRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	incident, err := h.service.CreateIncident(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, incident)
}

func (h *BreachHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	incident, sla, err := h.service.GetIncident(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	// Enrich response with SLA data (using a wrapper map or just responding with the object if SLA is embedded?
	// The service returns incident and sla map separately.
	// Let's combine them into a response wrapper or just add SLA to metadata if possible.
	// For simplicity, we can return a map combining them.

	response := map[string]interface{}{
		"incident": incident,
		"sla":      sla,
	}

	httputil.JSON(w, http.StatusOK, response)
}

func (h *BreachHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req service.UpdateIncidentRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	incident, err := h.service.UpdateIncident(r.Context(), id, req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, incident)
}

func (h *BreachHandler) List(w http.ResponseWriter, r *http.Request) {
	pagination := httputil.ParsePagination(r)

	filter := breach.Filter{}
	if status := r.URL.Query().Get("status"); status != "" {
		s := breach.IncidentStatus(status)
		filter.Status = &s
	}
	if severity := r.URL.Query().Get("severity"); severity != "" {
		s := breach.IncidentSeverity(severity)
		filter.Severity = &s
	}

	result, err := h.service.ListIncidents(r.Context(), filter, pagination)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSONWithPagination(w, result.Items, result.Page, result.PageSize, result.Total)
}

func (h *BreachHandler) GetCertInReport(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	report, err := h.service.GenerateCertInReport(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, report)
}
