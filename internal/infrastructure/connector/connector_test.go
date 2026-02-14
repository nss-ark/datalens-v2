package connector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

func TestConnectors_ConnectSafety(t *testing.T) {
	cfg := &config.Config{}
	registry := NewConnectorRegistry(cfg, nil, nil)
	supportedTypes := registry.SupportedTypes()

	for _, dsType := range supportedTypes {
		t.Run(string(dsType), func(t *testing.T) {
			conn, err := registry.GetConnector(dsType)
			require.NoError(t, err, "failed to get connector for %s", dsType)
			require.NotNil(t, conn)

			// Create a dummy DataSource
			ds := &discovery.DataSource{
				TenantEntity: types.TenantEntity{
					BaseEntity: types.BaseEntity{
						ID: types.NewID(),
					},
				},
				Name: "Test DS",
				Type: dsType,
				// Provide dummy credentials to pass basic validation if any
				Credentials: "user:pass",
				Host:        "localhost",
				Port:        1234,
				Database:    "testdb",
				Config:      "{}",
			}

			// Connect should likely return an error (connection refused, etc.) but MUST NOT panic
			assert.NotPanics(t, func() {
				err := conn.Connect(context.Background(), ds)
				// We expect an error in most cases since there is no real DB
				// But some connectors (like S3) might succeed if they just init the client without calling out immediately
				// So we don't strictly assert error, just no panic.
				if err != nil {
					t.Logf("Connect() returned error as expected: %v", err)
				}
			}, "Connect() panic for %s", dsType)
		})
	}
}
