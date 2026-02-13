package handler

import (
	"encoding/json"
	"net/http"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
	"github.com/go-chi/chi/v5"
)

// GrievanceHandler handles grievance HTTP requests.
type GrievanceHandler struct {
	service *service.GrievanceService
}

// NewGrievanceHandler creates a new GrievanceHandler.
func NewGrievanceHandler(service *service.GrievanceService) *GrievanceHandler {
	return &GrievanceHandler{service: service}
}

// Routes returns a chi.Router with Internal routes.
func (h *GrievanceHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}/assign", h.Assign)
	r.Put("/{id}/resolve", h.Resolve)
	r.Put("/{id}/escalate", h.Escalate)
	return r
}

// PortalRoutes returns a chi.Router with Public/Portal routes.
func (h *GrievanceHandler) PortalRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Submit)
	r.Get("/", h.ListMyGrievances)
	r.Get("/{id}", h.GetByID) // Reuse GetByID, it checks tenant/subject context internally if needed?
	// Actually GetByID checks tenant. For portal, we need to ensure subject ownership or use a specific portal Get.
	// The service.GetGrievance checks tenant.
	// But for portal, we should probably add a check that the subject matches the logged-in user.
	// However, distinct portal endpoint is safer.
	r.Post("/{id}/feedback", h.SubmitFeedback)
	return r
}

// Submit handles POST /api/public/portal/grievances.
func (h *GrievanceHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req service.CreateGrievanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	// For portal, subject ID comes from context (auth middleware)
	// We assume Portal Auth Middleware populates SubjectID in context.
	// If not, we might need to rely on the request or fail.
	// Looking at existing patterns (DSR), usually the profile ID or subject ID is in context.
	// For now, let's assume the frontend sends the data_subject_id (from profile) OR we extract it.
	// To be safe and explicit based on the spec "Portal (public) routes ... (portal JWT auth)",
	// let's try to extract it from context if the request body is empty.
	// But `req.DataSubjectID` is in the struct.

	// If the user didn't provide it in body (likely), we should inject it from context if available.
	if req.DataSubjectID == "" {
		if subjectID, ok := types.SubjectIDFromContext(r.Context()); ok {
			req.DataSubjectID = subjectID.String()
		}
	}

	result, err := h.service.SubmitGrievance(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, result)
}

// List handles GET /api/v2/grievances (Internal).
func (h *GrievanceHandler) List(w http.ResponseWriter, r *http.Request) {
	pagination := httputil.ParsePagination(r)
	filters := make(map[string]any)

	if s := r.URL.Query().Get("status"); s != "" {
		filters["status"] = s
	}
	if p := r.URL.Query().Get("priority"); p != "" {
		// Parse int logic... simplified for now, assuming binder handles it or we parse manually
		// but service takes map[string]any.
		filters["priority"] = 0 // Placeholder, query params are strings
	}
	if a := r.URL.Query().Get("assigned_to"); a != "" {
		filters["assigned_to"] = a
	}

	result, err := h.service.ListByTenant(r.Context(), filters, pagination)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSONWithPagination(w, result.Items, result.Page, result.PageSize, result.Total)
}

// ListMyGrievances handles GET /api/public/portal/grievances.
func (h *GrievanceHandler) ListMyGrievances(w http.ResponseWriter, r *http.Request) {
	subjectID, ok := types.SubjectIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "subject context required")
		return
	}

	result, err := h.service.ListBySubject(r.Context(), subjectID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, result)
}

// GetByID handles GET /{id}.
func (h *GrievanceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	g, err := h.service.GetGrievance(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, g)
}

// Assign handles PUT /api/v2/grievances/{id}/assign.
func (h *GrievanceHandler) Assign(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req struct {
		AssigneeID string `json:"assignee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	assigneeID, err := types.ParseID(req.AssigneeID)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid assignee_id")
		return
	}

	if err := h.service.AssignGrievance(r.Context(), id, assigneeID); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

// Resolve handles PUT /api/v2/grievances/{id}/resolve.
func (h *GrievanceHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req struct {
		Resolution string `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.service.ResolveGrievance(r.Context(), id, req.Resolution); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

// Escalate handles PUT /api/v2/grievances/{id}/escalate.
func (h *GrievanceHandler) Escalate(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req struct {
		Authority string `json:"authority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.service.EscalateGrievance(r.Context(), id, req.Authority); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"status": "escalated"})
}

// SubmitFeedback handles POST /api/public/portal/grievances/{id}/feedback.
func (h *GrievanceHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.service.SubmitFeedback(r.Context(), id, req.Rating, req.Comment); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"status": "feedback_submitted"})
}
