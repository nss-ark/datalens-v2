package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
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
	r.Get("/users", h.ListUsers)
	r.Get("/users/{id}", h.GetUser)
	r.Patch("/users/{id}/status", h.UpdateUserStatus)
	r.Put("/users/{id}/roles", h.AssignRoles)
	r.Get("/roles", h.ListRoles)

	// Compliance / DSRs (Cross-tenant)
	r.Get("/dsr", h.ListDSRs)
	r.Get("/dsr/{id}", h.GetDSR)

	// Retention Policies
	r.Post("/retention-policies", h.CreateRetentionPolicy)
	r.Get("/retention-policies", h.ListRetentionPolicies)
	r.Get("/retention-policies/{id}", h.GetRetentionPolicy)
	r.Put("/retention-policies/{id}", h.UpdateRetentionPolicy)

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

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	pagination := httputil.ParsePagination(r)

	filter := identity.UserFilter{
		Limit:  pagination.PageSize,
		Offset: pagination.Offset(),
		Search: r.URL.Query().Get("search"),
	}

	if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
		id, err := httputil.ParseID(tenantID)
		if err == nil {
			filter.TenantID = &id
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := identity.UserStatus(status)
		filter.Status = &s
	}

	users, total, err := h.service.ListUsers(r.Context(), filter)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSONWithPagination(w, users, pagination.Page, pagination.PageSize, total)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	user, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, user)
}

type UpdateStatusRequest struct {
	Status identity.UserStatus `json:"status"`
}

func (h *AdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	var req UpdateStatusRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Validate status enum
	switch req.Status {
	case identity.UserActive, identity.UserSuspended, identity.UserInvited:
		// valid
	default:
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_STATUS", "Invalid user status")
		return
	}

	if req.Status == identity.UserActive {
		err = h.service.ActivateUser(r.Context(), id)
	} else if req.Status == identity.UserSuspended {
		err = h.service.SuspendUser(r.Context(), id)
	} else {
		// For now we only support suspend/activate via this endpoint
		// Invite is handled via re-invite flow usually
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_OPERATION", "Only ACTIVE and SUSPENDED status updates are supported")
		return
	}

	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"message": "user status updated"})
}

type AssignRolesRequest struct {
	RoleIDs []string `json:"role_ids"`
}

func (h *AdminHandler) AssignRoles(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	var req AssignRolesRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	var roleIDs []types.ID
	for _, rid := range req.RoleIDs {
		parsedID, err := httputil.ParseID(rid)
		if err != nil {
			httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ROLE_ID", "Invalid role ID: "+rid)
			return
		}
		roleIDs = append(roleIDs, parsedID)
	}

	if err := h.service.AssignRoles(r.Context(), id, roleIDs); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"message": "roles assigned"})
}

func (h *AdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.ListRoles(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, roles)
}

func (h *AdminHandler) ListDSRs(w http.ResponseWriter, r *http.Request) {
	pagination := httputil.ParsePagination(r)

	var statusFilter *compliance.DSRStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := compliance.DSRStatus(s)
		statusFilter = &st
	}

	var typeFilter *compliance.DSRRequestType
	if t := r.URL.Query().Get("type"); t != "" {
		rt := compliance.DSRRequestType(t)
		typeFilter = &rt
	}

	dsrs, err := h.service.GetAllDSRs(r.Context(), pagination, statusFilter, typeFilter)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, dsrs)
}

func (h *AdminHandler) GetDSR(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid DSR ID")
		return
	}

	dsr, err := h.service.GetDSR(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, dsr)
}

// -------------------------------------------------------------------------
// Retention Policies
// -------------------------------------------------------------------------

func (h *AdminHandler) CreateRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	var req compliance.RetentionPolicy
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	// Admin API: TenantID usually comes from context if tenant-scoped, but admin might explicitly set it?
	// AdminService expects TenantID in req.
	// If this is super-admin creating policy for a tenant, req.TenantID is needed.
	if req.TenantID == (types.ID{}) {
		httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "tenant_id is required")
		return
	}

	policy, err := h.service.CreateRetentionPolicy(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, policy)
}

func (h *AdminHandler) GetRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID")
		return
	}

	policy, err := h.service.GetRetentionPolicy(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, policy)
}

func (h *AdminHandler) ListRetentionPolicies(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.URL.Query().Get("tenant_id")
	if tenantIDStr == "" {
		httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "tenant_id query param is required")
		return
	}

	tenantID, err := httputil.ParseID(tenantIDStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid tenant_id")
		return
	}

	policies, err := h.service.ListRetentionPolicies(r.Context(), tenantID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, policies)
}

func (h *AdminHandler) UpdateRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID")
		return
	}

	var req compliance.RetentionPolicy
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	policy, err := h.service.UpdateRetentionPolicy(r.Context(), id, req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, policy)
}
