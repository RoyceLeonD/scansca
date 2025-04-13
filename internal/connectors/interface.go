package connectors

import (
	"context"
)

// ResultSet represents a database query result
type ResultSet struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Error   string          `json:"error,omitempty"`
}

// DatabaseConfig holds configuration for a database connection
type DatabaseConfig struct {
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	ConnectionString string                 `json:"connection_string"`
	Options          map[string]interface{} `json:"options,omitempty"`
}

// Connector defines the interface for database connectors
type Connector interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context, connectionString string) error

	// Disconnect closes the database connection
	Disconnect(ctx context.Context) error

	// ListSchemas returns a list of schemas in the database
	ListSchemas(ctx context.Context) ([]string, error)

	// ListTables returns a list of tables in the specified schema
	ListTables(ctx context.Context, schema string) ([]string, error)

	// GetTableColumns returns information about the columns in a table
	GetTableColumns(ctx context.Context, schema, table string) (map[string]string, error)

	// ExecuteQuery executes a query and returns the results
	ExecuteQuery(ctx context.Context, query string) (ResultSet, error)

	// GetConnectorType returns the type of the connector (e.g., "postgresql", "mysql")
	GetConnectorType() string

	// Ping checks if the database connection is alive
	Ping(ctx context.Context) error
}

// ConnectorFactory creates database connectors
type ConnectorFactory interface {
	// CreateConnector creates a connector based on the database type
	CreateConnector(dbType string) (Connector, error)
}