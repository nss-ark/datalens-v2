package connector

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/complyark/datalens/internal/domain/discovery"
)

func TestMongoDBConnector_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 1. Setup MongoDB (Container or CI Service)
	var uri, host string
	var port int
	var container testcontainers.Container
	var err error

	if os.Getenv("MONGODB_URL") != "" {
		uri = os.Getenv("MONGODB_URL")
		host = "localhost"
		port = 27017
	} else {
		req := testcontainers.ContainerRequest{
			Image:        "mongo:7.0",
			ExposedPorts: []string{"27017/tcp"},
			WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(60 * time.Second),
		}

		container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		require.NoError(t, err)
		defer container.Terminate(ctx)

		host, err = container.Host(ctx)
		require.NoError(t, err)

		p, err := container.MappedPort(ctx, "27017")
		require.NoError(t, err)
		port = p.Int()

		uri = fmt.Sprintf("mongodb://%s:%d", host, port)
	}

	// 2. Seed test data directly via mongo driver
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	db := client.Database("testdb")

	// Insert into "users" collection with nested address
	_, err = db.Collection("users").InsertMany(ctx, []interface{}{
		bson.M{
			"name":  "Alice",
			"email": "alice@example.com",
			"age":   30,
			"address": bson.M{
				"city": "New York",
				"zip":  "10001",
			},
		},
		bson.M{
			"name":  "Bob",
			"email": "bob@example.com",
			"age":   25,
			"address": bson.M{
				"city": "San Francisco",
				"zip":  "94102",
			},
		},
	})
	require.NoError(t, err)

	// Insert into "orders" collection
	_, err = db.Collection("orders").InsertOne(ctx, bson.M{
		"user":   "Alice",
		"amount": 99.99,
		"status": "completed",
	})
	require.NoError(t, err)

	// 3. Test Connector
	connector := NewMongoDBConnector()
	ds := &discovery.DataSource{
		Host:     host,
		Port:     port,
		Database: "testdb",
	}

	// Test Connect
	err = connector.Connect(ctx, ds)
	require.NoError(t, err)
	defer connector.Close()

	// Test Capabilities
	caps := connector.Capabilities()
	assert.True(t, caps.CanDiscover)
	assert.True(t, caps.CanSample)

	// Test DiscoverSchema
	inv, entities, err := connector.DiscoverSchema(ctx, discovery.DiscoveryInput{})
	require.NoError(t, err)
	assert.NotNil(t, inv)
	assert.Equal(t, 2, len(entities), "Should discover 'users' and 'orders' collections")

	// Verify collection names
	names := make(map[string]bool)
	for _, e := range entities {
		names[e.Name] = true
	}
	assert.True(t, names["users"])
	assert.True(t, names["orders"])

	// Test GetFields — verify nested dot-notation
	fields, err := connector.GetFields(ctx, "users")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(fields), 5, "Should have _id, name, email, age, address.city, address.zip")

	fieldMap := make(map[string]string)
	for _, f := range fields {
		fieldMap[f.Name] = f.DataType
	}

	// Check top-level fields
	assert.Equal(t, "string", fieldMap["name"])
	assert.Equal(t, "string", fieldMap["email"])
	assert.Equal(t, "int", fieldMap["age"])

	// Check nested fields (dot-notation)
	assert.Equal(t, "string", fieldMap["address.city"], "Nested fields should use dot notation")
	assert.Equal(t, "string", fieldMap["address.zip"], "Nested fields should use dot notation")

	// Test SampleData — top-level field
	samples, err := connector.SampleData(ctx, "users", "email", 10)
	require.NoError(t, err)
	assert.Len(t, samples, 2)
	assert.Contains(t, samples, "alice@example.com")
	assert.Contains(t, samples, "bob@example.com")

	// Test SampleData — nested field (dot-notation)
	citySamples, err := connector.SampleData(ctx, "users", "address.city", 10)
	require.NoError(t, err)
	assert.Len(t, citySamples, 2)
	assert.Contains(t, citySamples, "New York")
	assert.Contains(t, citySamples, "San Francisco")
}
