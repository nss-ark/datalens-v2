package handler

import (
	"net/http"

	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
	"github.com/go-chi/chi/v5"
)

type NoticeHandler struct {
	service            *service.NoticeService
	translationService *service.TranslationService
}

func NewNoticeHandler(service *service.NoticeService, translationService *service.TranslationService) *NoticeHandler {
	return &NoticeHandler{
		service:            service,
		translationService: translationService,
	}
}

func (h *NoticeHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Post("/{id}/publish", h.Publish)
	r.Post("/{id}/archive", h.Archive)
	r.Post("/{id}/bind", h.Bind)

	// Translation Routes
	r.Post("/{id}/translate", h.Translate)
	r.Get("/{id}/translations", h.GetTranslations)
	r.Put("/{id}/translations/{lang}", h.OverrideTranslation)
	return r
}

func (h *NoticeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateNoticeRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	notice, err := h.service.Create(r.Context(), req)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, notice)
}

func (h *NoticeHandler) List(w http.ResponseWriter, r *http.Request) {
	notices, err := h.service.List(r.Context())
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, notices)
}

func (h *NoticeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	notice, err := h.service.Get(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, notice)
}

func (h *NoticeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req service.UpdateNoticeRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	req.ID = id

	if err := h.service.Update(r.Context(), req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *NoticeHandler) Publish(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	notice, err := h.service.Publish(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, notice)
}

func (h *NoticeHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if err := h.service.Archive(r.Context(), id); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *NoticeHandler) Bind(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	var req struct {
		WidgetIDs []types.ID `json:"widget_ids"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if err := h.service.Bind(r.Context(), id, req.WidgetIDs); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *NoticeHandler) Translate(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	results, err := h.translationService.TranslateNotice(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, results)
}

func (h *NoticeHandler) GetTranslations(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	translations, err := h.translationService.GetTranslations(r.Context(), id)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, translations)
}

func (h *NoticeHandler) OverrideTranslation(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	lang := chi.URLParam(r, "lang")
	if lang == "" {
		httputil.ErrorFromDomain(w, types.NewValidationError("language code required", nil))
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if err := h.translationService.OverrideTranslation(r.Context(), id, lang, req.Text); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]bool{"success": true})
}
