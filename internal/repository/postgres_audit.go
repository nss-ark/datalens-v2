package repository

import (
	"context"
	"fmt"

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
	query := `INSERT INTO audit_logs (id, tenant_id, actor_id, action, resource_type, resource_id, changes, ip_address, user_agent, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, query,
		log.ID, log.TenantID, log.ActorID, log.Action,
		log.ResourceType, log.ResourceID, log.Changes,
		log.IPAddress, log.UserAgent, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

// GetByTenant retrieves audit logs for a tenant.
// TODO: Add pagination and filters in future iteration.
func (r *PostgresAuditRepository) GetByTenant(ctx context.Context, tenantID types.ID, limit int) ([]audit.AuditLog, error) {
	query := `SELECT id, tenant_id, actor_id, action, resource_type, resource_id, changes, ip_address, user_agent, created_at
			  FROM audit_logs
			  WHERE tenant_id = $1
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
			&l.ID, &l.TenantID, &l.ActorID, &l.Action,
			&l.ResourceType, &l.ResourceID, &l.Changes,
			&l.IPAddress, &l.UserAgent, &l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}
