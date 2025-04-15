# Scansca API Documentation

## Overview

Scansca provides a RESTful API for managing database connections and executing queries. The API is divided into two main sections:

1. **REST API**: Standard HTTP endpoints for database operations
2. **MCP API**: Model Context Protocol endpoints for LLM integration

## REST API

### Databases

#### Register a Database

```
POST /api/v1/databases
```

Register a new database connection.

**Request Body:**

```json
{
  "name": "my-postgres",
  "type": "postgresql",
  "connection_string": "postgres://user:password@localhost:5432/dbname",
  "options": {
    "max_connections": 10
  }
}
```

**Response:**

```json
{
  "message": "Database registered successfully"
}
```

#### List Databases

```
GET /api/v1/databases
```

List all registered databases.

**Response:**

```json
{
  "databases": [
    {
      "name": "my-postgres",
      "type": "postgresql",
      "connection_string": "[REDACTED]"
    }
  ]
}
```

#### Get Database Info

```
GET /api/v1/databases/:name
```

Get detailed information about a specific database.

**Response:**

```json
{
  "name": "my-postgres",
  "type": "postgresql",
  "status": {
    "connected": true,
    "acquired_connections": 1,
    "idle_connections": 2,
    "total_connections": 3,
    "max_connections": 10,
    "database_version": "PostgreSQL 13.4"
  }
}
```

#### Remove Database

```
DELETE /api/v1/databases/:name
```

Remove a database connection.

**Response:**

```json
{
  "message": "Database removed successfully"
}
```

### Schema Introspection

#### List Schemas

```
GET /api/v1/databases/:name/schemas
```

List all schemas in a database.

**Response:**

```json
{
  "database": "my-postgres",
  "schemas": ["public", "app", "auth"]
}
```

#### List Tables

```
GET /api/v1/databases/:name/schemas/:schema/tables
```

List all tables in a schema.

**Response:**

```json
{
  "database": "my-postgres",
  "schema": "public",
  "tables": ["users", "products", "orders"],
  "views": ["active_users", "recent_orders"]
}
```

#### Get Table Info

```
GET /api/v1/databases/:name/schemas/:schema/tables/:table
```

Get detailed information about a table.

**Response:**

```json
{
  "database": "my-postgres",
  "schema": "public",
  "table": "users",
  "tableInfo": {
    "schema": "public",
    "name": "users",
    "columns": [
      {
        "name": "id",
        "data_type": "integer",
        "is_nullable": false,
        "is_primary_key": true
      },
      {
        "name": "email",
        "data_type": "varchar",
        "is_nullable": false
      }
    ],
    "primary_key": ["id"],
    "estimated_row_count": 1000
  }
}
```

### Query Execution

#### Execute Query

```
POST /api/v1/query
```

Execute a SQL query.

**Request Body:**

```json
{
  "database": "my-postgres",
  "query": "SELECT * FROM users WHERE id = $1",
  "params": [123]
}
```

**Response:**

```json
{
  "result": {
    "columns": ["id", "email", "name"],
    "rows": [
      [123, "user@example.com", "John Doe"]
    ],
    "affected_rows": 1,
    "execution_time": 0.005
  }
}
```

## MCP API

### List Tools

```
GET /mcp/v1/tools
```

List all available MCP tools.

**Response:**

```json
{
  "tools": [
    {
      "name": "execute_query",
      "description": "Execute a SQL query on a registered database",
      "parameters": {
        "database": {
          "type": "string",
          "description": "Database name"
        },
        "query": {
          "type": "string",
          "description": "SQL query to execute"
        },
        "params": {
          "type": "array",
          "description": "Query parameters (optional)"
        }
      }
    },
    {
      "name": "list_schemas",
      "description": "List all schemas in a database",
      "parameters": {
        "database": {
          "type": "string",
          "description": "Database name"
        }
      }
    }
  ]
}
```

### Invoke Tool

```
POST /mcp/v1/tools/:name/invoke
```

Invoke a specific MCP tool.

**Request Body:**

```json
{
  "database": "my-postgres",
  "query": "SELECT current_database(), current_user"
}
```

**Response:**

```json
{
  "status": "success",
  "result": {
    "columns": ["current_database", "current_user"],
    "rows": [
      ["scansca", "scansca_user"]
    ],
    "affected_rows": 1,
    "execution_time": 0.002
  }
}
```

## Error Handling

All API endpoints return appropriate HTTP status codes:

- `200 OK`: Successful operation
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid parameters
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

Error responses include a JSON body with an error message:

```json
{
  "error": "Database 'unknown-db' not found"
}
```

## Future Enhancements

Planned API enhancements include:

1. Authentication/Authorization
2. Batch operations
3. Additional database connectors
4. Query history and saved queries