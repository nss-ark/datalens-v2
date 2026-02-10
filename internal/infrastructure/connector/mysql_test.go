package connector

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/complyark/datalens/internal/domain/discovery"
)

func TestMySQLConnector_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 1. Start MySQL Container
	req := testcontainers.ContainerRequest{
		Image:        "mysql:8",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "testdb",
		},
		WaitingFor: wait.ForLog("port: 3306  MySQL Community Server").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)

	// 2. Connector Setup
	connector := NewMySQLConnector()
	ds := &discovery.DataSource{
		Host:        host,
		Port:        port.Int(),
		Database:    "testdb",
		Credentials: "root:password",
	}

	// 3. Connect
	err = connector.Connect(ctx, ds)
	require.NoError(t, err)
	defer connector.Close()

	// 4. Seed Data
	db, err := sql.Open("mysql", fmt.Sprintf("root:password@tcp(%s:%d)/testdb", host, port.Int()))
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100), email VARCHAR(100))`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO users (id, name, email) VALUES (1, 'John', 'john@example.com')`)
	require.NoError(t, err)

	// 5. DiscoverSchema
	inv, entities, err := connector.DiscoverSchema(ctx, discovery.DiscoveryInput{})
	require.NoError(t, err)
	assert.NotNil(t, inv)
	assert.Len(t, entities, 1)
	assert.Equal(t, "users", entities[0].Name)

	// 6. GetFields
	fields, err := connector.GetFields(ctx, "users")
	require.NoError(t, err)
	assert.Len(t, fields, 3) // id, name, email

	// Verify Field Types
	fieldMap := make(map[string]discovery.DataField)
	for _, f := range fields {
		fieldMap[f.Name] = f
	}
	assert.Equal(t, "int", fieldMap["id"].DataType)
	assert.Equal(t, "varchar", fieldMap["name"].DataType)

	// 7. SampleData
	samples, err := connector.SampleData(ctx, "users", "email", 10)
	require.NoError(t, err)
	assert.Len(t, samples, 1)
	assert.Equal(t, "john@example.com", samples[0])
}
