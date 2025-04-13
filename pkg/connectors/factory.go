package connectors

import (
	"fmt"

	"github.com/royceleond/antska/internal/connectors"
	"github.com/royceleond/antska/internal/connectors/postgresql"
)

// ConnectorFactory creates database connectors
type ConnectorFactory struct {
	// Add any factory-specific fields if needed
}

// NewConnectorFactory creates a new connector factory
func NewConnectorFactory() connectors.ConnectorFactory {
	return &ConnectorFactory{}
}

// CreateConnector creates a connector based on the database type
func (f *ConnectorFactory) CreateConnector(dbType string) (connectors.Connector, error) {
	switch dbType {
	case "postgresql":
		return postgresql.NewPostgresConnector(), nil
	// Add other connector types as they're implemented
	// case "mysql":
	//     return mysql.NewMySQLConnector(), nil
	// case "sqlite":
	//     return sqlite.NewSQLiteConnector(), nil
	// case "dynamodb":
	//     return dynamodb.NewDynamoDBConnector(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}