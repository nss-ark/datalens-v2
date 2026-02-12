package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
)

// ConsentRenewalRepo implements consent.ConsentRenewalRepository.
type ConsentRenewalRepo struct {
	pool *pgxpool.Pool
}

// NewConsentRenewalRepo creates a new ConsentRenewalRepo.
func NewConsentRenewalRepo(pool *pgxpool.Pool) *ConsentRenewalRepo {
	return &ConsentRenewalRepo{pool: pool}
}

// Create persists a new consent renewal log.
func (r *ConsentRenewalRepo) Create(ctx context.Context, l *consent.ConsentRenewalLog) error {
	query := `
		INSERT INTO consent_renewal_logs (
			id, tenant_id, subject_id, purpose_id, original_expiry,
			status, reminder_sent_at, renewed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		l.ID, l.TenantID, l.SubjectID, l.PurposeID, l.OriginalExpiry,
		l.Status, l.ReminderSentAt, l.RenewedAt, l.CreatedAt, l.UpdatedAt,
	).Scan(&l.CreatedAt, &l.UpdatedAt)
}

// GetBySubject retrieves renewal logs for a subject.
func (r *ConsentRenewalRepo) GetBySubject(ctx context.Context, tenantID, subjectID types.ID) ([]consent.ConsentRenewalLog, error) {
	query := `
		SELECT id, tenant_id, subject_id, purpose_id, original_expiry,
		       status, reminder_sent_at, renewed_at, created_at, updated_at
		FROM consent_renewal_logs
		WHERE tenant_id = $1 AND subject_id = $2
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID, subjectID)
	if err != nil {
		return nil, fmt.Errorf("list renewal logs: %w", err)
	}
	defer rows.Close()

	var logs []consent.ConsentRenewalLog
	for rows.Next() {
		var l consent.ConsentRenewalLog
		if err := rows.Scan(
			&l.ID, &l.TenantID, &l.SubjectID, &l.PurposeID, &l.OriginalExpiry,
			&l.Status, &l.ReminderSentAt, &l.RenewedAt, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan renewal log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// Update persists changes to an existing consent renewal log.
func (r *ConsentRenewalRepo) Update(ctx context.Context, l *consent.ConsentRenewalLog) error {
	query := `
		UPDATE consent_renewal_logs
		SET status = $1, reminder_sent_at = $2, renewed_at = $3, updated_at = NOW()
		WHERE id = $4 AND tenant_id = $5
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		l.Status, l.ReminderSentAt, l.RenewedAt,
		l.ID, l.TenantID,
	).Scan(&l.UpdatedAt)
}

// GetPending retrieves renewal logs that are pending.
func (r *ConsentRenewalRepo) GetPending(ctx context.Context, tenantID types.ID) ([]consent.ConsentRenewalLog, error) {
	query := `
		SELECT id, tenant_id, subject_id, purpose_id, original_expiry,
		       status, reminder_sent_at, renewed_at, created_at, updated_at
		FROM consent_renewal_logs
		WHERE tenant_id = $1 AND status = 'PENDING'
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list pending renewal logs: %w", err)
	}
	defer rows.Close()

	var logs []consent.ConsentRenewalLog
	for rows.Next() {
		var l consent.ConsentRenewalLog
		if err := rows.Scan(
			&l.ID, &l.TenantID, &l.SubjectID, &l.PurposeID, &l.OriginalExpiry,
			&l.Status, &l.ReminderSentAt, &l.RenewedAt, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan renewal log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// Compile-time interface check.
var _ consent.ConsentRenewalRepository = (*ConsentRenewalRepo)(nil)
