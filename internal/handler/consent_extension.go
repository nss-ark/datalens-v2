package handler

import (
	"encoding/json"
	"net/http"

	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

func (h *ConsentHandler) renewConsent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SubjectID  types.ID   `json:"subject_id"`
		PurposeIDs []types.ID `json:"purpose_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	// In a real implementation, we should verify that the request is authenticated
	// (e.g., via Portal session) or authorized (e.g. via signed link).
	// For now, trusting the SubjectID as per task scope for "Public API".

	// Tenant ID is in context from WidgetAuthMiddleware (or should be).
	// Wait, PublicRoutes uses `WidgetAPIKeyAuth`.
	// Does `renew` use Widget Auth?
	// The `renew` endpoint is in `PublicRoutes` which has `Use(mw.WidgetAuthMiddleware)`.
	// So we have tenant_id in context.

	if err := h.expirySvc.RenewConsent(r.Context(), req.SubjectID, req.PurposeIDs); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
