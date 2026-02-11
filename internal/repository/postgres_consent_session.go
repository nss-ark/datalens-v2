package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
)

// ConsentSessionRepo implements consent.ConsentSessionRepository.
type ConsentSessionRepo struct {
	pool *pgxpool.Pool
}

// NewConsentSessionRepo creates a new ConsentSessionRepo.
func NewConsentSessionRepo(pool *pgxpool.Pool) *ConsentSessionRepo {
	return &ConsentSessionRepo{pool: pool}
}

// Create persists a new consent session (append-only — no updates or deletes).
func (r *ConsentSessionRepo) Create(ctx context.Context, s *consent.ConsentSession) error {
	decisionsJSON, err := json.Marshal(s.Decisions)
	if err != nil {
		return fmt.Errorf("marshal consent decisions: %w", err)
	}

	query := `
		INSERT INTO consent_sessions (
			id, tenant_id, widget_id, subject_id, decisions,
			ip_address, user_agent, page_url, widget_version,
			notice_version, signature, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at`

	return r.pool.QueryRow(ctx, query,
		s.ID, s.TenantID, s.WidgetID, s.SubjectID, decisionsJSON,
		s.IPAddress, s.UserAgent, s.PageURL, s.WidgetVersion,
		s.NoticeVersion, s.Signature, s.CreatedAt,
	).Scan(&s.CreatedAt)
}

// GetBySubject retrieves all consent sessions for a subject within a tenant.
func (r *ConsentSessionRepo) GetBySubject(ctx context.Context, tenantID, subjectID types.ID) ([]consent.ConsentSession, error) {
	query := `
		SELECT id, tenant_id, widget_id, subject_id, decisions,
		       ip_address, user_agent, page_url, widget_version,
		       notice_version, signature, created_at
		FROM consent_sessions
		WHERE tenant_id = $1 AND subject_id = $2
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID, subjectID)
	if err != nil {
		return nil, fmt.Errorf("list consent sessions: %w", err)
	}
	defer rows.Close()

	var sessions []consent.ConsentSession
	for rows.Next() {
		var s consent.ConsentSession
		var decisionsJSON []byte
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.WidgetID, &s.SubjectID, &decisionsJSON,
			&s.IPAddress, &s.UserAgent, &s.PageURL, &s.WidgetVersion,
			&s.NoticeVersion, &s.Signature, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan consent session: %w", err)
		}
		if err := json.Unmarshal(decisionsJSON, &s.Decisions); err != nil {
			return nil, fmt.Errorf("unmarshal consent decisions: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// Compile-time interface check.
var _ consent.ConsentSessionRepository = (*ConsentSessionRepo)(nil)
