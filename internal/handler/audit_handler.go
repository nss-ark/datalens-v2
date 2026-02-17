package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/domain/audit"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// AuditHandler handles audit log HTTP requests for the Control Centre.
type AuditHandler struct {
	service *service.AuditService
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(service *service.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

// Routes returns a chi.Router with audit log routes.
func (h *AuditHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	return r
}

// List handles GET /api/v2/audit-logs.
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	// 1. Extract tenantID from context
	tenantID, ok := types.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "tenant context required")
		return
	}

	// 2. Parse pagination
	pagination := httputil.ParsePagination(r)

	// 3. Parse filters from query params
	var filters audit.AuditFilters

	filters.EntityType = r.URL.Query().Get("entity_type")
	filters.Action = r.URL.Query().Get("action")

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		uid, err := types.ParseID(userIDStr)
		if err != nil {
			httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid user_id format")
			return
		}
		filters.UserID = &uid
	}

	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid start_date format, use RFC3339")
			return
		}
		filters.StartDate = &t
	}

	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid end_date format, use RFC3339")
			return
		}
		filters.EndDate = &t
	}

	// 4. Call service
	result, err := h.service.ListByTenant(r.Context(), tenantID, filters, pagination)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	// 5. Return paginated response
	httputil.JSONWithPagination(w, result.Items, pagination.Page, pagination.PageSize, result.Total)
}
