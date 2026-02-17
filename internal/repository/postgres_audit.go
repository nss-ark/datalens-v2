package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/complyark/datalens/internal/domain/audit"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAuditRepository implements audit.Repository using PostgreSQL.
type PostgresAuditRepository struct {
	db *pgxpool.Pool
}

// NewPostgresAuditRepository creates a new PostgresAuditRepository.
func NewPostgresAuditRepository(db *pgxpool.Pool) *PostgresAuditRepository {
	return &PostgresAuditRepository{db: db}
}

// Create persists a new audit log entry.
func (r *PostgresAuditRepository) Create(ctx context.Context, log *audit.AuditLog) error {
	// Map TenantID -> client_id, UserID -> user_id
	query := `INSERT INTO audit_logs (id, client_id, user_id, action, resource_type, resource_id, old_values, new_values, ip_address, user_agent, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.Exec(ctx, query,
		log.ID, log.TenantID, log.UserID, log.Action,
		log.ResourceType, log.ResourceID, log.OldValues, log.NewValues,
		log.IPAddress, log.UserAgent, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

// GetByTenant retrieves audit logs for a tenant.
func (r *PostgresAuditRepository) GetByTenant(ctx context.Context, tenantID types.ID, limit int) ([]audit.AuditLog, error) {
	query := `SELECT id, client_id, user_id, action, resource_type, resource_id, old_values, new_values, ip_address, user_agent, created_at
			  FROM audit_logs
			  WHERE client_id = $1
			  ORDER BY created_at DESC
			  LIMIT $2`

	rows, err := r.db.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []audit.AuditLog
	for rows.Next() {
		var l audit.AuditLog
		if err := rows.Scan(
			&l.ID, &l.TenantID, &l.UserID, &l.Action,
			&l.ResourceType, &l.ResourceID, &l.OldValues, &l.NewValues,
			&l.IPAddress, &l.UserAgent, &l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// ListByTenant retrieves paginated, filtered audit logs for a tenant.
func (r *PostgresAuditRepository) ListByTenant(ctx context.Context, tenantID types.ID, filters audit.AuditFilters, pagination types.Pagination) (*types.PaginatedResult[audit.AuditLog], error) {
	// Build dynamic WHERE clause
	var conditions []string
	var args []any
	argIdx := 1

	// Base tenant filter
	conditions = append(conditions, "client_id = $"+strconv.Itoa(argIdx))
	args = append(args, tenantID)
	argIdx++

	if filters.EntityType != "" {
		conditions = append(conditions, "resource_type = $"+strconv.Itoa(argIdx))
		args = append(args, filters.EntityType)
		argIdx++
	}

	if filters.Action != "" {
		conditions = append(conditions, "action = $"+strconv.Itoa(argIdx))
		args = append(args, filters.Action)
		argIdx++
	}

	if filters.UserID != nil {
		conditions = append(conditions, "COALESCE(user_id, actor_id) = $"+strconv.Itoa(argIdx))
		args = append(args, *filters.UserID)
		argIdx++
	}

	if filters.StartDate != nil {
		conditions = append(conditions, "created_at >= $"+strconv.Itoa(argIdx))
		args = append(args, *filters.StartDate)
		argIdx++
	}

	if filters.EndDate != nil {
		conditions = append(conditions, "created_at <= $"+strconv.Itoa(argIdx))
		args = append(args, *filters.EndDate)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := "SELECT COUNT(*) FROM audit_logs WHERE " + whereClause
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count audit logs: %w", err)
	}

	// Data query with pagination
	limit := pagination.Limit()
	offset := pagination.Offset()

	dataQuery := fmt.Sprintf(
		`SELECT id, client_id, COALESCE(user_id, actor_id), action, resource_type, resource_id, old_values, new_values, ip_address, user_agent, created_at
		 FROM audit_logs
		 WHERE %s
		 ORDER BY created_at DESC
		 LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []audit.AuditLog
	for rows.Next() {
		var l audit.AuditLog
		if err := rows.Scan(
			&l.ID, &l.TenantID, &l.UserID, &l.Action,
			&l.ResourceType, &l.ResourceID, &l.OldValues, &l.NewValues,
			&l.IPAddress, &l.UserAgent, &l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, l)
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return &types.PaginatedResult[audit.AuditLog]{
		Items:      logs,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   limit,
		TotalPages: totalPages,
	}, nil
}
