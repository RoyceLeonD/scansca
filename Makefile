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
BINARY_NAME = scansca-server
DOCKER_COMPOSE = docker-compose -f docker/docker-compose.yaml

# Main build target
build:
	@echo "Building Scansca MCP Server..."
	@mkdir -p bin
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./cmd/...

# Database connection parameters for tests
DB_TEST_PARAMS = SCANSCA_TEST_DB_HOST=127.0.0.1 SCANSCA_TEST_DB_PORT=5433 SCANSCA_TEST_DB_USER=test_user SCANSCA_TEST_DB_PASSWORD=test_password SCANSCA_TEST_DB_NAME=scansca_test

# Run all integration tests with databases
test-integration: test-db-up test-data-load wait-for-db
	@echo "Running integration tests..."
	$(DB_TEST_PARAMS) $(GOTEST) -v -tags=integration ./cmd/... ./internal/...
	@$(MAKE) test-db-down

# Run integration tests for a specific database connector
test-postgres: test-db-up test-data-load wait-for-db
	@echo "Running PostgreSQL integration tests..."
	$(DB_TEST_PARAMS) $(GOTEST) -v -tags=integration ./internal/connectors/postgresql
	@$(MAKE) test-db-down

# Run integration tests for MySQL (when implemented)
test-mysql: test-mysql-up test-mysql-data-load wait-for-mysql
	@echo "Running MySQL integration tests..."
	SCANSCA_TEST_DB_HOST=127.0.0.1 SCANSCA_TEST_DB_PORT=3307 SCANSCA_TEST_DB_USER=test_user SCANSCA_TEST_DB_PASSWORD=test_password SCANSCA_TEST_DB_NAME=scansca_test $(GOTEST) -v -tags=integration ./internal/connectors/mysql
	@$(MAKE) test-mysql-down

# Load test data for PostgreSQL
test-data-load: wait-for-db
	@echo "Loading test data into PostgreSQL..."
	@docker exec -i docker-postgres-test-1 psql -U test_user -d scansca_test < docker/test_data/postgresql/init-test-data.sql

# Load test data for MySQL (when implemented)
test-mysql-data-load: wait-for-mysql
	@echo "Loading test data into MySQL..."
	@docker exec -i docker-mysql-test-1 mysql -u test_user -ptest_password scansca_test < docker/test_data/mysql/init-test-data.sql

# Wait for PostgreSQL to be ready
wait-for-db:
	@echo "Waiting for PostgreSQL to be ready..."
	@for i in 1 2 3 4 5; do \
		if docker exec docker-postgres-test-1 pg_isready -U test_user -d scansca_test > /dev/null 2>&1; then \
			break; \
		fi; \
		echo "Waiting for PostgreSQL to start... ($$i/5)"; \
		sleep 2; \
	done

# Wait for MySQL to be ready
wait-for-mysql:
	@echo "Waiting for MySQL to be ready..."
	@for i in 1 2 3 4 5; do \
		if docker exec docker-mysql-test-1 mysqladmin ping -h 127.0.0.1 -u test_user -ptest_password > /dev/null 2>&1; then \
			break; \
		fi; \
		echo "Waiting for MySQL to start... ($$i/5)"; \
		sleep 2; \
	done

# Start MySQL test database
test-mysql-up:
	@echo "Starting MySQL test database..."
	@docker-compose -f docker/docker-compose.test.yaml --profile mysql up -d mysql-test

# Stop MySQL test database
test-mysql-down:
	@echo "Stopping MySQL test database..."
	@docker-compose -f docker/docker-compose.test.yaml --profile mysql down mysql-test

# Run code coverage
cover:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run application
run: build
	@echo "Starting Scansca MCP Server..."
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
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/server

# Generate API documentation
generate-docs:
	@echo "Generating API docs..."
	swag init -g cmd/server/main.go -o api/docs

# Quick and comprehensive targets
quick: build test

# Run migration to new project structure
migrate:
	@echo "Running migration to Scansca project structure..."
	./migrate_to_scansca.sh

# Full developer workflow
all: lint test build

# Get started with development
dev-setup: docker-up setup
	@echo "Development environment is set up and ready!"
	@echo "Run 'make dev' to start the server"