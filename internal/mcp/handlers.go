package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// handleExecuteQuery handles the execute_query tool
func (s *MCPServer) handleExecuteQuery(ctx context.Context, rawParams json.RawMessage) (interface{}, error) {
	startTime := time.Now()
	
	// Parse parameters
	var queryParams ExecuteQueryRequest
	if err := json.Unmarshal(rawParams, &queryParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	
	// Find the connector
	connector, err := s.getConnector(queryParams.Database)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	
	// Execute query
	result, err := connector.ExecuteQuery(ctx, queryParams.Query, queryParams.Params...)
	
	// Prepare response
	response := map[string]interface{}{
		"columns": result.Columns,
		"rows":    result.Rows,
		"count":   len(result.Rows),
		"duration": time.Since(startTime).String(),
	}
	
	if err != nil {
		response["error"] = err.Error()
	}
	
	return response, nil
}

// handleListSchemas handles the list_schemas tool
func (s *MCPServer) handleListSchemas(ctx context.Context, rawParams json.RawMessage) (interface{}, error) {
	// Parse parameters
	var schemaParams ListSchemasRequest
	if err := json.Unmarshal(rawParams, &schemaParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	
	// Find the connector
	connector, err := s.getConnector(schemaParams.Database)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	
	// List schemas
	schemas, err := connector.ListSchemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	
	// Return result
	return map[string]interface{}{
		"database": schemaParams.Database,
		"schemas":  schemas,
	}, nil
}

// handleListTables handles the list_tables tool
func (s *MCPServer) handleListTables(ctx context.Context, rawParams json.RawMessage) (interface{}, error) {
	// Parse parameters
	var tableParams ListTablesRequest
	if err := json.Unmarshal(rawParams, &tableParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	
	// Find the connector
	connector, err := s.getConnector(tableParams.Database)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	
	// List tables
	tables, err := connector.ListTables(ctx, tableParams.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	
	// Try to list views (optional)
	views, viewErr := connector.ListViews(ctx, tableParams.Schema)
	if viewErr != nil {
		log.Warn().Err(viewErr).Msg("Failed to list views")
		views = []string{} // Use empty list on error
	}
	
	// Return result
	return map[string]interface{}{
		"database": tableParams.Database,
		"schema":   tableParams.Schema,
		"tables":   tables,
		"views":    views,
	}, nil
}

// handleGetTableInfo handles the get_table_info tool
func (s *MCPServer) handleGetTableInfo(ctx context.Context, rawParams json.RawMessage) (interface{}, error) {
	// Parse parameters
	var tableInfoParams TableInfoRequest
	if err := json.Unmarshal(rawParams, &tableInfoParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	
	// Find the connector
	connector, err := s.getConnector(tableInfoParams.Database)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	
	// Get table info
	tableInfo, err := connector.GetTableInfo(ctx, tableInfoParams.Schema, tableInfoParams.Table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table info: %w", err)
	}
	
	// Return result
	return map[string]interface{}{
		"database":  tableInfoParams.Database,
		"schema":    tableInfoParams.Schema,
		"table":     tableInfoParams.Table,
		"tableInfo": tableInfo,
	}, nil
}

// getConnector finds a database connector by name
func (s *MCPServer) getConnector(name string) (Connector, error) {
	// Look through all the databases in the registry
	for _, config := range s.dbRegistry.ListDatabases() {
		if config.Name == name {
			// Found the database, create a connector
			connector, err := s.dbRegistry.Create(config.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to create connector: %w", err)
			}
			
			// Connect to the database
			err = connector.Connect(context.Background(), config.ConnectionString)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to database: %w", err)
			}
			
			return connector, nil
		}
	}
	
	return nil, fmt.Errorf("database not found: %s", name)
}