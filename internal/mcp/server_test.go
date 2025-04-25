package mcp

import (
	"testing"
)

// MockConnectorRegistry implements the ConnectorRegistry interface for testing
type MockConnectorRegistry struct {
	databases []DatabaseConfig
	factories map[string]func() Connector
}

// NewMockConnectorRegistry creates a new mock connector registry
func NewMockConnectorRegistry() *MockConnectorRegistry {
	return &MockConnectorRegistry{
		databases: []DatabaseConfig{},
		factories: make(map[string]func() Connector),
	}
}

// ListDatabases returns the list of registered databases
func (r *MockConnectorRegistry) ListDatabases() []DatabaseConfig {
	return r.databases
}

// Create creates a new connector for the given database type
func (r *MockConnectorRegistry) Create(dbType string) (Connector, error) {
	factory, exists := r.factories[dbType]
	if !exists {
		return nil, ErrUnsupportedDatabaseType
	}
	return factory(), nil
}

// AddDatabase adds a database to the registry
func (r *MockConnectorRegistry) AddDatabase(config DatabaseConfig) {
	r.databases = append(r.databases, config)
}

// Register registers a connector factory
func (r *MockConnectorRegistry) Register(dbType string, factory func() Connector) {
	r.factories[dbType] = factory
}

// Error variable for unsupported database type
var ErrUnsupportedDatabaseType = &DatabaseTypeError{Message: "unsupported database type"}

// DatabaseTypeError represents an error for unsupported database types
type DatabaseTypeError struct {
	Message string
}

// Error implements the error interface
func (e *DatabaseTypeError) Error() string {
	return e.Message
}

// TestNewMCPServer tests creating a new MCP server
func TestNewMCPServer(t *testing.T) {
	registry := NewMockConnectorRegistry()
	
	server := NewMCPServer("localhost", 8080, registry)
	
	if server == nil {
		t.Fatal("Server is nil")
	}
	
	if server.dbRegistry != registry {
		t.Error("Server has incorrect registry")
	}
}