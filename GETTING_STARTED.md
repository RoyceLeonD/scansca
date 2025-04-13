# Getting Started with Antska

This guide will help you get started with development on the Antska MCP Server project.

## Quick Start

```bash
# Clone the repository (if you haven't already)
git clone https://github.com/royceleond/antska.git
cd antska

# Set up dependencies and start containers
make dev-setup

# Build and run the server
make run
```

## Development Workflow

### Essential Commands

- `make build` - Build the project
- `make test` - Run unit tests
- `make run` - Build and run the server
- `make docker-up` - Start Docker containers with PostgreSQL
- `make docker-down` - Stop Docker containers

### Testing

- `make test` - Run basic unit tests
- `make test-integration` - Run integration tests with a test database

### Docker Environment

The project uses Docker to provide a consistent development environment:

- `make docker-up` - Start development containers
- `make docker-down` - Stop development containers  
- `make docker-rebuild` - Rebuild Docker containers

### Complete Workflow

A typical development cycle might look like this:

```bash
# Start the database
make docker-up

# Make code changes...

# Build and test
make build test

# Run the server
make run

# When finished
make docker-down
```

## Project Structure

```
antska/
├── cmd/
│   └── antska-server/   # Main application entry point
├── internal/            # Internal packages
│   ├── connectors/      # Database connector implementations
│   ├── aml/             # Antska Management Layer
│   └── mci/             # Model Context Interface
├── pkg/                 # Public packages
├── docker/              # Docker configurations
└── documentation/       # Project documentation
```

## Configuration

Configuration is stored in `config/antska.yaml`. You can customize settings there.

## Next Steps

- Check the main [README.md](README.md) for a project overview
- Review the [documentation](documentation/) for more details
- See the implementation plan in [documentation/plan.md](documentation/plan.md)