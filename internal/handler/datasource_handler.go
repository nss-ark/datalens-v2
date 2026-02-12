// Package handler defines HTTP handlers that map REST API routes
// to domain service operations.
package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/robfig/cron/v3"

	"github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// DataSourceHandler handles data source REST endpoints.
type DataSourceHandler struct {
	svc *service.DataSourceService
}

// NewDataSourceHandler creates a new DataSourceHandler.
func NewDataSourceHandler(svc *service.DataSourceService) *DataSourceHandler {
	return &DataSourceHandler{svc: svc}
}

// Routes returns a chi.Router with data source routes mounted.
func (h *DataSourceHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	// Scan scheduling
	r.Put("/{id}/scan/schedule", h.SetSchedule)
	r.Delete("/{id}/scan/schedule", h.ClearSchedule)

	// M365 specific routes
	r.Get("/{id}/m365/users", h.ListM365Users)
	r.Get("/{id}/m365/sites", h.ListM365Sites)

	return r
}

func (h *DataSourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusForbidden, "TENANT_REQUIRED", "tenant context is required")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
		Host        string `json:"host"`
		Port        int    `json:"port"`
		Database    string `json:"database"`
		Credentials string `json:"credentials"`
		Config      string `json:"config"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	ds, err := h.svc.Create(r.Context(), service.CreateDataSourceInput{
		TenantID:    tenantID,
		Name:        req.Name,
		Type:        types.DataSourceType(req.Type),
		Description: req.Description,
		Host:        req.Host,
		Port:        req.Port,
		Database:    req.Database,
		Credentials: req.Credentials,
		Config:      req.Config,
	})
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, ds)
}

func (h *DataSourceHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusForbidden, "TENANT_REQUIRED", "tenant context is required")
		return
	}

	sources, err := h.svc.ListByTenant(r.Context(), tenantID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, sources)
}

func (h *DataSourceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	ds, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, ds)
}

func (h *DataSourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Host        string `json:"host"`
		Port        *int   `json:"port"`
		Database    string `json:"database"`
		Credentials string `json:"credentials"`
		Config      string `json:"config"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	ds, err := h.svc.Update(r.Context(), service.UpdateDataSourceInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Host:        req.Host,
		Port:        req.Port,
		Database:    req.Database,
		Credentials: req.Credentials,
		Config:      req.Config,
	})
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, ds)
}

func (h *DataSourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusNoContent, nil)
}

// SetSchedule sets a cron expression for automatic scanning.
func (h *DataSourceHandler) SetSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req struct {
		Cron string `json:"cron"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if req.Cron == "" {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_CRON", "cron expression is required")
		return
	}

	// Validate cron expression
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	if _, err := parser.Parse(req.Cron); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_CRON", "invalid cron expression: "+err.Error())
		return
	}

	ds, err := h.svc.SetSchedule(r.Context(), id, req.Cron)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, ds)
}

// ClearSchedule removes the scan schedule from a data source.
func (h *DataSourceHandler) ClearSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if err := h.svc.ClearSchedule(r.Context(), id); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusNoContent, nil)
}

// ListM365Users lists users from M365.
func (h *DataSourceHandler) ListM365Users(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	users, err := h.svc.ListM365Users(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, users)
}

// ListM365Sites lists sites from M365.
func (h *DataSourceHandler) ListM365Sites(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	sites, err := h.svc.ListM365Sites(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, sites)
}
