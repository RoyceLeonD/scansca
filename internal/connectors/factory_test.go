package connectors

import (
	"context"
	"testing"
)

func TestCreateConnector(t *testing.T) {
	factory := NewConnectorFactory()
	
	tests := []struct {
		name          string
		dbType        string
		shouldSucceed bool
		expectedType  string
	}{
		{
			name:          "PostgreSQL",
			dbType:        "postgresql",
			shouldSucceed: true,
			expectedType:  "postgresql",
		},
		{
			name:          "Postgres alias",
			dbType:        "postgres",
			shouldSucceed: true,
			expectedType:  "postgresql",
		},
		{
			name:          "Unsupported database",
			dbType:        "unsupported-db",
			shouldSucceed: false,
		},
		// Add more tests as more connectors are implemented
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			connector, err := factory.CreateConnector(tc.dbType)
			
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("Expected success for %s, but got error: %v", tc.dbType, err)
				}
				
				if connector == nil {
					t.Errorf("Expected non-nil connector for %s", tc.dbType)
					return
				}
				
				actualType := connector.GetConnectorType()
				if actualType != tc.expectedType {
					t.Errorf("Expected connector type %s, but got %s", tc.expectedType, actualType)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for unsupported database type %s, but got success", tc.dbType)
				}
				
				if connector != nil {
					t.Errorf("Expected nil connector for unsupported database type %s", tc.dbType)
				}
			}
		})
	}
}

func TestCreateAdvancedConnector(t *testing.T) {
	factory := NewAdvancedConnectorFactory()
	
	tests := []struct {
		name          string
		dbType        string
		shouldSucceed bool
		expectedType  string
	}{
		{
			name:          "PostgreSQL Advanced",
			dbType:        "postgresql",
			shouldSucceed: true,
			expectedType:  "postgresql",
		},
		{
			name:          "Postgres alias Advanced",
			dbType:        "postgres",
			shouldSucceed: true,
			expectedType:  "postgresql",
		},
		{
			name:          "Unsupported database Advanced",
			dbType:        "unsupported-db",
			shouldSucceed: false,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			connector, err := factory.CreateAdvancedConnector(tc.dbType)
			
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("Expected success for %s, but got error: %v", tc.dbType, err)
				}
				
				if connector == nil {
					t.Errorf("Expected non-nil connector for %s", tc.dbType)
					return
				}
				
				actualType := connector.GetConnectorType()
				if actualType != tc.expectedType {
					t.Errorf("Expected connector type %s, but got %s", tc.expectedType, actualType)
				}
				
				// Test that advanced methods exist
				ctx := context.Background()
				_, err := connector.GetStatus(ctx)
				if err != nil && err.Error() != "not connected to database" {
					t.Errorf("Expected advanced method GetStatus to exist, but got unexpected error: %v", err)
				}
				
				_, err = connector.ListViews(ctx, "test")
				if err != nil && err.Error() != "not connected to database" {
					t.Errorf("Expected advanced method ListViews to exist, but got unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for unsupported database type %s, but got success", tc.dbType)
				}
				
				if connector != nil {
					t.Errorf("Expected nil connector for unsupported database type %s", tc.dbType)
				}
			}
		})
	}
}

// TestConnectorInterface verifies that the adapter properly implements both interfaces
func TestConnectorInterface(t *testing.T) {
	// Create a PostgreSQL adapter
	adapter := &PostgresConnectorAdapter{}
	
	// Test it satisfies both interfaces
	var _ Connector = adapter         // Should compile if adapter implements Connector
	var _ AdvancedConnector = adapter // Should compile if adapter implements AdvancedConnector
}