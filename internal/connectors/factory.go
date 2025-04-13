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