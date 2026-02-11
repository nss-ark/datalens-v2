package handler

import (
	"net/http"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/internal/service"
	govService "github.com/complyark/datalens/internal/service/governance"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
	"github.com/go-chi/chi/v5"
)

type GovernanceHandler struct {
	contextEngine *govService.ContextEngine
	policyService *service.PolicyService
}

func NewGovernanceHandler(
	engine *govService.ContextEngine,
	policyService *service.PolicyService,
) *GovernanceHandler {
	return &GovernanceHandler{
		contextEngine: engine,
		policyService: policyService,
	}
}

func (h *GovernanceHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Context Engine Routes
	r.Post("/suggest", h.SuggestPurposes)

	// Policy Engine Routes
	r.Get("/policies", h.ListPolicies)
	r.Post("/policies", h.CreatePolicy)
	r.Get("/violations", h.ListViolations)
	r.Post("/scan", h.TriggerScan)

	return r
}

type SuggestPurposesRequest struct {
	Items []SuggestItem `json:"items"`
	UseAI bool          `json:"use_ai"`
}

type SuggestItem struct {
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name"`
	DataType   string `json:"data_type"`
}

// SuggestPurposes triggers the context engine analysis.
// POST /api/v2/governance/suggest
func (h *GovernanceHandler) SuggestPurposes(w http.ResponseWriter, r *http.Request) {
	var req SuggestPurposesRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	// Map request to service items
	var serviceItems []govService.PurposeSuggestionItem
	for _, item := range req.Items {
		serviceItems = append(serviceItems, govService.PurposeSuggestionItem{
			TableName:  item.TableName,
			ColumnName: item.ColumnName,
			DataType:   item.DataType,
		})
	}

	suggestions, err := h.contextEngine.SuggestPurposes(r.Context(), serviceItems, req.UseAI)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, suggestions)
}

// ListPolicies retrieves all active policies.
func (h *GovernanceHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.policyService.GetPolicies(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, policies)
}

// CreatePolicy creates a new policy.
func (h *GovernanceHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req governance.Policy
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if req.Name == "" {
		httputil.ErrorFromDomain(w, types.NewValidationError("name is required", nil))
		return
	}

	if err := h.policyService.CreatePolicy(r.Context(), &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, req)
}

// ListViolations retrieves violations, optionally filtered by status.
func (h *GovernanceHandler) ListViolations(w http.ResponseWriter, r *http.Request) {
	statusStr := r.URL.Query().Get("status")
	var status *governance.ViolationStatus

	if statusStr != "" {
		s := governance.ViolationStatus(statusStr)
		status = &s
	}

	violations, err := h.policyService.GetViolations(r.Context(), status)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, violations)
}

// TriggerScan triggers an immediate policy evaluation scan.
func (h *GovernanceHandler) TriggerScan(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := types.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorFromDomain(w, types.NewForbiddenError("tenant context required"))
		return
	}

	if err := h.policyService.EvaluatePolicies(r.Context(), tenantID); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	// Return the fresh violations
	violations, err := h.policyService.GetViolations(r.Context(), nil)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, violations)
}
