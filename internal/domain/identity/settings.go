package identity

import (
	"context"
	"encoding/json"
	"time"
)

// PlatformSetting represents a key-value configuration for the platform.
type PlatformSetting struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// PlatformSettingsRepository defines the interface for managing platform settings.
type PlatformSettingsRepository interface {
	Get(ctx context.Context, key string) (*PlatformSetting, error)
	Set(ctx context.Context, setting *PlatformSetting) error
	GetAll(ctx context.Context) ([]PlatformSetting, error)
}
