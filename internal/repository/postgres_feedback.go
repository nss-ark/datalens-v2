package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// =============================================================================
// DetectionFeedbackRepo
// =============================================================================

// DetectionFeedbackRepo implements discovery.DetectionFeedbackRepository.
type DetectionFeedbackRepo struct {
	pool *pgxpool.Pool
}

// NewDetectionFeedbackRepo creates a new DetectionFeedbackRepo.
func NewDetectionFeedbackRepo(pool *pgxpool.Pool) *DetectionFeedbackRepo {
	return &DetectionFeedbackRepo{pool: pool}
}

func (r *DetectionFeedbackRepo) Create(ctx context.Context, fb *discovery.DetectionFeedback) error {
	fb.ID = types.NewID()
	query := `
		INSERT INTO detection_feedback (
			id, classification_id, tenant_id, feedback_type,
			original_category, original_type, original_confidence, original_method,
			corrected_category, corrected_type,
			corrected_by, corrected_at, notes,
			column_name, table_name, data_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		fb.ID, fb.ClassificationID, fb.TenantID, fb.FeedbackType,
		fb.OriginalCategory, fb.OriginalType, fb.OriginalConfidence, fb.OriginalMethod,
		fb.CorrectedCategory, fb.CorrectedType,
		fb.CorrectedBy, fb.CorrectedAt, fb.Notes,
		fb.ColumnName, fb.TableName, fb.DataType,
	).Scan(&fb.CreatedAt, &fb.UpdatedAt)
}

func (r *DetectionFeedbackRepo) GetByID(ctx context.Context, id types.ID) (*discovery.DetectionFeedback, error) {
	query := `
		SELECT id, classification_id, tenant_id, feedback_type,
		       original_category, original_type, original_confidence, original_method,
		       corrected_category, corrected_type,
		       corrected_by, corrected_at, notes,
		       column_name, table_name, data_type,
		       created_at, updated_at
		FROM detection_feedback
		WHERE id = $1`

	fb := &discovery.DetectionFeedback{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&fb.ID, &fb.ClassificationID, &fb.TenantID, &fb.FeedbackType,
		&fb.OriginalCategory, &fb.OriginalType, &fb.OriginalConfidence, &fb.OriginalMethod,
		&fb.CorrectedCategory, &fb.CorrectedType,
		&fb.CorrectedBy, &fb.CorrectedAt, &fb.Notes,
		&fb.ColumnName, &fb.TableName, &fb.DataType,
		&fb.CreatedAt, &fb.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("DetectionFeedback", id)
		}
		return nil, fmt.Errorf("get detection feedback: %w", err)
	}
	return fb, nil
}

func (r *DetectionFeedbackRepo) GetByClassification(ctx context.Context, classificationID types.ID) ([]discovery.DetectionFeedback, error) {
	query := `
		SELECT id, classification_id, tenant_id, feedback_type,
		       original_category, original_type, original_confidence, original_method,
		       corrected_category, corrected_type,
		       corrected_by, corrected_at, notes,
		       column_name, table_name, data_type,
		       created_at, updated_at
		FROM detection_feedback
		WHERE classification_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, classificationID)
	if err != nil {
		return nil, fmt.Errorf("list feedback by classification: %w", err)
	}
	defer rows.Close()

	var results []discovery.DetectionFeedback
	for rows.Next() {
		var fb discovery.DetectionFeedback
		if err := rows.Scan(
			&fb.ID, &fb.ClassificationID, &fb.TenantID, &fb.FeedbackType,
			&fb.OriginalCategory, &fb.OriginalType, &fb.OriginalConfidence, &fb.OriginalMethod,
			&fb.CorrectedCategory, &fb.CorrectedType,
			&fb.CorrectedBy, &fb.CorrectedAt, &fb.Notes,
			&fb.ColumnName, &fb.TableName, &fb.DataType,
			&fb.CreatedAt, &fb.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan detection feedback: %w", err)
		}
		results = append(results, fb)
	}
	return results, rows.Err()
}

func (r *DetectionFeedbackRepo) GetByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[discovery.DetectionFeedback], error) {
	countQuery := `SELECT COUNT(*) FROM detection_feedback WHERE tenant_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count detection feedback: %w", err)
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query := `
		SELECT id, classification_id, tenant_id, feedback_type,
		       original_category, original_type, original_confidence, original_method,
		       corrected_category, corrected_type,
		       corrected_by, corrected_at, notes,
		       column_name, table_name, data_type,
		       created_at, updated_at
		FROM detection_feedback
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, tenantID, pagination.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list detection feedback: %w", err)
	}
	defer rows.Close()

	var items []discovery.DetectionFeedback
	for rows.Next() {
		var fb discovery.DetectionFeedback
		if err := rows.Scan(
			&fb.ID, &fb.ClassificationID, &fb.TenantID, &fb.FeedbackType,
			&fb.OriginalCategory, &fb.OriginalType, &fb.OriginalConfidence, &fb.OriginalMethod,
			&fb.CorrectedCategory, &fb.CorrectedType,
			&fb.CorrectedBy, &fb.CorrectedAt, &fb.Notes,
			&fb.ColumnName, &fb.TableName, &fb.DataType,
			&fb.CreatedAt, &fb.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan detection feedback: %w", err)
		}
		items = append(items, fb)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := total / pagination.PageSize
	if total%pagination.PageSize > 0 {
		totalPages++
	}

	return &types.PaginatedResult[discovery.DetectionFeedback]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *DetectionFeedbackRepo) GetCorrectionPatterns(ctx context.Context, tenantID types.ID, columnPattern string) ([]discovery.DetectionFeedback, error) {
	query := `
		SELECT id, classification_id, tenant_id, feedback_type,
		       original_category, original_type, original_confidence, original_method,
		       corrected_category, corrected_type,
		       corrected_by, corrected_at, notes,
		       column_name, table_name, data_type,
		       created_at, updated_at
		FROM detection_feedback
		WHERE tenant_id = $1
		  AND feedback_type = 'CORRECTED'
		  AND column_name ILIKE $2
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID, "%"+columnPattern+"%")
	if err != nil {
		return nil, fmt.Errorf("get correction patterns: %w", err)
	}
	defer rows.Close()

	var results []discovery.DetectionFeedback
	for rows.Next() {
		var fb discovery.DetectionFeedback
		if err := rows.Scan(
			&fb.ID, &fb.ClassificationID, &fb.TenantID, &fb.FeedbackType,
			&fb.OriginalCategory, &fb.OriginalType, &fb.OriginalConfidence, &fb.OriginalMethod,
			&fb.CorrectedCategory, &fb.CorrectedType,
			&fb.CorrectedBy, &fb.CorrectedAt, &fb.Notes,
			&fb.ColumnName, &fb.TableName, &fb.DataType,
			&fb.CreatedAt, &fb.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan correction pattern: %w", err)
		}
		results = append(results, fb)
	}
	return results, rows.Err()
}

func (r *DetectionFeedbackRepo) GetAccuracyStats(ctx context.Context, tenantID types.ID, method types.DetectionMethod) (*discovery.AccuracyStats, error) {
	query := `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE feedback_type = 'VERIFIED') AS verified,
			COUNT(*) FILTER (WHERE feedback_type = 'CORRECTED') AS corrected,
			COUNT(*) FILTER (WHERE feedback_type = 'REJECTED') AS rejected
		FROM detection_feedback
		WHERE tenant_id = $1 AND original_method = $2`

	stats := &discovery.AccuracyStats{Method: method}
	err := r.pool.QueryRow(ctx, query, tenantID, method).Scan(
		&stats.Total, &stats.Verified, &stats.Corrected, &stats.Rejected,
	)
	if err != nil {
		return nil, fmt.Errorf("get accuracy stats: %w", err)
	}

	if stats.Total > 0 {
		stats.Accuracy = float64(stats.Verified) / float64(stats.Total)
	}

	return stats, nil
}

// Compile-time check.
var _ discovery.DetectionFeedbackRepository = (*DetectionFeedbackRepo)(nil)
