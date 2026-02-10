package connector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// MockConnector
type MockConnector struct {
	Type types.DataSourceType
	mock.Mock
}

func (m *MockConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	args := m.Called(ctx, ds)
	return args.Error(0)
}

func (m *MockConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*discovery.DataInventory), args.Get(1).([]discovery.DataEntity), args.Error(2)
}

func (m *MockConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	args := m.Called(ctx, entityID)
	return args.Get(0).([]discovery.DataField), args.Error(1)
}

func (m *MockConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	args := m.Called(ctx, entity, field, limit)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{CanDiscover: true, CanSample: true}
}

func (m *MockConnector) Close() error {
	return m.Called().Error(0)
}

func TestConnectorRegistry_RegisterAndGet(t *testing.T) {
	registry := NewConnectorRegistry()

	// Verify built-in connectors are registered
	pgConn, err := registry.GetConnector(types.DataSourcePostgreSQL)
	require.NoError(t, err)
	assert.NotNil(t, pgConn)

	mysqlConn, err := registry.GetConnector(types.DataSourceMySQL)
	require.NoError(t, err)
	assert.NotNil(t, mysqlConn)
}

func TestConnectorRegistry_UnknownType(t *testing.T) {
	registry := NewConnectorRegistry()

	_, err := registry.GetConnector("UNKNOWN_TYPE")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported data source type")
}

func TestConnectorRegistry_SupportedTypes(t *testing.T) {
	registry := NewConnectorRegistry()

	supportedTypes := registry.SupportedTypes()
	assert.Greater(t, len(supportedTypes), 0)
	assert.Contains(t, supportedTypes, types.DataSourcePostgreSQL)
	assert.Contains(t, supportedTypes, types.DataSourceMySQL)
}
