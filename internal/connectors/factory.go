package connectors

import (
	"fmt"
)

// DefaultConnectorFactory is the standard factory implementation
type DefaultConnectorFactory struct{}

// NewConnectorFactory creates a new database connector factory
func NewConnectorFactory() *DefaultConnectorFactory {
	return &DefaultConnectorFactory{}
}

// CreateConnector creates a connector based on database type
func (f *DefaultConnectorFactory) CreateConnector(dbType string) (Connector, error) {
	switch dbType {
	case "postgresql", "postgres":
		// Using a function to avoid import cycle
		return NewPostgresConnector(), nil
	// Add cases for other database types as they are implemented
	// case "mysql":
	//     return NewMySQLConnector(), nil
	// case "sqlite":
	//     return NewSQLiteConnector(), nil
	// case "dynamodb":
	//     return NewDynamoDBConnector(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// AdvancedConnectorFactory is a factory for creating advanced connectors
type AdvancedConnectorFactory struct{}

// NewAdvancedConnectorFactory creates a new advanced database connector factory
func NewAdvancedConnectorFactory() *AdvancedConnectorFactory {
	return &AdvancedConnectorFactory{}
}

// CreateAdvancedConnector creates an advanced connector based on database type
func (f *AdvancedConnectorFactory) CreateAdvancedConnector(dbType string) (AdvancedConnector, error) {
	switch dbType {
	case "postgresql", "postgres":
		return NewAdvancedPostgresConnector(), nil
	// Add cases for other database types as they are implemented
	// case "mysql":
	//     return NewAdvancedMySQLConnector(), nil
	// case "sqlite":
	//     return NewAdvancedSQLiteConnector(), nil
	// case "dynamodb":
	//     return NewAdvancedDynamoDBConnector(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}