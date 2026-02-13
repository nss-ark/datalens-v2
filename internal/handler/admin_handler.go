package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
)

type AdminHandler struct {
	service *service.AdminService
}

func NewAdminHandler(service *service.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/tenants", h.ListTenants)
	r.Post("/tenants", h.OnboardTenant)
	r.Get("/stats", h.GetStats)
	return r
}

func (h *AdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	pagination := httputil.ParsePagination(r)

	filter := identity.TenantFilter{
		Limit:  pagination.PageSize,
		Offset: pagination.Offset(),
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := identity.TenantStatus(status)
		filter.Status = &s
	}

	tenants, total, err := h.service.ListTenants(r.Context(), filter)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSONWithPagination(w, tenants, pagination.Page, pagination.PageSize, total)
}

func (h *AdminHandler) OnboardTenant(w http.ResponseWriter, r *http.Request) {
	var input service.OnboardInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	result, err := h.service.OnboardTenant(r.Context(), input)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, result)
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, stats)
}
