package breach

import (
	"context"

	"github.com/complyark/datalens/pkg/types"
)

type Filter struct {
	Status   *IncidentStatus
	Severity *IncidentSeverity
}

type Repository interface {
	Create(ctx context.Context, incident *BreachIncident) error
	GetByID(ctx context.Context, id types.ID) (*BreachIncident, error)
	Update(ctx context.Context, incident *BreachIncident) error
	List(ctx context.Context, tenantID types.ID, filter Filter, pagination types.Pagination) (*types.PaginatedResult[BreachIncident], error)
}
