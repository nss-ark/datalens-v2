package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// PurposeHandler handles purpose REST endpoints.
type PurposeHandler struct {
	svc *service.PurposeService
}

// NewPurposeHandler creates a new PurposeHandler.
func NewPurposeHandler(svc *service.PurposeService) *PurposeHandler {
	return &PurposeHandler{svc: svc}
}

// Routes returns a chi.Router with purpose routes mounted.
func (h *PurposeHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *PurposeHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusForbidden, "TENANT_REQUIRED", "tenant context is required")
		return
	}

	var req struct {
		Code            string `json:"code"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		LegalBasis      string `json:"legal_basis"`
		RetentionDays   int    `json:"retention_days"`
		RequiresConsent bool   `json:"requires_consent"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	p, err := h.svc.Create(r.Context(), service.CreatePurposeInput{
		TenantID:        tenantID,
		Code:            req.Code,
		Name:            req.Name,
		Description:     req.Description,
		LegalBasis:      types.LegalBasis(req.LegalBasis),
		RetentionDays:   req.RetentionDays,
		RequiresConsent: req.RequiresConsent,
	})
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, p)
}

func (h *PurposeHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusForbidden, "TENANT_REQUIRED", "tenant context is required")
		return
	}

	purposes, err := h.svc.ListByTenant(r.Context(), tenantID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, purposes)
}

func (h *PurposeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, p)
}

func (h *PurposeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		LegalBasis      string `json:"legal_basis"`
		RetentionDays   int    `json:"retention_days"`
		IsActive        *bool  `json:"is_active"`
		RequiresConsent *bool  `json:"requires_consent"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	p, err := h.svc.Update(r.Context(), service.UpdatePurposeInput{
		ID:              id,
		Name:            req.Name,
		Description:     req.Description,
		LegalBasis:      types.LegalBasis(req.LegalBasis),
		RetentionDays:   req.RetentionDays,
		IsActive:        req.IsActive,
		RequiresConsent: req.RequiresConsent,
	})
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, p)
}

func (h *PurposeHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
