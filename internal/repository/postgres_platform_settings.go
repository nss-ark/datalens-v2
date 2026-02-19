package repository

import (
	"context"
	"fmt"

	"github.com/complyark/datalens/internal/domain/identity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlatformSettingsRepo struct {
	pool *pgxpool.Pool
}

func NewPlatformSettingsRepo(pool *pgxpool.Pool) *PlatformSettingsRepo {
	return &PlatformSettingsRepo{pool: pool}
}

func (r *PlatformSettingsRepo) Get(ctx context.Context, key string) (*identity.PlatformSetting, error) {
	query := `SELECT key, value, updated_at FROM platform_settings WHERE key = $1`
	var s identity.PlatformSetting
	err := r.pool.QueryRow(ctx, query, key).Scan(&s.Key, &s.Value, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("setting not found: %s", key)
		}
		return nil, fmt.Errorf("scan setting: %w", err)
	}
	return &s, nil
}

func (r *PlatformSettingsRepo) GetAll(ctx context.Context) ([]identity.PlatformSetting, error) {
	query := `SELECT key, value, updated_at FROM platform_settings`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}
	defer rows.Close()

	var settings []identity.PlatformSetting
	for rows.Next() {
		var s identity.PlatformSetting
		if err := rows.Scan(&s.Key, &s.Value, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan setting row: %w", err)
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func (r *PlatformSettingsRepo) Set(ctx context.Context, setting *identity.PlatformSetting) error {
	query := `
		INSERT INTO platform_settings (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query, setting.Key, setting.Value)
	if err != nil {
		return fmt.Errorf("exec set setting: %w", err)
	}
	return nil
}
