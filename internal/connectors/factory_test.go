package connectors

import (
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