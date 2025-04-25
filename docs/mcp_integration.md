# MCP Integration Guide

Antska makes your databases accessible to Large Language Models (LLMs) through the Model Context Protocol (MCP). This guide shows you how to use these capabilities in your applications.

## What is the Model Context Protocol?

The Model Context Protocol (MCP) is an open standard that allows LLMs to interact with external tools and data. Think of it as a universal connector between AI and data sources.

With MCP, your LLM applications can:

- **Discover** what databases and tools are available
- **Query** those databases directly
- **Receive** structured data in a consistent format
- **Maintain** context across multiple interactions

Antska implements MCP so you can connect any compatible LLM to your databases without custom integration work.

## Connecting LLMs to Databases

Antska connects LLMs to databases through a simple, standardized interface:

![Antska Architecture](https://raw.githubusercontent.com/royceleond/scansca/main/docs/assets/mcp-architecture.png)

1. **LLM clients** send requests to Antska's MCP server
2. **Antska** processes those requests and connects to the appropriate database
3. **Databases** return data which Antska formats and returns to the client

Antska supports real-time data streaming using Server-Sent Events (SSE), making it ideal for interactive LLM applications.

## Getting Started

### 1. Start the Antska Server

```bash
# Using docker-compose
docker-compose up -d

# Or using the binary
./bin/scansca
```

### 2. Register a Database

```bash
curl -X POST http://localhost:8080/api/v1/databases \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-postgres",
    "type": "postgresql",
    "connection_string": "postgres://user:password@localhost:5432/mydb"
  }'
```

### 3. Integrate with Your LLM Application

Antska provides two integration methods:

#### Option A: Direct HTTP Requests

```python
import requests

# Execute a SQL query
response = requests.post(
    "http://localhost:8080/api/v1/tools/execute_query",
    json={
        "database": "my-postgres",
        "query": "SELECT * FROM users LIMIT 10"
    }
)

results = response.json()
print(f"Found {len(results['result']['rows'])} users")
```

#### Option B: Server-Sent Events for Streaming

```javascript
// Connect to the SSE endpoint
const eventSource = new EventSource('http://localhost:8080/sse');

// Listen for events
eventSource.addEventListener('query_result', (event) => {
  const data = JSON.parse(event.data);
  console.log(`Received ${data.count} rows`);
  updateUI(data.rows);
});

// Send a query
function executeQuery(database, query) {
  fetch('/api/v1/tools/execute_query', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ database, query })
  });
}
```

## Available Tools

Antska provides these database tools for LLMs:

| Tool | Description | Example Use |
|------|-------------|-------------|
| `execute_query` | Run SQL queries | "Show me the top 10 users by activity" |
| `list_schemas` | List available schemas | "What schemas are in this database?" |
| `list_tables` | List tables in a schema | "Show me tables in the public schema" |
| `get_table_info` | Get table structure | "What columns are in the users table?" |

## Example: Database Exploration Workflow

Here's how an LLM might explore a database using Antska:

1. **List available databases**
   ```json
   GET /api/v1/databases
   ```

2. **Discover schemas in a database**
   ```json
   POST /api/v1/tools/list_schemas
   {
     "database": "my-postgres"
   }
   ```

3. **Find tables in a schema**
   ```json
   POST /api/v1/tools/list_tables
   {
     "database": "my-postgres",
     "schema": "public"
   }
   ```

4. **Examine table structure**
   ```json
   POST /api/v1/tools/get_table_info
   {
     "database": "my-postgres",
     "schema": "public",
     "table": "users"
   }
   ```

5. **Query the table**
   ```json
   POST /api/v1/tools/execute_query
   {
     "database": "my-postgres",
     "query": "SELECT id, name, email FROM users WHERE active = true LIMIT 5"
   }
   ```

## Best Practices

To get the most out of Antska with LLMs:

1. **Start simple**: Begin with basic database queries before complex operations
2. **Use limits**: Always limit query results to prevent overwhelming responses
3. **Handle errors**: Check status codes and error messages in responses
4. **Stream large results**: Use SSE for queries that might return large datasets
5. **Cache metadata**: Store database structure information to reduce repeated calls
6. **Secure connections**: Use authentication and encryption in production environments

## Need Help?

- **Documentation**: Full [API Reference](api.md)  
- **Examples**: See [example clients](https://github.com/royceleond/scansca/tree/main/examples)
- **Community**: Join our [Discord](https://discord.gg/antska)
- **Issues**: Report bugs on [GitHub](https://github.com/royceleond/scansca/issues)