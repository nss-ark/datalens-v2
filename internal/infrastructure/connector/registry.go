package connector

import (
	"fmt"
	"sync"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// ConnectorFactory is a constructor function that returns a new Connector instance.
type ConnectorFactory func() discovery.Connector

// ConnectorRegistry maps data source types to their connector constructors.
// It is the single lookup point used by DiscoveryService to resolve the
// right connector for a given DataSource.
type ConnectorRegistry struct {
	mu        sync.RWMutex
	factories map[types.DataSourceType]ConnectorFactory
}

// NewConnectorRegistry creates a registry pre-loaded with built-in connectors.
func NewConnectorRegistry() *ConnectorRegistry {
	r := &ConnectorRegistry{
		factories: make(map[types.DataSourceType]ConnectorFactory),
	}

	// Register built-in connectors
	r.Register(types.DataSourcePostgreSQL, func() discovery.Connector {
		return NewPostgresConnector()
	})
	r.Register(types.DataSourceMySQL, func() discovery.Connector {
		return NewMySQLConnector()
	})
	r.Register(types.DataSourceMongoDB, func() discovery.Connector {
		return NewMongoDBConnector()
	})
	// TODO: Implement S3 connector
	r.Register(types.DataSourceS3, func() discovery.Connector {
		return NewS3Connector()
	})

	return r
}

// Register adds a connector factory for a data source type.
// Overwrites any existing factory for the same type.
func (r *ConnectorRegistry) Register(dsType types.DataSourceType, factory ConnectorFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[dsType] = factory
}

// GetConnector returns a new Connector for the given data source type.
func (r *ConnectorRegistry) GetConnector(dsType types.DataSourceType) (discovery.Connector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[dsType]
	if !ok {
		return nil, fmt.Errorf("unsupported data source type: %s (registered: %v)", dsType, r.SupportedTypes())
	}
	return factory(), nil
}

// SupportedTypes returns a slice of all registered data source types.
func (r *ConnectorRegistry) SupportedTypes() []types.DataSourceType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]types.DataSourceType, 0, len(r.factories))
	for t := range r.factories {
		result = append(result, t)
	}
	return result
}
