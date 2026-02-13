package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresGrievanceRepository implements compliance.GrievanceRepository.
type PostgresGrievanceRepository struct {
	db *pgxpool.Pool
}

// NewPostgresGrievanceRepository creates a new PostgresGrievanceRepository.
func NewPostgresGrievanceRepository(db *pgxpool.Pool) *PostgresGrievanceRepository {
	return &PostgresGrievanceRepository{db: db}
}

// Create persists a new grievance.
func (r *PostgresGrievanceRepository) Create(ctx context.Context, g *compliance.Grievance) error {
	query := `
		INSERT INTO grievances (
			id, tenant_id, data_subject_id, subject, description, category, status, priority,
			assigned_to, resolution, submitted_at, due_date, resolved_at, escalated_to,
			feedback_rating, feedback_comment, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18
		)
	`
	_, err := r.db.Exec(ctx, query,
		g.ID, g.TenantID, g.DataSubjectID, g.Subject, g.Description, g.Category, g.Status, g.Priority,
		g.AssignedTo, g.Resolution, g.SubmittedAt, g.DueDate, g.ResolvedAt, g.EscalatedTo,
		g.FeedbackRating, g.FeedbackComment, g.CreatedAt, g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create grievance: %w", err)
	}
	return nil
}

// GetByID retrieves a grievance by ID.
func (r *PostgresGrievanceRepository) GetByID(ctx context.Context, id types.ID) (*compliance.Grievance, error) {
	query := `
		SELECT
			id, tenant_id, data_subject_id, subject, description, category, status, priority,
			assigned_to, resolution, submitted_at, due_date, resolved_at, escalated_to,
			feedback_rating, feedback_comment, created_at, updated_at
		FROM grievances
		WHERE id = $1
	`
	var g compliance.Grievance
	err := r.db.QueryRow(ctx, query, id).Scan(
		&g.ID, &g.TenantID, &g.DataSubjectID, &g.Subject, &g.Description, &g.Category, &g.Status, &g.Priority,
		&g.AssignedTo, &g.Resolution, &g.SubmittedAt, &g.DueDate, &g.ResolvedAt, &g.EscalatedTo,
		&g.FeedbackRating, &g.FeedbackComment, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("grievance", map[string]any{"id": id})
		}
		return nil, fmt.Errorf("get grievance by id: %w", err)
	}
	return &g, nil
}

// ListByTenant retrieves grievances for a tenant with filters and pagination.
func (r *PostgresGrievanceRepository) ListByTenant(ctx context.Context, tenantID types.ID, filters map[string]any, pagination types.Pagination) (*types.PaginatedResult[compliance.Grievance], error) {
	var conditions []string
	args := []any{tenantID}

	// Always filter by tenant
	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", len(args)))

	// Apply filters
	if status, ok := filters["status"].(string); ok && status != "" {
		args = append(args, status)
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	if priority, ok := filters["priority"].(int); ok {
		args = append(args, priority)
		conditions = append(conditions, fmt.Sprintf("priority = $%d", len(args)))
	}
	if assignedTo, ok := filters["assigned_to"].(string); ok && assignedTo != "" {
		id, err := types.ParseID(assignedTo)
		if err == nil {
			args = append(args, id)
			conditions = append(conditions, fmt.Sprintf("assigned_to = $%d", len(args)))
		}
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count query
	countQuery := "SELECT COUNT(*) FROM grievances " + whereClause
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count grievances: %w", err)
	}

	// Data query
	offset := (pagination.Page - 1) * pagination.PageSize
	args = append(args, pagination.PageSize, offset)
	query := fmt.Sprintf(`
		SELECT
			id, tenant_id, data_subject_id, subject, description, category, status, priority,
			assigned_to, resolution, submitted_at, due_date, resolved_at, escalated_to,
			feedback_rating, feedback_comment, created_at, updated_at
		FROM grievances
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)-1, len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list grievances: %w", err)
	}
	defer rows.Close()

	var grievances []compliance.Grievance
	for rows.Next() {
		var g compliance.Grievance
		if err := rows.Scan(
			&g.ID, &g.TenantID, &g.DataSubjectID, &g.Subject, &g.Description, &g.Category, &g.Status, &g.Priority,
			&g.AssignedTo, &g.Resolution, &g.SubmittedAt, &g.DueDate, &g.ResolvedAt, &g.EscalatedTo,
			&g.FeedbackRating, &g.FeedbackComment, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan grievance: %w", err)
		}
		grievances = append(grievances, g)
	}

	return &types.PaginatedResult[compliance.Grievance]{
		Items:    grievances,
		Total:    total,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}, nil
}

// ListBySubject retrieves all grievances for a specific subject ID (portal view).
func (r *PostgresGrievanceRepository) ListBySubject(ctx context.Context, tenantID, subjectID types.ID) ([]compliance.Grievance, error) {
	query := `
		SELECT
			id, tenant_id, data_subject_id, subject, description, category, status, priority,
			assigned_to, resolution, submitted_at, due_date, resolved_at, escalated_to,
			feedback_rating, feedback_comment, created_at, updated_at
		FROM grievances
		WHERE tenant_id = $1 AND data_subject_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, subjectID)
	if err != nil {
		return nil, fmt.Errorf("list grievances by subject: %w", err)
	}
	defer rows.Close()

	var grievances []compliance.Grievance
	for rows.Next() {
		var g compliance.Grievance
		if err := rows.Scan(
			&g.ID, &g.TenantID, &g.DataSubjectID, &g.Subject, &g.Description, &g.Category, &g.Status, &g.Priority,
			&g.AssignedTo, &g.Resolution, &g.SubmittedAt, &g.DueDate, &g.ResolvedAt, &g.EscalatedTo,
			&g.FeedbackRating, &g.FeedbackComment, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan grievance: %w", err)
		}
		grievances = append(grievances, g)
	}
	return grievances, nil
}

// Update persists changes to an existing grievance.
func (r *PostgresGrievanceRepository) Update(ctx context.Context, g *compliance.Grievance) error {
	query := `
		UPDATE grievances SET
			status = $1,
			assigned_to = $2,
			resolution = $3,
			resolved_at = $4,
			escalated_to = $5,
			feedback_rating = $6,
			feedback_comment = $7,
			updated_at = $8
		WHERE id = $9 AND tenant_id = $10
	`
	_, err := r.db.Exec(ctx, query,
		g.Status, g.AssignedTo, g.Resolution, g.ResolvedAt, g.EscalatedTo,
		g.FeedbackRating, g.FeedbackComment, g.UpdatedAt,
		g.ID, g.TenantID,
	)
	if err != nil {
		return fmt.Errorf("update grievance: %w", err)
	}
	return nil
}
