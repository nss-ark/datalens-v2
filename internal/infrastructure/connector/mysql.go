package connector

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql" // MySQL driver

	"github.com/complyark/datalens/internal/domain/discovery"
)

// MySQLConnector implements discovery.Connector for MySQL databases.
type MySQLConnector struct {
	db *sql.DB
}

// NewMySQLConnector creates a new MySQLConnector.
func NewMySQLConnector() *MySQLConnector {
	return &MySQLConnector{}
}

// Compile-time check
var _ discovery.Connector = (*MySQLConnector)(nil)

// Capabilities returns the supported operations for MySQL.
func (c *MySQLConnector) Capabilities() discovery.ConnectorCapabilities {
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

// Connect establishes a connection to the MySQL database.
func (c *MySQLConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	// Credentials is expected to be "user:password" format.
	credentials := ds.Credentials
	if credentials == "" {
		return fmt.Errorf("credentials required")
	}

	// Build DSN: user:password@tcp(host:port)/dbname?parseTime=true
	dsn := fmt.Sprintf("%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		credentials, ds.Host, ds.Port, ds.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("ping mysql: %w", err)
	}

	c.db = db
	return nil
}

// DiscoverSchema queries information_schema to build the inventory.
func (c *MySQLConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.db == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	// Get tables from the connected database (DATABASE() returns current DB)
	query := `
		SELECT TABLE_NAME, TABLE_SCHEMA, TABLE_TYPE
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_TYPE = 'BASE TABLE'`

	var args []interface{}
	if !input.ChangedSince.IsZero() {
		query += " AND (UPDATE_TIME > ? OR CREATE_TIME > ?)"
		args = append(args, input.ChangedSince, input.ChangedSince)
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()

	var entities []discovery.DataEntity
	for rows.Next() {
		var name, schema, kind string
		if err := rows.Scan(&name, &schema, &kind); err != nil {
			return nil, nil, err
		}

		entities = append(entities, discovery.DataEntity{
			Name:   name,
			Schema: schema,
			Type:   discovery.EntityTypeTable,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	inventory := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
	}

	return inventory, entities, nil
}

// GetFields retrieves columns for a table.
func (c *MySQLConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := `
		SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := c.db.QueryContext(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("query columns: %w", err)
	}
	defer rows.Close()

	var fields []discovery.DataField
	for rows.Next() {
		var name, dtype, nullableStr, columnKey string
		if err := rows.Scan(&name, &dtype, &nullableStr, &columnKey); err != nil {
			return nil, err
		}

		fields = append(fields, discovery.DataField{
			Name:         name,
			DataType:     normalizeMySQLType(dtype),
			Nullable:     nullableStr == "YES",
			IsPrimaryKey: columnKey == "PRI",
			IsForeignKey: columnKey == "MUL",
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fields, nil
}

// SampleData retrieves sample values from a specific entity/field.
func (c *MySQLConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Use backtick quoting for MySQL identifiers
	safeEntity := quoteMySQL(entity)
	safeField := quoteMySQL(field)

	query := fmt.Sprintf("SELECT CAST(%s AS CHAR) FROM %s WHERE %s IS NOT NULL LIMIT %d",
		safeField, safeEntity, safeField, limit)

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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return samples, nil
}

// Close releases the MySQL connection.
func (c *MySQLConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// =============================================================================
// Helpers
// =============================================================================

// quoteMySQL wraps an identifier in backticks and escapes internal backticks.
func quoteMySQL(identifier string) string {
	escaped := strings.ReplaceAll(identifier, "`", "``")
	return "`" + escaped + "`"
}

// normalizeMySQLType maps verbose MySQL column types to simpler names
// for consistency with the PII detection engine.
func normalizeMySQLType(mysqlType string) string {
	lower := strings.ToLower(mysqlType)

	// Strip length/precision specifiers for cleaner type names
	// e.g., "varchar(255)" → "varchar", "int(11)" → "int"
	if idx := strings.Index(lower, "("); idx != -1 {
		base := lower[:idx]
		switch base {
		case "enum", "set":
			return mysqlType // Keep ENUM/SET values for detection context
		default:
			return base
		}
	}

	return lower
}
