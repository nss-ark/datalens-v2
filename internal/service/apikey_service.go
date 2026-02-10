package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/complyark/datalens/pkg/types"
)

// APIKeyInfo is the public representation of an API key (never includes the raw key).
type APIKeyInfo struct {
	ID          types.ID              `json:"id"`
	TenantID    types.ID              `json:"tenant_id"`
	Name        string                `json:"name"`
	Prefix      string                `json:"prefix"`
	Permissions []identity.Permission `json:"permissions"`
	ExpiresAt   *time.Time            `json:"expires_at,omitempty"`
	LastUsed    *time.Time            `json:"last_used,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	RevokedAt   *time.Time            `json:"revoked_at,omitempty"`
}

// CreateKeyResult is returned when a new key is created.
// The RawKey is shown only once at creation time.
type CreateKeyResult struct {
	APIKeyInfo
	RawKey string `json:"raw_key"`
}

// APIKeyService manages API keys for agent authentication.
type APIKeyService struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewAPIKeyService creates a new APIKeyService.
func NewAPIKeyService(pool *pgxpool.Pool, logger *slog.Logger) *APIKeyService {
	return &APIKeyService{
		pool:   pool,
		logger: logger.With("service", "apikey"),
	}
}

// CreateKey generates a new API key for a tenant.
func (s *APIKeyService) CreateKey(ctx context.Context, tenantID types.ID, name string, permissions []identity.Permission) (*CreateKeyResult, error) {
	if name == "" {
		return nil, types.NewValidationError("key name is required", nil)
	}

	// Generate a cryptographically secure random key
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	rawKey := "dlk_" + hex.EncodeToString(rawBytes) // dlk_ prefix for identification
	prefix := rawKey[:12]                           // "dlk_" + 8 hex chars

	hash, err := bcrypt.GenerateFromPassword([]byte(rawKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash key: %w", err)
	}

	id := types.NewID()
	query := `
		INSERT INTO api_keys (id, tenant_id, name, key_hash, prefix, permissions)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	var createdAt, updatedAt time.Time
	err = s.pool.QueryRow(ctx, query, id, tenantID, name, string(hash), prefix, permissions).
		Scan(&createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert api key: %w", err)
	}

	s.logger.InfoContext(ctx, "api key created", "id", id, "tenant_id", tenantID, "name", name)

	return &CreateKeyResult{
		APIKeyInfo: APIKeyInfo{
			ID:          id,
			TenantID:    tenantID,
			Name:        name,
			Prefix:      prefix,
			Permissions: permissions,
			CreatedAt:   createdAt,
		},
		RawKey: rawKey,
	}, nil
}

// ValidateKey checks a raw API key against stored hashes.
// Returns the tenant ID and permissions if valid.
func (s *APIKeyService) ValidateKey(ctx context.Context, rawKey string) (types.ID, []identity.Permission, error) {
	if len(rawKey) < 12 {
		return types.ID{}, nil, types.NewUnauthorizedError("invalid api key")
	}
	prefix := rawKey[:12]

	query := `
		SELECT id, tenant_id, key_hash, permissions, expires_at, revoked_at
		FROM api_keys
		WHERE prefix = $1`

	rows, err := s.pool.Query(ctx, query, prefix)
	if err != nil {
		return types.ID{}, nil, fmt.Errorf("query api keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id          types.ID
			tenantID    types.ID
			keyHash     string
			permissions []identity.Permission
			expiresAt   *time.Time
			revokedAt   *time.Time
		)
		if err := rows.Scan(&id, &tenantID, &keyHash, &permissions, &expiresAt, &revokedAt); err != nil {
			continue
		}

		// Skip revoked keys
		if revokedAt != nil {
			continue
		}

		// Skip expired keys
		if expiresAt != nil && expiresAt.Before(time.Now().UTC()) {
			continue
		}

		// Check hash
		if err := bcrypt.CompareHashAndPassword([]byte(keyHash), []byte(rawKey)); err != nil {
			continue
		}

		// Match found — update last_used
		_, _ = s.pool.Exec(ctx, `UPDATE api_keys SET last_used = NOW() WHERE id = $1`, id)

		return tenantID, permissions, nil
	}

	return types.ID{}, nil, types.NewUnauthorizedError("invalid or expired api key")
}

// RevokeKey marks an API key as revoked.
func (s *APIKeyService) RevokeKey(ctx context.Context, keyID types.ID) error {
	query := `UPDATE api_keys SET revoked_at = NOW(), updated_at = NOW() WHERE id = $1 AND revoked_at IS NULL`
	ct, err := s.pool.Exec(ctx, query, keyID)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return types.NewNotFoundError("APIKey", keyID)
	}
	s.logger.InfoContext(ctx, "api key revoked", "id", keyID)
	return nil
}

// ListKeys returns all API keys for a tenant (without secrets).
func (s *APIKeyService) ListKeys(ctx context.Context, tenantID types.ID) ([]APIKeyInfo, error) {
	query := `
		SELECT id, tenant_id, name, prefix, permissions, expires_at, last_used, created_at, revoked_at
		FROM api_keys
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKeyInfo
	for rows.Next() {
		var k APIKeyInfo
		if err := rows.Scan(
			&k.ID, &k.TenantID, &k.Name, &k.Prefix, &k.Permissions,
			&k.ExpiresAt, &k.LastUsed, &k.CreatedAt, &k.RevokedAt,
		); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// GetKeyByID returns a single API key info by its ID.
func (s *APIKeyService) GetKeyByID(ctx context.Context, keyID types.ID) (*APIKeyInfo, error) {
	query := `
		SELECT id, tenant_id, name, prefix, permissions, expires_at, last_used, created_at, revoked_at
		FROM api_keys
		WHERE id = $1`

	k := &APIKeyInfo{}
	err := s.pool.QueryRow(ctx, query, keyID).Scan(
		&k.ID, &k.TenantID, &k.Name, &k.Prefix, &k.Permissions,
		&k.ExpiresAt, &k.LastUsed, &k.CreatedAt, &k.RevokedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, types.NewNotFoundError("APIKey", keyID)
		}
		return nil, fmt.Errorf("get api key: %w", err)
	}
	return k, nil
}
