package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DepartmentRepo implements governance.DepartmentRepository.
type DepartmentRepo struct {
	pool *pgxpool.Pool
}

// NewDepartmentRepo creates a new DepartmentRepo.
func NewDepartmentRepo(pool *pgxpool.Pool) *DepartmentRepo {
	return &DepartmentRepo{pool: pool}
}

func (r *DepartmentRepo) Create(ctx context.Context, d *governance.Department) error {
	d.ID = types.NewID()
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	if !d.IsActive {
		d.IsActive = true
	}

	query := `
		INSERT INTO departments (
			id, tenant_id, name, description, owner_id, owner_name, owner_email,
			responsibilities, notification_enabled, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.pool.Exec(ctx, query,
		d.ID, d.TenantID, d.Name, d.Description, d.OwnerID, d.OwnerName, d.OwnerEmail,
		d.Responsibilities, d.NotificationEnabled, d.IsActive, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create department: %w", err)
	}
	return nil
}

func (r *DepartmentRepo) GetByID(ctx context.Context, id types.ID) (*governance.Department, error) {
	query := `
		SELECT id, tenant_id, name, description, owner_id, owner_name, owner_email,
		       responsibilities, notification_enabled, is_active, created_at, updated_at
		FROM departments WHERE id = $1
	`
	var d governance.Department
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.TenantID, &d.Name, &d.Description, &d.OwnerID, &d.OwnerName, &d.OwnerEmail,
		&d.Responsibilities, &d.NotificationEnabled, &d.IsActive, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("Department", id)
		}
		return nil, fmt.Errorf("get department: %w", err)
	}
	return &d, nil
}

func (r *DepartmentRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]governance.Department, error) {
	query := `
		SELECT id, tenant_id, name, description, owner_id, owner_name, owner_email,
		       responsibilities, notification_enabled, is_active, created_at, updated_at
		FROM departments WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query departments: %w", err)
	}
	defer rows.Close()

	var results []governance.Department
	for rows.Next() {
		var d governance.Department
		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.Name, &d.Description, &d.OwnerID, &d.OwnerName, &d.OwnerEmail,
			&d.Responsibilities, &d.NotificationEnabled, &d.IsActive, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan department: %w", err)
		}
		results = append(results, d)
	}
	return results, nil
}

func (r *DepartmentRepo) GetByOwner(ctx context.Context, ownerID types.ID) ([]governance.Department, error) {
	query := `
		SELECT id, tenant_id, name, description, owner_id, owner_name, owner_email,
		       responsibilities, notification_enabled, is_active, created_at, updated_at
		FROM departments WHERE owner_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("query departments by owner: %w", err)
	}
	defer rows.Close()

	var results []governance.Department
	for rows.Next() {
		var d governance.Department
		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.Name, &d.Description, &d.OwnerID, &d.OwnerName, &d.OwnerEmail,
			&d.Responsibilities, &d.NotificationEnabled, &d.IsActive, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan department: %w", err)
		}
		results = append(results, d)
	}
	return results, nil
}

func (r *DepartmentRepo) Update(ctx context.Context, d *governance.Department) error {
	d.UpdatedAt = time.Now()

	query := `
		UPDATE departments SET
			name = $2, description = $3, owner_id = $4, owner_name = $5,
			owner_email = $6, responsibilities = $7, notification_enabled = $8,
			is_active = $9, updated_at = $10
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query,
		d.ID, d.Name, d.Description, d.OwnerID, d.OwnerName,
		d.OwnerEmail, d.Responsibilities, d.NotificationEnabled,
		d.IsActive, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update department: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return types.NewNotFoundError("Department", d.ID)
	}
	return nil
}

func (r *DepartmentRepo) Delete(ctx context.Context, id types.ID) error {
	query := `DELETE FROM departments WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete department: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return types.NewNotFoundError("Department", id)
	}
	return nil
}
