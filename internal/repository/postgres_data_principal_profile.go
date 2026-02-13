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

// PostgresDataPrincipalProfileRepository implements consent.DataPrincipalProfileRepository.
type PostgresDataPrincipalProfileRepository struct {
	db *pgxpool.Pool
}

// NewDataPrincipalProfileRepo creates a new PostgresDataPrincipalProfileRepository.
func NewDataPrincipalProfileRepo(db *pgxpool.Pool) *PostgresDataPrincipalProfileRepository {
	return &PostgresDataPrincipalProfileRepository{db: db}
}

// Create persists a new DataPrincipalProfile.
func (r *PostgresDataPrincipalProfileRepository) Create(ctx context.Context, p *consent.DataPrincipalProfile) error {
	query := `
		INSERT INTO data_principal_profiles (
			id, tenant_id, email, phone, verification_status, verified_at,
			verification_method, subject_id, last_access_at, preferred_lang,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.TenantID, p.Email, p.Phone, p.VerificationStatus, p.VerifiedAt,
		p.VerificationMethod, p.SubjectID, p.LastAccessAt, p.PreferredLang,
		p.CreatedAt, p.UpdatedAt,
	)
	return err
}

// GetByID retrieves a DataPrincipalProfile by its ID.
func (r *PostgresDataPrincipalProfileRepository) GetByID(ctx context.Context, id types.ID) (*consent.DataPrincipalProfile, error) {
	query := `
		SELECT
			id, tenant_id, email, phone, verification_status, verified_at,
			verification_method, subject_id, last_access_at, preferred_lang,
			created_at, updated_at
		FROM data_principal_profiles
		WHERE id = $1
	`
	var p consent.DataPrincipalProfile
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.TenantID, &p.Email, &p.Phone, &p.VerificationStatus, &p.VerifiedAt,
		&p.VerificationMethod, &p.SubjectID, &p.LastAccessAt, &p.PreferredLang,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("profile not found", nil)
		}
		return nil, fmt.Errorf("get profile by id: %w", err)
	}
	return &p, nil
}

// GetByEmail retrieves a DataPrincipalProfile by email within a tenant.
func (r *PostgresDataPrincipalProfileRepository) GetByEmail(ctx context.Context, tenantID types.ID, email string) (*consent.DataPrincipalProfile, error) {
	query := `
		SELECT
			id, tenant_id, email, phone, verification_status, verified_at,
			verification_method, subject_id, last_access_at, preferred_lang,
			created_at, updated_at
		FROM data_principal_profiles
		WHERE tenant_id = $1 AND email = $2
	`
	var p consent.DataPrincipalProfile
	err := r.db.QueryRow(ctx, query, tenantID, email).Scan(
		&p.ID, &p.TenantID, &p.Email, &p.Phone, &p.VerificationStatus, &p.VerifiedAt,
		&p.VerificationMethod, &p.SubjectID, &p.LastAccessAt, &p.PreferredLang,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("profile not found", nil)
		}
		return nil, fmt.Errorf("get profile by email: %w", err)
	}
	return &p, nil
}

// Update persists changes to an existing DataPrincipalProfile.
func (r *PostgresDataPrincipalProfileRepository) Update(ctx context.Context, p *consent.DataPrincipalProfile) error {
	p.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE data_principal_profiles
		SET
			email = $1, phone = $2, verification_status = $3, verified_at = $4,
			verification_method = $5, subject_id = $6, last_access_at = $7,
			preferred_lang = $8, updated_at = $9
		WHERE id = $10 AND tenant_id = $11
	`
	cmdTag, err := r.db.Exec(ctx, query,
		p.Email, p.Phone, p.VerificationStatus, p.VerifiedAt,
		p.VerificationMethod, p.SubjectID, p.LastAccessAt, p.PreferredLang,
		p.UpdatedAt, p.ID, p.TenantID,
	)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return types.NewNotFoundError("profile not found or update failed", nil)
	}
	return nil
}

// ListByTenant retrieves a paginated list of DataPrincipalProfiles for a tenant.
func (r *PostgresDataPrincipalProfileRepository) ListByTenant(ctx context.Context, tenantID types.ID, pagination types.Pagination) (*types.PaginatedResult[consent.DataPrincipalProfile], error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 {
		pagination.PageSize = 20
	}
	offset := (pagination.Page - 1) * pagination.PageSize

	// Count
	var total int
	countQuery := `SELECT COUNT(*) FROM data_principal_profiles WHERE tenant_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count profiles: %w", err)
	}

	// List
	query := `
		SELECT
			id, tenant_id, email, phone, verification_status, verified_at,
			verification_method, subject_id, last_access_at, preferred_lang,
			created_at, updated_at
		FROM data_principal_profiles
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, pagination.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	defer rows.Close()

	var items []consent.DataPrincipalProfile
	for rows.Next() {
		var p consent.DataPrincipalProfile
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.Email, &p.Phone, &p.VerificationStatus, &p.VerifiedAt,
			&p.VerificationMethod, &p.SubjectID, &p.LastAccessAt, &p.PreferredLang,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}
		items = append(items, p)
	}

	return &types.PaginatedResult[consent.DataPrincipalProfile]{
		Items:       items,
		Total:       total,
		Page:        pagination.Page,
		PageSize:    pagination.PageSize,
		TotalPages:  (total + pagination.PageSize - 1) / pagination.PageSize,
		HasNextPage: total > offset+pagination.PageSize,
	}, nil
}
