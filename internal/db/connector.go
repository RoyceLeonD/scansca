package db

import (
	"context"
)

// Connector defines the interface for database interactions
// All database-specific implementations must implement this interface
type Connector interface {
	// Connection Management
	Connect(ctx context.Context, connectionString string) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	GetStatus(ctx context.Context) (*DatabaseStatus, error)
	
	// Schema Introspection
	ListSchemas(ctx context.Context) ([]string, error)
	ListTables(ctx context.Context, schema string) ([]string, error)
	ListViews(ctx context.Context, schema string) ([]string, error)
	GetTableColumns(ctx context.Context, schema, table string) (map[string]string, error)
	GetTableInfo(ctx context.Context, schema, table string) (*TableInfo, error)
	
	// Query Execution
	ExecuteQuery(ctx context.Context, query string, args ...interface{}) (ResultSet, error)
	ExecuteStatement(ctx context.Context, statement string, args ...interface{}) (int64, error)
	ExecuteTransaction(ctx context.Context, stmts []string) error
	
	// Type Information
	GetConnectorType() string
}

// BaseConnector provides a common structure for connector implementations
type BaseConnector struct {
	Config ConnectionPoolConfig
}

// ConnectorRegistry manages database connector registration and creation
type ConnectorRegistry struct {
	factories map[string]func() Connector
}

// NewConnectorRegistry creates a new connector registry
func NewConnectorRegistry() *ConnectorRegistry {
	return &ConnectorRegistry{
		factories: make(map[string]func() Connector),
	}
}

// Register adds a connector factory for a specific database type
func (r *ConnectorRegistry) Register(dbType string, factory func() Connector) {
	r.factories[dbType] = factory
}

// Create instantiates a connector for the specified database type
func (r *ConnectorRegistry) Create(dbType string) (Connector, error) {
	factory, exists := r.factories[dbType]
	if !exists {
		return nil, ErrUnsupportedDatabaseType(dbType)
	}
	return factory(), nil
}

// ErrUnsupportedDatabaseType returns an error for unsupported database types
func ErrUnsupportedDatabaseType(dbType string) error {
	return &UnsupportedDatabaseTypeError{DBType: dbType}
}

// UnsupportedDatabaseTypeError is returned when an unsupported database type is requested
type UnsupportedDatabaseTypeError struct {
	DBType string
}

// Error implements the error interface
func (e *UnsupportedDatabaseTypeError) Error() string {
	return "unsupported database type: " + e.DBType
}