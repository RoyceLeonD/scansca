package connectors

import (
	"context"
	"time"
)

// ResultSet represents a database query result
type ResultSet struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Error   string          `json:"error,omitempty"`
	// Additional metadata about the query
	AffectedRows int64    `json:"affected_rows,omitempty"`
	ExecutionTime float64 `json:"execution_time,omitempty"` // in seconds
}

// ColumnInfo contains detailed information about a database column
type ColumnInfo struct {
	Name            string `json:"name"`
	DataType        string `json:"data_type"`
	IsNullable      bool   `json:"is_nullable"`
	DefaultValue    string `json:"default_value,omitempty"`
	IsPrimaryKey    bool   `json:"is_primary_key"`
	IsForeignKey    bool   `json:"is_foreign_key"`
	ReferencesTable string `json:"references_table,omitempty"`
	ReferencesColumn string `json:"references_column,omitempty"`
}

// TableInfo contains detailed information about a database table
type TableInfo struct {
	Schema           string       `json:"schema"`
	Name             string       `json:"name"`
	Columns          []ColumnInfo `json:"columns"`
	PrimaryKey       []string     `json:"primary_key,omitempty"`
	EstimatedRowCount int64       `json:"estimated_row_count"`
	CreateTime       time.Time    `json:"create_time"`
	Description      string       `json:"description,omitempty"`
}

// DatabaseConfig holds configuration for a database connection
type DatabaseConfig struct {
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	ConnectionString string                 `json:"connection_string"`
	Options          map[string]interface{} `json:"options,omitempty"`
}

// Connector defines the base interface for all database connectors
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

// AdvancedConnector extends the base Connector interface with additional capabilities
type AdvancedConnector interface {
	Connector

	// Reconnect attempts to reconnect to the database using the stored connection string
	Reconnect(ctx context.Context) error

	// GetStatus returns the current status of the database connection
	GetStatus(ctx context.Context) (map[string]interface{}, error)

	// ListViews returns a list of views in the specified schema
	ListViews(ctx context.Context, schema string) ([]string, error)

	// GetTableInfo returns detailed information about a table
	GetTableInfo(ctx context.Context, schema, table string) (*TableInfo, error)
	
	// ExecuteStatement executes a SQL statement (INSERT, UPDATE, DELETE) and returns affected rows
	ExecuteStatement(ctx context.Context, statement string, args ...interface{}) (int64, error)
	
	// ExecuteTransaction executes multiple statements in a transaction
	ExecuteTransaction(ctx context.Context, stmts []string) error
	
	// ExecuteBatch executes multiple statements efficiently using a batch
	ExecuteBatch(ctx context.Context, stmts []string) error
}