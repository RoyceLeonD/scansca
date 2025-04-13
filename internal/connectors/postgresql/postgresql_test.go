package postgresql

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func getTestConnectionString() string {
	host := os.Getenv("ANTSKA_TEST_DB_HOST")
	if host == "" {
		host = "127.0.0.1" // Use IP address instead of localhost to avoid IPv6 issues
	}
	
	port := os.Getenv("ANTSKA_TEST_DB_PORT")
	if port == "" {
		port = "5433" // Default test port from docker-compose.test.yaml
	}
	
	user := os.Getenv("ANTSKA_TEST_DB_USER")
	if user == "" {
		user = "test_user"
	}
	
	password := os.Getenv("ANTSKA_TEST_DB_PASSWORD")
	if password == "" {
		password = "test_password"
	}
	
	dbName := os.Getenv("ANTSKA_TEST_DB_NAME")
	if dbName == "" {
		dbName = "antska_test"
	}
	
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbName)
}

// This test requires a running PostgreSQL database
// Run with: make test-integration
func TestPostgresConnector_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	connector := NewPostgresConnector()
	
	// Connect to the database
	connStr := getTestConnectionString()
	
	// Retry connection a few times in case the database is still starting up
	var err error
	for i := 0; i < 5; i++ {
		err = connector.Connect(ctx, connStr)
		if err == nil {
			break
		}
		t.Logf("Connection attempt %d failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	
	// Clean up after the test
	defer connector.Disconnect(ctx)
	
	// Test Ping
	if err := connector.Ping(ctx); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
	
	// Test ListSchemas
	schemas, err := connector.ListSchemas(ctx)
	if err != nil {
		t.Errorf("ListSchemas failed: %v", err)
	}
	t.Logf("Found schemas: %v", schemas)
	
	// Look for public schema
	hasPublic := false
	for _, schema := range schemas {
		if schema == "public" {
			hasPublic = true
			break
		}
	}
	if !hasPublic {
		t.Errorf("Public schema not found")
	}
	
	// Test ListTables in public schema
	tables, err := connector.ListTables(ctx, "public")
	if err != nil {
		t.Errorf("ListTables failed: %v", err)
	}
	t.Logf("Found tables in public schema: %v", tables)
	
	// Check for tables that should be created by migrations
	expectedTables := map[string]bool{
		"services":    false,
		"jobs":        false,
		"job_results": false,
	}
	
	for _, table := range tables {
		if _, exists := expectedTables[table]; exists {
			expectedTables[table] = true
		}
	}
	
	for table, found := range expectedTables {
		if !found {
			t.Errorf("Expected table %s not found", table)
		}
	}
	
	// Test GetTableColumns
	if len(tables) > 0 {
		// Get a sample table
		sampleTable := "services"
		
		columns, err := connector.GetTableColumns(ctx, "public", sampleTable)
		if err != nil {
			t.Errorf("GetTableColumns failed: %v", err)
		}
		
		t.Logf("Columns in %s table: %v", sampleTable, columns)
		
		// Verify expected columns
		expectedColumns := map[string]bool{
			"id":         false,
			"name":       false,
			"type":       false,
			"config":     false,
			"created_at": false,
			"updated_at": false,
		}
		
		for column := range columns {
			if _, exists := expectedColumns[column]; exists {
				expectedColumns[column] = true
			}
		}
		
		for column, found := range expectedColumns {
			if !found {
				t.Errorf("Expected column %s not found in table %s", column, sampleTable)
			}
		}
	}
	
	// Test ExecuteQuery - simple query
	result, err := connector.ExecuteQuery(ctx, "SELECT 1 as test")
	if err != nil {
		t.Errorf("ExecuteQuery failed: %v", err)
	}
	
	if len(result.Columns) != 1 || result.Columns[0] != "test" {
		t.Errorf("Expected column 'test', got %v", result.Columns)
	}
	
	if len(result.Rows) != 1 || result.Rows[0][0] != int32(1) {
		t.Errorf("Expected row with value 1, got %v", result.Rows)
	}
	
	// Test ExecuteQuery - query the services table
	servicesResult, err := connector.ExecuteQuery(ctx, "SELECT * FROM services LIMIT 5")
	if err != nil {
		t.Errorf("ExecuteQuery for services failed: %v", err)
	}
	
	t.Logf("Services query columns: %v", servicesResult.Columns)
	t.Logf("Service rows count: %d", len(servicesResult.Rows))
	
	// The table might be empty, which is fine for this test
	// We're just checking that the query executes without error
}