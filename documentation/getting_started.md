# Getting Started with Antska

This guide will help you set up and start using Antska, an MCP server for database intelligence.

## Prerequisites

Before you begin, make sure you have the following installed:

- Go 1.22 or higher
- Docker and Docker Compose
- Git

## Installation

### Option 1: Clone and Build from Source

```bash
# Clone the repository
git clone https://github.com/royceleond/antska.git
cd antska

# Install dependencies
go mod download

# Build the binary
go build -o antska-server cmd/antska-server/main.go

# Run the server
./antska-server
```

### Option 2: Using Docker

```bash
# Clone the repository
git clone https://github.com/royceleond/antska.git
cd antska

# Build and run with Docker Compose
docker-compose -f docker/docker-compose.yaml up -d
```

## Configuration

Antska uses a configuration file to manage settings. By default, it looks for a file named `config.yaml` in the current directory.

Create a configuration file with the following structure:

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  debug: false

storage:
  type: "postgres"
  connection_string: "postgres://antska_user:antska_password@localhost:5432/antska"

security:
  enable_auth: true
  jwt_secret: "your-secret-key"
  token_expiry: "24h"

logging:
  level: "info"
  format: "json"
```

## Setting Up Your First Database Connection

### 1. Start the Antska Server

```bash
./antska-server --config config.yaml
```

### 2. Register a Database

Use the REST API to register a database:

```bash
curl -X POST http://localhost:8080/api/v1/databases \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-postgres",
    "type": "postgresql",
    "connection_string": "postgres://user:password@localhost:5432/dbname"
  }'
```

Or use the Antska CLI:

```bash
antska-cli database add \
  --name "my-postgres" \
  --type "postgresql" \
  --connection "postgres://user:password@localhost:5432/dbname"
```

### 3. Verify the Connection

```bash
curl -X GET http://localhost:8080/api/v1/databases/my-postgres/ping
```

You should receive a successful response if the connection is established.

## Using Antska with MCP Clients

Antska implements the Model Context Protocol (MCP), allowing it to be used with any MCP-compatible client.

### Example: Using with an LLM Client

1. Connect your LLM client to the Antska MCP endpoint:
   ```
   http://localhost:8080/mcp
   ```

2. The LLM can now discover and use the database tools provided by Antska.

3. Example request from an LLM to query a database:
   ```json
   {
     "tool": "execute_query",
     "parameters": {
       "database_id": "my-postgres",
       "query": "SELECT * FROM users LIMIT 10"
     }
   }
   ```

## Setting Up Scheduled Jobs

Antska can schedule recurring database operations:

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "daily-health-check",
    "database_id": "my-postgres",
    "type": "health_check",
    "schedule": "0 0 * * *",
    "parameters": {
      "timeout": "30s"
    }
  }'
```

## Schema Exploration

Explore the schema of a registered database:

```bash
# List all schemas
curl -X GET http://localhost:8080/api/v1/databases/my-postgres/schemas

# List tables in a specific schema
curl -X GET http://localhost:8080/api/v1/databases/my-postgres/schemas/public/tables

# Get details of a specific table
curl -X GET http://localhost:8080/api/v1/databases/my-postgres/schemas/public/tables/users
```

## Executing Queries

Execute SQL queries against a registered database:

```bash
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "database_id": "my-postgres",
    "query": "SELECT * FROM users WHERE created_at > $1",
    "parameters": ["2023-01-01"]
  }'
```

## Cross-Database Queries

Antska can execute queries across multiple databases:

```bash
curl -X POST http://localhost:8080/api/v1/cross-query \
  -H "Content-Type: application/json" \
  -d '{
    "name": "user-order-join",
    "steps": [
      {
        "database_id": "postgres-users",
        "query": "SELECT id, email FROM users WHERE region = $1",
        "parameters": ["us-west"],
        "output": "users"
      },
      {
        "database_id": "mysql-orders",
        "query": "SELECT * FROM orders WHERE user_id IN ($1)",
        "parameters_from": {
          "source": "users",
          "column": "id",
          "join_as": "user_id"
        },
        "output": "orders"
      }
    ],
    "result": "orders"
  }'
```

## Monitoring and Management

Antska provides endpoints for monitoring and managing the server:

```bash
# Server status
curl -X GET http://localhost:8080/api/v1/status

# View logs
curl -X GET http://localhost:8080/api/v1/logs?level=error&limit=100

# List active connections
curl -X GET http://localhost:8080/api/v1/connections
```

## Next Steps

- Explore the [API Reference](api_reference.md) for detailed endpoint documentation
- Learn about [MCP Integration](mcp_integration.md) for LLM clients
- Check the [Use Cases](use_cases.md) for inspiration on using Antska
- View the [Technology Stack](tech_stack.md) for implementation details