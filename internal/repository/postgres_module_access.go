package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// ModuleAccessRepo implements identity.ModuleAccessRepository.
type ModuleAccessRepo struct {
	pool *pgxpool.Pool
}

// NewModuleAccessRepo creates a new ModuleAccessRepo.
func NewModuleAccessRepo(pool *pgxpool.Pool) *ModuleAccessRepo {
	return &ModuleAccessRepo{pool: pool}
}

func (r *ModuleAccessRepo) GetByTenantID(ctx context.Context, tenantID types.ID) ([]identity.ModuleAccess, error) {
	query := `
		SELECT id, tenant_id, module_name, enabled, created_at, updated_at
		FROM module_access
		WHERE tenant_id = $1
		ORDER BY module_name ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list module access: %w", err)
	}
	defer rows.Close()

	var results []identity.ModuleAccess
	for rows.Next() {
		var m identity.ModuleAccess
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.ModuleName, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan module access: %w", err)
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (r *ModuleAccessRepo) SetModules(ctx context.Context, tenantID types.ID, modules []identity.ModuleAccess) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing rows for this tenant
	if _, err := tx.Exec(ctx, "DELETE FROM module_access WHERE tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("delete existing modules: %w", err)
	}

	// Insert new rows
	for _, m := range modules {
		id := types.NewID()
		_, err := tx.Exec(ctx,
			`INSERT INTO module_access (id, tenant_id, module_name, enabled) VALUES ($1, $2, $3, $4)`,
			id, tenantID, m.ModuleName, m.Enabled,
		)
		if err != nil {
			return fmt.Errorf("insert module %s: %w", m.ModuleName, err)
		}
	}

	return tx.Commit(ctx)
}

var _ identity.ModuleAccessRepository = (*ModuleAccessRepo)(nil)
