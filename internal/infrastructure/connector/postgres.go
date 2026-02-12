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
// Capabilities returns the supported operations.
func (c *PostgresConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:             true,
		CanSample:               true,
		CanDelete:               true,
		CanUpdate:               false,
		CanExport:               true,
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

// Delete deletes entities matching the filter.
func (c *PostgresConnector) Delete(ctx context.Context, entity string, filter map[string]string) (int64, error) {
	if c.conn == nil {
		return 0, fmt.Errorf("not connected")
	}

	if len(filter) == 0 {
		return 0, fmt.Errorf("refusing to delete with empty filter")
	}

	// 1. Sanitize Table Name
	// We assume entity is "schema.table" or just "table"
	parts := strings.Split(entity, ".")
	var safeTable string
	if len(parts) == 2 {
		safeTable = pgx.Identifier{parts[0], parts[1]}.Sanitize()
	} else {
		safeTable = pgx.Identifier{entity}.Sanitize()
	}

	// 2. Build WHERE clause dynamically
	var conditions []string
	var args []interface{}
	argIdx := 1

	// Sort keys for deterministic order (good for testing/logging, not strictly needed for SQL)
	// But map iteration is random.
	// For now, simple iteration.
	for col, val := range filter {
		safeCol := pgx.Identifier{col}.Sanitize()
		conditions = append(conditions, fmt.Sprintf("%s = $%d", safeCol, argIdx))
		args = append(args, val)
		argIdx++
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", safeTable, strings.Join(conditions, " AND "))

	// 3. Execute
	tag, err := c.conn.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("delete failed: %w", err)
	}

	return tag.RowsAffected(), nil
}

// Export retrieves all data for entities matching the filter.
func (c *PostgresConnector) Export(ctx context.Context, entity string, filter map[string]string) ([]map[string]interface{}, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// 1. Sanitize Table Name
	parts := strings.Split(entity, ".")
	var safeTable string
	if len(parts) == 2 {
		safeTable = pgx.Identifier{parts[0], parts[1]}.Sanitize()
	} else {
		safeTable = pgx.Identifier{entity}.Sanitize()
	}

	// 2. Build WHERE clause dynamically
	var conditions []string
	var args []interface{}
	argIdx := 1

	for col, val := range filter {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", pgx.Identifier{col}.Sanitize(), argIdx))
		args = append(args, val)
		argIdx++
	}

	var query string
	if len(conditions) > 0 {
		query = fmt.Sprintf("SELECT * FROM %s WHERE %s", safeTable, strings.Join(conditions, " AND "))
	} else {
		// If no filter, return everything? Or error? Connector interface implies filter is for narrowing.
		// DSR access usually means "exported data for THIS user".
		// But if filter is empty, maybe it means all data?
		// Let's allow empty filter for now, but in practice DSR executor should provide one.
		query = fmt.Sprintf("SELECT * FROM %s", safeTable)
	}

	// 3. Execute
	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("export query failed: %w", err)
	}
	defer rows.Close()

	// 4. Parse results into maps
	var results []map[string]interface{}
	fields := rows.FieldDescriptions()
	columnNames := make([]string, len(fields))
	for i, fd := range fields {
		columnNames[i] = string(fd.Name)
	}

	for rows.Next() {
		// Create a slice of interface{} to hold values
		values := make([]interface{}, len(columnNames))
		valuePtrs := make([]interface{}, len(columnNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		// Convert to map
		rowMap := make(map[string]interface{})
		for i, col := range columnNames {
			val := values[i]
			// Handle []byte for text? pgx handles many types.
			// DSR export usually creates JSON, so we rely on json.Marshal handling these types.
			rowMap[col] = val
		}
		results = append(results, rowMap)
	}

	return results, nil
}
