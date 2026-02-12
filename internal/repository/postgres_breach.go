package repository

import (
	"context"
	"fmt"

	"github.com/complyark/datalens/internal/domain/breach"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresBreachRepository struct {
	db *pgxpool.Pool
}

func NewPostgresBreachRepository(db *pgxpool.Pool) *PostgresBreachRepository {
	return &PostgresBreachRepository{db: db}
}

func (r *PostgresBreachRepository) Create(ctx context.Context, b *breach.BreachIncident) error {
	query := `
		INSERT INTO breach_incidents (
			id, tenant_id, title, description, type, severity, status,
			detected_at, occurred_at, reported_to_cert_in_at, reported_to_dpb_at, closed_at,
			affected_systems, affected_data_subject_count, pii_categories,
			is_reportable_cert_in, is_reportable_dpb,
			poc_name, poc_role, poc_email,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15,
			$16, $17,
			$18, $19, $20,
			$21, $22
		)
	`
	_, err := r.db.Exec(ctx, query,
		b.ID, b.TenantID, b.Title, b.Description, b.Type, b.Severity, b.Status,
		b.DetectedAt, b.OccurredAt, b.ReportedToCertInAt, b.ReportedToDPBAt, b.ClosedAt,
		b.AffectedSystems, b.AffectedDataSubjectCount, b.PiiCategories,
		b.IsReportableToCertIn, b.IsReportableToDPB,
		b.PoCName, b.PoCRole, b.PoCEmail,
		b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create breach incident: %w", err)
	}
	return nil
}

func (r *PostgresBreachRepository) GetByID(ctx context.Context, id types.ID) (*breach.BreachIncident, error) {
	query := `
		SELECT 
			id, tenant_id, title, description, type, severity, status,
			detected_at, occurred_at, reported_to_cert_in_at, reported_to_dpb_at, closed_at,
			affected_systems, affected_data_subject_count, pii_categories,
			is_reportable_cert_in, is_reportable_dpb,
			poc_name, poc_role, poc_email,
			created_at, updated_at
		FROM breach_incidents
		WHERE id = $1
	`
	var b breach.BreachIncident
	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.TenantID, &b.Title, &b.Description, &b.Type, &b.Severity, &b.Status,
		&b.DetectedAt, &b.OccurredAt, &b.ReportedToCertInAt, &b.ReportedToDPBAt, &b.ClosedAt,
		&b.AffectedSystems, &b.AffectedDataSubjectCount, &b.PiiCategories,
		&b.IsReportableToCertIn, &b.IsReportableToDPB,
		&b.PoCName, &b.PoCRole, &b.PoCEmail,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, types.NewNotFoundError("breach incident", map[string]any{"id": id})
	}
	return &b, nil
}

func (r *PostgresBreachRepository) Update(ctx context.Context, b *breach.BreachIncident) error {
	query := `
		UPDATE breach_incidents SET
			title = $1, description = $2, type = $3, severity = $4, status = $5,
			detected_at = $6, occurred_at = $7, reported_to_cert_in_at = $8, reported_to_dpb_at = $9, closed_at = $10,
			affected_systems = $11, affected_data_subject_count = $12, pii_categories = $13,
			is_reportable_cert_in = $14, is_reportable_dpb = $15,
			poc_name = $16, poc_role = $17, poc_email = $18,
			updated_at = $19
		WHERE id = $20
	`
	_, err := r.db.Exec(ctx, query,
		b.Title, b.Description, b.Type, b.Severity, b.Status,
		b.DetectedAt, b.OccurredAt, b.ReportedToCertInAt, b.ReportedToDPBAt, b.ClosedAt,
		b.AffectedSystems, b.AffectedDataSubjectCount, b.PiiCategories,
		b.IsReportableToCertIn, b.IsReportableToDPB,
		b.PoCName, b.PoCRole, b.PoCEmail,
		b.UpdatedAt, b.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update breach incident: %w", err)
	}
	return nil
}

func (r *PostgresBreachRepository) List(ctx context.Context, tenantID types.ID, filter breach.Filter, pagination types.Pagination) (*types.PaginatedResult[breach.BreachIncident], error) {
	query := `
		SELECT 
			id, tenant_id, title, description, type, severity, status,
			detected_at, occurred_at, reported_to_cert_in_at, reported_to_dpb_at, closed_at,
			affected_systems, affected_data_subject_count, pii_categories,
			is_reportable_cert_in, is_reportable_dpb,
			poc_name, poc_role, poc_email,
			created_at, updated_at
		FROM breach_incidents
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIdx := 2 // $1 is tenantID

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}

	if filter.Severity != nil {
		query += fmt.Sprintf(" AND severity = $%d", argIdx)
		args = append(args, *filter.Severity)
		argIdx++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS count_q"
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count breach incidents: %w", err)
	}

	// Pagination
	query += " ORDER BY created_at DESC"
	offset := (pagination.Page - 1) * pagination.PageSize
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, pagination.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list breach incidents: %w", err)
	}
	defer rows.Close()

	var incidents []breach.BreachIncident
	for rows.Next() {
		var b breach.BreachIncident
		err := rows.Scan(
			&b.ID, &b.TenantID, &b.Title, &b.Description, &b.Type, &b.Severity, &b.Status,
			&b.DetectedAt, &b.OccurredAt, &b.ReportedToCertInAt, &b.ReportedToDPBAt, &b.ClosedAt,
			&b.AffectedSystems, &b.AffectedDataSubjectCount, &b.PiiCategories,
			&b.IsReportableToCertIn, &b.IsReportableToDPB,
			&b.PoCName, &b.PoCRole, &b.PoCEmail,
			&b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan breach incident: %w", err)
		}
		incidents = append(incidents, b)
	}

	return &types.PaginatedResult[breach.BreachIncident]{
		Items:      incidents,
		Total:      int(total),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	}, nil
}
