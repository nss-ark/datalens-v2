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

// PurposeAssignmentRepo implements governance.PurposeAssignmentRepository.
type PurposeAssignmentRepo struct {
	pool *pgxpool.Pool
}

// NewPurposeAssignmentRepo creates a new PurposeAssignmentRepo.
func NewPurposeAssignmentRepo(pool *pgxpool.Pool) *PurposeAssignmentRepo {
	return &PurposeAssignmentRepo{pool: pool}
}

func (r *PurposeAssignmentRepo) Create(ctx context.Context, a *governance.PurposeAssignment) error {
	a.ID = types.NewID()
	a.AssignedAt = time.Now()

	query := `
		INSERT INTO purpose_assignments (
			id, tenant_id, purpose_id, scope_type, scope_id,
			scope_name, inherited, overridden_by, assigned_by, assigned_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		a.ID, a.TenantID, a.PurposeID, a.ScopeType, a.ScopeID,
		a.ScopeName, a.Inherited, a.OverriddenBy, a.AssignedBy, a.AssignedAt,
	)
	if err != nil {
		return fmt.Errorf("create purpose assignment: %w", err)
	}
	return nil
}

func (r *PurposeAssignmentRepo) Delete(ctx context.Context, id types.ID) error {
	query := `DELETE FROM purpose_assignments WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete purpose assignment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return types.NewNotFoundError("PurposeAssignment", id)
	}
	return nil
}

func (r *PurposeAssignmentRepo) GetByScope(ctx context.Context, tenantID types.ID, scopeType governance.ScopeType, scopeID string) ([]governance.PurposeAssignment, error) {
	query := `
		SELECT id, tenant_id, purpose_id, scope_type, scope_id,
		       scope_name, inherited, overridden_by, assigned_by, assigned_at
		FROM purpose_assignments
		WHERE tenant_id = $1 AND scope_type = $2 AND scope_id = $3
		ORDER BY assigned_at DESC
	`
	return r.scanRows(ctx, query, tenantID, scopeType, scopeID)
}

func (r *PurposeAssignmentRepo) GetByPurpose(ctx context.Context, tenantID types.ID, purposeID types.ID) ([]governance.PurposeAssignment, error) {
	query := `
		SELECT id, tenant_id, purpose_id, scope_type, scope_id,
		       scope_name, inherited, overridden_by, assigned_by, assigned_at
		FROM purpose_assignments
		WHERE tenant_id = $1 AND purpose_id = $2
		ORDER BY assigned_at DESC
	`
	return r.scanRows(ctx, query, tenantID, purposeID)
}

func (r *PurposeAssignmentRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]governance.PurposeAssignment, error) {
	query := `
		SELECT id, tenant_id, purpose_id, scope_type, scope_id,
		       scope_name, inherited, overridden_by, assigned_by, assigned_at
		FROM purpose_assignments
		WHERE tenant_id = $1
		ORDER BY assigned_at DESC
	`
	return r.scanRows(ctx, query, tenantID)
}

func (r *PurposeAssignmentRepo) scanRows(ctx context.Context, query string, args ...any) ([]governance.PurposeAssignment, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query purpose assignments: %w", err)
	}
	defer rows.Close()

	var results []governance.PurposeAssignment
	for rows.Next() {
		var a governance.PurposeAssignment
		if err := rows.Scan(
			&a.ID, &a.TenantID, &a.PurposeID, &a.ScopeType, &a.ScopeID,
			&a.ScopeName, &a.Inherited, &a.OverriddenBy, &a.AssignedBy, &a.AssignedAt,
		); err != nil {
			return nil, fmt.Errorf("scan purpose assignment: %w", err)
		}
		results = append(results, a)
	}
	return results, nil
}
