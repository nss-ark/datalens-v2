package service

import (
	"context"

	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresClientRepository implements ClientRepository using pgx.
type PostgresClientRepository struct {
	db *pgxpool.Pool
}

func NewPostgresClientRepository(db *pgxpool.Pool) *PostgresClientRepository {
	return &PostgresClientRepository{db: db}
}

func (r *PostgresClientRepository) GetClient(ctx context.Context, tenantID types.ID) (*Client, error) {
	query := `SELECT id, name, logo_url, primary_color, support_email, portal_url FROM clients WHERE id = $1`
	var c Client
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&c.ID, &c.Name, &c.LogoURL, &c.PrimaryColor, &c.SupportEmail, &c.PortalURL)
	if err == nil {
		return &c, nil
	}

	// Fallback as per original logic
	return &Client{ID: tenantID, Name: "DataLens", SupportEmail: func(s string) *string { return &s }("support@datalens.com")}, nil
}
