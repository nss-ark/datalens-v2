package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoPARepo implements compliance.RoPARepository.
type RoPARepo struct {
	pool *pgxpool.Pool
}

// NewRoPARepo creates a new RoPARepo.
func NewRoPARepo(pool *pgxpool.Pool) *RoPARepo {
	return &RoPARepo{pool: pool}
}

func (r *RoPARepo) Create(ctx context.Context, version *compliance.RoPAVersion) error {
	version.ID = types.NewID()
	version.CreatedAt = time.Now()

	contentJSON, err := json.Marshal(version.Content)
	if err != nil {
		return fmt.Errorf("marshal ropa content: %w", err)
	}

	query := `
		INSERT INTO ropa_versions (
			id, tenant_id, version, generated_by, status,
			content, change_summary, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = r.pool.Exec(ctx, query,
		version.ID, version.TenantID, version.Version, version.GeneratedBy, version.Status,
		contentJSON, version.ChangeSummary, version.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create ropa version: %w", err)
	}
	return nil
}

func (r *RoPARepo) GetLatest(ctx context.Context, tenantID types.ID) (*compliance.RoPAVersion, error) {
	query := `
		SELECT id, tenant_id, version, generated_by, status,
		       content, change_summary, created_at
		FROM ropa_versions WHERE tenant_id = $1
		ORDER BY created_at DESC LIMIT 1
	`
	var v compliance.RoPAVersion
	var contentJSON []byte
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&v.ID, &v.TenantID, &v.Version, &v.GeneratedBy, &v.Status,
		&contentJSON, &v.ChangeSummary, &v.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No versions yet — not an error
		}
		return nil, fmt.Errorf("get latest ropa version: %w", err)
	}
	if err := json.Unmarshal(contentJSON, &v.Content); err != nil {
		return nil, fmt.Errorf("unmarshal ropa content: %w", err)
	}
	return &v, nil
}

func (r *RoPARepo) GetByVersion(ctx context.Context, tenantID types.ID, version string) (*compliance.RoPAVersion, error) {
	query := `
		SELECT id, tenant_id, version, generated_by, status,
		       content, change_summary, created_at
		FROM ropa_versions WHERE tenant_id = $1 AND version = $2
	`
	var v compliance.RoPAVersion
	var contentJSON []byte
	err := r.pool.QueryRow(ctx, query, tenantID, version).Scan(
		&v.ID, &v.TenantID, &v.Version, &v.GeneratedBy, &v.Status,
		&contentJSON, &v.ChangeSummary, &v.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("RoPAVersion", version)
		}
		return nil, fmt.Errorf("get ropa version: %w", err)
	}
	if err := json.Unmarshal(contentJSON, &v.Content); err != nil {
		return nil, fmt.Errorf("unmarshal ropa content: %w", err)
	}
	return &v, nil
}

func (r *RoPARepo) ListVersions(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[compliance.RoPAVersion], error) {
	// Count
	var total int
	countQuery := `SELECT COUNT(*) FROM ropa_versions WHERE tenant_id = $1`
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count ropa versions: %w", err)
	}

	// Select
	selectQuery := `
		SELECT id, tenant_id, version, generated_by, status,
		       content, change_summary, created_at
		FROM ropa_versions WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, selectQuery, tenantID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, fmt.Errorf("query ropa versions: %w", err)
	}
	defer rows.Close()

	var items []compliance.RoPAVersion
	for rows.Next() {
		var v compliance.RoPAVersion
		var contentJSON []byte
		if err := rows.Scan(
			&v.ID, &v.TenantID, &v.Version, &v.GeneratedBy, &v.Status,
			&contentJSON, &v.ChangeSummary, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ropa version: %w", err)
		}
		if err := json.Unmarshal(contentJSON, &v.Content); err != nil {
			return nil, fmt.Errorf("unmarshal ropa content: %w", err)
		}
		items = append(items, v)
	}

	return &types.PaginatedResult[compliance.RoPAVersion]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: (total + pagination.PageSize - 1) / pagination.PageSize,
	}, nil
}

func (r *RoPARepo) UpdateStatus(ctx context.Context, id types.ID, status compliance.RoPAStatus) error {
	query := `UPDATE ropa_versions SET status = $1 WHERE id = $2`
	tag, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update ropa status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return types.NewNotFoundError("RoPAVersion", id)
	}
	return nil
}
