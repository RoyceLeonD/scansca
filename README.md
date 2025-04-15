# Scansca - Database Intelligence Platform

Scansca is a self-hostable server that connects Large Language Model (LLM) clients with database systems through the Model Context Protocol (MCP). It enables technical users to gain integrated insights from diverse data environments using natural language.

## Features

- **Multi-Database Support**: Connect to PostgreSQL (with MySQL, SQLite, and DynamoDB coming soon)
- **MCP Integration**: Seamless integration with MCP-compatible LLM clients
- **Automatic Schema Discovery**: Expose database structures with minimal configuration
- **Simple API**: RESTful interface for database operations
- **Docker Ready**: Easy deployment with Docker Compose

## Quick Start

### Prerequisites

- Go 1.22+
- Docker and Docker Compose (optional, for database dependencies)

### Installation

```bash
# Clone the repository
git clone https://github.com/royceleond/scansca.git
cd scansca

# Install dependencies
make deps

# Start a PostgreSQL database (optional)
make docker-compose

# Build and run the server
make run
```

## Basic Usage

### Register a Database

```bash
curl -X POST http://localhost:8080/api/v1/databases \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-postgres",
    "type": "postgresql",
    "connection_string": "postgres://scansca_user:scansca_password@localhost:5432/scansca"
  }'
```

### Execute a Query

```bash
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "database": "my-postgres",
    "query": "SELECT current_database(), current_user"
  }'
```

### Use with LLM Clients

For MCP-compatible LLM clients, Scansca exposes endpoints at:

```
GET  /mcp/v1/tools               # Lists available tools
POST /mcp/v1/tools/:name/invoke  # Invokes a specific tool
```

### Documentation

View the documentation by running:

```bash
make docs
```

This will start a documentation server on port 8080 (configurable with DOCS_PORT).

For more information, see the [online documentation](docs/getting-started.md).

## Architecture

```
┌──────────────────────────┐
│      MCP Client           │
│  (LLM, CLI, Application)  │
└──────────────┬───────────┘
               │
               ▼
┌──────────────────────────┐
│    Scansca MCP Server     │  HTTP API for database operations
│                           │  and MCP protocol support
├──────────────────────────┤
│    Database Connectors    │  Uniform interface for different
└──────────────┬───────────┘  database systems
               │
               ▼
┌──────────────────────────┐
│      Database Servers     │
│  (Postgres, MySQL, etc.)   │
└──────────────────────────┘
```

## Development

```bash
# Build the project
make build

# Run tests
make test

# Start the server
make run

# Start PostgreSQL using Docker Compose
make docker-compose

# Stop Docker Compose services
make docker-compose-down

# Start documentation server
make docs
```

## Project Structure

```
scansca/
├── cmd/          # Application entry point
├── config/       # Configuration files
├── docker/       # Docker configurations
├── docs/         # Documentation
├── internal/     # Private implementation
│   ├── db/       # Database connections and models
│   ├── server/   # HTTP server and API
│   └── mcp/      # MCP protocol implementation
└── Makefile      # Build automation
```

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.