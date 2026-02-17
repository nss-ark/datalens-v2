package connector

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complyark/datalens/internal/domain/discovery"
)

func TestSQLServerConnector_Integration(t *testing.T) {
	// Plan said: "skip if MSSQL_DSN env var is missing"
	dsnEnv := os.Getenv("MSSQL_DSN")
	if dsnEnv == "" {
		t.Skip("Skipping integration test: MSSQL_DSN Not Set")
	}

	// Parsing DSN for test setup
	// Expected format: sqlserver://sa:StrongPass1!@localhost:1433?database=master
	// We need to extract creds to pass to Connector
	// For simplicity, we assume the DSN is valid and we just use it directly for setup
	// but the Connector expects specific fields in DataSource.

	// Helper to parse DSN string to DataSource struct?
	// Or we configure the test with specific env vars: MSSQL_HOST, MSSQL_PORT, MSSQL_USER, MSSQL_PASS
	// Let's use specific vars to make it easier to construct DataSource.

	host := os.Getenv("MSSQL_HOST")
	if host == "" {
		host = "localhost"
	}
	port := 1433
	user := os.Getenv("MSSQL_USER")
	if user == "" {
		user = "sa"
	}
	pass := os.Getenv("MSSQL_PASSWORD")
	if pass == "" {
		t.Log("MSSQL_PASSWORD not set, assuming default test password if DSN was set, but better to skip if not sure.")
	}
	dbName := os.Getenv("MSSQL_DB")
	if dbName == "" {
		dbName = "master" // Default to master to create test db
	}

	ctx := context.Background()

	// 1. Setup Data - We need to connect to create a test table
	// We use direct sql.Open to setup
	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", user, pass, host, port, dbName)
	setupDB, err := sql.Open("sqlserver", connStr)
	require.NoError(t, err)
	defer setupDB.Close()

	if err := setupDB.PingContext(ctx); err != nil {
		t.Skipf("Could not connect to MSSQL: %v", err)
	}

	tableName := "TestUsers"
	// Drop if exists
	_, err = setupDB.ExecContext(ctx, fmt.Sprintf("IF OBJECT_ID('%s', 'U') IS NOT NULL DROP TABLE %s", tableName, tableName))
	require.NoError(t, err)

	// Create Table
	_, err = setupDB.ExecContext(ctx, fmt.Sprintf(`CREATE TABLE %s (
		ID INT PRIMARY KEY,
		Name NVARCHAR(100),
		Email VARCHAR(100)
	)`, tableName))
	require.NoError(t, err)

	// Insert Data
	_, err = setupDB.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (ID, Name, Email) VALUES (1, 'John Doe', 'john@example.com')", tableName))
	require.NoError(t, err)

	defer func() {
		// Cleanup
		setupDB.ExecContext(ctx, fmt.Sprintf("DROP TABLE %s", tableName))
	}()

	// 2. Connector Setup
	connector := NewSQLServerConnector()
	ds := &discovery.DataSource{
		Host:        host,
		Port:        port,
		Database:    dbName,
		Credentials: fmt.Sprintf("%s:%s", user, pass),
	}

	// 3. Connect
	err = connector.Connect(ctx, ds)
	require.NoError(t, err)
	defer connector.Close()

	// 4. DiscoverSchema
	inv, entities, err := connector.DiscoverSchema(ctx, discovery.DiscoveryInput{})
	require.NoError(t, err)
	assert.NotNil(t, inv)

	// Find our table
	var found *discovery.DataEntity
	for i, e := range entities {
		if strings.EqualFold(e.Name, tableName) {
			found = &entities[i]
			break
		}
	}
	require.NotNil(t, found, "Table %s not found in discovery", tableName)

	// 5. GetFields
	fields, err := connector.GetFields(ctx, found.Name)
	require.NoError(t, err)

	fieldMap := make(map[string]discovery.DataField)
	for _, f := range fields {
		fieldMap[f.Name] = f
	}

	// Check Types
	// ID -> int
	// Name -> nvarchar
	// Email -> varchar
	assert.Contains(t, fieldMap, "ID")
	assert.Contains(t, strings.ToLower(fieldMap["ID"].DataType), "int")

	assert.Contains(t, fieldMap, "Name")
	assert.Contains(t, strings.ToLower(fieldMap["Name"].DataType), "nvarchar")

	assert.Contains(t, fieldMap, "Email")
	assert.Contains(t, strings.ToLower(fieldMap["Email"].DataType), "varchar")

	// 6. SampleData
	samples, err := connector.SampleData(ctx, found.Name, "Email", 5)
	require.NoError(t, err)
	assert.NotEmpty(t, samples)
	assert.Equal(t, "john@example.com", samples[0])
}
