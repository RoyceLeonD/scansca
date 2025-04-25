package mcp

import (
	"context"
	"time"
)

// Connector defines the interface for a database connector
type Connector interface {
	// Connection management
	Connect(ctx context.Context, connectionString string) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	GetStatus(ctx context.Context) (*DatabaseStatus, error)
	
	// Schema introspection
	ListSchemas(ctx context.Context) ([]string, error)
	ListTables(ctx context.Context, schema string) ([]string, error)
	ListViews(ctx context.Context, schema string) ([]string, error)
	GetTableColumns(ctx context.Context, schema, table string) (map[string]string, error)
	GetTableInfo(ctx context.Context, schema, table string) (*TableInfo, error)
	
	// Query execution
	ExecuteQuery(ctx context.Context, query string, args ...interface{}) (ResultSet, error)
	ExecuteStatement(ctx context.Context, statement string, args ...interface{}) (int64, error)
	ExecuteTransaction(ctx context.Context, stmts []string) error
	
	// Type info
	GetConnectorType() string
}

// DatabaseConfig holds configuration for a database connection
type DatabaseConfig struct {
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	ConnectionString string                 `json:"connection_string"`
	Options          map[string]interface{} `json:"options,omitempty"`
}

// DatabaseStatus represents the current status of a database connection
type DatabaseStatus struct {
	Connected         bool                   `json:"connected"`
	AcquiredConns     int                    `json:"acquired_connections"`
	IdleConns         int                    `json:"idle_connections"`
	TotalConns        int                    `json:"total_connections"`
	MaxConns          int                    `json:"max_connections"`
	DatabaseVersion   string                 `json:"database_version,omitempty"`
	AdditionalDetails map[string]interface{} `json:"additional_details,omitempty"`
}

// ColumnInfo contains detailed information about a database column
type ColumnInfo struct {
	Name             string `json:"name"`
	DataType         string `json:"data_type"`
	IsNullable       bool   `json:"is_nullable"`
	DefaultValue     string `json:"default_value,omitempty"`
	IsPrimaryKey     bool   `json:"is_primary_key"`
	IsForeignKey     bool   `json:"is_foreign_key"`
	ReferencesTable  string `json:"references_table,omitempty"`
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

// ResultSet represents a database query result
type ResultSet struct {
	Columns       []string        `json:"columns"`
	Rows          [][]interface{} `json:"rows"`
	Error         string          `json:"error,omitempty"`
	AffectedRows  int64           `json:"affected_rows,omitempty"`
	ExecutionTime float64         `json:"execution_time,omitempty"` // in seconds
}

// MCP Request Types

// ExecuteQueryRequest represents parameters for a query execution request
type ExecuteQueryRequest struct {
	Database string        `json:"database" binding:"required"`
	Query    string        `json:"query" binding:"required"`
	Params   []interface{} `json:"params,omitempty"`
}

// ListSchemasRequest represents parameters for listing schemas
type ListSchemasRequest struct {
	Database string `json:"database" binding:"required"`
}

// ListTablesRequest represents parameters for listing tables
type ListTablesRequest struct {
	Database string `json:"database" binding:"required"`
	Schema   string `json:"schema" binding:"required"`
}

// TableInfoRequest represents parameters for getting table info
type TableInfoRequest struct {
	Database string `json:"database" binding:"required"`
	Schema   string `json:"schema" binding:"required"`
	Table    string `json:"table" binding:"required"`
}