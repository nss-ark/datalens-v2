package azure

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/complyark/datalens/internal/domain/discovery"
	"github.com/complyark/datalens/internal/infrastructure/connector/shared"
	_ "github.com/microsoft/go-mssqldb" // SQL Server driver
)

// AzureSQLConnector differs from generic SQL Server in connection metadata and perhaps auth.
type AzureSQLConnector struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewAzureSQLConnector() *AzureSQLConnector {
	return &AzureSQLConnector{
		logger: slog.Default().With("connector", "azure_sql"),
	}
}

func (c *AzureSQLConnector) Connect(ctx context.Context, ds *discovery.DataSource) error {
	if c.db != nil {
		return nil
	}

	// ds.Credentials should contain connection string or components as JSON
	creds, err := shared.ParseCredentials(ds.Credentials)
	if err != nil {
		return fmt.Errorf("parse credentials: %w", err)
	}

	connStr, ok := creds["connection_string"].(string)
	if !ok || connStr == "" {
		return fmt.Errorf("credentials (connection_string) required")
	}

	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return fmt.Errorf("open connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	c.db = db
	return nil
}

func (c *AzureSQLConnector) DiscoverSchema(ctx context.Context, input discovery.DiscoveryInput) (*discovery.DataInventory, []discovery.DataEntity, error) {
	if c.db == nil {
		return nil, nil, fmt.Errorf("not connected")
	}

	// Query Information Schema
	query := `
		SELECT TABLE_SCHEMA, TABLE_NAME, TABLE_TYPE 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_TYPE IN ('BASE TABLE', 'VIEW')
	`
	// If input.ChangedSince is set, we might check sys.objects.modify_date?
	// SQL Server has create_date and modify_date in sys.objects.

	if !input.ChangedSince.IsZero() {
		// T-SQL syntax for date comparison
		// We'd need to join with sys.objects or sys.tables
		// For simplicity, we'll scan all for now or improved query:
		// SELECT t.TABLE_SCHEMA, t.TABLE_NAME, t.TABLE_TYPE FROM INFORMATION_SCHEMA.TABLES t JOIN sys.objects o ON t.TABLE_NAME = o.name AND SCHEMA_NAME(o.schema_id) = t.TABLE_SCHEMA WHERE o.modify_date > @p1
	}

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("query schema: %w", err)
	}
	defer rows.Close()

	var entities []discovery.DataEntity
	for rows.Next() {
		var schema, name, tableType string
		if err := rows.Scan(&schema, &name, &tableType); err != nil {
			return nil, nil, fmt.Errorf("scan row: %w", err)
		}

		fullTableName := fmt.Sprintf("%s.%s", schema, name)
		entity := discovery.DataEntity{
			Name:   fullTableName,
			Type:   discovery.EntityTypeTable,
			Schema: schema,
		}
		if tableType == "VIEW" {
			entity.Type = discovery.EntityTypeView
		}
		entities = append(entities, entity)
	}

	inv := &discovery.DataInventory{
		TotalEntities: len(entities),
		SchemaVersion: "1.0",
	}

	return inv, entities, nil
}

func (c *AzureSQLConnector) GetFields(ctx context.Context, entityID string) ([]discovery.DataField, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Parse schema.table from entityID
	parts := splitSchemaTable(entityID)
	schema := parts[0]
	table := parts[1]

	query := `
        SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = @p1 AND TABLE_NAME = @p2
        ORDER BY ORDINAL_POSITION
    `
	rows, err := c.db.QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("query columns: %w", err)
	}
	defer rows.Close()

	var fields []discovery.DataField
	for rows.Next() {
		var colName, dataType, isNullable string
		if err := rows.Scan(&colName, &dataType, &isNullable); err != nil {
			return nil, err
		}

		fields = append(fields, discovery.DataField{
			Name:     colName,
			DataType: dataType,
			Nullable: isNullable == "YES",
		})
	}

	return fields, nil
}

func (c *AzureSQLConnector) SampleData(ctx context.Context, entity, field string, limit int) ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := fmt.Sprintf("SELECT TOP %d %s FROM %s", limit, field, entity)
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("sample data: %w", err)
	}
	defer rows.Close()

	var samples []string
	for rows.Next() {
		var val interface{}
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		if val != nil {
			samples = append(samples, fmt.Sprintf("%v", val))
		}
	}

	return samples, nil
}

func splitSchemaTable(full string) []string {
	// simplistic split, assumes schema.table format with no dots in names
	// ideally handle [schema].[table] quoting
	parts := func(s string) []string {
		// TODO: robust parsing
		p := make([]string, 2)
		idx := -1
		// find last dot?
		for i := len(s) - 1; i >= 0; i-- {
			if s[i] == '.' {
				idx = i
				break
			}
		}
		if idx != -1 {
			p[0] = s[:idx]
			p[1] = s[idx+1:]
		} else {
			p[0] = "dbo"
			p[1] = s
		}
		return p
	}(full)
	return parts
}

func (c *AzureSQLConnector) Capabilities() discovery.ConnectorCapabilities {
	return discovery.ConnectorCapabilities{
		CanDiscover:             true,
		CanSample:               true,
		SupportsSchemaDiscovery: true,
		SupportsDataSampling:    true,
	}
}

func (c *AzureSQLConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
