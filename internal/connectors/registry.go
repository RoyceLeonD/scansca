package connectors

import (
	"context"

	"github.com/royceleond/antska/internal/connectors/postgresql"
)

// PostgresConnectorAdapter adapts the PostgreSQL connector to the Connector interface
type PostgresConnectorAdapter struct {
	pg *postgresql.PostgresConnector
}

// NewPostgresConnector creates a new PostgreSQL connector adapter
func NewPostgresConnector() Connector {
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
		Columns: pgResult.Columns,
		Rows:    pgResult.Rows,
		Error:   pgResult.Error,
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

// Additional connector creation functions will be added here as they are implemented:
// func NewMySQLConnector() Connector { ... }
// func NewSQLiteConnector() Connector { ... }
// func NewDynamoDBConnector() Connector { ... }