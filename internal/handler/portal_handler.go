package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// PortalHandler handles HTTP requests for the Data Principal Portal.
type PortalHandler struct {
	authService      *service.PortalAuthService
	principalService *service.DataPrincipalService
	middleware       *middleware.PortalAuthMiddleware
}

// NewPortalHandler creates a new PortalHandler.
func NewPortalHandler(
	authService *service.PortalAuthService,
	principalService *service.DataPrincipalService,
) *PortalHandler {
	return &PortalHandler{
		authService:      authService,
		principalService: principalService,
		middleware:       middleware.NewPortalAuthMiddleware(authService),
	}
}

// Routes returns the router for portal endpoints.
// Mounted at /api/public/portal
func (h *PortalHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Public: Verification
	r.Post("/verify", h.initiateLogin)
	r.Post("/verify/confirm", h.verifyLogin)

	// Protected: Profile, Consent, DPR
	r.Group(func(r chi.Router) {
		r.Use(h.middleware.PortalJWTAuth)
		r.Get("/profile", h.getProfile)
		r.Get("/consent-history", h.getConsentHistory)
		r.Post("/dpr", h.submitDPR)
		r.Get("/dpr", h.listDPRs)
		r.Get("/dpr/{id}", h.getDPR)
	})

	// Guardian Verification (DPDPA Section 9)
	r.Group(func(r chi.Router) {
		r.Use(h.middleware.PortalJWTAuth)
		r.Post("/guardian/verify-init", h.initiateGuardianVerify)
		r.Post("/guardian/verify", h.verifyGuardian)
	})

	return r
}

type verifyRequest struct {
	TenantID types.ID `json:"tenant_id"`
	Email    string   `json:"email"`
	Phone    string   `json:"phone"`
}

func (h *PortalHandler) initiateLogin(w http.ResponseWriter, r *http.Request) {
	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	if err := h.authService.InitiateLogin(r.Context(), req.TenantID, req.Email, req.Phone); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"message": "verification code sent"})
}

type confirmRequest struct {
	TenantID types.ID `json:"tenant_id"`
	Email    string   `json:"email"`
	Phone    string   `json:"phone"`
	Code     string   `json:"code"`
}

func (h *PortalHandler) verifyLogin(w http.ResponseWriter, r *http.Request) {
	var req confirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	token, profile, err := h.authService.VerifyLogin(r.Context(), req.TenantID, req.Email, req.Phone, req.Code)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"token":   token,
		"profile": profile,
	})
}

func (h *PortalHandler) getProfile(w http.ResponseWriter, r *http.Request) {
	principalID, ok := r.Context().Value(types.ContextKey("principal_id")).(types.ID)
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "principal context missing")
		return
	}

	profile, err := h.principalService.GetProfile(r.Context(), principalID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, profile)
}

func (h *PortalHandler) getConsentHistory(w http.ResponseWriter, r *http.Request) {
	principalID, ok := r.Context().Value(types.ContextKey("principal_id")).(types.ID)
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "principal context missing")
		return
	}

	// Simple pagination parsing, could use httputil.ParsePagination
	page := 1
	pageSize := 20
	// TODO: Parse query params

	pagination := types.Pagination{Page: page, PageSize: pageSize}
	history, err := h.principalService.GetConsentHistory(r.Context(), principalID, pagination)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, history)
}

func (h *PortalHandler) submitDPR(w http.ResponseWriter, r *http.Request) {
	principalID, ok := r.Context().Value(types.ContextKey("principal_id")).(types.ID)
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "principal context missing")
		return
	}

	var req service.CreateDPRRequestInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	dpr, err := h.principalService.SubmitDPR(r.Context(), principalID, req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, dpr)
}

func (h *PortalHandler) listDPRs(w http.ResponseWriter, r *http.Request) {
	principalID, ok := r.Context().Value(types.ContextKey("principal_id")).(types.ID)
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "principal context missing")
		return
	}

	list, err := h.principalService.ListDPRs(r.Context(), principalID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, list)
}

func (h *PortalHandler) getDPR(w http.ResponseWriter, r *http.Request) {
	principalID, ok := r.Context().Value(types.ContextKey("principal_id")).(types.ID)
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "principal context missing")
		return
	}

	dprIDStr := chi.URLParam(r, "id")
	dprID, err := types.ParseID(dprIDStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid dpr id")
		return
	}

	dpr, err := h.principalService.GetDPR(r.Context(), principalID, dprID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, dpr)
}

// Guardian Verification Handlers

type initiateGuardianRequest struct {
	Contact string `json:"contact"` // Email or Phone
}

func (h *PortalHandler) initiateGuardianVerify(w http.ResponseWriter, r *http.Request) {
	principalID, ok := r.Context().Value(types.ContextKey("principal_id")).(types.ID)
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "principal context missing")
		return
	}

	var req initiateGuardianRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	if err := h.principalService.InitiateGuardianVerification(r.Context(), principalID, req.Contact); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"message": "guardian verification code sent"})
}

type verifyGuardianRequest struct {
	Code string `json:"code"`
}

func (h *PortalHandler) verifyGuardian(w http.ResponseWriter, r *http.Request) {
	principalID, ok := r.Context().Value(types.ContextKey("principal_id")).(types.ID)
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "principal context missing")
		return
	}

	var req verifyGuardianRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	if err := h.principalService.VerifyGuardian(r.Context(), principalID, req.Code); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"message": "guardian verified successfully"})
}
