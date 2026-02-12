package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresIdentityProfileRepo struct {
	db *pgxpool.Pool
}

func NewIdentityProfileRepo(db *pgxpool.Pool) *PostgresIdentityProfileRepo {
	return &PostgresIdentityProfileRepo{db: db}
}

func (r *PostgresIdentityProfileRepo) Create(ctx context.Context, profile *identity.IdentityProfile) error {
	query := `
		INSERT INTO identity_profiles (
			id, tenant_id, subject_id, assurance_level, verification_status, 
			documents, last_verified_at, next_verification_due, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		profile.ID, profile.TenantID, profile.SubjectID, profile.AssuranceLevel, profile.VerificationStatus,
		profile.Documents, profile.LastVerifiedAt, profile.NextVerificationDue, profile.CreatedAt, profile.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create identity profile: %w", err)
	}
	return nil
}

func (r *PostgresIdentityProfileRepo) GetBySubject(ctx context.Context, tenantID, subjectID types.ID) (*identity.IdentityProfile, error) {
	query := `
		SELECT id, tenant_id, subject_id, assurance_level, verification_status, 
		       documents, last_verified_at, next_verification_due, created_at, updated_at
		FROM identity_profiles
		WHERE tenant_id = $1 AND subject_id = $2
	`
	var profile identity.IdentityProfile
	err := r.db.QueryRow(ctx, query, tenantID, subjectID).Scan(
		&profile.ID, &profile.TenantID, &profile.SubjectID, &profile.AssuranceLevel, &profile.VerificationStatus,
		&profile.Documents, &profile.LastVerifiedAt, &profile.NextVerificationDue, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("IdentityProfile", subjectID)
		}
		return nil, fmt.Errorf("get identity profile: %w", err)
	}
	return &profile, nil
}

func (r *PostgresIdentityProfileRepo) Update(ctx context.Context, profile *identity.IdentityProfile) error {
	query := `
		UPDATE identity_profiles
		SET assurance_level = $1, verification_status = $2, documents = $3, 
		    last_verified_at = $4, next_verification_due = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8
	`
	_, err := r.db.Exec(ctx, query,
		profile.AssuranceLevel, profile.VerificationStatus, profile.Documents,
		profile.LastVerifiedAt, profile.NextVerificationDue, profile.UpdatedAt,
		profile.ID, profile.TenantID,
	)
	if err != nil {
		return fmt.Errorf("update identity profile: %w", err)
	}
	return nil
}
