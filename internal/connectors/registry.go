package connectors

import (
	"context"

	"github.com/royceleond/antska/internal/connectors/postgresql"
)

// PostgresConnectorAdapter adapts the PostgreSQL connector to the Connector interfaces
type PostgresConnectorAdapter struct {
	pg *postgresql.PostgresConnector
}

// NewPostgresConnector creates a new PostgreSQL connector adapter
func NewPostgresConnector() Connector {
	return &PostgresConnectorAdapter{
		pg: postgresql.NewPostgresConnector(),
	}
}

// NewAdvancedPostgresConnector creates a new PostgreSQL connector adapter with the advanced interface
func NewAdvancedPostgresConnector() AdvancedConnector {
	return &PostgresConnectorAdapter{
		pg: postgresql.NewPostgresConnector(),
	}
}

// Connect adapts the PostgresConnector.Connect method
func (a *PostgresConnectorAdapter) Connect(ctx context.Context, connectionString string) error {
	return a.pg.Connect(ctx, connectionString)
}

// Disconnect adapts the PostgresConnector.Disconnect method
func (a *PostgresConnectorAdapter) Disconnect(ctx context.Context) error {
	return a.pg.Disconnect(ctx)
}

// ListSchemas adapts the PostgresConnector.ListSchemas method
func (a *PostgresConnectorAdapter) ListSchemas(ctx context.Context) ([]string, error) {
	return a.pg.ListSchemas(ctx)
}

// ListTables adapts the PostgresConnector.ListTables method
func (a *PostgresConnectorAdapter) ListTables(ctx context.Context, schema string) ([]string, error) {
	return a.pg.ListTables(ctx, schema)
}

// GetTableColumns adapts the PostgresConnector.GetTableColumns method
func (a *PostgresConnectorAdapter) GetTableColumns(ctx context.Context, schema, table string) (map[string]string, error) {
	return a.pg.GetTableColumns(ctx, schema, table)
}

// ExecuteQuery adapts the PostgresConnector.ExecuteQuery method
func (a *PostgresConnectorAdapter) ExecuteQuery(ctx context.Context, query string) (ResultSet, error) {
	pgResult, err := a.pg.ExecuteQuery(ctx, query)
	
	// Convert the postgresql.ResultSet to connectors.ResultSet
	return ResultSet{
		Columns:      pgResult.Columns,
		Rows:         pgResult.Rows,
		Error:        pgResult.Error,
		AffectedRows: pgResult.AffectedRows,
		ExecutionTime: pgResult.ExecutionTime,
	}, err
}

// GetConnectorType adapts the PostgresConnector.GetConnectorType method
func (a *PostgresConnectorAdapter) GetConnectorType() string {
	return a.pg.GetConnectorType()
}

// Ping adapts the PostgresConnector.Ping method
func (a *PostgresConnectorAdapter) Ping(ctx context.Context) error {
	return a.pg.Ping(ctx)
}

// AdvancedConnector Interface Methods

// Reconnect implements the AdvancedConnector.Reconnect method
func (a *PostgresConnectorAdapter) Reconnect(ctx context.Context) error {
	return a.pg.Reconnect(ctx)
}

// GetStatus implements the AdvancedConnector.GetStatus method
func (a *PostgresConnectorAdapter) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	return a.pg.GetStatus(ctx)
}

// ListViews implements the AdvancedConnector.ListViews method
func (a *PostgresConnectorAdapter) ListViews(ctx context.Context, schema string) ([]string, error) {
	return a.pg.ListViews(ctx, schema)
}

// GetTableInfo implements the AdvancedConnector.GetTableInfo method
func (a *PostgresConnectorAdapter) GetTableInfo(ctx context.Context, schema, table string) (*TableInfo, error) {
	pgTableInfo, err := a.pg.GetTableInfo(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	
	// Convert PostgreSQL TableInfo to our interface TableInfo
	columns := make([]ColumnInfo, len(pgTableInfo.Columns))
	for i, col := range pgTableInfo.Columns {
		columns[i] = ColumnInfo{
			Name:            col.Name,
			DataType:        col.DataType,
			IsNullable:      col.IsNullable,
			DefaultValue:    col.DefaultValue,
			IsPrimaryKey:    col.IsPrimaryKey,
			IsForeignKey:    col.IsForeignKey,
			ReferencesTable: col.ReferencesTable,
			ReferencesColumn: col.ReferencesColumn,
		}
	}
	
	return &TableInfo{
		Schema:           pgTableInfo.Schema,
		Name:             pgTableInfo.Name,
		Columns:          columns,
		PrimaryKey:       pgTableInfo.PrimaryKey,
		EstimatedRowCount: pgTableInfo.EstimatedRowCount,
		CreateTime:       pgTableInfo.CreateTime,
		Description:      pgTableInfo.Description,
	}, nil
}

// ExecuteStatement implements the AdvancedConnector.ExecuteStatement method
func (a *PostgresConnectorAdapter) ExecuteStatement(ctx context.Context, statement string, args ...interface{}) (int64, error) {
	return a.pg.ExecuteStatement(ctx, statement, args...)
}

// ExecuteTransaction implements the AdvancedConnector.ExecuteTransaction method
func (a *PostgresConnectorAdapter) ExecuteTransaction(ctx context.Context, stmts []string) error {
	return a.pg.ExecuteTransaction(ctx, stmts)
}

// ExecuteBatch implements the AdvancedConnector.ExecuteBatch method
func (a *PostgresConnectorAdapter) ExecuteBatch(ctx context.Context, stmts []string) error {
	return a.pg.ExecuteBatch(ctx, stmts)
}

// Additional connector creation functions will be added here as they are implemented:
// func NewMySQLConnector() Connector { ... }
// func NewSQLiteConnector() Connector { ... }
// func NewDynamoDBConnector() Connector { ... }