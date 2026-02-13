package handler

import (
	"encoding/json"
	"net/http"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/go-chi/chi/v5"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Notifications
	r.Get("/", h.ListNotifications)

	// Templates
	r.Route("/templates", func(r chi.Router) {
		r.Post("/", h.CreateTemplate)
		r.Get("/", h.ListTemplates)
		r.Get("/{id}", h.GetTemplate)
		r.Put("/{id}", h.UpdateTemplate)
	})

	return r
}

// Notifications

func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	pagination := httputil.ParsePagination(r)

	filter := consent.NotificationFilter{}
	if recipientID := r.URL.Query().Get("recipient_id"); recipientID != "" {
		filter.RecipientID = &recipientID
	}
	if eventType := r.URL.Query().Get("event_type"); eventType != "" {
		filter.EventType = &eventType
	}
	if channel := r.URL.Query().Get("channel"); channel != "" {
		filter.Channel = &channel
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	result, err := h.service.ListNotifications(r.Context(), filter, pagination)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSONWithPagination(w, result.Items, result.Page, result.PageSize, result.Total)
}

// Templates

func (h *NotificationHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req service.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	tmpl, err := h.service.CreateTemplate(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, tmpl)
}

func (h *NotificationHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req service.UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	tmpl, err := h.service.UpdateTemplate(r.Context(), id, req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, tmpl)
}

func (h *NotificationHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.service.ListTemplates(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, templates)
}

func (h *NotificationHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	tmpl, err := h.service.GetTemplate(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, tmpl)
}
