# MCP Integration Guide

This document explains how Scansca implements the Model Context Protocol (MCP) to enable seamless integration between LLM clients and database systems.

## What is the Model Context Protocol (MCP)?

The Model Context Protocol (MCP) is a standardized protocol for enabling LLMs to interact with external tools and data sources. It allows LLMs to:

1. Discover available tools and resources
2. Invoke operations on these resources
3. Access structured data in a consistent format
4. Maintain context across interactions

By implementing MCP, Antska provides a standardized way for any LLM client to interact with database systems, without requiring custom integration for each LLM provider.

## Antska MCP Implementation

Antska implements the MCP specification using a custom implementation with Server-Sent Events (SSE) for streaming. This provides a lightweight, standard-compliant way to build MCP servers in Go without external dependencies.

### MCP Architecture

Antska's MCP implementation consists of several key components:

1. **MCPServer**: The core server that manages SSE connections and HTTP handlers
2. **ConnectorRegistry**: Registry for database connectors that allows creating connections to databases
3. **Connector Interface**: Interface that database drivers implement to provide standardized access
4. **Tool Registry**: Manages available MCP tools and their handlers

The architecture allows for:

1. **Database Connectors**: Each connected database can be accessed via the connector registry
2. **Query Tools**: Tools for executing queries against databases
3. **Schema Tools**: Tools for exploring database schemas
4. **Management Tools**: Tools for managing database connections and jobs

Example of the MCPServer implementation:

```go
// MCPServer represents the MCP server implementation
type MCPServer struct {
    addr          string
    dbRegistry    ConnectorRegistry
    clients       sync.Map
    clientCounter int
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(host string, port int, registry ConnectorRegistry) *MCPServer {
    return &MCPServer{
        addr:         fmt.Sprintf("%s:%d", host, port),
        dbRegistry:   registry,
        clients:      sync.Map{},
    }
}

// Start starts the MCP server
func (s *MCPServer) Start() error {
    mux := http.NewServeMux()
    
    // SSE endpoint for tool events
    mux.HandleFunc("/sse", s.handleSSE)
    
    // HTTP endpoints for tool calls
    mux.HandleFunc("/api/v1/tools/execute_query", s.handleExecuteQueryHTTP)
    mux.HandleFunc("/api/v1/tools/list_schemas", s.handleListSchemasHTTP)
    mux.HandleFunc("/api/v1/tools/list_tables", s.handleListTablesHTTP)
    mux.HandleFunc("/api/v1/tools/get_table_info", s.handleGetTableInfoHTTP)
    
    // Tool discovery endpoint
    mux.HandleFunc("/api/v1/tools", s.handleToolDiscoveryHTTP)
    
    log.Printf("Starting MCP server on %s\n", s.addr)
    return http.ListenAndServe(s.addr, mux)
}
```

### MCP Tool Registration

Antska registers tools that LLMs can invoke to perform operations on databases:

```go
// ToolRegistry holds available MCP tools
type ToolRegistry struct {
    tools          map[string]Tool
    connectorMgr   ConnectorManager
}

// Tool represents an MCP tool
type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Parameters  interface{} `json:"parameters"`
    Handler     ToolHandler
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

    // Register other tools...
}
```

### MCP Request Flow

When a client sends an MCP request to Antska, the following flow occurs:

1. The HTTP request is received by the appropriate endpoint handler
2. The handler parses the request parameters and validates them
3. The handler calls the corresponding tool handler function
4. The tool handler retrieves the appropriate database connector from the registry
5. The connector executes the requested operation on the database
6. Results are formatted as JSON and returned to the client
7. For SSE clients, events are also pushed via the event stream

Here's a simplified example of an HTTP handler:

```go
// handleExecuteQueryHTTP handles execute_query tool calls
func (s *MCPServer) handleExecuteQueryHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Read and parse request body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
        return
    }
    
    var params ExecuteQueryRequest
    if err := json.Unmarshal(body, &params); err != nil {
        http.Error(w, fmt.Sprintf("Invalid request format: %v", err), http.StatusBadRequest)
        return
    }
    
    // Execute query
    result, err := s.handleExecuteQuery(r.Context(), body)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error executing query: %v", err), http.StatusInternalServerError)
        return
    }
    
    // Return result
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "result": result,
    })
}
```

## Integration Examples

### Example 1: Querying a Database

An LLM client might make an HTTP POST request to the execute_query endpoint:

```http
POST /api/v1/tools/execute_query HTTP/1.1
Host: antska.example.com
Content-Type: application/json

{
  "database": "postgres-main",
  "query": "SELECT * FROM users LIMIT 10"
}
```

Antska processes this request and returns the results in a structured format:

```json
{
  "status": "success",
  "result": {
    "columns": ["id", "name", "email", "created_at"],
    "rows": [
      [1, "John Doe", "john@example.com", "2023-01-01T00:00:00Z"],
      [2, "Jane Smith", "jane@example.com", "2023-01-02T00:00:00Z"],
      ...
    ],
    "count": 10,
    "duration": "10.245ms"
  }
}
```

### Example 2: Schema Exploration

An LLM client might make an HTTP POST request to list the tables in a schema:

```http
POST /api/v1/tools/list_tables HTTP/1.1
Host: antska.example.com
Content-Type: application/json

{
  "database": "postgres-main",
  "schema": "public"
}
```

Antska would return:

```json
{
  "status": "success",
  "result": {
    "database": "postgres-main",
    "schema": "public",
    "tables": ["users", "orders", "products", "categories"],
    "views": ["active_users", "order_summary"]
  }
}
```

## Client Implementation Guide

To interact with Antska from your LLM application:

1. Use standard HTTP clients to make POST requests to the tool endpoints
2. Discover available tools via the `/api/v1/tools` endpoint
3. Generate appropriate tool calls based on user requests
4. Process tool responses and present results
5. For event streaming, connect to the SSE endpoint

Example using standard HTTP requests:

```python
import requests
import json

# Get available tools
tools_response = requests.get("https://antska.example.com/api/v1/tools")
tools = tools_response.json()["tools"]

# Execute a database query
query_response = requests.post(
    "https://antska.example.com/api/v1/tools/execute_query",
    json={
        "database": "postgres-main",
        "query": "SELECT * FROM users LIMIT 10"
    }
)

# Process and display results
result = query_response.json()["result"]
print(json.dumps(result, indent=2))
```

Example of connecting to the SSE endpoint:

```javascript
// Browser example
const eventSource = new EventSource('https://antska.example.com/sse');

eventSource.addEventListener('init', (event) => {
  const data = JSON.parse(event.data);
  console.log('Connected to Antska MCP server:', data);
});

eventSource.addEventListener('query_result', (event) => {
  const result = JSON.parse(event.data);
  console.log('Query result received:', result);
});

eventSource.addEventListener('error', (e) => {
  console.error('SSE connection error:', e);
});
```

## Best Practices

1. **Connection Management**: For long-running processes, use the SSE endpoint for real-time updates
2. **Error Handling**: Always check the status field in responses and handle errors appropriately
3. **Parameter Validation**: Validate parameters before sending requests to avoid validation errors
4. **Authentication**: Include authentication headers with requests when required
5. **Rate Limiting**: Respect rate limits to prevent service degradation
6. **Performance**: Use query parameters to limit result sets when dealing with large data
7. **Persistence**: Store connection information and reuse database connections when possible
