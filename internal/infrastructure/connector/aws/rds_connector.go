package aws

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/complyark/datalens/internal/domain/discovery"
)

// RDSClientInterface defines the subset of RDS operations we need.
type RDSClientInterface interface {
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
}

// RDSConnector implements the Connector interface for Amazon RDS.
// It primarily discovers DB instances. Content discovery within databases
// should be handled by the respective database connectors (PostgreSQL/MySQL)
// by adding them as data sources.
type RDSConnector struct {
	client RDSClientInterface
	logger *slog.Logger
}

// NewRDSConnector creates a new RDS connector.
func NewRDSConnector() *RDSConnector {
	return &RDSConnector{
		logger: slog.Default().With("connector", "rds"),
	}
}

// Connect establishes a connection to AWS RDS service.
func (c *RDSConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	if c.client != nil {
		return nil
	}

	cfg, err := LoadConfig(ctx, ds)
	if err != nil {
		return err
	}

	c.client = rds.NewFromConfig(cfg)
	return nil
}

// DiscoverSchema lists RDS instances.
func (c *RDSConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.client == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	var entities []discovery.DataEntity
	paginator := rds.NewDescribeDBInstancesPaginator(c.client, &rds.DescribeDBInstancesInput{})

	count := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("describe db instances: %w", err)
		}

		for _, instance := range page.DBInstances {
			// Check creation time if strictly needed, but RDS doesn't support "ChangedSince" in the same way
			// We could filter by InstanceCreateTime if input.ChangedSince is set?
			if !input.ChangedSince.IsZero() && instance.InstanceCreateTime != nil && instance.InstanceCreateTime.Before(input.ChangedSince) {
				continue
			}

			name := aws.ToString(instance.DBInstanceIdentifier)
			// engine := aws.ToString(instance.Engine)
			// status := aws.ToString(instance.DBInstanceStatus)

			entity := discovery.DataEntity{
				Name: name,
				Type: discovery.EntityTypeDatabase,
				// Metadata removed as it's not in DataEntity struct
			}
			entities = append(entities, entity)
			count++
		}
	}

	inv := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
	}

	return inv, entities, nil
}

// GetFields is not applicable for RDS discovery (it discovers instances, not tables).
func (c *RDSConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	return []discovery.DataField{}, nil
}

// SampleData is not applicable for RDS discovery.
func (c *RDSConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	return []string{}, nil
}

func (c *RDSConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:         true,
		CanSample:           false, // Sampling happens at the DB level, not RDS level
		SupportsIncremental: false,
	}
}

// Delete is a stub for RDS.
func (c *RDSConnector) Delete(ctx context.Context, entity string, filter map[string]string) (int64, error) {
	return 0, fmt.Errorf("delete not supported for rds")
}

// Export is a stub for RDS.
func (c *RDSConnector) Export(ctx context.Context, entity string, filter map[string]string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("export not supported for rds")
}

func (c *RDSConnector) Close() error {
	return nil
}
