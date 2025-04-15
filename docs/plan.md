# Scansca MCP Server Implementation Document

## Goal

Scansca aspires to be an advanced, self-hostable MCP (Model Context Protocol) server, serving as a robust bridge between diverse database systems and modern language model (LLM) clients. By seamlessly orchestrating complex cross-database queries and providing intuitive, yet powerful management functionalities, Antska will empower technical users to unlock deep, integrated insights from heterogeneous data environments. This will dramatically reduce the complexity of handling disparate databases, fostering a richer and more integrated approach to data intelligence and operational efficiency.

## High-Level Design

```plaintext
+----------------------------------+
|          MCP Client              |
| (LLM Integration, Query requests)|
+----------------+-----------------+
                 |
                 v
+----------------------------------+
|        Scansca MCP Server         |
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
|    Scansca Management Layer (AML) |
| (Database registration,          |
| chron scheduling, state handling)|
+----------------+-----------------+
                 |
                 v
+----------------------------------+
|         Database Connectors      |
|  (Postgres, MariaDB, SQLite,     |
|    MySQL, DynamoDB connectors)   |
+----------------------------------+
```

## Implementation Plan

### Step 1: Database Connectors

#### 1.1 Unified Connector Interface

- Implement interface defining methods for essential database operations:

```go
type Connector interface {
    Connect(connectionString string) error
    ListSchemas() ([]string, error)
    ListTables(schema string) ([]string, error)
    ExecuteQuery(query string) (ResultSet, error)
}
```

#### 1.2 Connectors Implementation

- PostgreSQL Connector using `pgx`:
  - Efficient handling of connection pools and transactions.
- MariaDB/MySQL Connector using `go-sql-driver/mysql`:
  - Robust handling of queries and reconnections.
- SQLite Connector using `mattn/go-sqlite3`:
  - Lightweight storage for local deployments or small datasets.
- DynamoDB Connector using AWS SDK:
  - Implement Dynamo-specific data retrieval and scans.

### Step 2: Scansca Management Layer (AML)

#### 2.1 Database Registry

- Create CRUD operations to manage connector instances dynamically:
  - Add, list, modify, remove database connectors at runtime.

#### 2.2 Chron Job Management

- Implement job scheduling using `robfig/cron` for automated tasks:
  - Regular schema introspection and synchronization.
  - Automated database health checks and alerting.

#### 2.3 State Management

- Lightweight JSON/YAML-based state management (`statemanager.go`):
  - Persistent storage of configurations and system state.
  - Efficient state recovery and initialization processes.

### Step 3: Model Context Interface (MCI)

#### 3.1 RESTful API

- Connector management endpoints:
  - Register new database connectors.
  - Retrieve list of registered connectors.
- Query execution endpoints:
  - Accept queries and route them through relevant connectors.
- Schema introspection endpoints:
  - Provide comprehensive schema details to assist clients.

#### 3.2 Middleware Integration

- Implement middleware for robust request handling:
  - Authentication and Authorization
  - Structured Logging and Metrics
  - Error Handling and Recovery

### Step 4: API Server

#### 4.1 HTTP Server Implementation

- Utilize the Gin web framework for high performance and maintainability.
- Define structured and consistent API endpoints for database operations and introspection.
- Ensure endpoints adhere to best practices for RESTful design.

### Step 5: Main Server Initialization

#### 5.1 Server Startup

- Configure main entry (`cmd/Scansca-server/main.go`):
  - Initialize configurations using dynamic registry (no hardcoded credentials).
  - Load connectors and AML modules efficiently.
  - Launch and monitor the API server instance with graceful shutdown.

### Step 6: MCP Integration

#### 6.1 MCP Protocol Compliance

- Leverage `mark3labs/mcp-go` SDK for MCP server functionalities:
  - Define and register protocol-compliant tools/resources.
  - Ensure seamless interaction between LLM clients and backend services.

#### 6.2 Documentation

- Comprehensive documentation for MCP clients:
  - SDK usage guides and integration examples.
  - Clear documentation of APIs, configurations, and error handling.

## Sanity Checking Steps

### Unit Testing

- Write exhaustive unit tests for each connector implementation.
- Ensure connector stability and predictable behavior under various scenarios.

### Integration Testing

- Develop comprehensive integration tests covering:
  - AML components (Registry, Chron management, State persistence).
  - API endpoints and middleware reliability.
  - MCP SDK integration and client-server interactions.

### Operational Checks

- Automate regular schema introspection validation.
- Conduct cross-database query execution accuracy tests.
- Continuous health monitoring of all database connections and scheduled tasks.

## Dependencies

```plaintext
go get github.com/mark3labs/mcp-go
go get github.com/gin-gonic/gin
go get github.com/spf13/viper
go get github.com/jackc/pgx/v5
go get github.com/go-sql-driver/mysql
go get github.com/mattn/go-sqlite3
go get github.com/aws/aws-sdk-go
go get github.com/robfig/cron/v3
```

This document provides an extensive, technical roadmap and clear, actionable steps for building a sophisticated, extensible Scansca MCP server designed for seamless database integration and interaction with modern LLMs.
