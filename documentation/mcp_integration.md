# MCP Integration Guide

This document explains how Antska implements the Model Context Protocol (MCP) to enable seamless integration between LLM clients and database systems.

## What is the Model Context Protocol (MCP)?

The Model Context Protocol (MCP) is a standardized protocol for enabling LLMs to interact with external tools and data sources. It allows LLMs to:

1. Discover available tools and resources
2. Invoke operations on these resources
3. Access structured data in a consistent format
4. Maintain context across interactions

By implementing MCP, Antska provides a standardized way for any LLM client to interact with database systems, without requiring custom integration for each LLM provider.

## Antska MCP Implementation

Antska implements the MCP specification using the [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) SDK, which provides the foundation for building MCP-compliant servers in Go.

### MCP Resource Registration

Antska registers the following types of resources with the MCP server:

1. **Database Connectors**: Each connected database is registered as a resource
2. **Query Tools**: Tools for executing queries against databases
3. **Schema Tools**: Tools for exploring database schemas
4. **Management Tools**: Tools for managing database connections and jobs

Example of registering database resources:

```go
func registerDatabaseResources(server *mcp.Server, registry *aml.Registry) error {
    databases, err := registry.ListDatabases()
    if err != nil {
        return err
    }
    
    for _, db := range databases {
        // Register database as a resource
        resource := mcp.Resource{
            ID:          db.ID,
            Name:        db.Name,
            Description: fmt.Sprintf("Connection to %s database", db.Type),
            Type:        string(db.Type),
            Metadata:    db.Metadata,
        }
        
        server.RegisterResource(resource)
    }
    
    return nil
}
```

### MCP Tool Registration

Antska registers tools that LLMs can invoke to perform operations on databases:

```go
func registerDatabaseTools(server *mcp.Server, registry *aml.Registry) error {
    // Register query execution tool
    queryTool := mcp.Tool{
        Name:        "execute_query",
        Description: "Execute SQL query against a database",
        Parameters: []mcp.Parameter{
            {
                Name:        "database_id",
                Description: "ID of the database to query",
                Type:        "string",
                Required:    true,
            },
            {
                Name:        "query",
                Description: "SQL query to execute",
                Type:        "string",
                Required:    true,
            },
        },
        Handler: handleExecuteQuery(registry),
    }
    
    server.RegisterTool(queryTool)
    
    // Register other tools...
    
    return nil
}
```

### MCP Request Flow

When a client sends an MCP request to Antska, the following flow occurs:

1. The MCP server receives the request and validates it
2. The server routes the request to the appropriate tool handler
3. The tool handler executes the requested operation using the Antska Management Layer
4. Results are formatted according to the MCP specification and returned to the client

## Integration Examples

### Example 1: Querying a Database

An LLM client might generate the following MCP request:

```json
{
  "tool": "execute_query",
  "parameters": {
    "database_id": "postgres-main",
    "query": "SELECT * FROM users LIMIT 10"
  }
}
```

Antska processes this request and returns the results in a structured format:

```json
{
  "status": "success",
  "data": {
    "columns": ["id", "name", "email", "created_at"],
    "rows": [
      [1, "John Doe", "john@example.com", "2023-01-01T00:00:00Z"],
      [2, "Jane Smith", "jane@example.com", "2023-01-02T00:00:00Z"],
      ...
    ],
    "rowCount": 10,
    "executionTime": "10ms"
  }
}
```

### Example 2: Schema Exploration

An LLM client might generate a request to explore database schema:

```json
{
  "tool": "list_tables",
  "parameters": {
    "database_id": "postgres-main",
    "schema": "public"
  }
}
```

Antska would return:

```json
{
  "status": "success",
  "data": {
    "tables": [
      {
        "name": "users",
        "schema": "public",
        "rowCount": 1000,
        "sizeBytes": 8192000
      },
      {
        "name": "orders",
        "schema": "public",
        "rowCount": 5000,
        "sizeBytes": 16384000
      },
      ...
    ]
  }
}
```

## Client Implementation Guide

To interact with Antska from your LLM application:

1. Use an MCP client library or implement the protocol directly
2. Discover available resources and tools
3. Generate appropriate tool calls based on user requests
4. Process tool responses and present results

Example using a hypothetical MCP client:

```python
from mcp_client import MCPClient

# Connect to Antska MCP server
client = MCPClient("https://antska.example.com/mcp")

# Discover available tools
tools = client.list_tools()

# Execute a database query
result = client.execute_tool(
    "execute_query",
    parameters={
        "database_id": "postgres-main",
        "query": "SELECT * FROM users LIMIT 10"
    }
)

# Process and display results
print(result.data)
```

## Best Practices

1. **Context Management**: Use resource IDs consistently to maintain context across requests
2. **Error Handling**: Properly handle error responses from the server
3. **Parameter Validation**: Validate parameters before sending requests
4. **Authentication**: Include authentication tokens with requests if required
5. **Rate Limiting**: Respect rate limits to prevent overloading the server