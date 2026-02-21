package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// ConsentHandler handles HTTP requests for consent management.
type ConsentHandler struct {
	service     *service.ConsentService
	expirySvc   *service.ConsentExpiryService
	sdkFilePath string // Absolute path to consent.min.js
}

// NewConsentHandler creates a new ConsentHandler.
func NewConsentHandler(s *service.ConsentService, expirySvc *service.ConsentExpiryService) *ConsentHandler {
	// Resolve SDK file path relative to executable
	sdkPath := resolveSDKPath()
	return &ConsentHandler{
		service:     s,
		expirySvc:   expirySvc,
		sdkFilePath: sdkPath,
	}
}

// Routes returns the router for internal (protected) consent endpoints.
// Mounted at /api/v2/consent (requires JWT auth).
func (h *ConsentHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/widgets", h.createWidget)
	r.Get("/widgets", h.listWidgets)
	r.Get("/widgets/{id}", h.getWidget)
	r.Put("/widgets/{id}", h.updateWidget)
	r.Delete("/widgets/{id}", h.deleteWidget)
	r.Put("/widgets/{id}/activate", h.activateWidget)
	r.Put("/widgets/{id}/pause", h.pauseWidget)
	r.Get("/widgets/{id}/embed-code", h.getEmbedCode)

	r.Get("/sessions", h.listSessions) // Actually getSessionsBySubject as per service, but simplified
	r.Get("/history/{subjectId}", h.getHistory)

	return r
}

// PublicRoutes returns the router for public (widget) consent endpoints.
// Mounted at /api/public/consent (requires Widget API Key auth).
func (h *ConsentHandler) PublicRoutes() chi.Router {
	r := chi.NewRouter()

	// SDK file (no auth required, served with aggressive caching)
	r.Get("/sdk/consent.min.js", h.serveSDKFile)

	// Widget config (public, requires API Key)
	r.Get("/widget/config", h.getWidgetConfig) // Using API key from header to identify widget

	// Consent operations
	r.Post("/sessions", h.recordConsent)
	r.Get("/check", h.checkConsent)
	r.Post("/withdraw", h.withdrawConsent)
	r.Post("/renew", h.renewConsent)

	return r
}

// =============================================================================
// Internal Handlers
// =============================================================================

func (h *ConsentHandler) createWidget(w http.ResponseWriter, r *http.Request) {
	var req service.CreateWidgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	widget, err := h.service.CreateWidget(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, widget)
}

func (h *ConsentHandler) listWidgets(w http.ResponseWriter, r *http.Request) {
	widgets, err := h.service.ListWidgets(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, widgets)
}

func (h *ConsentHandler) getWidget(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := types.ParseID(idStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid widget id")
		return
	}

	widget, err := h.service.GetWidget(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, widget)
}

func (h *ConsentHandler) updateWidget(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := types.ParseID(idStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid widget id")
		return
	}

	var req service.UpdateWidgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	widget, err := h.service.UpdateWidget(r.Context(), id, req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, widget)
}

func (h *ConsentHandler) deleteWidget(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := types.ParseID(idStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid widget id")
		return
	}

	if err := h.service.DeleteWidget(r.Context(), id); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConsentHandler) activateWidget(w http.ResponseWriter, r *http.Request) {
	h.setWidgetStatus(w, r, h.service.ActivateWidget)
}

func (h *ConsentHandler) pauseWidget(w http.ResponseWriter, r *http.Request) {
	h.setWidgetStatus(w, r, h.service.PauseWidget)
}

func (h *ConsentHandler) setWidgetStatus(w http.ResponseWriter, r *http.Request, fn func(context.Context, types.ID) (*consent.ConsentWidget, error)) {
	idStr := chi.URLParam(r, "id")
	id, err := types.ParseID(idStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid widget id")
		return
	}

	widget, err := fn(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, widget)
}

func (h *ConsentHandler) listSessions(w http.ResponseWriter, r *http.Request) {
	subjectIDStr := r.URL.Query().Get("subject_id")

	// If subject_id is provided, use the existing subject-based listing
	if subjectIDStr != "" {
		subjectID, err := types.ParseID(subjectIDStr)
		if err != nil {
			httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid subject id")
			return
		}

		sessions, err := h.service.GetSessionsBySubject(r.Context(), subjectID)
		if err != nil {
			httputil.ErrorFromDomain(w, err)
			return
		}

		httputil.JSON(w, http.StatusOK, sessions)
		return
	}

	// Tenant-wide listing with optional filters and pagination
	pagination := httputil.ParsePagination(r)

	var filters consent.ConsentSessionFilters
	if purposeStr := r.URL.Query().Get("purpose_id"); purposeStr != "" {
		purposeID, err := types.ParseID(purposeStr)
		if err != nil {
			httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid purpose_id")
			return
		}
		filters.PurposeID = &purposeID
	}
	filters.Status = r.URL.Query().Get("status")

	result, err := h.service.ListSessionsByTenant(r.Context(), filters, pagination)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSONWithPagination(w, result.Items, pagination.Page, pagination.PageSize, result.Total)
}

func (h *ConsentHandler) getHistory(w http.ResponseWriter, r *http.Request) {
	subjectIDStr := chi.URLParam(r, "subjectId")
	subjectID, err := types.ParseID(subjectIDStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid subject id")
		return
	}

	// Parsing pagination
	// Default page=1, page_size=20 (simplified, should use helper if available)
	page := 1
	pageSize := 20
	// TODO: Parse from query params if needed, defaulting for now

	pagination := types.Pagination{Page: page, PageSize: pageSize}
	result, err := h.service.GetHistory(r.Context(), subjectID, pagination)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, result)
}

// =============================================================================
// Public Handlers
// =============================================================================

func (h *ConsentHandler) getWidgetConfig(w http.ResponseWriter, r *http.Request) {
	// API Key is validated by middleware, but we need it to look up config
	// OR we used the context widget.
	// Let's use the API key from header as the service method expects it,
	// or create a service method that takes ID if we have it in context.
	// Service has `GetWidgetConfig(ctx, apiKey)`.
	apiKey := r.Header.Get("X-Widget-Key")

	config, err := h.service.GetWidgetConfig(r.Context(), apiKey)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, config)
}

func (h *ConsentHandler) recordConsent(w http.ResponseWriter, r *http.Request) {
	var req service.RecordConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	// Ensure WidgetID matches authenticated widget
	// Using the ID from context (injected by middleware) to enforce security
	ctxWidgetID, ok := r.Context().Value(types.ContextKeyWidgetID).(types.ID)
	if ok {
		req.WidgetID = ctxWidgetID
	} else {
		// Should be caught by middleware
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "widget context missing")
		return
	}

	session, err := h.service.RecordConsent(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, session)
}

func (h *ConsentHandler) checkConsent(w http.ResponseWriter, r *http.Request) {
	subjectIDStr := r.URL.Query().Get("subject_id")
	purposeIDStr := r.URL.Query().Get("purpose_id")

	if subjectIDStr == "" || purposeIDStr == "" {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "subject_id and purpose_id required")
		return
	}

	subjectID, err := types.ParseID(subjectIDStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid subject_id")
		return
	}
	purposeID, err := types.ParseID(purposeIDStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid purpose_id")
		return
	}

	// Tenant ID is in context from middleware
	tenantID, ok := types.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "tenant context missing")
		return
	}

	granted, err := h.service.CheckConsent(r.Context(), tenantID, subjectID, purposeID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]bool{"granted": granted})
}

func (h *ConsentHandler) withdrawConsent(w http.ResponseWriter, r *http.Request) {
	var req service.WithdrawConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	if err := h.service.WithdrawConsent(r.Context(), req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// serveSDKFile serves the consent widget JS SDK file with aggressive caching.
func (h *ConsentHandler) serveSDKFile(w http.ResponseWriter, r *http.Request) {
	if h.sdkFilePath == "" {
		httputil.ErrorResponse(w, http.StatusNotFound, "SDK_NOT_FOUND", "consent sdk file not found")
		return
	}

	data, err := os.ReadFile(h.sdkFilePath)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusNotFound, "SDK_NOT_FOUND", "consent sdk file not available")
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400, s-maxage=604800") // 1d browser, 7d CDN
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// getEmbedCode generates the embed snippet for a consent widget.
func (h *ConsentHandler) getEmbedCode(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := types.ParseID(idStr)
	if err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "invalid widget id")
		return
	}

	widget, err := h.service.GetWidget(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	// Generate embed code
	// Host is derived from the request (works for local dev and production)
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	host := r.Host
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	embedCode := fmt.Sprintf(
		`<script src="%s/api/public/consent/sdk/consent.min.js" data-widget-id="%s" data-api-key="%s" defer></script>`,
		baseURL, widget.ID, widget.APIKey,
	)

	httputil.JSON(w, http.StatusOK, map[string]string{
		"embed_code": embedCode,
		"widget_id":  widget.ID.String(),
		"version":    fmt.Sprintf("%d", widget.Version),
	})
}

// resolveSDKPath finds the consent.min.js file.
// It checks multiple locations to work in both dev and production.
func resolveSDKPath() string {
	// Try paths relative to the current working directory
	candidates := []string{
		"sdk/consent/dist/consent.min.js",
		"../../sdk/consent/dist/consent.min.js", // When running from cmd/api/
	}

	// Also try relative to the source file (for development)
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		srcDir := filepath.Dir(filename)
		candidates = append(candidates,
			filepath.Join(srcDir, "..", "..", "..", "sdk", "consent", "dist", "consent.min.js"),
		)
	}

	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	return ""
}
