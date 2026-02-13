package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// DSRHandler handles DSR HTTP requests.
type DSRHandler struct {
	service  *service.DSRService
	executor *service.DSRExecutor
}

// NewDSRHandler creates a new DSRHandler.
func NewDSRHandler(service *service.DSRService, executor *service.DSRExecutor) *DSRHandler {
	return &DSRHandler{
		service:  service,
		executor: executor,
	}
}

// Routes returns a chi.Router with DSR routes.
func (h *DSRHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}/approve", h.Approve)
	r.Put("/{id}/reject", h.Reject)
	r.Get("/{id}/result", h.GetResult)
	r.Post("/{id}/execute", h.ExecuteManual)
	r.Patch("/{id}/status", h.UpdateStatus)
	return r
}

// Create handles POST /api/v2/dsr.
func (h *DSRHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateDSRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if req.SubjectEmail == "" {
		httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "subject_email is required")
		return
	}

	dsr, err := h.service.CreateDSR(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, dsr)
}

// GetByID handles GET /api/v2/dsr/{id}.
func (h *DSRHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	dsr, err := h.service.GetDSR(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, dsr)
}

// List handles GET /api/v2/dsr.
func (h *DSRHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

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

	result, err := h.service.GetDSRs(r.Context(), types.Pagination{Page: page, PageSize: pageSize}, statusFilter, typeFilter)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, result)
}

// UpdateStatus handles PATCH /api/v2/dsr/{id}/status.
func (h *DSRHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req struct {
		Status compliance.DSRStatus `json:"status"`
		Notes  string               `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if req.Status == "" {
		httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "status is required")
		return
	}

	dsr, err := h.service.UpdateStatus(r.Context(), id, req.Status, req.Notes)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, dsr)
}

// Approve handles PUT /api/v2/dsr/{id}/approve.
func (h *DSRHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	dsr, err := h.service.ApproveDSR(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, dsr)
}

// Reject handles PUT /api/v2/dsr/{id}/reject.
func (h *DSRHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if req.Reason == "" {
		httputil.ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "reason is required")
		return
	}

	dsr, err := h.service.RejectDSR(r.Context(), id, req.Reason)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, dsr)
}

// GetResult handles GET /api/v2/dsr/{id}/result.
func (h *DSRHandler) GetResult(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	result, err := h.executor.GetExecutionResult(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, result)
}

// ExecuteManual handles POST /api/v2/dsr/{id}/execute.
func (h *DSRHandler) ExecuteManual(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	// Execute synchronously (admin override)
	if err := h.executor.ExecuteDSR(r.Context(), id); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"status": "executed"})
}
