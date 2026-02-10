// Package connector provides the connector registry that manages
// all available data source connectors.
package connector

import (
	"fmt"
	"sync"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/pkg/types"
)

// Registry manages available data source connectors.
// Connectors register themselves at startup, and the registry
// provides the correct connector for a given data source type.
type Registry struct {
	mu        sync.RWMutex
	factories map[types.DataSourceType]ConnectorFactory
}

// ConnectorFactory creates a new connector instance for a data source.
type ConnectorFactory func() discovery.Connector

// NewRegistry creates an empty connector registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[types.DataSourceType]ConnectorFactory),
	}
}

// Register adds a connector factory for a specific data source type.
func (r *Registry) Register(dsType types.DataSourceType, factory ConnectorFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[dsType] = factory
}

// Get returns a new connector instance for the given data source type.
func (r *Registry) Get(dsType types.DataSourceType) (discovery.Connector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[dsType]
	if !ok {
		return nil, fmt.Errorf("no connector registered for type: %s", dsType)
	}

	return factory(), nil
}

// SupportedTypes returns all registered data source types.
func (r *Registry) SupportedTypes() []types.DataSourceType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	supported := make([]types.DataSourceType, 0, len(r.factories))
	for t := range r.factories {
		supported = append(supported, t)
	}
	return supported
}
