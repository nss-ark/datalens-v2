package repository

import (
	"context"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresNoticeRepository struct {
	db *pgxpool.Pool
}

func NewPostgresNoticeRepository(db *pgxpool.Pool) *PostgresNoticeRepository {
	return &PostgresNoticeRepository{db: db}
}

func (r *PostgresNoticeRepository) Create(ctx context.Context, n *consent.ConsentNotice) error {
	query := `INSERT INTO consent_notices (
		id, tenant_id, series_id, title, content, version, status, purposes, widget_ids, regulation, published_at, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(ctx, query,
		n.ID, n.TenantID, n.SeriesID, n.Title, n.Content, n.Version, n.Status, n.Purposes, n.WidgetIDs, n.Regulation, n.PublishedAt, n.CreatedAt, n.UpdatedAt,
	)
	return err
}

func (r *PostgresNoticeRepository) GetByID(ctx context.Context, id types.ID) (*consent.ConsentNotice, error) {
	query := `SELECT 
		id, tenant_id, series_id, title, content, version, status, purposes, widget_ids, regulation, published_at, created_at, updated_at
	FROM consent_notices WHERE id = $1`

	var n consent.ConsentNotice
	err := r.db.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.TenantID, &n.SeriesID, &n.Title, &n.Content, &n.Version, &n.Status, &n.Purposes, &n.WidgetIDs, &n.Regulation, &n.PublishedAt, &n.CreatedAt, &n.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, types.NewNotFoundError("notice not found", map[string]any{"id": id})
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *PostgresNoticeRepository) GetByTenant(ctx context.Context, tenantID types.ID) ([]consent.ConsentNotice, error) {
	query := `SELECT 
		id, tenant_id, series_id, title, content, version, status, purposes, widget_ids, regulation, published_at, created_at, updated_at
	FROM consent_notices WHERE tenant_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notices []consent.ConsentNotice
	for rows.Next() {
		var n consent.ConsentNotice
		if err := rows.Scan(
			&n.ID, &n.TenantID, &n.SeriesID, &n.Title, &n.Content, &n.Version, &n.Status, &n.Purposes, &n.WidgetIDs, &n.Regulation, &n.PublishedAt, &n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		notices = append(notices, n)
	}
	return notices, nil
}

func (r *PostgresNoticeRepository) Update(ctx context.Context, n *consent.ConsentNotice) error {
	query := `UPDATE consent_notices SET
		title = $1, content = $2, purposes = $3, widget_ids = $4, regulation = $5, updated_at = $6
	WHERE id = $7 AND tenant_id = $8`

	cmd, err := r.db.Exec(ctx, query,
		n.Title, n.Content, n.Purposes, n.WidgetIDs, n.Regulation, time.Now(),
		n.ID, n.TenantID,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return types.NewNotFoundError("notice not found", map[string]any{"id": n.ID})
	}
	return nil
}

func (r *PostgresNoticeRepository) Publish(ctx context.Context, id types.ID) (int, error) {
	// Updates status to PUBLISHED and sets published_at
	// We assume version is already set correctly on Create (e.g. v1, v2)
	// But if we need to auto-increment here, we might need more logic.
	// Current requirement: "increments version" -> If it was DRAFT (v0?), make it v1?
	// But schema default is 1.
	// Let's assume we just publish the current version.

	query := `UPDATE consent_notices SET
		status = $1, published_at = $2, updated_at = $2
	WHERE id = $3 RETURNING version`

	var version int
	err := r.db.QueryRow(ctx, query, consent.NoticeStatusPublished, time.Now(), id).Scan(&version)
	if err == pgx.ErrNoRows {
		return 0, types.NewNotFoundError("notice not found", map[string]any{"id": id})
	}
	if err != nil {
		return 0, err
	}
	return version, nil
}

func (r *PostgresNoticeRepository) Archive(ctx context.Context, id types.ID) error {
	query := `UPDATE consent_notices SET
		status = $1, updated_at = $2
	WHERE id = $3`

	cmd, err := r.db.Exec(ctx, query, consent.NoticeStatusArchived, time.Now(), id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return types.NewNotFoundError("notice not found", map[string]any{"id": id})
	}
	return nil
}

func (r *PostgresNoticeRepository) BindToWidgets(ctx context.Context, noticeID types.ID, widgetIDs []types.ID) error {
	query := `UPDATE consent_notices SET
		widget_ids = $1, updated_at = $2
	WHERE id = $3`

	cmd, err := r.db.Exec(ctx, query, widgetIDs, time.Now(), noticeID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return types.NewNotFoundError("notice not found", map[string]any{"id": noticeID})
	}
	return nil
}

// GetLatestVersion returns the latest version number for a given series.
func (r *PostgresNoticeRepository) GetLatestVersion(ctx context.Context, seriesID types.ID) (int, error) {
	query := `SELECT COALESCE(MAX(version), 0) FROM consent_notices WHERE series_id = $1`
	var v int
	err := r.db.QueryRow(ctx, query, seriesID).Scan(&v)
	return v, err
}
