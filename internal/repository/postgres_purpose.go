package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/types"
)

// PurposeRepo implements governance.PurposeRepository.
type PurposeRepo struct {
	pool *pgxpool.Pool
}

// NewPurposeRepo creates a new PurposeRepo.
func NewPurposeRepo(pool *pgxpool.Pool) *PurposeRepo {
	return &PurposeRepo{pool: pool}
}

func (r *PurposeRepo) Create(ctx context.Context, p *governance.Purpose) error {
	p.ID = types.NewID()
	query := `
		INSERT INTO purposes (id, tenant_id, code, name, description, legal_basis,
		    retention_days, is_active, requires_consent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		p.ID, p.TenantID, p.Code, p.Name, p.Description, p.LegalBasis,
		p.RetentionDays, p.IsActive, p.RequiresConsent,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
}

func (r *PurposeRepo) GetByID(ctx context.Context, id types.ID) (*governance.Purpose, error) {
	query := `
		SELECT id, tenant_id, code, name, description, legal_basis,
		       retention_days, is_active, requires_consent, created_at, updated_at
		FROM purposes
		WHERE id = $1`

	p := &governance.Purpose{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.TenantID, &p.Code, &p.Name, &p.Description, &p.LegalBasis,
		&p.RetentionDays, &p.IsActive, &p.RequiresConsent, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("Purpose", id)
		}
		return nil, fmt.Errorf("get purpose: %w", err)
	}
	return p, nil
}

func (r *PurposeRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]governance.Purpose, error) {
	query := `
		SELECT id, tenant_id, code, name, description, legal_basis,
		       retention_days, is_active, requires_consent, created_at, updated_at
		FROM purposes
		WHERE tenant_id = $1
		ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list purposes: %w", err)
	}
	defer rows.Close()

	var results []governance.Purpose
	for rows.Next() {
		var p governance.Purpose
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.Code, &p.Name, &p.Description, &p.LegalBasis,
			&p.RetentionDays, &p.IsActive, &p.RequiresConsent, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan purpose: %w", err)
		}
		results = append(results, p)
	}
	return results, rows.Err()
}

func (r *PurposeRepo) GetByCode(ctx context.Context, tenantID types.ID, code string) (*governance.Purpose, error) {
	query := `
		SELECT id, tenant_id, code, name, description, legal_basis,
		       retention_days, is_active, requires_consent, created_at, updated_at
		FROM purposes
		WHERE tenant_id = $1 AND code = $2`

	p := &governance.Purpose{}
	err := r.pool.QueryRow(ctx, query, tenantID, code).Scan(
		&p.ID, &p.TenantID, &p.Code, &p.Name, &p.Description, &p.LegalBasis,
		&p.RetentionDays, &p.IsActive, &p.RequiresConsent, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("Purpose", code)
		}
		return nil, fmt.Errorf("get purpose by code: %w", err)
	}
	return p, nil
}

func (r *PurposeRepo) Update(ctx context.Context, p *governance.Purpose) error {
	query := `
		UPDATE purposes
		SET name = $2, description = $3, legal_basis = $4, retention_days = $5,
		    is_active = $6, requires_consent = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		p.ID, p.Name, p.Description, p.LegalBasis, p.RetentionDays, p.IsActive, p.RequiresConsent,
	).Scan(&p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.NewNotFoundError("Purpose", p.ID)
		}
		return fmt.Errorf("update purpose: %w", err)
	}
	return nil
}

func (r *PurposeRepo) Delete(ctx context.Context, id types.ID) error {
	query := `DELETE FROM purposes WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete purpose: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return types.NewNotFoundError("Purpose", id)
	}
	return nil
}

// Compile-time check.
var _ governance.PurposeRepository = (*PurposeRepo)(nil)
