package repository

import (
	"context"
	"fmt"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresLineageRepository struct {
	db *pgxpool.Pool
}

func NewPostgresLineageRepository(db *pgxpool.Pool) *PostgresLineageRepository {
	return &PostgresLineageRepository{db: db}
}

func (r *PostgresLineageRepository) Create(ctx context.Context, flow *governance.DataFlow) error {
	query := `
		INSERT INTO data_flows (id, tenant_id, source_id, destination_id, data_type, data_path, purpose_id, status, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		flow.ID, flow.TenantID, flow.SourceID, flow.DestinationID,
		flow.DataType, flow.DataPath, flow.PurposeID, flow.Status, flow.Description,
		flow.CreatedAt, flow.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create data flow: %w", err)
	}
	return nil
}

func (r *PostgresLineageRepository) GetByTenant(ctx context.Context, tenantID types.ID) ([]governance.DataFlow, error) {
	query := `
		SELECT id, tenant_id, source_id, destination_id, data_type, data_path, purpose_id, status, description, created_at, updated_at
		FROM data_flows
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query data flows: %w", err)
	}
	defer rows.Close()

	var flows []governance.DataFlow
	for rows.Next() {
		var f governance.DataFlow
		if err := rows.Scan(
			&f.ID, &f.TenantID, &f.SourceID, &f.DestinationID,
			&f.DataType, &f.DataPath, &f.PurposeID, &f.Status, &f.Description,
			&f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan data flow: %w", err)
		}
		flows = append(flows, f)
	}
	return flows, nil
}

func (r *PostgresLineageRepository) GetBySource(ctx context.Context, sourceID types.ID) ([]governance.DataFlow, error) {
	query := `
		SELECT id, tenant_id, source_id, destination_id, data_type, data_path, purpose_id, status, description, created_at, updated_at
		FROM data_flows
		WHERE source_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, sourceID)
	if err != nil {
		return nil, fmt.Errorf("query data flows by source: %w", err)
	}
	defer rows.Close()

	var flows []governance.DataFlow
	for rows.Next() {
		var f governance.DataFlow
		if err := rows.Scan(
			&f.ID, &f.TenantID, &f.SourceID, &f.DestinationID,
			&f.DataType, &f.DataPath, &f.PurposeID, &f.Status, &f.Description,
			&f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan data flow: %w", err)
		}
		flows = append(flows, f)
	}
	return flows, nil
}

func (r *PostgresLineageRepository) GetByDestination(ctx context.Context, destID types.ID) ([]governance.DataFlow, error) {
	query := `
		SELECT id, tenant_id, source_id, destination_id, data_type, data_path, purpose_id, status, description, created_at, updated_at
		FROM data_flows
		WHERE destination_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, destID)
	if err != nil {
		return nil, fmt.Errorf("query data flows by destination: %w", err)
	}
	defer rows.Close()

	var flows []governance.DataFlow
	for rows.Next() {
		var f governance.DataFlow
		if err := rows.Scan(
			&f.ID, &f.TenantID, &f.SourceID, &f.DestinationID,
			&f.DataType, &f.DataPath, &f.PurposeID, &f.Status, &f.Description,
			&f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan data flow: %w", err)
		}
		flows = append(flows, f)
	}
	return flows, nil
}
