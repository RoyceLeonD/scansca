package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// MockConnector implements the Connector interface for testing
type MockConnector struct{}

func (m *MockConnector) Connect(ctx context.Context, connectionString string) error {
	return nil
}

func (m *MockConnector) Disconnect(ctx context.Context) error {
	return nil
}

func (m *MockConnector) Ping(ctx context.Context) error {
	return nil
}

func (m *MockConnector) GetStatus(ctx context.Context) (*DatabaseStatus, error) {
	return &DatabaseStatus{
		Connected: true,
		MaxConns:  10,
	}, nil
}

func (m *MockConnector) ListSchemas(ctx context.Context) ([]string, error) {
	return []string{"public", "schema1", "schema2"}, nil
}

func (m *MockConnector) ListTables(ctx context.Context, schema string) ([]string, error) {
	return []string{"table1", "table2", "table3"}, nil
}

func (m *MockConnector) ListViews(ctx context.Context, schema string) ([]string, error) {
	return []string{"view1", "view2"}, nil
}

func (m *MockConnector) GetTableColumns(ctx context.Context, schema, table string) (map[string]string, error) {
	return map[string]string{
		"id":    "integer",
		"name":  "text",
		"value": "numeric",
	}, nil
}

func (m *MockConnector) GetTableInfo(ctx context.Context, schema, table string) (*TableInfo, error) {
	return &TableInfo{
		Schema: schema,
		Name:   table,
		Columns: []ColumnInfo{
			{Name: "id", DataType: "integer", IsPrimaryKey: true},
			{Name: "name", DataType: "text"},
			{Name: "value", DataType: "numeric"},
		},
		PrimaryKey:       []string{"id"},
		EstimatedRowCount: 1000,
		CreateTime:       time.Now(),
	}, nil
}

func (m *MockConnector) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (ResultSet, error) {
	return ResultSet{
		Columns: []string{"id", "name", "value"},
		Rows: [][]interface{}{
			{1, "test1", 10.5},
			{2, "test2", 20.5},
		},
	}, nil
}

func (m *MockConnector) ExecuteStatement(ctx context.Context, statement string, args ...interface{}) (int64, error) {
	return 1, nil
}

func (m *MockConnector) ExecuteTransaction(ctx context.Context, stmts []string) error {
	return nil
}

func (m *MockConnector) GetConnectorType() string {
	return "mock"
}

// createTestServer creates a server with a mock connector for testing
func createTestServer(t *testing.T) *MCPServer {
	// Create registry and mock connector
	registry := NewMockConnectorRegistry()
	factory := func() Connector {
		return &MockConnector{}
	}
	registry.Register("mock", factory)
	
	// Add test database to registry
	registry.AddDatabase(DatabaseConfig{
		Name: "test-db",
		Type: "mock",
		ConnectionString: "mock://test",
	})
	
	// Create server
	server := NewMCPServer("localhost", 8080, registry)
	return server
}

// TestHandleExecuteQuery tests the execute_query handler
func TestHandleExecuteQuery(t *testing.T) {
	server := createTestServer(t)
	ctx := context.Background()
	
	// Create test request
	params := ExecuteQueryRequest{
		Database: "test-db",
		Query:    "SELECT * FROM test_table",
		Params:   []interface{}{},
	}
	
	// Marshal parameters
	rawParams, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal parameters: %v", err)
	}
	
	// Execute handler
	result, err := server.handleExecuteQuery(ctx, rawParams)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	
	// Check response
	response, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map response, got %T", result)
	}
	
	// Check columns
	columns, ok := response["columns"].([]string)
	if !ok {
		t.Error("Missing or invalid columns in result")
	} else if len(columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(columns))
	}
	
	// Check rows
	rows, ok := response["rows"].([][]interface{})
	if !ok {
		t.Error("Missing or invalid rows in result")
	} else if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}
}

// TestHandleListSchemas tests the list_schemas handler
func TestHandleListSchemas(t *testing.T) {
	server := createTestServer(t)
	ctx := context.Background()
	
	// Create test request
	params := ListSchemasRequest{
		Database: "test-db",
	}
	
	// Marshal parameters
	rawParams, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal parameters: %v", err)
	}
	
	// Execute handler
	result, err := server.handleListSchemas(ctx, rawParams)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	
	// Check response
	response, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map response, got %T", result)
	}
	
	// Check database
	database, ok := response["database"].(string)
	if !ok || database != "test-db" {
		t.Errorf("Expected database 'test-db', got %v", response["database"])
	}
	
	// Check schemas
	schemas, ok := response["schemas"].([]string)
	if !ok {
		t.Error("Missing or invalid schemas in result")
	} else if len(schemas) != 3 {
		t.Errorf("Expected 3 schemas, got %d", len(schemas))
	}
}