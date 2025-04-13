.PHONY: build test run clean docker-up docker-down lint vet dev release-binary generate-docs

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOTIDY = $(GOCMD) mod tidy
GOVET = $(GOCMD) vet
GOFMT = gofmt
BINARY_NAME = antska-server
DOCKER_COMPOSE = docker-compose -f docker/docker-compose.yaml

# Main build target
build:
	@echo "Building Antska MCP Server..."
	@mkdir -p bin
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/antska-server

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./cmd/...

# Run integration tests with database
test-integration: test-db-up
	@echo "Running integration tests..."
	@sleep 5  # Give the database time to start
	ANTSKA_TEST_DB_HOST=127.0.0.1 ANTSKA_TEST_DB_PORT=5433 ANTSKA_TEST_DB_USER=test_user ANTSKA_TEST_DB_PASSWORD=test_password ANTSKA_TEST_DB_NAME=antska_test $(GOTEST) -v -tags=integration ./cmd/... ./internal/...
	@$(MAKE) test-db-down

# Run code coverage
cover:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run application
run: build
	@echo "Starting Antska MCP Server..."
	./bin/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf bin
	@rm -f coverage.out coverage.html

# Setup and install dependencies
setup:
	@echo "Setting up dependencies..."
	$(GOMOD) download
	$(GOTIDY)

# Run linting tools
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w -s .

# Start Docker containers for development
docker-up:
	@echo "Starting Docker containers..."
	$(DOCKER_COMPOSE) up -d

# Start test database
test-db-up:
	@echo "Starting test database..."
	docker-compose -f docker/docker-compose.test.yaml up -d

# Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	$(DOCKER_COMPOSE) down

# Stop test database
test-db-down:
	@echo "Stopping test database..."
	docker-compose -f docker/docker-compose.test.yaml down

# Clean and rebuild Docker containers
docker-rebuild:
	@echo "Rebuilding Docker containers..."
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d --build

# Start development environment
dev: docker-up
	@echo "Starting development environment..."
	@$(MAKE) run

# Build a release binary
release-binary:
	@echo "Building release binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/antska-server
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/antska-server
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/antska-server

# Generate API documentation
generate-docs:
	@echo "Generating API docs..."
	swag init -g cmd/antska-server/main.go -o api/docs

# Quick and comprehensive targets
quick: build test

# Full developer workflow
all: lint test build

# Get started with development
dev-setup: docker-up setup
	@echo "Development environment is set up and ready!"
	@echo "Run 'make dev' to start the server"