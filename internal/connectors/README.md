# Database Connectors for Antska

This package provides a unified interface for connecting to various database systems. Each connector implements a common interface, allowing Antska to interact with different database types through a consistent API.

## Architecture

### Core Interfaces

- **Connector**: Basic database connector interface with essential operations
- **AdvancedConnector**: Extended interface with advanced features and operations

### Connector Types

Currently implemented:
- **PostgreSQL**: Full-featured connector for PostgreSQL databases

Planned:
- MySQL/MariaDB
- SQLite
- DynamoDB

### Factory Pattern

The package uses a factory pattern to create the appropriate connector:

```go
// Basic factory
factory := connectors.NewConnectorFactory()
connector, err := factory.CreateConnector("postgresql")

// Advanced factory
advFactory := connectors.NewAdvancedConnectorFactory()
advConnector, err := advFactory.CreateAdvancedConnector("postgresql")
```

## Testing Framework

The connectors have two types of tests:

1. **Unit Tests**: Regular tests that don't require a database connection
2. **Integration Tests**: Tests that require a live database connection

### Running Tests

Use the provided Makefile targets:

```bash
# Run unit tests only
make test

# Run integration tests for all databases
make test-integration

# Run PostgreSQL integration tests specifically
make test-postgres

# Run MySQL integration tests (when implemented)
make test-mysql
```

### Test Database Infrastructure

The integration tests use Docker containers to provide isolated database environments:

- The containers are defined in `docker/docker-compose.test.yaml`
- Test data is loaded from `docker/test_data/{database}/init-test-data.sql`
- The `wait-for-db` target ensures the database is ready before running tests

### Test Data Structure

The test data includes:

1. Sample tables with primary/foreign keys
2. Test schema with multiple tables
3. Various data types to test type conversion
4. Example views to test view introspection

## Adding a New Connector

To add a new database connector:

1. Create a new package in `internal/connectors/{database}`
2. Implement the required interfaces
3. Add a factory method to `registry.go`
4. Update the factory cases in `factory.go`
5. Add test data in `docker/test_data/{database}/init-test-data.sql`
6. Add Docker config in `docker-compose.test.yaml`
7. Add Makefile targets for testing

Example structure for a new connector:

```
internal/connectors/
  ├── sqlite/
  │   ├── sqlite.go
  │   └── sqlite_test.go
  └── registry.go  (update with NewSQLiteConnector)
```

## Connection Configuration

Each connector can have specific configuration options. See the individual connector documentation for details:

- [PostgreSQL Configuration](postgresql/README.md)