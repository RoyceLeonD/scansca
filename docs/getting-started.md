# Getting Started with Scansca

This guide will help you set up and run Scansca on your local machine.

## Installation

### Prerequisites

1. **Go 1.22+**: [Install Go](https://golang.org/doc/install)
2. **Docker** (optional): [Install Docker](https://docs.docker.com/get-docker/)
3. **Docker Compose** (optional): [Install Docker Compose](https://docs.docker.com/compose/install/)

### Clone the Repository

```bash
git clone https://github.com/royceleond/scansca.git
cd scansca
```

### Install Dependencies

```bash
make deps
```

## Running Scansca

### Start a Database (Optional)

Scansca includes a Docker Compose file to start a PostgreSQL database for testing:

```bash
make docker-compose
```

This command starts a PostgreSQL instance with:
- Username: `scansca_user`
- Password: `scansca_password`
- Database: `scansca`
- Port: `5432`

### Build and Run the Server

```bash
make build  # Build the binary
make run    # Run the server
```

The server will start on `http://localhost:8080` by default.

## Basic Operations

### Register a Database

Register your database with the Scansca server:

```bash
curl -X POST http://localhost:8080/api/v1/databases \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-postgres",
    "type": "postgresql",
    "connection_string": "postgres://scansca_user:scansca_password@localhost:5432/scansca"
  }'
```

### List Available Databases

Check your registered databases:

```bash
curl http://localhost:8080/api/v1/databases
```

### Explore Database Schema

List schemas:

```bash
curl http://localhost:8080/api/v1/databases/my-postgres/schemas
```

List tables in a schema:

```bash
curl http://localhost:8080/api/v1/databases/my-postgres/schemas/public/tables
```

Get table information:

```bash
curl http://localhost:8080/api/v1/databases/my-postgres/schemas/public/tables/users
```

### Execute a Query

Run a SQL query against your database:

```bash
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "database": "my-postgres",
    "query": "SELECT current_database(), current_user"
  }'
```

## Configuration

Scansca uses a configuration file located at `config/scansca.yaml`. You can modify this file to change server settings:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
```

You can also use environment variables to override these settings:

```bash
SCANSCA_SERVER_PORT=9000 make run
```

## Using with LLM Clients

Scansca implements the Model Context Protocol (MCP), making it compatible with MCP-enabled LLM clients. For LLM integration, use the MCP endpoints:

```bash
# List available tools
curl http://localhost:8080/mcp/v1/tools

# Execute a query using the MCP interface
curl -X POST http://localhost:8080/mcp/v1/tools/execute_query/invoke \
  -H "Content-Type: application/json" \
  -d '{
    "database": "my-postgres",
    "query": "SELECT current_database(), current_user"
  }'
```

## Next Steps

- Check the [API documentation](api.md) for detailed API information
- Explore the project structure to understand components
- Connect multiple databases to test cross-database functionality
- Try connecting an MCP-compatible LLM client

## Troubleshooting

### Common Issues

1. **Connection Refused**: Ensure your database server is running and accessible.
2. **Authentication Failed**: Check your database connection string credentials.
3. **Missing Dependencies**: Run `make deps` to ensure all dependencies are installed.

### Logs

Scansca logs to standard output. Increase verbosity by setting the log level in the configuration.

### Getting Help

For more help, please open an issue on the GitHub repository.