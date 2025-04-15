package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/royceleond/scansca/internal/db"
)

// Tool represents an MCP tool
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
	Handler     ToolHandler
}

// ToolHandler is a function that handles tool execution
type ToolHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// QueryParams represents parameters for query execution
type QueryParams struct {
	Database string        `json:"database" binding:"required"`
	Query    string        `json:"query" binding:"required"`
	Params   []interface{} `json:"params"`
}

// SchemaParams represents parameters for schema listing
type SchemaParams struct {
	Database string `json:"database" binding:"required"`
}

// TableParams represents parameters for table operations
type TableParams struct {
	Database string `json:"database" binding:"required"`
	Schema   string `json:"schema" binding:"required"`
}

// TableInfoParams represents parameters for table info retrieval
type TableInfoParams struct {
	Database string `json:"database" binding:"required"`
	Schema   string `json:"schema" binding:"required"`
	Table    string `json:"table" binding:"required"`
}

// ToolRegistry holds available MCP tools
type ToolRegistry struct {
	tools          map[string]Tool
	connectorMgr   ConnectorManager
}

// ConnectorManager defines the interface for database connection management
type ConnectorManager interface {
	GetConnector(name string) (db.Connector, error)
	ListDatabases() []db.DatabaseConfig
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(connMgr ConnectorManager) *ToolRegistry {
	registry := &ToolRegistry{
		tools:        make(map[string]Tool),
		connectorMgr: connMgr,
	}

	// Register default tools
	registry.registerDefaultTools()

	return registry
}

// registerDefaultTools registers the standard MCP tools
func (r *ToolRegistry) registerDefaultTools() {
	// Database query tool
	r.Register(Tool{
		Name:        "execute_query",
		Description: "Execute a SQL query on a registered database",
		Parameters: map[string]interface{}{
			"database": map[string]string{
				"type":        "string",
				"description": "Database name",
			},
			"query": map[string]string{
				"type":        "string",
				"description": "SQL query to execute",
			},
			"params": map[string]string{
				"type":        "array",
				"description": "Query parameters (optional)",
			},
		},
		Handler: r.handleExecuteQuery,
	})

	// List schemas tool
	r.Register(Tool{
		Name:        "list_schemas",
		Description: "List all schemas in a database",
		Parameters: map[string]interface{}{
			"database": map[string]string{
				"type":        "string",
				"description": "Database name",
			},
		},
		Handler: r.handleListSchemas,
	})

	// List tables tool
	r.Register(Tool{
		Name:        "list_tables",
		Description: "List all tables in a database schema",
		Parameters: map[string]interface{}{
			"database": map[string]string{
				"type":        "string",
				"description": "Database name",
			},
			"schema": map[string]string{
				"type":        "string",
				"description": "Schema name",
			},
		},
		Handler: r.handleListTables,
	})

	// Get table info tool
	r.Register(Tool{
		Name:        "get_table_info",
		Description: "Get detailed information about a database table",
		Parameters: map[string]interface{}{
			"database": map[string]string{
				"type":        "string",
				"description": "Database name",
			},
			"schema": map[string]string{
				"type":        "string",
				"description": "Schema name",
			},
			"table": map[string]string{
				"type":        "string",
				"description": "Table name",
			},
		},
		Handler: r.handleGetTableInfo,
	})
}

// Register adds a tool to the registry
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Name] = tool
}

// GetTool returns a tool by name
func (r *ToolRegistry) GetTool(name string) (Tool, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

// ListTools returns all registered tools
func (r *ToolRegistry) ListTools() []Tool {
	result := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	return result
}

// ExecuteTool executes a tool with the given parameters
func (r *ToolRegistry) ExecuteTool(ctx context.Context, name string, rawParams json.RawMessage) (interface{}, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	// Parse parameters
	var params map[string]interface{}
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Execute tool handler
	startTime := time.Now()
	result, err := tool.Handler(ctx, params)
	duration := time.Since(startTime)

	log.Debug().
		Str("tool", name).
		Dur("duration", duration).
		Interface("params", params).
		Bool("success", err == nil).
		Msg("Tool execution")

	return result, err
}

// Tool handlers

// handleExecuteQuery handles the execute_query tool
func (r *ToolRegistry) handleExecuteQuery(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	dbName, ok := params["database"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid database parameter")
	}

	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid query parameter")
	}

	// Get query parameters
	var queryParams []interface{}
	if rawParams, ok := params["params"]; ok {
		if paramsSlice, ok := rawParams.([]interface{}); ok {
			queryParams = paramsSlice
		}
	}

	// Get database connector
	connector, err := r.connectorMgr.GetConnector(dbName)
	if err != nil {
		return nil, err
	}

	// Execute query
	result, err := connector.ExecuteQuery(ctx, query, queryParams...)
	if err != nil {
		// Still return the result, as it may contain partial data and error info
		result.Error = err.Error()
	}

	return result, nil
}

// handleListSchemas handles the list_schemas tool
func (r *ToolRegistry) handleListSchemas(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	dbName, ok := params["database"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid database parameter")
	}

	// Get database connector
	connector, err := r.connectorMgr.GetConnector(dbName)
	if err != nil {
		return nil, err
	}

	// List schemas
	schemas, err := connector.ListSchemas(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"database": dbName,
		"schemas":  schemas,
	}, nil
}

// handleListTables handles the list_tables tool
func (r *ToolRegistry) handleListTables(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	dbName, ok := params["database"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid database parameter")
	}

	schema, ok := params["schema"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid schema parameter")
	}

	// Get database connector
	connector, err := r.connectorMgr.GetConnector(dbName)
	if err != nil {
		return nil, err
	}

	// List tables
	tables, err := connector.ListTables(ctx, schema)
	if err != nil {
		return nil, err
	}

	// Try to list views (optional)
	views, viewErr := connector.ListViews(ctx, schema)
	if viewErr != nil {
		log.Warn().Err(viewErr).Msg("Failed to list views")
		views = []string{} // Use empty list on error
	}

	return map[string]interface{}{
		"database": dbName,
		"schema":   schema,
		"tables":   tables,
		"views":    views,
	}, nil
}

// handleGetTableInfo handles the get_table_info tool
func (r *ToolRegistry) handleGetTableInfo(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	dbName, ok := params["database"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid database parameter")
	}

	schema, ok := params["schema"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid schema parameter")
	}

	table, ok := params["table"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid table parameter")
	}

	// Get database connector
	connector, err := r.connectorMgr.GetConnector(dbName)
	if err != nil {
		return nil, err
	}

	// Get table info
	tableInfo, err := connector.GetTableInfo(ctx, schema, table)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"database":  dbName,
		"schema":    schema,
		"table":     table,
		"tableInfo": tableInfo,
	}, nil
}