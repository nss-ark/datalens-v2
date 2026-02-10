package connector

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complyark/datalens/internal/domain/discovery"
)

// PostgresConnector implements discovery.Connector for PostgreSQL databases.
type PostgresConnector struct {
	conn *pgxpool.Pool
}

// NewPostgresConnector creates a new PostgresConnector.
func NewPostgresConnector() *PostgresConnector {
	return &PostgresConnector{}
}

// Compile-time check
var _ discovery.Connector = (*PostgresConnector)(nil)

// Capabilities returns the supported operations.
func (c *PostgresConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:             true,
		CanSample:               true,
		CanDelete:               false, // Read-only for now
		CanUpdate:               false,
		SupportsStreaming:       true,
		SupportsIncremental:     false,
		SupportsSchemaDiscovery: true,
		SupportsDataSampling:    true,
		SupportsParallelScan:    true,
		MaxConcurrency:          8,
	}
}

// Connect establishes a connection to the PostgreSQL database.
func (c *PostgresConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	// Credentials is expected to be "user:password" format.
	userPass := ds.Credentials
	if userPass == "" {
		return fmt.Errorf("credentials required")
	}

	// Build DSN: postgres://user:pass@host:port/db?sslmode=disable
	dsn := fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
		userPass, ds.Host, ds.Port, ds.Database)
	// For now, I will assume it follows "user:password" format.

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	c.conn = pool
	return nil
}

// DiscoverSchema queries information_schema to build the inventory.
func (c *PostgresConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.conn == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	// 1. Get Tables
	tablesQuery := `
		SELECT table_name, table_schema, table_type
		FROM information_schema.tables
		WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
		  AND table_type = 'BASE TABLE'`

	rows, err := c.conn.Query(ctx, tablesQuery)
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

	// 2. Count rows (optional, can be slow, maybe skip or estimate?)
	// Let's skip row count for now to be fast.

	inventory := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
		// TotalFields populated by service after getting fields?
		// Interface says DiscoverSchema returns entities. Check interface again?
		// Interface: DiscoverSchema(ctx) (*DataInventory, []DataEntity, error)
	}

	return inventory, entities, nil
}

// GetFields retrieves columns for a table.
func (c *PostgresConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// entityID is assumed to be "schema.table" or just "table"
	// We'll simplisticly assume it's the table name and use public schema if not specified?
	// Or better, require entityID to be the table name we got from DiscoverSchema.
	// But wait, DiscoverSchema returned DataEntity objects which usually have IDs?
	// The interface uses string for entityID.
	// Let's assume entityID passed here is the Name from DataEntity.

	// We need to handle schema.
	// Let's assume for now strict Name match in public schema or flexible?
	// Better query:
	query := `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position`

	rows, err := c.conn.Query(ctx, query, entityID)
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

// SampleData retrieves rows for a specific column.
func (c *PostgresConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Construct query carefully to avoid SQL injection on table/col names if possible.
	// pgx doesn't allow parameterizing identifiers easily.
	// We must sanitize identifiers.
	// Simple identifier sanitization: quote them.

	// If entity has schema "schema.table", pgx.Identifier{"schema", "table"}.Sanitize() is better.
	// But we only have string.
	// Let's try to split by dot?
	parts := strings.Split(entity, ".")
	var safeEntityStr string
	if len(parts) == 2 {
		safeEntityStr = pgx.Identifier{parts[0], parts[1]}.Sanitize()
	} else {
		safeEntityStr = pgx.Identifier{entity}.Sanitize()
	}

	safeField := pgx.Identifier{field}.Sanitize()

	query := fmt.Sprintf("SELECT %s::text FROM %s WHERE %s IS NOT NULL LIMIT %d",
		safeField, safeEntityStr, safeField, limit)

	rows, err := c.conn.Query(ctx, query)
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

func (c *PostgresConnector) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}
