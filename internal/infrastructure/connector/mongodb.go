package connector

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/complyark/datalens/internal/domain/discovery"
)

// MongoDBConnector implements discovery.Connector for MongoDB.
type MongoDBConnector struct {
	client *mongo.Client
	dbName string
}

// NewMongoDBConnector creates a new MongoDBConnector.
func NewMongoDBConnector() *MongoDBConnector {
	return &MongoDBConnector{}
}

// Compile-time check
var _ discovery.Connector = (*MongoDBConnector)(nil)

// Capabilities returns the supported operations.
func (c *MongoDBConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:             true,
		CanSample:               true,
		CanDelete:               false,
		CanUpdate:               false,
		SupportsStreaming:       true,
		SupportsIncremental:     false, // MongoDB schema is dynamic, hard to track "modified tables" efficiently without oplog
		SupportsSchemaDiscovery: true,
		SupportsDataSampling:    true,
		SupportsParallelScan:    true,
		MaxConcurrency:          4,
	}
}

// Connect establishes a connection to the MongoDB cluster.
func (c *MongoDBConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	// Credentials logic (assuming URI or constructed from parts)
	// If ConnectionString is provided in Config, use it.
	// Otherwise build from parts.
	// The DataSource struct usually has Host, Port, Database, Credentials.
	// We'll prioritize constructing a standard URI if not explicitly provided in a "connection_string" extra field.
	// For now, let's assume standard logic: mongodb://user:pass@host:port/db

	var uri string
	if ds.Port == 0 {
		ds.Port = 27017
	}

	if ds.Credentials != "" {
		// Expecting "user:password"
		uri = fmt.Sprintf("mongodb://%s@%s:%d", ds.Credentials, ds.Host, ds.Port)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d", ds.Host, ds.Port)
	}

	clientOpts := options.Client().ApplyURI(uri)
	// Set timeout
	clientOpts.SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("ping mongo: %w", err)
	}

	c.client = client
	c.dbName = ds.Database
	if c.dbName == "" {
		return fmt.Errorf("database name required")
	}

	return nil
}

// DiscoverSchema lists collections in the database.
func (c *MongoDBConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.client == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	db := c.client.Database(c.dbName)
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, nil, fmt.Errorf("list collections: %w", err)
	}

	var entities []discovery.DataEntity
	for _, name := range collections {
		entities = append(entities, discovery.DataEntity{
			Name:   name,
			Schema: c.dbName,                  // MongoDB doesn't have "schemas" like SQL, use DB name
			Type:   discovery.EntityTypeTable, // Collection ~ Table
		})
	}

	inventory := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
	}

	return inventory, entities, nil
}

// GetFields infers fields by sampling documents from the collection.
func (c *MongoDBConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	// entityID is collection name
	coll := c.client.Database(c.dbName).Collection(entityID)

	// Sample 10 documents to infer schema
	opts := options.Find().SetLimit(10)
	cursor, err := coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("find sample: %w", err)
	}
	defer cursor.Close(ctx)

	fieldMap := make(map[string]string) // name -> type

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		flattenFields(doc, "", fieldMap)
	}

	var fields []discovery.DataField
	for name, dtype := range fieldMap {
		fields = append(fields, discovery.DataField{
			Name:     name,
			DataType: dtype,
			Nullable: true, // No strict schema
		})
	}

	// Sort for stability
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	return fields, nil
}

// flattenFields recursively walks the document and records field types using dot notation.
func flattenFields(doc bson.M, prefix string, result map[string]string) {
	for k, v := range doc {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		if v == nil {
			continue
		}

		// Handle nested objects
		if nested, ok := v.(bson.M); ok {
			flattenFields(nested, key, result)
			continue
		} else if nested, ok := v.(primitive.D); ok {
			flattenFields(nested.Map(), key, result)
			continue
		}

		// Record type
		result[key] = detectMongoType(v)
	}
}

func detectMongoType(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case string:
		return "string"
	case int, int32, int64:
		return "int"
	case float32, float64:
		return "double"
	case bool:
		return "bool"
	case primitive.DateTime, time.Time:
		return "date"
	case primitive.ObjectID:
		return "objectId"
	case primitive.Binary:
		return "binary"
	case []interface{}, primitive.A:
		return "array"
	default:
		return reflect.TypeOf(v).String()
	}
}

// SampleData retrieves values for a specific field.
func (c *MongoDBConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	coll := c.client.Database(c.dbName).Collection(entity)

	// Projection: include only the requested field (and _id:0 to exclude id)
	// Handle dot notation projection automatically supported by MongoDB
	projection := bson.D{{Key: field, Value: 1}, {Key: "_id", Value: 0}}
	opts := options.Find().SetLimit(int64(limit)).SetProjection(projection)

	// Filter where field exists and is not null
	filter := bson.D{{Key: field, Value: bson.D{{Key: "$ne", Value: nil}}}}

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find samples: %w", err)
	}
	defer cursor.Close(ctx)

	var samples []string
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		// Extract value using dot notation helper
		val := getNestedValue(doc, field)
		if val != nil {
			samples = append(samples, fmt.Sprintf("%v", val))
		}
	}

	return samples, nil
}

func getNestedValue(doc bson.M, path string) interface{} {
	parts := strings.Split(path, ".")
	current := doc
	for i, part := range parts {
		val, ok := current[part]
		if !ok {
			return nil
		}
		if i == len(parts)-1 {
			return val
		}
		if next, ok := val.(bson.M); ok {
			current = next
		} else if next, ok := val.(primitive.D); ok {
			current = next.Map()
		} else {
			return nil // Cannot traverse further
		}
	}
	return nil
}

// Close disconnects the client.
func (c *MongoDBConnector) Close() error {
	if c.client != nil {
		return c.client.Disconnect(context.Background())
	}
	return nil
}

// Delete is a stub for MongoDB.
func (c *MongoDBConnector) Delete(ctx context.Context, entity string, filter map[string]string) (int64, error) {
	return 0, fmt.Errorf("delete not supported for mongodb yet")
}

// Export is a stub for MongoDB.
func (c *MongoDBConnector) Export(ctx context.Context, entity string, filter map[string]string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("export not supported for mongodb yet")
}
