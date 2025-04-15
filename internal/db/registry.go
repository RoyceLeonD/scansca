package db

// ConnectorFactory is a function type for creating connector instances
type ConnectorFactory func() Connector

// InitializeRegistry registers all available database connectors
func InitializeRegistry() *ConnectorRegistry {
	registry := NewConnectorRegistry()
	
	// The actual registration will happen in the main package
	// to avoid import cycles
	
	return registry
}