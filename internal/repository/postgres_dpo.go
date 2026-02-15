package repository

import (
	"context"
	"fmt"

	"github.com/complyark/datalens/internal/domain/compliance"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDPOContactRepository implements compliance.DPOContactRepository using PostgreSQL.
type PostgresDPOContactRepository struct {
	db *pgxpool.Pool
}

// NewPostgresDPOContactRepository creates a new instance of PostgresDPOContactRepository.
func NewPostgresDPOContactRepository(db *pgxpool.Pool) *PostgresDPOContactRepository {
	return &PostgresDPOContactRepository{db: db}
}

// Upsert creates or updates a DPO contact record.
func (r *PostgresDPOContactRepository) Upsert(ctx context.Context, contact *compliance.DPOContact) error {
	query := `
		INSERT INTO dpo_contacts (
			tenant_id, org_name, dpo_name, dpo_email, dpo_phone, address, website_url, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		ON CONFLICT (tenant_id) DO UPDATE SET
			org_name = EXCLUDED.org_name,
			dpo_name = EXCLUDED.dpo_name,
			dpo_email = EXCLUDED.dpo_email,
			dpo_phone = EXCLUDED.dpo_phone,
			address = EXCLUDED.address,
			website_url = EXCLUDED.website_url,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		contact.TenantID,
		contact.OrgName,
		contact.DPOName,
		contact.DPOEmail,
		contact.DPOPhone,
		contact.Address,
		contact.WebsiteURL,
		contact.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("upsert dpo contact: %w", err)
	}

	return nil
}

// Get retrieves a DPO contact record by tenant ID.
func (r *PostgresDPOContactRepository) Get(ctx context.Context, tenantID types.ID) (*compliance.DPOContact, error) {
	query := `
		SELECT 
			tenant_id, org_name, dpo_name, dpo_email, dpo_phone, address, website_url, updated_at
		FROM dpo_contacts
		WHERE tenant_id = $1
	`

	var contact compliance.DPOContact
	err := r.db.QueryRow(ctx, query, tenantID).Scan(
		&contact.TenantID,
		&contact.OrgName,
		&contact.DPOName,
		&contact.DPOEmail,
		&contact.DPOPhone,
		&contact.Address,
		&contact.WebsiteURL,
		&contact.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("dpo contact not found", map[string]any{"tenant_id": tenantID})
		}
		return nil, fmt.Errorf("get dpo contact: %w", err)
	}

	return &contact, nil
}
