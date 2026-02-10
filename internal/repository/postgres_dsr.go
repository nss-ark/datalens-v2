package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/pkg/types"
)

// DSRRepo implements compliance.DSRRepository.
type DSRRepo struct {
	pool *pgxpool.Pool
}

// NewDSRRepo creates a new DSRRepo.
func NewDSRRepo(pool *pgxpool.Pool) *DSRRepo {
	return &DSRRepo{pool: pool}
}

// Create persists a new DSR.
func (r *DSRRepo) Create(ctx context.Context, dsr *compliance.DSR) error {
	query := `
		INSERT INTO dsr_requests (
			id, tenant_id, request_type, status,
			subject_name, subject_email, subject_identifiers,
			priority, sla_deadline, assigned_to, reason,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		dsr.ID, dsr.TenantID, dsr.RequestType, dsr.Status,
		dsr.SubjectName, dsr.SubjectEmail, dsr.SubjectIdentifiers,
		dsr.Priority, dsr.SLADeadline, dsr.AssignedTo, dsr.Reason,
		dsr.CreatedAt, dsr.UpdatedAt,
	).Scan(&dsr.CreatedAt, &dsr.UpdatedAt)
}

// GetByID retrieves a DSR by ID.
func (r *DSRRepo) GetByID(ctx context.Context, id types.ID) (*compliance.DSR, error) {
	query := `
		SELECT id, tenant_id, request_type, status,
		       subject_name, subject_email, subject_identifiers,
		       priority, sla_deadline, assigned_to, reason,
		       created_at, updated_at, completed_at
		FROM dsr_requests
		WHERE id = $1`

	dsr := &compliance.DSR{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&dsr.ID, &dsr.TenantID, &dsr.RequestType, &dsr.Status,
		&dsr.SubjectName, &dsr.SubjectEmail, &dsr.SubjectIdentifiers,
		&dsr.Priority, &dsr.SLADeadline, &dsr.AssignedTo, &dsr.Reason,
		&dsr.CreatedAt, &dsr.UpdatedAt, &dsr.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("DSR", id)
		}
		return nil, fmt.Errorf("get dsr: %w", err)
	}
	return dsr, nil
}

// GetByTenant lists DSRs for a tenant with pagination and optional status filtering.
func (r *DSRRepo) GetByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination, statusFilter *compliance.DSRStatus) (*types.PaginatedResult[compliance.DSR], error) {
	baseQuery := `FROM dsr_requests WHERE tenant_id = $1`
	args := []any{tenantID}
	argIdx := 2

	if statusFilter != nil {
		baseQuery += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, *statusFilter)
		argIdx++
	}

	// Count total
	countQuery := `SELECT COUNT(*) ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count dsr: %w", err)
	}

	// Fetch items
	offset := (pagination.Page - 1) * pagination.PageSize
	query := fmt.Sprintf(`
		SELECT id, tenant_id, request_type, status,
		       subject_name, subject_email, subject_identifiers,
		       priority, sla_deadline, assigned_to, reason,
		       created_at, updated_at, completed_at
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, baseQuery, argIdx, argIdx+1)

	args = append(args, pagination.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list dsr: %w", err)
	}
	defer rows.Close()

	var items []compliance.DSR
	for rows.Next() {
		var dsr compliance.DSR
		if err := rows.Scan(
			&dsr.ID, &dsr.TenantID, &dsr.RequestType, &dsr.Status,
			&dsr.SubjectName, &dsr.SubjectEmail, &dsr.SubjectIdentifiers,
			&dsr.Priority, &dsr.SLADeadline, &dsr.AssignedTo, &dsr.Reason,
			&dsr.CreatedAt, &dsr.UpdatedAt, &dsr.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dsr: %w", err)
		}
		items = append(items, dsr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := total / pagination.PageSize
	if total%pagination.PageSize > 0 {
		totalPages++
	}

	return &types.PaginatedResult[compliance.DSR]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Update updates an existing DSR.
func (r *DSRRepo) Update(ctx context.Context, dsr *compliance.DSR) error {
	query := `
		UPDATE dsr_requests
		SET status = $1, assigned_to = $2, reason = $3, 
		    updated_at = NOW(), completed_at = $4
		WHERE id = $5 AND tenant_id = $6
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		dsr.Status, dsr.AssignedTo, dsr.Reason, dsr.CompletedAt,
		dsr.ID, dsr.TenantID,
	).Scan(&dsr.UpdatedAt)
}

// CreateTask persists a new DSRTask.
func (r *DSRRepo) CreateTask(ctx context.Context, task *compliance.DSRTask) error {
	query := `
		INSERT INTO dsr_tasks (
			id, dsr_id, data_source_id, tenant_id,
			task_type, status, result, error,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		task.ID, task.DSRID, task.DataSourceID, task.TenantID,
		task.TaskType, task.Status, task.Result, task.Error,
		task.CreatedAt, task.UpdatedAt,
	).Scan(&task.CreatedAt, &task.UpdatedAt)
}

// GetTasksByDSR retrieves all tasks for a DSR.
func (r *DSRRepo) GetTasksByDSR(ctx context.Context, dsrID types.ID) ([]compliance.DSRTask, error) {
	query := `
		SELECT id, dsr_id, data_source_id, tenant_id,
		       task_type, status, result, error,
		       created_at, updated_at, completed_at
		FROM dsr_tasks
		WHERE dsr_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, dsrID)
	if err != nil {
		return nil, fmt.Errorf("list dsr tasks: %w", err)
	}
	defer rows.Close()

	var tasks []compliance.DSRTask
	for rows.Next() {
		var task compliance.DSRTask
		if err := rows.Scan(
			&task.ID, &task.DSRID, &task.DataSourceID, &task.TenantID,
			&task.TaskType, &task.Status, &task.Result, &task.Error,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dsr task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// UpdateTask updates a DSRTask.
func (r *DSRRepo) UpdateTask(ctx context.Context, task *compliance.DSRTask) error {
	query := `
		UPDATE dsr_tasks
		SET status = $1, result = $2, error = $3,
		    updated_at = NOW(), completed_at = $4
		WHERE id = $5
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		task.Status, task.Result, task.Error, task.CompletedAt,
		task.ID,
	).Scan(&task.UpdatedAt)
}

// Compile-time check
var _ compliance.DSRRepository = (*DSRRepo)(nil)
