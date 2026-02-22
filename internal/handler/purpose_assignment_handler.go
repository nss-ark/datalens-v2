package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// PurposeAssignmentHandler handles HTTP requests for purpose scope assignments.
type PurposeAssignmentHandler struct {
	service *service.PurposeAssignmentService
}

// NewPurposeAssignmentHandler creates a new PurposeAssignmentHandler.
func NewPurposeAssignmentHandler(s *service.PurposeAssignmentService) *PurposeAssignmentHandler {
	return &PurposeAssignmentHandler{service: s}
}

// Routes returns a chi.Router with purpose assignment routes.
// Mounted at /api/v2/purpose-assignments.
func (h *PurposeAssignmentHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Assign)
	r.Delete("/{id}", h.Remove)
	r.Get("/", h.GetByScope)            // ?scope_type=TABLE&scope_id=db.schema.users
	r.Get("/effective", h.GetEffective) // ?scope_type=COLUMN&scope_id=db.schema.users.email
	r.Get("/all", h.GetAll)
	return r
}

// Assign handles POST /api/v2/purpose-assignments — assign a purpose at a scope level.
func (h *PurposeAssignmentHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var req service.AssignPurposeInput
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	assignment, err := h.service.Assign(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, assignment)
}

// Remove handles DELETE /api/v2/purpose-assignments/{id}.
func (h *PurposeAssignmentHandler) Remove(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if err := h.service.Remove(r.Context(), id); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetByScope handles GET /api/v2/purpose-assignments — direct assignments at scope.
func (h *PurposeAssignmentHandler) GetByScope(w http.ResponseWriter, r *http.Request) {
	scopeType := r.URL.Query().Get("scope_type")
	scopeID := r.URL.Query().Get("scope_id")

	if scopeType == "" || scopeID == "" {
		httputil.ErrorFromDomain(w, types.NewValidationError("scope_type and scope_id query parameters are required", nil))
		return
	}

	if !isValidScope(scopeType) {
		httputil.ErrorFromDomain(w, types.NewValidationError("invalid scope_type", map[string]any{
			"scope_type": scopeType,
			"valid":      governance.ValidScopeTypes,
		}))
		return
	}

	assignments, err := h.service.GetByScope(r.Context(), governance.ScopeType(scopeType), scopeID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, assignments)
}

// GetEffective handles GET /api/v2/purpose-assignments/effective — resolves inheritance.
func (h *PurposeAssignmentHandler) GetEffective(w http.ResponseWriter, r *http.Request) {
	scopeType := r.URL.Query().Get("scope_type")
	scopeID := r.URL.Query().Get("scope_id")

	if scopeType == "" || scopeID == "" {
		httputil.ErrorFromDomain(w, types.NewValidationError("scope_type and scope_id query parameters are required", nil))
		return
	}

	if !isValidScope(scopeType) {
		httputil.ErrorFromDomain(w, types.NewValidationError("invalid scope_type", map[string]any{
			"scope_type": scopeType,
			"valid":      governance.ValidScopeTypes,
		}))
		return
	}

	assignments, err := h.service.GetEffective(r.Context(), governance.ScopeType(scopeType), scopeID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, assignments)
}

// GetAll handles GET /api/v2/purpose-assignments/all — all assignments for tenant.
func (h *PurposeAssignmentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	assignments, err := h.service.GetAll(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, assignments)
}

// isValidScope checks if a scope_type string is valid.
func isValidScope(s string) bool {
	for _, v := range governance.ValidScopeTypes {
		if string(v) == s {
			return true
		}
	}
	return false
}
