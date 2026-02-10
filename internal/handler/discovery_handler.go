package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/domain/discovery"
	mw "github.com/complyark/datalens/internal/middleware"
	"github.com/complyark/datalens/internal/service"
	"github.com/complyark/datalens/pkg/httputil"
	"github.com/complyark/datalens/pkg/types"
)

// DiscoveryHandler handles discovery-related REST endpoints
// for data inventories, entities, and fields.
type DiscoveryHandler struct {
	service       service.DiscoveryOrchestrator
	scanService   service.ScanOrchestrator // <--- Interface
	inventoryRepo discovery.DataInventoryRepository
	entityRepo    discovery.DataEntityRepository
	fieldRepo     discovery.DataFieldRepository
}

// NewDiscoveryHandler creates a new DiscoveryHandler.
func NewDiscoveryHandler(
	service service.DiscoveryOrchestrator,
	scanService service.ScanOrchestrator, // <--- Interface
	inventoryRepo discovery.DataInventoryRepository,
	entityRepo discovery.DataEntityRepository,
	fieldRepo discovery.DataFieldRepository,
) *DiscoveryHandler {
	return &DiscoveryHandler{
		service:       service,
		scanService:   scanService,
		inventoryRepo: inventoryRepo,
		entityRepo:    entityRepo,
		fieldRepo:     fieldRepo,
	}
}

// Routes returns a chi.Router with discovery routes mounted.
// These are nested under /data-sources/{sourceID}/...
func (h *DiscoveryHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Trigger Scan (Async)
	r.Post("/data-sources/{sourceID}/scan", h.ScanDataSource)

	// Test Connection
	r.Post("/data-sources/{sourceID}/test", h.TestConnection)

	// Scan Status & History
	r.Get("/data-sources/{sourceID}/scan/status", h.GetScanStatus)
	r.Get("/data-sources/{sourceID}/scan/history", h.GetScanHistory)

	// Inventory for a data source
	r.Get("/data-sources/{sourceID}/inventory", h.GetInventory)

	// Entities for an inventory
	r.Get("/inventories/{inventoryID}/entities", h.ListEntities)
	r.Get("/entities/{entityID}", h.GetEntity)

	// Fields for an entity
	r.Get("/entities/{entityID}/fields", h.ListFields)
	r.Get("/fields/{fieldID}", h.GetField)

	// Classifications (Review Queue)
	r.Get("/classifications", h.GetClassifications)

	return r
}

// GetClassifications returns a paginated list of PII classifications with filters.
// GET /api/v2/discovery/classifications
func (h *DiscoveryHandler) GetClassifications(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := mw.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "tenant context missing")
		return
	}

	filter := discovery.ClassificationFilter{
		Pagination: httputil.ParsePagination(r),
	}

	if dsIDStr := r.URL.Query().Get("data_source_id"); dsIDStr != "" {
		id, err := httputil.ParseID(dsIDStr)
		if err != nil {
			httputil.ErrorFromDomain(w, err)
			return
		}
		filter.DataSourceID = &id
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := types.VerificationStatus(statusStr)
		filter.Status = &s
	}

	if methodStr := r.URL.Query().Get("detection_method"); methodStr != "" {
		m := types.DetectionMethod(methodStr)
		filter.DetectionMethod = &m
	}

	result, err := h.service.GetClassifications(r.Context(), tenantID, filter)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, result)
}

// TestConnection tests connectivity to a data source.
// POST /api/v2/data-sources/{sourceID}/test
func (h *DiscoveryHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	sourceID, err := httputil.ParseID(chi.URLParam(r, "sourceID"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if err := h.service.TestConnection(r.Context(), sourceID); err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Connection successful",
	})
}

// ScanDataSource triggers a background scan of the data source.
// POST /api/v2/data-sources/{sourceID}/scan
func (h *DiscoveryHandler) ScanDataSource(w http.ResponseWriter, r *http.Request) {
	sourceID, err := httputil.ParseID(chi.URLParam(r, "sourceID"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	tenantID, ok := mw.TenantIDFromContext(r.Context())
	if !ok {
		httputil.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "tenant context missing")
		return
	}

	// Queue the scan job
	run, err := h.scanService.EnqueueScan(r.Context(), sourceID, tenantID, discovery.ScanTypeFull)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	// Return 202 Accepted with Job ID
	httputil.JSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "Scan queued",
		"job_id":  run.ID,
		"status":  run.Status,
	})
}

// GetScanStatus returns the status of a specific scan job or the latest for a source.
// GET /api/v2/data-sources/{sourceID}/scan/status
func (h *DiscoveryHandler) GetScanStatus(w http.ResponseWriter, r *http.Request) {
	// For now, this just gets the latest history since we don't have a "latest" pointer easily matching "status" semantics.
	// Actually, let's implement getting the LATEST scan for the source.
	// We'll reuse GetHistory and take the first one.

	sourceID, err := httputil.ParseID(chi.URLParam(r, "sourceID"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	history, err := h.scanService.GetHistory(r.Context(), sourceID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	if len(history) == 0 {
		httputil.ErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "no scans found for data source")
		return
	}

	httputil.JSON(w, http.StatusOK, history[0]) // Return latest
}

// GetScanHistory returns the scan history for a data source.
// GET /api/v2/data-sources/{sourceID}/scan/history
func (h *DiscoveryHandler) GetScanHistory(w http.ResponseWriter, r *http.Request) {
	sourceID, err := httputil.ParseID(chi.URLParam(r, "sourceID"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	history, err := h.scanService.GetHistory(r.Context(), sourceID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, history)
}

// GetInventory returns the data inventory for a given data source.
func (h *DiscoveryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	sourceID, err := httputil.ParseID(chi.URLParam(r, "sourceID"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	inv, err := h.inventoryRepo.GetByDataSource(r.Context(), sourceID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, inv)
}

// ListEntities returns all data entities for a given inventory.
func (h *DiscoveryHandler) ListEntities(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := httputil.ParseID(chi.URLParam(r, "inventoryID"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	entities, err := h.entityRepo.GetByInventory(r.Context(), inventoryID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, entities)
}

// GetEntity returns a single data entity by ID.
func (h *DiscoveryHandler) GetEntity(w http.ResponseWriter, r *http.Request) {
	entityID, err := httputil.ParseID(chi.URLParam(r, "entityID"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	entity, err := h.entityRepo.GetByID(r.Context(), entityID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, entity)
}

// ListFields returns all data fields for a given entity.
func (h *DiscoveryHandler) ListFields(w http.ResponseWriter, r *http.Request) {
	entityID, err := httputil.ParseID(chi.URLParam(r, "entityID"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	fields, err := h.fieldRepo.GetByEntity(r.Context(), entityID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, fields)
}

// GetField returns a single data field by ID.
func (h *DiscoveryHandler) GetField(w http.ResponseWriter, r *http.Request) {
	fieldID, err := httputil.ParseID(chi.URLParam(r, "fieldID"))
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	field, err := h.fieldRepo.GetByID(r.Context(), fieldID)
	if err != nil {
		httputil.ErrorFromDomain(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, field)
}
