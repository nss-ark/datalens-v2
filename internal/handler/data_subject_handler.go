package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// DataSubjectHandler handles HTTP requests for data subject (Data Principal Profile) listing.
type DataSubjectHandler struct {
	profileRepo consent.DataPrincipalProfileRepository
}

// NewDataSubjectHandler creates a new DataSubjectHandler.
func NewDataSubjectHandler(profileRepo consent.DataPrincipalProfileRepository) *DataSubjectHandler {
	return &DataSubjectHandler{profileRepo: profileRepo}
}

// Routes returns a chi.Router with data subject routes.
// Mounted at /api/v2/subjects.
func (h *DataSubjectHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	return r
}

// List handles GET /api/v2/subjects.
// Supports optional ?search= query param for ILIKE search on email/phone.
// Returns paginated results via ?page= and ?page_size= params.
func (h *DataSubjectHandler) List(w http.ResponseWriter, r *http.Request) {
	// 1. Extract tenantID from context
	tenantID, ok := types.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "tenant context required")
		return
	}

	// 2. Parse pagination
	pagination := httputil.ParsePagination(r)

	// 3. Parse search query
	search := r.URL.Query().Get("search")

	// 4. Call repo — search or list
	if search != "" {
		result, err := h.profileRepo.SearchByTenant(r.Context(), tenantID, search, pagination)
		if err != nil {
			httputil.ErrorFromDomain(w, err)
			return
		}
		httputil.JSONWithPagination(w, result.Items, pagination.Page, pagination.PageSize, result.Total)
	} else {
		result, err := h.profileRepo.ListByTenant(r.Context(), tenantID, pagination)
		if err != nil {
			httputil.ErrorFromDomain(w, err)
			return
		}
		httputil.JSONWithPagination(w, result.Items, pagination.Page, pagination.PageSize, result.Total)
	}
}
