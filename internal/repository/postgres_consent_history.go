package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
)

// ConsentHistoryRepo implements consent.ConsentHistoryRepository.
type ConsentHistoryRepo struct {
	pool *pgxpool.Pool
}

// NewConsentHistoryRepo creates a new ConsentHistoryRepo.
func NewConsentHistoryRepo(pool *pgxpool.Pool) *ConsentHistoryRepo {
	return &ConsentHistoryRepo{pool: pool}
}

// Create persists a new consent history entry (append-only — no updates or deletes).
func (r *ConsentHistoryRepo) Create(ctx context.Context, entry *consent.ConsentHistoryEntry) error {
	query := `
		INSERT INTO consent_history (
			id, tenant_id, subject_id, widget_id, purpose_id, purpose_name,
			previous_status, new_status, source, ip_address,
			user_agent, notice_version, signature, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at`

	return r.pool.QueryRow(ctx, query,
		entry.ID, entry.TenantID, entry.SubjectID, entry.WidgetID,
		entry.PurposeID, entry.PurposeName,
		entry.PreviousStatus, entry.NewStatus, entry.Source, entry.IPAddress,
		entry.UserAgent, entry.NoticeVersion, entry.Signature, entry.CreatedAt,
	).Scan(&entry.CreatedAt)
}

// GetBySubject retrieves paginated consent history for a subject within a tenant.
func (r *ConsentHistoryRepo) GetBySubject(ctx context.Context, tenantID, subjectID types.ID, pagination types.Pagination) (*types.PaginatedResult[consent.ConsentHistoryEntry], error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM consent_history WHERE tenant_id = $1 AND subject_id = $2`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID, subjectID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count consent history: %w", err)
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query := `
		SELECT id, tenant_id, subject_id, widget_id, purpose_id, purpose_name,
		       previous_status, new_status, source, ip_address,
		       user_agent, notice_version, signature, created_at
		FROM consent_history
		WHERE tenant_id = $1 AND subject_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, tenantID, subjectID, pagination.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list consent history: %w", err)
	}
	defer rows.Close()

	var items []consent.ConsentHistoryEntry
	for rows.Next() {
		var e consent.ConsentHistoryEntry
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.SubjectID, &e.WidgetID,
			&e.PurposeID, &e.PurposeName,
			&e.PreviousStatus, &e.NewStatus, &e.Source, &e.IPAddress,
			&e.UserAgent, &e.NoticeVersion, &e.Signature, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan consent history: %w", err)
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := total / pagination.PageSize
	if total%pagination.PageSize > 0 {
		totalPages++
	}

	return &types.PaginatedResult[consent.ConsentHistoryEntry]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByPurpose retrieves all consent history entries for a purpose within a tenant.
func (r *ConsentHistoryRepo) GetByPurpose(ctx context.Context, tenantID, purposeID types.ID) ([]consent.ConsentHistoryEntry, error) {
	query := `
		SELECT id, tenant_id, subject_id, widget_id, purpose_id, purpose_name,
		       previous_status, new_status, source, ip_address,
		       user_agent, notice_version, signature, created_at
		FROM consent_history
		WHERE tenant_id = $1 AND purpose_id = $2
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID, purposeID)
	if err != nil {
		return nil, fmt.Errorf("list consent history by purpose: %w", err)
	}
	defer rows.Close()

	var items []consent.ConsentHistoryEntry
	for rows.Next() {
		var e consent.ConsentHistoryEntry
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.SubjectID, &e.WidgetID,
			&e.PurposeID, &e.PurposeName,
			&e.PreviousStatus, &e.NewStatus, &e.Source, &e.IPAddress,
			&e.UserAgent, &e.NoticeVersion, &e.Signature, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan consent history: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// GetLatestState retrieves the most recent consent history entry for a subject and purpose.
func (r *ConsentHistoryRepo) GetLatestState(ctx context.Context, tenantID, subjectID, purposeID types.ID) (*consent.ConsentHistoryEntry, error) {
	query := `
		SELECT id, tenant_id, subject_id, widget_id, purpose_id, purpose_name,
		       previous_status, new_status, source, ip_address,
		       user_agent, notice_version, signature, created_at
		FROM consent_history
		WHERE tenant_id = $1 AND subject_id = $2 AND purpose_id = $3
		ORDER BY created_at DESC
		LIMIT 1`

	var e consent.ConsentHistoryEntry
	err := r.pool.QueryRow(ctx, query, tenantID, subjectID, purposeID).Scan(
		&e.ID, &e.TenantID, &e.SubjectID, &e.WidgetID,
		&e.PurposeID, &e.PurposeName,
		&e.PreviousStatus, &e.NewStatus, &e.Source, &e.IPAddress,
		&e.UserAgent, &e.NoticeVersion, &e.Signature, &e.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found = no consent state
		}
		return nil, fmt.Errorf("get latest consent state: %w", err)
	}

	return &e, nil
}

// Compile-time interface check.
var _ consent.ConsentHistoryRepository = (*ConsentHistoryRepo)(nil)
