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

// Setup helper for tests
func setupConnector(t *testing.T) (*PostgresConnector, context.Context, func()) {
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
	
	// Return cleanup function
	cleanup := func() {
		connector.Disconnect(ctx)
	}
	
	return connector, ctx, cleanup
}

// This test requires a running PostgreSQL database
// Run with: make test-postgres
func TestPostgresConnector_BasicOperations(t *testing.T) {
	connector, ctx, cleanup := setupConnector(t)
	defer cleanup()
	
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
	
	// Check that ExecutionTime is recorded
	if servicesResult.ExecutionTime <= 0 {
		t.Errorf("Expected ExecutionTime to be greater than 0, got %f", servicesResult.ExecutionTime)
	}
}

func TestPostgresConnector_GetStatus(t *testing.T) {
	connector, ctx, cleanup := setupConnector(t)
	defer cleanup()
	
	// Test GetStatus
	status, err := connector.GetStatus(ctx)
	if err != nil {
		t.Errorf("GetStatus failed: %v", err)
	}
	
	// Check status has expected keys
	expectedKeys := []string{
		"acquired_connections",
		"idle_connections",
		"total_connections",
		"max_connections",
		"database_version",
	}
	
	for _, key := range expectedKeys {
		if _, ok := status[key]; !ok {
			t.Errorf("Expected status key %s not found", key)
		}
	}
	
	t.Logf("Database version: %v", status["database_version"])
}

func TestPostgresConnector_ListViews(t *testing.T) {
	connector, ctx, cleanup := setupConnector(t)
	defer cleanup()
	
	// Create a test view
	_, err := connector.ExecuteStatement(ctx, `CREATE OR REPLACE VIEW public.test_view AS SELECT * FROM services`)
	if err != nil {
		t.Fatalf("Failed to create test view: %v", err)
	}
	
	// Test ListViews
	views, err := connector.ListViews(ctx, "public")
	if err != nil {
		t.Errorf("ListViews failed: %v", err)
	}
	
	// Check that our test view is in the list
	viewFound := false
	for _, view := range views {
		if view == "test_view" {
			viewFound = true
			break
		}
	}
	
	if !viewFound {
		t.Errorf("Expected view 'test_view' not found in views: %v", views)
	}
	
	// Clean up
	_, err = connector.ExecuteStatement(ctx, `DROP VIEW public.test_view`)
	if err != nil {
		t.Logf("Failed to drop test view: %v", err)
	}
}

func TestPostgresConnector_GetTableInfo(t *testing.T) {
	connector, ctx, cleanup := setupConnector(t)
	defer cleanup()
	
	// Test GetTableInfo on the services table
	tableInfo, err := connector.GetTableInfo(ctx, "public", "services")
	if err != nil {
		t.Fatalf("GetTableInfo failed: %v", err)
		return // Return early to avoid nil pointer dereference
	}
	
	// Check schema and name
	if tableInfo.Schema != "public" {
		t.Errorf("Expected schema 'public', got '%s'", tableInfo.Schema)
	}
	
	if tableInfo.Name != "services" {
		t.Errorf("Expected table name 'services', got '%s'", tableInfo.Name)
	}
	
	// Check primary key
	if len(tableInfo.PrimaryKey) == 0 {
		t.Errorf("Expected primary key to be non-empty")
	} else if tableInfo.PrimaryKey[0] != "id" {
		t.Errorf("Expected primary key column to be 'id', got '%s'", tableInfo.PrimaryKey[0])
	}
	
	// Check columns
	if len(tableInfo.Columns) < 5 {
		t.Errorf("Expected at least 5 columns, got %d", len(tableInfo.Columns))
	}
	
	// Check for specific columns
	idFound := false
	nameFound := false
	typeFound := false
	
	for _, col := range tableInfo.Columns {
		switch col.Name {
		case "id":
			idFound = true
			if !col.IsPrimaryKey {
				t.Errorf("Expected 'id' column to be a primary key")
			}
		case "name":
			nameFound = true
		case "type":
			typeFound = true
		}
	}
	
	if !idFound {
		t.Errorf("Expected column 'id' not found")
	}
	if !nameFound {
		t.Errorf("Expected column 'name' not found")
	}
	if !typeFound {
		t.Errorf("Expected column 'type' not found")
	}
}

func TestPostgresConnector_Statement(t *testing.T) {
	connector, ctx, cleanup := setupConnector(t)
	defer cleanup()
	
	// Create a test table
	_, err := connector.ExecuteStatement(ctx, `
		CREATE TABLE IF NOT EXISTS public.test_statements (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			value INT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	
	// Insert test data
	rowsAffected, err := connector.ExecuteStatement(ctx, `
		INSERT INTO public.test_statements (name, value) VALUES 
		('test1', 100),
		('test2', 200),
		('test3', 300)
	`)
	if err != nil {
		t.Errorf("ExecuteStatement INSERT failed: %v", err)
	}
	
	if rowsAffected != 3 {
		t.Errorf("Expected 3 rows affected, got %d", rowsAffected)
	}
	
	// Update test data
	rowsAffected, err = connector.ExecuteStatement(ctx, `
		UPDATE public.test_statements SET value = value * 2 WHERE name = $1
	`, "test1")
	if err != nil {
		t.Errorf("ExecuteStatement UPDATE failed: %v", err)
	}
	
	if rowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", rowsAffected)
	}
	
	// Query the updated data
	result, err := connector.ExecuteQuery(ctx, `
		SELECT * FROM public.test_statements ORDER BY id
	`)
	if err != nil {
		t.Errorf("ExecuteQuery after updates failed: %v", err)
	}
	
	if len(result.Rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(result.Rows))
	}
	
	// Clean up
	_, err = connector.ExecuteStatement(ctx, `DROP TABLE public.test_statements`)
	if err != nil {
		t.Logf("Failed to drop test table: %v", err)
	}
}

func TestPostgresConnector_Transaction(t *testing.T) {
	connector, ctx, cleanup := setupConnector(t)
	defer cleanup()
	
	// Create a test table
	_, err := connector.ExecuteStatement(ctx, `
		CREATE TABLE IF NOT EXISTS public.test_transactions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			value INT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	
	// Execute transaction
	err = connector.ExecuteTransaction(ctx, []string{
		`INSERT INTO public.test_transactions (name, value) VALUES ('tx1', 100)`,
		`INSERT INTO public.test_transactions (name, value) VALUES ('tx2', 200)`,
		`UPDATE public.test_transactions SET value = 150 WHERE name = 'tx1'`,
	})
	if err != nil {
		t.Errorf("ExecuteTransaction failed: %v", err)
	}
	
	// Query data to verify transaction
	result, err := connector.ExecuteQuery(ctx, `
		SELECT * FROM public.test_transactions ORDER BY id
	`)
	if err != nil {
		t.Errorf("ExecuteQuery after transaction failed: %v", err)
	}
	
	if len(result.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(result.Rows))
	}
	
	// Try a transaction that should fail
	err = connector.ExecuteTransaction(ctx, []string{
		`INSERT INTO public.test_transactions (name, value) VALUES ('tx3', 300)`,
		`INSERT INTO public.nonexistent_table (col) VALUES ('fail')`, // This should fail
	})
	if err == nil {
		t.Errorf("Expected transaction to fail, but it succeeded")
	}
	
	// Verify no data was committed in the failed transaction
	result, err = connector.ExecuteQuery(ctx, `
		SELECT COUNT(*) FROM public.test_transactions WHERE name = 'tx3'
	`)
	if err != nil {
		t.Errorf("ExecuteQuery to check transaction rollback failed: %v", err)
	}
	
	if len(result.Rows) == 0 || result.Rows[0][0].(int64) != 0 {
		t.Errorf("Expected 0 rows for 'tx3', got %v", result.Rows[0][0])
	}
	
	// Clean up
	_, err = connector.ExecuteStatement(ctx, `DROP TABLE public.test_transactions`)
	if err != nil {
		t.Logf("Failed to drop test table: %v", err)
	}
}

func TestPostgresConnector_BatchOperations(t *testing.T) {
	connector, ctx, cleanup := setupConnector(t)
	defer cleanup()
	
	// Create a test table
	_, err := connector.ExecuteStatement(ctx, `
		CREATE TABLE IF NOT EXISTS public.test_batch (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			value INT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	
	// Execute batch of statements
	err = connector.ExecuteBatch(ctx, []string{
		`INSERT INTO public.test_batch (name, value) VALUES ('batch1', 100)`,
		`INSERT INTO public.test_batch (name, value) VALUES ('batch2', 200)`,
		`INSERT INTO public.test_batch (name, value) VALUES ('batch3', 300)`,
		`UPDATE public.test_batch SET value = 150 WHERE name = 'batch1'`,
	})
	if err != nil {
		t.Errorf("ExecuteBatch failed: %v", err)
	}
	
	// Query data to verify batch execution
	result, err := connector.ExecuteQuery(ctx, `
		SELECT COUNT(*) FROM public.test_batch
	`)
	if err != nil {
		t.Errorf("ExecuteQuery after batch failed: %v", err)
	}
	
	if len(result.Rows) == 0 || result.Rows[0][0].(int64) != 3 {
		t.Errorf("Expected 3 rows, got %v", result.Rows[0][0])
	}
	
	// Clean up
	_, err = connector.ExecuteStatement(ctx, `DROP TABLE public.test_batch`)
	if err != nil {
		t.Logf("Failed to drop test table: %v", err)
	}
}

func TestPostgresConnector_CustomConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	
	// Create connector with custom configuration
	customConfig := ConnectionConfig{
		MaxConnections: 5,
		MinConnections: 1,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
	}
	
	connector := NewPostgresConnectorWithConfig(customConfig)
	
	// Connect to the database
	connStr := getTestConnectionString()
	err := connector.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect with custom config: %v", err)
	}
	defer connector.Disconnect(ctx)
	
	// Get status and verify max connections
	status, err := connector.GetStatus(ctx)
	if err != nil {
		t.Errorf("GetStatus failed: %v", err)
	}
	
	if maxConn, ok := status["max_connections"]; !ok || maxConn != 5 {
		t.Errorf("Expected max_connections to be 5, got %v", maxConn)
	}
}

func TestPostgresConnector_Reconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	connector := NewPostgresConnector()
	
	// Connect to the database
	connStr := getTestConnectionString()
	err := connector.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	
	// Disconnect
	connector.Disconnect(ctx)
	
	// Attempt to execute a query (should fail)
	_, err = connector.ExecuteQuery(ctx, "SELECT 1")
	if err == nil {
		t.Errorf("Expected query to fail after disconnection, but it succeeded")
	}
	
	// Reconnect
	err = connector.Reconnect(ctx)
	if err != nil {
		t.Errorf("Reconnect failed: %v", err)
	}
	
	// Try a query again
	_, err = connector.ExecuteQuery(ctx, "SELECT 1")
	if err != nil {
		t.Errorf("Query failed after reconnection: %v", err)
	}
	
	// Clean up
	connector.Disconnect(ctx)
}

// Test querying data from the test_schema created in the test data
func TestPostgresConnector_TestSchema(t *testing.T) {
	connector, ctx, cleanup := setupConnector(t)
	defer cleanup()
	
	// Check that test_schema exists
	schemas, err := connector.ListSchemas(ctx)
	if err != nil {
		t.Errorf("ListSchemas failed: %v", err)
	}
	
	schemaFound := false
	for _, schema := range schemas {
		if schema == "test_schema" {
			schemaFound = true
			break
		}
	}
	
	if !schemaFound {
		t.Skip("test_schema not found - skipping test (run 'make test-data-load' first)")
	}
	
	// Test ListTables in test_schema
	tables, err := connector.ListTables(ctx, "test_schema")
	if err != nil {
		t.Errorf("ListTables failed: %v", err)
	}
	
	expectedTestTables := map[string]bool{
		"users":      false,
		"posts":      false,
		"data_types": false,
	}
	
	for _, table := range tables {
		if _, exists := expectedTestTables[table]; exists {
			expectedTestTables[table] = true
		}
	}
	
	for table, found := range expectedTestTables {
		if !found {
			t.Errorf("Expected table %s not found in test_schema", table)
		}
	}
	
	// Test querying users table
	result, err := connector.ExecuteQuery(ctx, `
		SELECT * FROM test_schema.users ORDER BY id
	`)
	if err != nil {
		t.Errorf("ExecuteQuery on test_schema.users failed: %v", err)
	}
	
	if len(result.Rows) < 3 {
		t.Errorf("Expected at least 3 rows in test_schema.users, got %d", len(result.Rows))
	}
	
	// Test querying with a JOIN
	result, err = connector.ExecuteQuery(ctx, `
		SELECT u.username, p.title 
		FROM test_schema.users u
		JOIN test_schema.posts p ON u.id = p.user_id
		ORDER BY u.id, p.id
	`)
	if err != nil {
		t.Errorf("ExecuteQuery with JOIN failed: %v", err)
	}
	
	if len(result.Rows) < 4 {
		t.Errorf("Expected at least 4 rows in JOIN query, got %d", len(result.Rows))
	}
	
	// Test handling of different data types
	result, err = connector.ExecuteQuery(ctx, `
		SELECT * FROM test_schema.data_types LIMIT 1
	`)
	if err != nil {
		t.Errorf("ExecuteQuery on data_types failed: %v", err)
	}
	
	if len(result.Rows) != 1 {
		t.Errorf("Expected 1 row in data_types, got %d", len(result.Rows))
	}
}