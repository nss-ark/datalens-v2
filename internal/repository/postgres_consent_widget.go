package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
)

// ConsentWidgetRepo implements consent.ConsentWidgetRepository.
type ConsentWidgetRepo struct {
	pool *pgxpool.Pool
}

// NewConsentWidgetRepo creates a new ConsentWidgetRepo.
func NewConsentWidgetRepo(pool *pgxpool.Pool) *ConsentWidgetRepo {
	return &ConsentWidgetRepo{pool: pool}
}

// Create persists a new consent widget.
func (r *ConsentWidgetRepo) Create(ctx context.Context, w *consent.ConsentWidget) error {
	configJSON, err := json.Marshal(w.Config)
	if err != nil {
		return fmt.Errorf("marshal widget config: %w", err)
	}

	query := `
		INSERT INTO consent_widgets (
			id, tenant_id, name, type, domain, status, config,
			embed_code, api_key, allowed_origins, version,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		w.ID, w.TenantID, w.Name, w.Type, w.Domain, w.Status,
		configJSON, w.EmbedCode, w.APIKey, w.AllowedOrigins, w.Version,
		w.CreatedAt, w.UpdatedAt,
	).Scan(&w.CreatedAt, &w.UpdatedAt)
}

// GetByID retrieves a consent widget by ID.
func (r *ConsentWidgetRepo) GetByID(ctx context.Context, id types.ID) (*consent.ConsentWidget, error) {
	query := `
		SELECT id, tenant_id, name, type, domain, status, config,
		       embed_code, api_key, allowed_origins, version,
		       created_at, updated_at
		FROM consent_widgets
		WHERE id = $1`

	w := &consent.ConsentWidget{}
	var configJSON []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&w.ID, &w.TenantID, &w.Name, &w.Type, &w.Domain, &w.Status,
		&configJSON, &w.EmbedCode, &w.APIKey, &w.AllowedOrigins, &w.Version,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("consent widget", id)
		}
		return nil, fmt.Errorf("get consent widget: %w", err)
	}

	if err := json.Unmarshal(configJSON, &w.Config); err != nil {
		return nil, fmt.Errorf("unmarshal widget config: %w", err)
	}
	return w, nil
}

// GetByTenant lists all consent widgets for a tenant.
func (r *ConsentWidgetRepo) GetByTenant(ctx context.Context, tenantID types.ID) ([]consent.ConsentWidget, error) {
	query := `
		SELECT id, tenant_id, name, type, domain, status, config,
		       embed_code, api_key, allowed_origins, version,
		       created_at, updated_at
		FROM consent_widgets
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list consent widgets: %w", err)
	}
	defer rows.Close()

	var widgets []consent.ConsentWidget
	for rows.Next() {
		var w consent.ConsentWidget
		var configJSON []byte
		if err := rows.Scan(
			&w.ID, &w.TenantID, &w.Name, &w.Type, &w.Domain, &w.Status,
			&configJSON, &w.EmbedCode, &w.APIKey, &w.AllowedOrigins, &w.Version,
			&w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan consent widget: %w", err)
		}
		if err := json.Unmarshal(configJSON, &w.Config); err != nil {
			return nil, fmt.Errorf("unmarshal widget config: %w", err)
		}
		widgets = append(widgets, w)
	}
	return widgets, rows.Err()
}

// GetByAPIKey retrieves a consent widget by its API key.
func (r *ConsentWidgetRepo) GetByAPIKey(ctx context.Context, apiKey string) (*consent.ConsentWidget, error) {
	query := `
		SELECT id, tenant_id, name, type, domain, status, config,
		       embed_code, api_key, allowed_origins, version,
		       created_at, updated_at
		FROM consent_widgets
		WHERE api_key = $1`

	w := &consent.ConsentWidget{}
	var configJSON []byte
	err := r.pool.QueryRow(ctx, query, apiKey).Scan(
		&w.ID, &w.TenantID, &w.Name, &w.Type, &w.Domain, &w.Status,
		&configJSON, &w.EmbedCode, &w.APIKey, &w.AllowedOrigins, &w.Version,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewNotFoundError("consent widget", nil)
		}
		return nil, fmt.Errorf("get consent widget by api key: %w", err)
	}

	if err := json.Unmarshal(configJSON, &w.Config); err != nil {
		return nil, fmt.Errorf("unmarshal widget config: %w", err)
	}
	return w, nil
}

// Update persists changes to an existing consent widget.
func (r *ConsentWidgetRepo) Update(ctx context.Context, w *consent.ConsentWidget) error {
	configJSON, err := json.Marshal(w.Config)
	if err != nil {
		return fmt.Errorf("marshal widget config: %w", err)
	}

	query := `
		UPDATE consent_widgets
		SET name = $1, type = $2, domain = $3, status = $4, config = $5,
		    embed_code = $6, allowed_origins = $7, version = $8,
		    updated_at = NOW()
		WHERE id = $9 AND tenant_id = $10
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		w.Name, w.Type, w.Domain, w.Status, configJSON,
		w.EmbedCode, w.AllowedOrigins, w.Version,
		w.ID, w.TenantID,
	).Scan(&w.UpdatedAt)
}

// Delete removes a consent widget.
func (r *ConsentWidgetRepo) Delete(ctx context.Context, id types.ID) error {
	query := `DELETE FROM consent_widgets WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete consent widget: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return types.NewNotFoundError("consent widget", id)
	}
	return nil
}

// Compile-time interface check.
var _ consent.ConsentWidgetRepository = (*ConsentWidgetRepo)(nil)
