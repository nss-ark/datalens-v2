package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RetentionRepo struct {
	pool *pgxpool.Pool
}

func NewRetentionRepo(pool *pgxpool.Pool) *RetentionRepo {
	return &RetentionRepo{pool: pool}
}

func (r *RetentionRepo) Create(ctx context.Context, p *compliance.RetentionPolicy) error {
	p.ID = types.NewID()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	query := `
		INSERT INTO retention_policies (
			id, tenant_id, purpose_id, max_retention_days, data_categories, 
			status, auto_erase, description, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.TenantID, p.PurposeID, p.MaxRetentionDays, p.DataCategories,
		p.Status, p.AutoErase, p.Description, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create retention policy: %w", err)
	}
	return nil
}

func (r *RetentionRepo) GetByID(ctx context.Context, id types.ID) (*compliance.RetentionPolicy, error) {
	query := `
		SELECT id, tenant_id, purpose_id, max_retention_days, data_categories, 
		       status, auto_erase, description, created_at, updated_at
		FROM retention_policies WHERE id = $1
	`
	var p compliance.RetentionPolicy
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.TenantID, &p.PurposeID, &p.MaxRetentionDays, &p.DataCategories,
		&p.Status, &p.AutoErase, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("RetentionPolicy", id)
		}
		return nil, fmt.Errorf("get retention policy: %w", err)
	}
	return &p, nil
}

func (r *RetentionRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]compliance.RetentionPolicy, error) {
	query := `
		SELECT id, tenant_id, purpose_id, max_retention_days, data_categories, 
		       status, auto_erase, description, created_at, updated_at
		FROM retention_policies WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query retention policies: %w", err)
	}
	defer rows.Close()

	var policies []compliance.RetentionPolicy
	for rows.Next() {
		var p compliance.RetentionPolicy
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.PurposeID, &p.MaxRetentionDays, &p.DataCategories,
			&p.Status, &p.AutoErase, &p.Description, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan retention policy: %w", err)
		}
		policies = append(policies, p)
	}
	return policies, nil
}

func (r *RetentionRepo) Update(ctx context.Context, p *compliance.RetentionPolicy) error {
	p.UpdatedAt = time.Now()
	query := `
		UPDATE retention_policies SET
			purpose_id = $2, max_retention_days = $3, data_categories = $4,
			status = $5, auto_erase = $6, description = $7, updated_at = $8
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query,
		p.ID, p.PurposeID, p.MaxRetentionDays, p.DataCategories,
		p.Status, p.AutoErase, p.Description, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update retention policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return types.NewNotFoundError("RetentionPolicy", p.ID)
	}
	return nil
}

func (r *RetentionRepo) Delete(ctx context.Context, id types.ID) error {
	query := `DELETE FROM retention_policies WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete retention policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return types.NewNotFoundError("RetentionPolicy", id)
	}
	return nil
}

// CreateLog persists a new retention action log entry.
func (r *RetentionRepo) CreateLog(ctx context.Context, log *compliance.RetentionLog) error {
	log.ID = types.NewID()
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	query := `
		INSERT INTO retention_logs (id, tenant_id, policy_id, action, target, details, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		log.ID, log.TenantID, log.PolicyID, log.Action, log.Target, log.Details, log.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("create retention log: %w", err)
	}
	return nil
}

// GetLogs retrieves paginated retention log entries, optionally filtered by policy ID.
func (r *RetentionRepo) GetLogs(ctx context.Context, tenantID types.ID, policyID *types.ID, pagination types.Pagination) (*types.PaginatedResult[compliance.RetentionLog], error) {
	whereClause := "WHERE tenant_id = $1"
	args := []any{tenantID}
	argIdx := 2

	if policyID != nil {
		whereClause += fmt.Sprintf(" AND policy_id = $%d", argIdx)
		args = append(args, *policyID)
		argIdx++
	}

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM retention_logs %s", whereClause)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count retention logs: %w", err)
	}

	// Select
	selectQuery := fmt.Sprintf(`
		SELECT id, tenant_id, policy_id, action, target, details, timestamp
		FROM retention_logs %s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, pagination.Limit(), pagination.Offset())

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query retention logs: %w", err)
	}
	defer rows.Close()

	var items []compliance.RetentionLog
	for rows.Next() {
		var l compliance.RetentionLog
		if err := rows.Scan(&l.ID, &l.TenantID, &l.PolicyID, &l.Action, &l.Target, &l.Details, &l.Timestamp); err != nil {
			return nil, fmt.Errorf("scan retention log: %w", err)
		}
		items = append(items, l)
	}

	return &types.PaginatedResult[compliance.RetentionLog]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: (total + pagination.PageSize - 1) / pagination.PageSize,
	}, nil
}
