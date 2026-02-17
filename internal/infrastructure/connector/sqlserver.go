package connector

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/microsoft/go-mssqldb"

	"github.com/complyark/datalens/internal/domain/discovery"
)

// SQLServerConnector implements discovery.Connector for Microsoft SQL Server.
type SQLServerConnector struct {
	db *sql.DB
}

// NewSQLServerConnector creates a new SQLServerConnector.
func NewSQLServerConnector() *SQLServerConnector {
	return &SQLServerConnector{}
}

// Compile-time check
var _ discovery.Connector = (*SQLServerConnector)(nil)

// Capabilities returns the supported operations.
func (c *SQLServerConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:             true,
		CanSample:               true,
		CanDelete:               false,
		CanUpdate:               false,
		SupportsStreaming:       false,
		SupportsIncremental:     false,
		SupportsSchemaDiscovery: true,
		SupportsDataSampling:    true,
		SupportsParallelScan:    true,
		MaxConcurrency:          4,
	}
}

// Connect establishes a connection to the SQL Server database.
func (c *SQLServerConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	// Credentials is expected to be "user:password" format.
	if ds.Credentials == "" {
		return fmt.Errorf("credentials required")
	}

	// Host format: "host" or "host:port"
	// go-mssqldb expects "sqlserver://username:password@host:port?database=dbname"

	u := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword("", ""), // Will be set below
		Host:   fmt.Sprintf("%s:%d", ds.Host, ds.Port),
	}

	// Handle credentials
	parts := strings.SplitN(ds.Credentials, ":", 2)
	if len(parts) == 2 {
		u.User = url.UserPassword(parts[0], parts[1])
	} else {
		return fmt.Errorf("invalid credentials format, expected user:password")
	}

	q := u.Query()
	q.Set("database", ds.Database)
	// Add other useful defaults
	q.Set("encrypt", "disable") // Default to disable for easier local dev, or make configurable?
	// Existing postgres connector sets sslmode=disable.

	u.RawQuery = q.Encode()

	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return fmt.Errorf("open sqlserver: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("ping sqlserver: %w", err)
	}

	c.db = db
	return nil
}

// DiscoverSchema queries information_schema to build the inventory.
func (c *SQLServerConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.db == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	// SQL Server uses schema concept heavily.
	query := `
		SELECT TABLE_SCHEMA, TABLE_NAME, TABLE_TYPE
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_TYPE = 'BASE TABLE'
		AND TABLE_SCHEMA NOT IN ('sys', 'INFORMATION_SCHEMA')
	`
	// Note: SQL Server doesn't have "row" based changed_since easily accessible without CDC enabled.
	// sys.dm_db_index_usage_stats is for usage, not modification time.
	// For now, ignoring ChangedSince as it's not reliably available on standard tables.

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()

	var entities []discovery.DataEntity
	for rows.Next() {
		var schema, name, kind string
		if err := rows.Scan(&schema, &name, &kind); err != nil {
			return nil, nil, err
		}

		entities = append(entities, discovery.DataEntity{
			Name:   name,
			Schema: schema,
			Type:   discovery.EntityTypeTable,
		})
	}

	inventory := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
	}

	return inventory, entities, nil
}

// GetFields retrieves columns for a table.
func (c *SQLServerConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	// We assume entityID is table name.
	// If schema is needed, we might need to change how we pass entityID.
	// For now, assuming entityID is just table name and we search across schemas or current schema.
	// Better to match how we discovered it.
	// DiscoverSchema returns Name and Schema.
	// But GetFields takes only entityID string.
	// If multiple schemas have same table name, this is ambiguous.
	// However, interface limitation.
	// Let's assume entityID might be "schema.table" or we query by table name.

	query := `
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_NAME = @p1
		ORDER BY ORDINAL_POSITION`

	rows, err := c.db.QueryContext(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("query columns: %w", err)
	}
	defer rows.Close()

	var fields []discovery.DataField
	for rows.Next() {
		var name, dtype, nullableStr string
		if err := rows.Scan(&name, &dtype, &nullableStr); err != nil {
			return nil, err
		}

		fields = append(fields, discovery.DataField{
			Name:     name,
			DataType: dtype,
			Nullable: nullableStr == "YES",
		})
	}

	return fields, nil
}

// SampleData retrieves sample values from a specific entity/field.
func (c *SQLServerConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Quote identifiers for SQL Server
	safeEntity := quoteSQLServer(entity)
	safeField := quoteSQLServer(field)

	// SQL Server uses TOP instead of LIMIT
	query := fmt.Sprintf("SELECT TOP %d CAST(%s AS NVARCHAR(MAX)) FROM %s WHERE %s IS NOT NULL",
		limit, safeField, safeEntity, safeField)

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query samples: %w", err)
	}
	defer rows.Close()

	var samples []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		samples = append(samples, s)
	}

	return samples, nil
}

// Close releases the connection.
func (c *SQLServerConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Delete is a stub.
func (c *SQLServerConnector) Delete(ctx context.Context, entity string, filter map[string]string) (int64, error) {
	return 0, fmt.Errorf("delete not supported for sqlserver yet")
}

// Export is a stub.
func (c *SQLServerConnector) Export(ctx context.Context, entity string, filter map[string]string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("export not supported for sqlserver yet")
}

// Helpers

func quoteSQLServer(identifier string) string {
	// Basic bracket quoting
	return "[" + strings.ReplaceAll(identifier, "]", "]]") + "]"
}
