package connector

import (
	"fmt"
	"sync"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector/aws"
	"github.com/complyark/datalens/internal/infrastructure/connector/azure"
	"github.com/complyark/datalens/internal/infrastructure/connector/m365"
	"github.com/complyark/datalens/internal/service/ai"
	"github.com/complyark/datalens/internal/service/detection"
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
func NewConnectorRegistry(cfg *config.Config, detector *detection.ComposableDetector, parser ai.ParsingService) *ConnectorRegistry {
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
	r.Register(types.DataSourceSQLServer, func() discovery.Connector {
		return NewSQLServerConnector()
	})

	// AWS Connectors
	r.Register(types.DataSourceS3, func() discovery.Connector {
		return aws.NewS3Connector()
	})
	r.Register(types.DataSourceRDS, func() discovery.Connector {
		return aws.NewRDSConnector()
	})
	r.Register(types.DataSourceDynamoDB, func() discovery.Connector {
		return aws.NewDynamoDBConnector()
	})

	// Azure Connectors
	r.Register(types.DataSourceAzureBlob, func() discovery.Connector {
		return azure.NewBlobConnector()
	})
	r.Register(types.DataSourceAzureSQL, func() discovery.Connector {
		return azure.NewAzureSQLConnector()
	})

	// Microsoft Connectors
	r.Register(types.DataSourceOutlook, func() discovery.Connector {
		return m365.NewOutlookConnector(cfg)
	})
	r.Register(types.DataSourceMicrosoft365, func() discovery.Connector {
		// Prefer the one in connector package that uses detector/filescanner
		return NewM365Connector(detector)
	})
	r.Register(types.DataSourceOneDrive, func() discovery.Connector {
		return NewM365Connector(detector)
	})

	// Google Connectors
	r.Register(types.DataSourceGoogleDrive, func() discovery.Connector {
		return NewGoogleConnector(cfg, detector)
	})
	r.Register(types.DataSourceGoogleWorkspace, func() discovery.Connector {
		return NewGoogleConnector(cfg, detector)
	})

	// File Upload Connector
	r.Register(types.DataSourceFileUpload, NewFileUploadConnectorFactory(parser, detector))

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

	// Normalize the type before lookup (safety net for pre-existing data with wrong casing)
	normalized := types.NormalizeDataSourceType(string(dsType))

	factory, ok := r.factories[normalized]
	if !ok {
		return nil, fmt.Errorf("unsupported data source type: %s (normalized: %s, registered: %v)", dsType, normalized, r.SupportedTypes())
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
