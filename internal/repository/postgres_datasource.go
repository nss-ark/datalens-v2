// Package repository provides PostgreSQL implementations of domain
// repository interfaces using pgx/v5.
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// DataSourceRepo implements discovery.DataSourceRepository.
type DataSourceRepo struct {
	pool *pgxpool.Pool
}

// NewDataSourceRepo creates a new DataSourceRepo.
func NewDataSourceRepo(pool *pgxpool.Pool) *DataSourceRepo {
	return &DataSourceRepo{pool: pool}
}

func (r *DataSourceRepo) Create(ctx context.Context, ds *discovery.DataSource) error {
	ds.ID = types.NewID()
	query := `
		INSERT INTO data_sources (id, tenant_id, name, type, description, host, port, database_name, credentials, config, scan_schedule, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		ds.ID, ds.TenantID, ds.Name, ds.Type, ds.Description,
		ds.Host, ds.Port, ds.Database, ds.Credentials, ds.Config, ds.ScanSchedule, ds.Status,
	).Scan(&ds.CreatedAt, &ds.UpdatedAt)
}

func (r *DataSourceRepo) GetByID(ctx context.Context, id types.ID) (*discovery.DataSource, error) {
	query := `
		SELECT id, tenant_id, name, type, description, host, port, database_name, credentials,
		       config, scan_schedule, status, last_sync_at, error_message, created_at, updated_at
		FROM data_sources
		WHERE id = $1 AND deleted_at IS NULL`

	ds := &discovery.DataSource{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ds.ID, &ds.TenantID, &ds.Name, &ds.Type, &ds.Description,
		&ds.Host, &ds.Port, &ds.Database, &ds.Credentials,
		&ds.Config, &ds.ScanSchedule, &ds.Status, &ds.LastSyncAt, &ds.ErrorMessage, &ds.CreatedAt, &ds.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("DataSource", id)
		}
		return nil, fmt.Errorf("get data source: %w", err)
	}
	return ds, nil
}

func (r *DataSourceRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]discovery.DataSource, error) {
	query := `
		SELECT id, tenant_id, name, type, description, host, port, database_name, credentials,
		       config, scan_schedule, status, last_sync_at, error_message, created_at, updated_at
		FROM data_sources
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list data sources: %w", err)
	}
	defer rows.Close()

	var results []discovery.DataSource
	for rows.Next() {
		var ds discovery.DataSource
		if err := rows.Scan(
			&ds.ID, &ds.TenantID, &ds.Name, &ds.Type, &ds.Description,
			&ds.Host, &ds.Port, &ds.Database, &ds.Credentials,
			&ds.Config, &ds.ScanSchedule, &ds.Status, &ds.LastSyncAt, &ds.ErrorMessage, &ds.CreatedAt, &ds.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan data source: %w", err)
		}
		results = append(results, ds)
	}
	return results, rows.Err()
}

func (r *DataSourceRepo) Update(ctx context.Context, ds *discovery.DataSource) error {
	query := `
		UPDATE data_sources
		SET name = $2, type = $3, description = $4, host = $5, port = $6,
		    database_name = $7, credentials = $8, config = $9, scan_schedule = $10,
		    status = $11, last_sync_at = $12, error_message = $13, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		ds.ID, ds.Name, ds.Type, ds.Description, ds.Host, ds.Port,
		ds.Database, ds.Credentials, ds.Config, ds.ScanSchedule,
		ds.Status, ds.LastSyncAt, ds.ErrorMessage,
	).Scan(&ds.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.NewNotFoundError("DataSource", ds.ID)
		}
		return fmt.Errorf("update data source: %w", err)
	}
	return nil
}

func (r *DataSourceRepo) Delete(ctx context.Context, id types.ID) error {
	query := `UPDATE data_sources SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete data source: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return types.NewNotFoundError("DataSource", id)
	}
	return nil
}

// Compile-time check.
var _ discovery.DataSourceRepository = (*DataSourceRepo)(nil)
