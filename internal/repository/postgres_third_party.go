package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/complyark/datalens/internal/domain/governance"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ThirdPartyRepo implements governance.ThirdPartyRepository.
type ThirdPartyRepo struct {
	pool *pgxpool.Pool
}

// NewThirdPartyRepo creates a new ThirdPartyRepo.
func NewThirdPartyRepo(pool *pgxpool.Pool) *ThirdPartyRepo {
	return &ThirdPartyRepo{pool: pool}
}

func (r *ThirdPartyRepo) Create(ctx context.Context, tp *governance.ThirdParty) error {
	tp.ID = types.NewID()
	tp.CreatedAt = time.Now()
	tp.UpdatedAt = time.Now()
	if !tp.IsActive {
		tp.IsActive = true
	}

	purposeIDsJSON, err := json.Marshal(tp.PurposeIDs)
	if err != nil {
		return fmt.Errorf("marshal purpose_ids: %w", err)
	}

	query := `
		INSERT INTO third_parties (
			id, tenant_id, name, type, country, dpa_doc_path, is_active,
			purpose_ids, dpa_status, dpa_signed_at, dpa_expires_at, dpa_notes,
			contact_name, contact_email, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err = r.pool.Exec(ctx, query,
		tp.ID, tp.TenantID, tp.Name, tp.Type, tp.Country, tp.DPADocPath, tp.IsActive,
		purposeIDsJSON, tp.DPAStatus, tp.DPASignedAt, tp.DPAExpiresAt, tp.DPANotes,
		tp.ContactName, tp.ContactEmail, tp.CreatedAt, tp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create third party: %w", err)
	}
	return nil
}

func (r *ThirdPartyRepo) GetByID(ctx context.Context, id types.ID) (*governance.ThirdParty, error) {
	query := `
		SELECT id, tenant_id, name, type, country, dpa_doc_path, is_active,
		       purpose_ids, dpa_status, dpa_signed_at, dpa_expires_at, dpa_notes,
		       contact_name, contact_email, created_at, updated_at
		FROM third_parties WHERE id = $1
	`
	var tp governance.ThirdParty
	var purposeIDsJSON []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tp.ID, &tp.TenantID, &tp.Name, &tp.Type, &tp.Country, &tp.DPADocPath, &tp.IsActive,
		&purposeIDsJSON, &tp.DPAStatus, &tp.DPASignedAt, &tp.DPAExpiresAt, &tp.DPANotes,
		&tp.ContactName, &tp.ContactEmail, &tp.CreatedAt, &tp.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("ThirdParty", id)
		}
		return nil, fmt.Errorf("get third party: %w", err)
	}
	if err := json.Unmarshal(purposeIDsJSON, &tp.PurposeIDs); err != nil {
		return nil, fmt.Errorf("unmarshal purpose_ids: %w", err)
	}
	return &tp, nil
}

func (r *ThirdPartyRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]governance.ThirdParty, error) {
	query := `
		SELECT id, tenant_id, name, type, country, dpa_doc_path, is_active,
		       purpose_ids, dpa_status, dpa_signed_at, dpa_expires_at, dpa_notes,
		       contact_name, contact_email, created_at, updated_at
		FROM third_parties WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query third parties: %w", err)
	}
	defer rows.Close()

	var results []governance.ThirdParty
	for rows.Next() {
		var tp governance.ThirdParty
		var purposeIDsJSON []byte
		if err := rows.Scan(
			&tp.ID, &tp.TenantID, &tp.Name, &tp.Type, &tp.Country, &tp.DPADocPath, &tp.IsActive,
			&purposeIDsJSON, &tp.DPAStatus, &tp.DPASignedAt, &tp.DPAExpiresAt, &tp.DPANotes,
			&tp.ContactName, &tp.ContactEmail, &tp.CreatedAt, &tp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan third party: %w", err)
		}
		if err := json.Unmarshal(purposeIDsJSON, &tp.PurposeIDs); err != nil {
			return nil, fmt.Errorf("unmarshal purpose_ids: %w", err)
		}
		results = append(results, tp)
	}
	return results, nil
}

func (r *ThirdPartyRepo) Update(ctx context.Context, tp *governance.ThirdParty) error {
	tp.UpdatedAt = time.Now()

	purposeIDsJSON, err := json.Marshal(tp.PurposeIDs)
	if err != nil {
		return fmt.Errorf("marshal purpose_ids: %w", err)
	}

	query := `
		UPDATE third_parties SET
			name = $2, type = $3, country = $4, dpa_doc_path = $5,
			is_active = $6, purpose_ids = $7, dpa_status = $8,
			dpa_signed_at = $9, dpa_expires_at = $10, dpa_notes = $11,
			contact_name = $12, contact_email = $13, updated_at = $14
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query,
		tp.ID, tp.Name, tp.Type, tp.Country, tp.DPADocPath,
		tp.IsActive, purposeIDsJSON, tp.DPAStatus,
		tp.DPASignedAt, tp.DPAExpiresAt, tp.DPANotes,
		tp.ContactName, tp.ContactEmail, tp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update third party: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return types.NewNotFoundError("ThirdParty", tp.ID)
	}
	return nil
}

func (r *ThirdPartyRepo) Delete(ctx context.Context, id types.ID) error {
	query := `DELETE FROM third_parties WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete third party: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return types.NewNotFoundError("ThirdParty", id)
	}
	return nil
}
