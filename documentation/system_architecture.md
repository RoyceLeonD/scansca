# Antska System Architecture

## Overview

Antska is designed as a self-hostable Model Context Protocol (MCP) server that bridges the gap between diverse database systems and modern language model (LLM) clients. By implementing the MCP specification, Antska enables LLMs to dynamically interact with database systems through a standardized interface, allowing for complex cross-database queries and intuitive database management.

## System Architecture

```
+----------------------------------+
|          MCP Client              |
| (LLM Integration, Query requests)|
+----------------+-----------------+
                 |
                 v
+----------------------------------+
|        Antska MCP Server         |
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
|    Antska Management Layer (AML) |
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

## Component Details

### 1. MCP Client

- **Purpose**: External LLM clients that communicate with Antska using the MCP protocol
- **Responsibilities**:
  - Making query requests in natural language
  - Processing structured responses from the MCP server
  - Translating complex data needs into actionable queries
- **Examples**: LangChain agents, direct LLM integration, custom client applications

### 2. Antska MCP Server

- **Purpose**: Core server component that implements the MCP specification
- **Responsibilities**:
  - Integrating with the mark3labs/mcp-go SDK
  - Registering database tools and resources for LLM use
  - Managing protocol compliance and client communication
  - Translating natural language queries into structured database operations
- **Key Components**:
  - Protocol handlers for client communication
  - Tool registry for exposing database capabilities
  - Authentication and security layers

### 3. Model Context Interface (MCI)

- **Purpose**: HTTP API layer providing standardized access to database services
- **Responsibilities**:
  - Exposing RESTful endpoints for query execution
  - Providing schema introspection capabilities
  - Managing database resources and connections
  - Implementing middleware for security, logging, and error handling
- **Key Endpoints**:
  - `/api/v1/databases` - CRUD operations for database connections
  - `/api/v1/query` - Execute queries against registered databases
  - `/api/v1/schemas` - Retrieve database schema information
  - `/api/v1/jobs` - Manage scheduled database operations

### 4. Antska Management Layer (AML)

- **Purpose**: Core business logic for database management and operations
- **Responsibilities**:
  - Managing database registration and connection pooling
  - Scheduling and executing recurring database tasks
  - Maintaining system state and configuration
  - Orchestrating cross-database operations
- **Key Components**:
  - Database registry for connection management
  - Cron scheduler for automated tasks
  - State manager for persistent configuration

### 5. Database Connectors

- **Purpose**: Unified interfaces for connecting to different database systems
- **Responsibilities**:
  - Implementing standardized methods for database operations
  - Managing database-specific connection handling
  - Translating generic queries to database-specific syntax
  - Providing type conversion and result normalization
- **Supported Databases**:
  - PostgreSQL (using pgx)
  - MySQL/MariaDB (using go-sql-driver/mysql)
  - SQLite (using mattn/go-sqlite3)
  - DynamoDB (using AWS SDK)

## Data Flow

1. **Query Initiation**:
   - LLM client sends a natural language query via MCP protocol
   - MCP Server receives and processes the query

2. **Query Processing**:
   - MCP Server translates the natural language query to structured operations
   - Model Context Interface routes the request to appropriate endpoints

3. **Database Operations**:
   - Antska Management Layer identifies target databases and operations
   - Database Connectors execute the operations on respective databases

4. **Result Processing**:
   - Database results are collected and normalized
   - MCI formats the results according to the request specifications

5. **Response Delivery**:
   - MCP Server formats the response according to MCP protocol
   - Response is returned to the LLM client

## Security Considerations

- **Authentication**: OAuth2/JWT-based authentication for API access
- **Authorization**: Role-based access control for database operations
- **Encryption**: TLS for all communications and encrypted storage for credentials
- **Audit Logging**: Comprehensive logging of all database operations and access patterns
- **Input Validation**: Strict validation and sanitization of all inputs to prevent injection attacks

## Scalability Design

- **Horizontal Scaling**: Stateless design allows for multiple server instances
- **Connection Pooling**: Efficient database connection management
- **Caching**: Response caching for frequently accessed data
- **Asynchronous Processing**: Background job execution for resource-intensive operations

## Monitoring and Observability

- **Metrics**: Prometheus integration for performance metrics
- **Logging**: Structured logging with correlation IDs
- **Tracing**: OpenTelemetry integration for distributed tracing
- **Alerting**: Configurable alerts for system health and database issues