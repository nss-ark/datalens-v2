// Package handler defines HTTP handlers that map REST API routes
// to domain service operations.
package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

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
