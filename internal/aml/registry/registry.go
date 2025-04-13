package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/royceleond/antska/internal/connectors"
)

// Registry manages database connections
type Registry struct {
	connectors      map[string]connectors.Connector
	connectorsMutex sync.RWMutex
	factory         connectors.ConnectorFactory
}

// NewRegistry creates a new database registry
func NewRegistry(factory connectors.ConnectorFactory) *Registry {
	return &Registry{
		connectors: make(map[string]connectors.Connector),
		factory:    factory,
	}
}

// RegisterDatabase registers a new database connection
func (r *Registry) RegisterDatabase(ctx context.Context, config connectors.DatabaseConfig) error {
	r.connectorsMutex.Lock()
	defer r.connectorsMutex.Unlock()

	// Check if a connector with this name already exists
	if _, exists := r.connectors[config.Name]; exists {
		return fmt.Errorf("database with name '%s' already registered", config.Name)
	}

	// Create connector based on database type
	connector, err := r.factory.CreateConnector(config.Type)
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}

	// Connect to the database
	if err := connector.Connect(ctx, config.ConnectionString); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Store the connector
	r.connectors[config.Name] = connector
	return nil
}

// UnregisterDatabase removes a database connection
func (r *Registry) UnregisterDatabase(ctx context.Context, name string) error {
	r.connectorsMutex.Lock()
	defer r.connectorsMutex.Unlock()

	connector, exists := r.connectors[name]
	if !exists {
		return fmt.Errorf("database with name '%s' not found", name)
	}

	// Disconnect from the database
	if err := connector.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from database: %w", err)
	}

	// Remove the connector from the map
	delete(r.connectors, name)
	return nil
}

// GetConnector returns a registered connector by name
func (r *Registry) GetConnector(name string) (connectors.Connector, error) {
	r.connectorsMutex.RLock()
	defer r.connectorsMutex.RUnlock()

	connector, exists := r.connectors[name]
	if !exists {
		return nil, fmt.Errorf("database with name '%s' not found", name)
	}

	return connector, nil
}

// ListDatabases returns a list of registered database names
func (r *Registry) ListDatabases() []string {
	r.connectorsMutex.RLock()
	defer r.connectorsMutex.RUnlock()

	databases := make([]string, 0, len(r.connectors))
	for name := range r.connectors {
		databases = append(databases, name)
	}

	return databases
}

// CloseAll closes all database connections
func (r *Registry) CloseAll(ctx context.Context) {
	r.connectorsMutex.Lock()
	defer r.connectorsMutex.Unlock()

	for name, connector := range r.connectors {
		err := connector.Disconnect(ctx)
		if err != nil {
			// Just log the error, but continue closing other connections
			fmt.Printf("Error disconnecting from database '%s': %v\n", name, err)
		}
	}

	// Clear the map
	r.connectors = make(map[string]connectors.Connector)
}