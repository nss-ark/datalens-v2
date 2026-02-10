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

// =============================================================================
// DataInventoryRepo
// =============================================================================

// DataInventoryRepo implements discovery.DataInventoryRepository.
type DataInventoryRepo struct {
	pool *pgxpool.Pool
}

// NewDataInventoryRepo creates a new DataInventoryRepo.
func NewDataInventoryRepo(pool *pgxpool.Pool) *DataInventoryRepo {
	return &DataInventoryRepo{pool: pool}
}

func (r *DataInventoryRepo) Create(ctx context.Context, inv *discovery.DataInventory) error {
	inv.ID = types.NewID()
	query := `
		INSERT INTO data_inventories (id, data_source_id, total_entities, total_fields, pii_fields_count, last_scanned_at, schema_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		inv.ID, inv.DataSourceID, inv.TotalEntities, inv.TotalFields,
		inv.PIIFieldsCount, inv.LastScannedAt, inv.SchemaVersion,
	).Scan(&inv.CreatedAt, &inv.UpdatedAt)
}

func (r *DataInventoryRepo) GetByID(ctx context.Context, id types.ID) (*discovery.DataInventory, error) {
	query := `
		SELECT id, data_source_id, total_entities, total_fields, pii_fields_count,
		       last_scanned_at, schema_version, created_at, updated_at
		FROM data_inventories
		WHERE id = $1`

	inv := &discovery.DataInventory{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&inv.ID, &inv.DataSourceID, &inv.TotalEntities, &inv.TotalFields,
		&inv.PIIFieldsCount, &inv.LastScannedAt, &inv.SchemaVersion,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("DataInventory", id)
		}
		return nil, fmt.Errorf("get data inventory: %w", err)
	}
	return inv, nil
}

func (r *DataInventoryRepo) GetByDataSource(ctx context.Context, dataSourceID types.ID) (*discovery.DataInventory, error) {
	query := `
		SELECT id, data_source_id, total_entities, total_fields, pii_fields_count,
		       last_scanned_at, schema_version, created_at, updated_at
		FROM data_inventories
		WHERE data_source_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	inv := &discovery.DataInventory{}
	err := r.pool.QueryRow(ctx, query, dataSourceID).Scan(
		&inv.ID, &inv.DataSourceID, &inv.TotalEntities, &inv.TotalFields,
		&inv.PIIFieldsCount, &inv.LastScannedAt, &inv.SchemaVersion,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("DataInventory for DataSource", dataSourceID)
		}
		return nil, fmt.Errorf("get data inventory by source: %w", err)
	}
	return inv, nil
}

func (r *DataInventoryRepo) Update(ctx context.Context, inv *discovery.DataInventory) error {
	query := `
		UPDATE data_inventories
		SET total_entities = $2, total_fields = $3, pii_fields_count = $4,
		    last_scanned_at = $5, schema_version = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		inv.ID, inv.TotalEntities, inv.TotalFields, inv.PIIFieldsCount,
		inv.LastScannedAt, inv.SchemaVersion,
	).Scan(&inv.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.NewNotFoundError("DataInventory", inv.ID)
		}
		return fmt.Errorf("update data inventory: %w", err)
	}
	return nil
}

// Compile-time check.
var _ discovery.DataInventoryRepository = (*DataInventoryRepo)(nil)

// =============================================================================
// DataEntityRepo
// =============================================================================

// DataEntityRepo implements discovery.DataEntityRepository.
type DataEntityRepo struct {
	pool *pgxpool.Pool
}

// NewDataEntityRepo creates a new DataEntityRepo.
func NewDataEntityRepo(pool *pgxpool.Pool) *DataEntityRepo {
	return &DataEntityRepo{pool: pool}
}

func (r *DataEntityRepo) Create(ctx context.Context, entity *discovery.DataEntity) error {
	entity.ID = types.NewID()
	query := `
		INSERT INTO data_entities (id, inventory_id, name, schema_name, type, row_count, pii_confidence)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		entity.ID, entity.InventoryID, entity.Name, entity.Schema,
		entity.Type, entity.RowCount, entity.PIIConfidence,
	).Scan(&entity.CreatedAt, &entity.UpdatedAt)
}

func (r *DataEntityRepo) GetByID(ctx context.Context, id types.ID) (*discovery.DataEntity, error) {
	query := `
		SELECT id, inventory_id, name, schema_name, type, row_count, pii_confidence,
		       created_at, updated_at
		FROM data_entities
		WHERE id = $1`

	entity := &discovery.DataEntity{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&entity.ID, &entity.InventoryID, &entity.Name, &entity.Schema,
		&entity.Type, &entity.RowCount, &entity.PIIConfidence,
		&entity.CreatedAt, &entity.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("DataEntity", id)
		}
		return nil, fmt.Errorf("get data entity: %w", err)
	}
	return entity, nil
}

func (r *DataEntityRepo) GetByInventory(ctx context.Context, inventoryID types.ID) ([]discovery.DataEntity, error) {
	query := `
		SELECT id, inventory_id, name, schema_name, type, row_count, pii_confidence,
		       created_at, updated_at
		FROM data_entities
		WHERE inventory_id = $1
		ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("list data entities: %w", err)
	}
	defer rows.Close()

	var results []discovery.DataEntity
	for rows.Next() {
		var entity discovery.DataEntity
		if err := rows.Scan(
			&entity.ID, &entity.InventoryID, &entity.Name, &entity.Schema,
			&entity.Type, &entity.RowCount, &entity.PIIConfidence,
			&entity.CreatedAt, &entity.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan data entity: %w", err)
		}
		results = append(results, entity)
	}
	return results, rows.Err()
}

func (r *DataEntityRepo) Update(ctx context.Context, entity *discovery.DataEntity) error {
	query := `
		UPDATE data_entities
		SET name = $2, schema_name = $3, type = $4, row_count = $5,
		    pii_confidence = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		entity.ID, entity.Name, entity.Schema, entity.Type,
		entity.RowCount, entity.PIIConfidence,
	).Scan(&entity.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.NewNotFoundError("DataEntity", entity.ID)
		}
		return fmt.Errorf("update data entity: %w", err)
	}
	return nil
}

func (r *DataEntityRepo) Delete(ctx context.Context, id types.ID) error {
	query := `DELETE FROM data_entities WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete data entity: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return types.NewNotFoundError("DataEntity", id)
	}
	return nil
}

// Compile-time check.
var _ discovery.DataEntityRepository = (*DataEntityRepo)(nil)

// =============================================================================
// DataFieldRepo
// =============================================================================

// DataFieldRepo implements discovery.DataFieldRepository.
type DataFieldRepo struct {
	pool *pgxpool.Pool
}

// NewDataFieldRepo creates a new DataFieldRepo.
func NewDataFieldRepo(pool *pgxpool.Pool) *DataFieldRepo {
	return &DataFieldRepo{pool: pool}
}

func (r *DataFieldRepo) Create(ctx context.Context, field *discovery.DataField) error {
	field.ID = types.NewID()
	query := `
		INSERT INTO data_fields (id, entity_id, name, data_type, nullable, is_primary_key, is_foreign_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		field.ID, field.EntityID, field.Name, field.DataType,
		field.Nullable, field.IsPrimaryKey, field.IsForeignKey,
	).Scan(&field.CreatedAt, &field.UpdatedAt)
}

func (r *DataFieldRepo) GetByID(ctx context.Context, id types.ID) (*discovery.DataField, error) {
	query := `
		SELECT id, entity_id, name, data_type, nullable, is_primary_key, is_foreign_key,
		       created_at, updated_at
		FROM data_fields
		WHERE id = $1`

	field := &discovery.DataField{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&field.ID, &field.EntityID, &field.Name, &field.DataType,
		&field.Nullable, &field.IsPrimaryKey, &field.IsForeignKey,
		&field.CreatedAt, &field.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("DataField", id)
		}
		return nil, fmt.Errorf("get data field: %w", err)
	}
	return field, nil
}

func (r *DataFieldRepo) GetByEntity(ctx context.Context, entityID types.ID) ([]discovery.DataField, error) {
	query := `
		SELECT id, entity_id, name, data_type, nullable, is_primary_key, is_foreign_key,
		       created_at, updated_at
		FROM data_fields
		WHERE entity_id = $1
		ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("list data fields: %w", err)
	}
	defer rows.Close()

	var results []discovery.DataField
	for rows.Next() {
		var field discovery.DataField
		if err := rows.Scan(
			&field.ID, &field.EntityID, &field.Name, &field.DataType,
			&field.Nullable, &field.IsPrimaryKey, &field.IsForeignKey,
			&field.CreatedAt, &field.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan data field: %w", err)
		}
		results = append(results, field)
	}
	return results, rows.Err()
}

func (r *DataFieldRepo) Update(ctx context.Context, field *discovery.DataField) error {
	query := `
		UPDATE data_fields
		SET name = $2, data_type = $3, nullable = $4,
		    is_primary_key = $5, is_foreign_key = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		field.ID, field.Name, field.DataType, field.Nullable,
		field.IsPrimaryKey, field.IsForeignKey,
	).Scan(&field.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.NewNotFoundError("DataField", field.ID)
		}
		return fmt.Errorf("update data field: %w", err)
	}
	return nil
}

func (r *DataFieldRepo) Delete(ctx context.Context, id types.ID) error {
	query := `DELETE FROM data_fields WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete data field: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return types.NewNotFoundError("DataField", id)
	}
	return nil
}

// Compile-time check.
var _ discovery.DataFieldRepository = (*DataFieldRepo)(nil)
