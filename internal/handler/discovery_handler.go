package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/httputil"
)

// DiscoveryHandler handles discovery-related REST endpoints
// for data inventories, entities, and fields.
type DiscoveryHandler struct {
	inventoryRepo discovery.DataInventoryRepository
	entityRepo    discovery.DataEntityRepository
	fieldRepo     discovery.DataFieldRepository
}

// NewDiscoveryHandler creates a new DiscoveryHandler.
func NewDiscoveryHandler(
	inventoryRepo discovery.DataInventoryRepository,
	entityRepo discovery.DataEntityRepository,
	fieldRepo discovery.DataFieldRepository,
) *DiscoveryHandler {
	return &DiscoveryHandler{
		inventoryRepo: inventoryRepo,
		entityRepo:    entityRepo,
		fieldRepo:     fieldRepo,
	}
}

// Routes returns a chi.Router with discovery routes mounted.
// These are nested under /data-sources/{sourceID}/...
func (h *DiscoveryHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Inventory for a data source
	r.Get("/data-sources/{sourceID}/inventory", h.GetInventory)

	// Entities for an inventory
	r.Get("/inventories/{inventoryID}/entities", h.ListEntities)
	r.Get("/entities/{entityID}", h.GetEntity)

	// Fields for an entity
	r.Get("/entities/{entityID}/fields", h.ListFields)
	r.Get("/fields/{fieldID}", h.GetField)

	return r
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
