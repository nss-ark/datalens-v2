package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =============================================================================
// PostgresNotificationRepository
// =============================================================================

type PostgresNotificationRepository struct {
	db *pgxpool.Pool
}

func NewPostgresNotificationRepository(db *pgxpool.Pool) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

func (r *PostgresNotificationRepository) Create(ctx context.Context, n *consent.ConsentNotification) error {
	query := `INSERT INTO consent_notifications (
		id, tenant_id, recipient_type, recipient_id, event_type, channel, template_id, payload, status, sent_at, error_message, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(ctx, query,
		n.ID, n.TenantID, n.RecipientType, n.RecipientID, n.EventType, n.Channel, n.TemplateID, n.Payload, n.Status, n.SentAt, n.ErrorMessage, n.CreatedAt, n.UpdatedAt,
	)
	return err
}

func (r *PostgresNotificationRepository) GetByID(ctx context.Context, id types.ID) (*consent.ConsentNotification, error) {
	query := `SELECT 
		id, tenant_id, recipient_type, recipient_id, event_type, channel, template_id, payload, status, sent_at, error_message, created_at, updated_at
	FROM consent_notifications WHERE id = $1`

	var n consent.ConsentNotification
	err := r.db.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.TenantID, &n.RecipientType, &n.RecipientID, &n.EventType, &n.Channel, &n.TemplateID, &n.Payload, &n.Status, &n.SentAt, &n.ErrorMessage, &n.CreatedAt, &n.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, types.NewNotFoundError("notification not found", map[string]any{"id": id})
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *PostgresNotificationRepository) ListByTenant(ctx context.Context, tenantID types.ID, filter consent.NotificationFilter, pagination types.Pagination) (*types.PaginatedResult[consent.ConsentNotification], error) {
	baseQuery := `SELECT 
		id, tenant_id, recipient_type, recipient_id, event_type, channel, template_id, payload, status, sent_at, error_message, created_at, updated_at
	FROM consent_notifications WHERE tenant_id = $1`
	args := []any{tenantID}
	argCount := 1

	if filter.RecipientID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND recipient_id = $%d", argCount)
		args = append(args, *filter.RecipientID)
	}
	if filter.EventType != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND event_type = $%d", argCount)
		args = append(args, *filter.EventType)
	}
	if filter.Channel != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND channel = $%d", argCount)
		args = append(args, *filter.Channel)
	}
	if filter.Status != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filter.Status)
	}

	// Count total
	countQuery := "SELECT COUNT(*) " + baseQuery[len("SELECT id, tenant_id, recipient_type, recipient_id, event_type, channel, template_id, payload, status, sent_at, error_message, created_at, updated_at "):]
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Add pagination
	baseQuery += " ORDER BY created_at DESC"
	argCount++
	baseQuery += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, pagination.PageSize)
	argCount++
	baseQuery += fmt.Sprintf(" OFFSET $%d", argCount)
	offset := (pagination.Page - 1) * pagination.PageSize
	args = append(args, offset)

	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []consent.ConsentNotification
	for rows.Next() {
		var n consent.ConsentNotification
		if err := rows.Scan(
			&n.ID, &n.TenantID, &n.RecipientType, &n.RecipientID, &n.EventType, &n.Channel, &n.TemplateID, &n.Payload, &n.Status, &n.SentAt, &n.ErrorMessage, &n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return &types.PaginatedResult[consent.ConsentNotification]{
		Items:    notifications,
		Total:    int(total),
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}, nil
}

func (r *PostgresNotificationRepository) UpdateStatus(ctx context.Context, id types.ID, status string, sentAt *time.Time, errorMessage *string) error {
	query := `UPDATE consent_notifications SET
		status = $1, sent_at = $2, error_message = $3, updated_at = $4
	WHERE id = $5`

	_, err := r.db.Exec(ctx, query, status, sentAt, errorMessage, time.Now().UTC(), id)
	return err
}

// =============================================================================
// PostgresNotificationTemplateRepository
// =============================================================================

type PostgresNotificationTemplateRepository struct {
	db *pgxpool.Pool
}

func NewPostgresNotificationTemplateRepository(db *pgxpool.Pool) *PostgresNotificationTemplateRepository {
	return &PostgresNotificationTemplateRepository{db: db}
}

func (r *PostgresNotificationTemplateRepository) Create(ctx context.Context, t *consent.NotificationTemplate) error {
	query := `INSERT INTO notification_templates (
		id, tenant_id, name, event_type, channel, subject, body_template, is_active, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(ctx, query,
		t.ID, t.TenantID, t.Name, t.EventType, t.Channel, t.Subject, t.BodyTemplate, t.IsActive, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (r *PostgresNotificationTemplateRepository) Update(ctx context.Context, t *consent.NotificationTemplate) error {
	query := `UPDATE notification_templates SET
		name = $1, subject = $2, body_template = $3, is_active = $4, updated_at = $5
	WHERE id = $6 AND tenant_id = $7`

	cmd, err := r.db.Exec(ctx, query,
		t.Name, t.Subject, t.BodyTemplate, t.IsActive, time.Now().UTC(), t.ID, t.TenantID,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return types.NewNotFoundError("template not found", map[string]any{"id": t.ID})
	}
	return nil
}

func (r *PostgresNotificationTemplateRepository) GetByID(ctx context.Context, id types.ID) (*consent.NotificationTemplate, error) {
	query := `SELECT 
		id, tenant_id, name, event_type, channel, subject, body_template, is_active, created_at, updated_at
	FROM notification_templates WHERE id = $1`

	var t consent.NotificationTemplate
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TenantID, &t.Name, &t.EventType, &t.Channel, &t.Subject, &t.BodyTemplate, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, types.NewNotFoundError("template not found", map[string]any{"id": id})
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresNotificationTemplateRepository) ListByTenant(ctx context.Context, tenantID types.ID) ([]consent.NotificationTemplate, error) {
	query := `SELECT 
		id, tenant_id, name, event_type, channel, subject, body_template, is_active, created_at, updated_at
	FROM notification_templates WHERE tenant_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []consent.NotificationTemplate
	for rows.Next() {
		var t consent.NotificationTemplate
		if err := rows.Scan(
			&t.ID, &t.TenantID, &t.Name, &t.EventType, &t.Channel, &t.Subject, &t.BodyTemplate, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *PostgresNotificationTemplateRepository) GetByEventAndChannel(ctx context.Context, tenantID types.ID, eventType, channel string) (*consent.NotificationTemplate, error) {
	query := `SELECT 
		id, tenant_id, name, event_type, channel, subject, body_template, is_active, created_at, updated_at
	FROM notification_templates 
	WHERE tenant_id = $1 AND event_type = $2 AND channel = $3 AND is_active = true
	LIMIT 1`

	var t consent.NotificationTemplate
	err := r.db.QueryRow(ctx, query, tenantID, eventType, channel).Scan(
		&t.ID, &t.TenantID, &t.Name, &t.EventType, &t.Channel, &t.Subject, &t.BodyTemplate, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, types.NewNotFoundError("active template not found", map[string]any{"event_type": eventType, "channel": channel})
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresNotificationTemplateRepository) Delete(ctx context.Context, id types.ID) error {
	query := `DELETE FROM notification_templates WHERE id = $1`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return types.NewNotFoundError("template not found", map[string]any{"id": id})
	}
	return nil
}
