package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
)

// PostgresDPRRequestRepository implements consent.DPRRequestRepository.
type PostgresDPRRequestRepository struct {
	db *pgxpool.Pool
}

// NewDPRRequestRepo creates a new PostgresDPRRequestRepository.
func NewDPRRequestRepo(db *pgxpool.Pool) *PostgresDPRRequestRepository {
	return &PostgresDPRRequestRepository{db: db}
}

// Create persists a new DPRRequest.
func (r *PostgresDPRRequestRepository) Create(ctx context.Context, req *consent.DPRRequest) error {
	query := `
		INSERT INTO dpr_requests (
			id, tenant_id, profile_id, dsr_id, type, description, status,
			submitted_at, deadline, verified_at, verification_ref, is_minor,
			guardian_name, guardian_email, guardian_relation, guardian_verified,
			completed_at, response_summary, download_url, appeal_of, appeal_reason,
			is_escalated, escalated_to, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25
		)
	`
	_, err := r.db.Exec(ctx, query,
		req.ID, req.TenantID, req.ProfileID, req.DSRID, req.Type, req.Description,
		req.Status, req.SubmittedAt, req.Deadline, req.VerifiedAt, req.VerificationRef,
		req.IsMinor, req.GuardianName, req.GuardianEmail, req.GuardianRelation,
		req.GuardianVerified, req.CompletedAt, req.ResponseSummary, req.DownloadURL,
		req.AppealOf, req.AppealReason, req.IsEscalated, req.EscalatedTo,
		req.CreatedAt, req.UpdatedAt,
	)
	return err
}

// GetByID retrieves a DPRRequest by its ID.
func (r *PostgresDPRRequestRepository) GetByID(ctx context.Context, id types.ID) (*consent.DPRRequest, error) {
	query := `
		SELECT
			id, tenant_id, profile_id, dsr_id, type, description, status,
			submitted_at, deadline, verified_at, verification_ref, is_minor,
			guardian_name, guardian_email, guardian_relation, guardian_verified,
			completed_at, response_summary, download_url, appeal_of, appeal_reason,
			is_escalated, escalated_to, created_at, updated_at
		FROM dpr_requests
		WHERE id = $1
	`
	var req consent.DPRRequest
	err := r.db.QueryRow(ctx, query, id).Scan(
		&req.ID, &req.TenantID, &req.ProfileID, &req.DSRID, &req.Type, &req.Description,
		&req.Status, &req.SubmittedAt, &req.Deadline, &req.VerifiedAt, &req.VerificationRef,
		&req.IsMinor, &req.GuardianName, &req.GuardianEmail, &req.GuardianRelation,
		&req.GuardianVerified, &req.CompletedAt, &req.ResponseSummary, &req.DownloadURL,
		&req.AppealOf, &req.AppealReason, &req.IsEscalated, &req.EscalatedTo,
		&req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("DPR request not found", nil)
		}
		return nil, fmt.Errorf("get DPR by id: %w", err)
	}
	return &req, nil
}

// GetByProfile retrieves all DPRRequests for a given profile.
func (r *PostgresDPRRequestRepository) GetByProfile(ctx context.Context, profileID types.ID) ([]consent.DPRRequest, error) {
	query := `
		SELECT
			id, tenant_id, profile_id, dsr_id, type, description, status,
			submitted_at, deadline, verified_at, verification_ref, is_minor,
			guardian_name, guardian_email, guardian_relation, guardian_verified,
			completed_at, response_summary, download_url, appeal_of, appeal_reason,
			is_escalated, escalated_to, created_at, updated_at
		FROM dpr_requests
		WHERE profile_id = $1
		ORDER BY submitted_at DESC
	`
	rows, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("query DPRs by profile: %w", err)
	}
	defer rows.Close()

	var requests []consent.DPRRequest
	for rows.Next() {
		var req consent.DPRRequest
		err := rows.Scan(
			&req.ID, &req.TenantID, &req.ProfileID, &req.DSRID, &req.Type, &req.Description,
			&req.Status, &req.SubmittedAt, &req.Deadline, &req.VerifiedAt, &req.VerificationRef,
			&req.IsMinor, &req.GuardianName, &req.GuardianEmail, &req.GuardianRelation,
			&req.GuardianVerified, &req.CompletedAt, &req.ResponseSummary, &req.DownloadURL,
			&req.AppealOf, &req.AppealReason, &req.IsEscalated, &req.EscalatedTo,
			&req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan DPR row: %w", err)
		}
		requests = append(requests, req)
	}
	return requests, nil
}

// GetByTenant retrieves paginated DPRRequests for a tenant.
func (r *PostgresDPRRequestRepository) GetByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[consent.DPRRequest], error) {
	query := `
		SELECT
			id, tenant_id, profile_id, dsr_id, type, description, status,
			submitted_at, deadline, verified_at, verification_ref, is_minor,
			guardian_name, guardian_email, guardian_relation, guardian_verified,
			completed_at, response_summary, download_url, appeal_of, appeal_reason,
			is_escalated, escalated_to, created_at, updated_at
		FROM dpr_requests
		WHERE tenant_id = $1
		ORDER BY submitted_at DESC
		LIMIT $2 OFFSET $3
	`
	offset := (pagination.Page - 1) * pagination.PageSize
	rows, err := r.db.Query(ctx, query, tenantID, pagination.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query DPRs by tenant: %w", err)
	}
	defer rows.Close()

	var requests []consent.DPRRequest
	for rows.Next() {
		var req consent.DPRRequest
		err := rows.Scan(
			&req.ID, &req.TenantID, &req.ProfileID, &req.DSRID, &req.Type, &req.Description,
			&req.Status, &req.SubmittedAt, &req.Deadline, &req.VerifiedAt, &req.VerificationRef,
			&req.IsMinor, &req.GuardianName, &req.GuardianEmail, &req.GuardianRelation,
			&req.GuardianVerified, &req.CompletedAt, &req.ResponseSummary, &req.DownloadURL,
			&req.AppealOf, &req.AppealReason, &req.IsEscalated, &req.EscalatedTo,
			&req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan DPR row: %w", err)
		}
		requests = append(requests, req)
	}

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM dpr_requests WHERE tenant_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count DPRs: %w", err)
	}

	return &types.PaginatedResult[consent.DPRRequest]{
		Items:      requests,
		Total:      int(total),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: (int(total) + pagination.PageSize - 1) / pagination.PageSize,
	}, nil
}

// Update persists changes to an existing DPRRequest.
func (r *PostgresDPRRequestRepository) Update(ctx context.Context, req *consent.DPRRequest) error {
	req.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE dpr_requests
		SET
			profile_id = $1, dsr_id = $2, type = $3, description = $4, status = $5,
			submitted_at = $6, deadline = $7, verified_at = $8, verification_ref = $9,
			is_minor = $10, guardian_name = $11, guardian_email = $12,
			guardian_relation = $13, guardian_verified = $14, completed_at = $15,
			response_summary = $16, download_url = $17, appeal_of = $18,
			appeal_reason = $19, is_escalated = $20, escalated_to = $21,
			updated_at = $22
		WHERE id = $23 AND tenant_id = $24
	`
	cmdTag, err := r.db.Exec(ctx, query,
		req.ProfileID, req.DSRID, req.Type, req.Description, req.Status,
		req.SubmittedAt, req.Deadline, req.VerifiedAt, req.VerificationRef,
		req.IsMinor, req.GuardianName, req.GuardianEmail,
		req.GuardianRelation, req.GuardianVerified, req.CompletedAt,
		req.ResponseSummary, req.DownloadURL, req.AppealOf,
		req.AppealReason, req.IsEscalated, req.EscalatedTo,
		req.UpdatedAt, req.ID, req.TenantID,
	)
	if err != nil {
		return fmt.Errorf("update DPR: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return types.NewNotFoundError("DPR request not found or update failed", nil)
	}
	return nil
}

// GetByDSRID retrieves a DPRRequest by its linked DSR ID.
func (r *PostgresDPRRequestRepository) GetByDSRID(ctx context.Context, dsrID types.ID) (*consent.DPRRequest, error) {
	query := `
		SELECT
			id, tenant_id, profile_id, dsr_id, type, description, status,
			submitted_at, deadline, verified_at, verification_ref, is_minor,
			guardian_name, guardian_email, guardian_relation, guardian_verified,
			completed_at, response_summary, download_url, appeal_of, appeal_reason,
			is_escalated, escalated_to, created_at, updated_at
		FROM dpr_requests
		WHERE dsr_id = $1
	`
	var req consent.DPRRequest
	err := r.db.QueryRow(ctx, query, dsrID).Scan(
		&req.ID, &req.TenantID, &req.ProfileID, &req.DSRID, &req.Type, &req.Description,
		&req.Status, &req.SubmittedAt, &req.Deadline, &req.VerifiedAt, &req.VerificationRef,
		&req.IsMinor, &req.GuardianName, &req.GuardianEmail, &req.GuardianRelation,
		&req.GuardianVerified, &req.CompletedAt, &req.ResponseSummary, &req.DownloadURL,
		&req.AppealOf, &req.AppealReason, &req.IsEscalated, &req.EscalatedTo,
		&req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("DPR request not found for DSR", dsrID)
		}
		return nil, fmt.Errorf("get DPR by dsr_id: %w", err)
	}
	return &req, nil
}
