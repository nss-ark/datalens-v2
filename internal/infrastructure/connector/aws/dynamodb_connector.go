package aws

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/complyark/datalens/internal/domain/discovery"
)

// DynamoDBClientInterface defines the subset of DynamoDB operations we need.
type DynamoDBClientInterface interface {
	ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// DynamoDBConnector implements the Connector interface for Amazon DynamoDB.
type DynamoDBConnector struct {
	client DynamoDBClientInterface
	logger *slog.Logger
}

// NewDynamoDBConnector creates a new DynamoDB connector.
func NewDynamoDBConnector() *DynamoDBConnector {
	return &DynamoDBConnector{
		logger: slog.Default().With("connector", "dynamodb"),
	}
}

// Connect establishes a connection to DynamoDB.
func (c *DynamoDBConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	if c.client != nil {
		return nil
	}

	cfg, err := LoadConfig(ctx, ds)
	if err != nil {
		return err
	}

	c.client = dynamodb.NewFromConfig(cfg)
	return nil
}

// DiscoverSchema lists DynamoDB tables.
func (c *DynamoDBConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.client == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	var entities []discovery.DataEntity
	paginator := dynamodb.NewListTablesPaginator(c.client, &dynamodb.ListTablesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list tables: %w", err)
		}

		for _, tableName := range page.TableNames {
			// Describe table to get more info (like item count, creation time) if needed
			// Note: This might be slow for many tables. For now, we just list names.
			// If we need creation time filtering, we must DescribeTable.

			// Let's do a lightweight DescribeTable if ChangedSince is set, or just return basic info.
			// For minimal calls, we assume all tables if no filtering.
			entity := discovery.DataEntity{
				Name: tableName,
				Type: discovery.EntityTypeTable,
			}

			// Optimization: only describe if we really need metadata or filtering
			if !input.ChangedSince.IsZero() {
				desc, err := c.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(tableName)})
				if err == nil && desc.Table != nil {
					if desc.Table.CreationDateTime != nil && desc.Table.CreationDateTime.Before(input.ChangedSince) {
						continue
					}
					entity.RowCount = desc.Table.ItemCount // Approximate
				}
			}

			entities = append(entities, entity)
		}
	}

	inv := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
	}

	return inv, entities, nil
}

// GetFields returns the attributes of a table.
// DynamoDB is schemaless, but we can look at KeySchema and maybe sample some items to find common attributes.
func (c *DynamoDBConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	desc, err := c.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(entityID)})
	if err != nil {
		return nil, fmt.Errorf("describe table: %w", err)
	}

	var fields []discovery.DataField

	// Add Key Schema fields
	for _, key := range desc.Table.KeySchema {
		fieldName := aws.ToString(key.AttributeName)
		fields = append(fields, discovery.DataField{
			Name:         fieldName,
			DataType:     "KEY", // We'd need to lookup AttributeDefinitions to get type (S/N/B)
			IsPrimaryKey: key.KeyType == types.KeyTypeHash || key.KeyType == types.KeyTypeRange,
			Nullable:     false,
		})
	}

	// Add other attributes from AttributeDefinitions
	for _, attr := range desc.Table.AttributeDefinitions {
		name := aws.ToString(attr.AttributeName)
		found := false
		for i := range fields {
			if fields[i].Name == name {
				// Update type
				fields[i].DataType = string(attr.AttributeType)
				found = true
				break
			}
		}
		if !found {
			fields = append(fields, discovery.DataField{
				Name:     name,
				DataType: string(attr.AttributeType),
			})
		}
	}

	return fields, nil
}

// SampleData scans a few items to get sample values.
func (c *DynamoDBConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Scan with limit
	out, err := c.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(entity),
		Limit:     aws.Int32(int32(limit)),
	})
	if err != nil {
		return nil, fmt.Errorf("scan table: %w", err)
	}

	var samples []string
	for _, item := range out.Items {
		if val, ok := item[field]; ok {
			// Extract string representation
			strVal := getAttributeValueString(val)
			if strVal != "" {
				samples = append(samples, strVal)
			}
		}
	}

	return samples, nil
}

func getAttributeValueString(av types.AttributeValue) string {
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		return v.Value
	case *types.AttributeValueMemberN:
		return v.Value
	case *types.AttributeValueMemberB:
		return fmt.Sprintf("%x", v.Value)
	case *types.AttributeValueMemberBOOL:
		return fmt.Sprintf("%v", v.Value)
	case *types.AttributeValueMemberNULL:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (c *DynamoDBConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:         true,
		CanSample:           true,
		SupportsIncremental: false, // Could support with Streams, but basic impl no
	}
}

func (c *DynamoDBConnector) Close() error {
	return nil
}
