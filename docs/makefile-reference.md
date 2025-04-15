# Makefile Reference

Scansca includes a comprehensive Makefile to simplify development, building, testing, and documentation workflows. This document explains all available `make` targets and how to use them.

## Quick Reference

| Category | Command | Description |
|----------|---------|-------------|
| **Building** | `make build` | Build the Scansca binary |
| | `make clean` | Remove build artifacts |
| | `make all` | Clean and build |
| | `make run` | Build and run Scansca |
| **Dependencies** | `make deps` | Download Go dependencies |
| | `make tidy` | Clean up Go module files |
| **Docker** | `make docker` | Build Docker image |
| | `make docker-compose` | Start services with Docker Compose |
| | `make docker-compose-down` | Stop Docker Compose services |
| **Code Quality** | `make lint` | Run linters |
| | `make fmt` | Format code |
| | `make vet` | Run Go vet |
| | `make test` | Run tests |
| **Documentation** | `make docs` | Start documentation server |
| | `make docs-stop` | Stop documentation server |

## Building and Running

### Building the Application

To build the Scansca binary:

```bash
make build
```

This creates a binary in the `bin/` directory.

### Cleaning Build Artifacts

To remove build artifacts:

```bash
make clean
```

### Building and Running in One Step

To build and immediately run Scansca:

```bash
make run
```

## Dependency Management

### Downloading Dependencies

To download all Go module dependencies:

```bash
make deps
```

### Tidying Dependencies

To clean up unused dependencies in Go module files:

```bash
make tidy
```

## Docker Operations

### Building a Docker Image

To build a Docker image for Scansca:

```bash
make docker
```

### Running with Docker Compose

To start Scansca and its dependencies (like PostgreSQL) using Docker Compose:

```bash
make docker-compose
```

### Stopping Docker Compose Services

To stop all services started with Docker Compose:

```bash
make docker-compose-down
```

## Code Quality

### Running Linters

To run code linters:

```bash
make lint
```

### Formatting Code

To automatically format Go code:

```bash
make fmt
```

### Vetting Code

To run Go's static analysis tool:

```bash
make vet
```

### Running Tests

To run all tests:

```bash
make test
```

## Documentation

### Starting the Documentation Server

To build and start the documentation server:

```bash
make docs
```

By default, this serves documentation at http://localhost:8090. You can change the port by setting the `DOCS_PORT` environment variable:

```bash
DOCS_PORT=8888 make docs
```

### Stopping the Documentation Server

To stop the running documentation server:

```bash
make docs-stop
```

## Environment Variables

The Makefile supports several environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `VERSION` | Git describe output or "dev" | Sets the version string in the binary |
| `DOCS_PORT` | 8090 | The port for the documentation server |

## Extending the Makefile

To add new targets to the Makefile, follow these guidelines:

1. Add the target name to the `.PHONY` list
2. Place the target in the appropriate section
3. Add comments to explain what the target does
4. Follow the established pattern for echo commands

Example of adding a new target:

```makefile
# In the appropriate section
new-target:
	@echo "Doing something new..."
	@command-to-run
```