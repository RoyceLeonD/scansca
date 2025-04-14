# Scansca - MCP Server for Database Intelligence

Scansca is a self-hostable Model Context Protocol (MCP) server that bridges diverse database systems with modern Large Language Model (LLM) clients. It empowers technical users to gain deep, integrated insights from heterogeneous data environments through natural language.

## Core Features

- **MCP Protocol Support**: Implements the Model Context Protocol for seamless integration with LLM clients
- **Multi-Database Connectivity**: Connect to PostgreSQL, MySQL, MariaDB, SQLite, and DynamoDB
- **Unified Query Interface**: Execute complex cross-database queries through a single interface
- **Schema Introspection**: Automatically discover and expose database structures
- **Scheduled Operations**: Configure recurring tasks for data monitoring and maintenance
- **RESTful API**: Comprehensive HTTP API for programmatic interaction

## Architecture

Scansca consists of several key components:

1. **MCP Server**: Handles client connections and implements the MCP protocol
2. **Model Context Interface (MCI)**: HTTP API for query execution and resource management
3. **Scansca Management Layer (SML)**: Manages database connections, scheduling, and state
4. **Database Connectors**: Unified interfaces for different database systems

```
+----------------------------------+
|          MCP Client              |
| (LLM Integration, Query requests)|
+----------------+-----------------+
                 |
                 v
+----------------------------------+
|        Scansca MCP Server        |
| (mark3labs/mcp-go SDK integration|
| tool/resource registration,      |
| MCP protocol compliance)         |
+----------------+-----------------+
                 |
                 v
+----------------------------------+
|    Model Context Interface (MCI) |
| (HTTP API: Query execution,      |
| schema introspection, resource   |
| management, middleware handling) |
+----------------+-----------------+
                 |
                 v
+----------------------------------+
|    Scansca Management Layer (SML) |
| (Database registration,           |
| chron scheduling, state handling) |
+----------------+-----------------+
                 |
                 v
+----------------------------------+
|         Database Connectors      |
|  (Postgres, MariaDB, SQLite,     |
|    MySQL, DynamoDB connectors)   |
+----------------------------------+
```

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose (for running databases locally)

### Installation

```bash
# Clone the repository
git clone https://github.com/royceleond/scansca.git
cd scansca

# Install dependencies
go mod download

# Start the server
go run cmd/server/main.go
```

## Usage

### Registering a Database

```bash
curl -X POST http://localhost:8080/api/v1/databases \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-postgres",
    "type": "postgresql",
    "connection_string": "postgres://user:password@localhost:5432/dbname"
  }'
```

### Executing a Query

```bash
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "database": "my-postgres",
    "query": "SELECT * FROM users LIMIT 10"
  }'
```

### Using with LLM Clients

Scansca can be used with any MCP-compatible LLM client. See the [documentation](./documentation) for integration examples.

## Development

### Project Structure

- `cmd/server/` - Server entry point
- `internal/` - Internal packages
  - `connectors/` - Database connector implementations
  - `sml/` - Scansca Management Layer
  - `mci/` - Model Context Interface
  - `mcp/` - MCP protocol implementation
- `pkg/` - Public packages for client usage
- `docker/` - Docker configurations
- `documentation/` - Project documentation

### Building from Source

```bash
go build -o scansca-server cmd/server/main.go
```

## Documentation

For detailed documentation, see the [documentation directory](./documentation).

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.