BINARY_NAME=scansca
BUILD_DIR=bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=${VERSION}"
DOCS_PORT?=8090

# Define all targets as phony
.PHONY: all build clean test run deps tidy \
        docker docker-compose docker-compose-down \
        lint fmt vet \
        docs docs-image docs-stop

#-------------------------------------------------------------------------------
# Build targets
#-------------------------------------------------------------------------------

# Default target: clean and build the binary
all: clean build

# Build the binary
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./cmd

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf ${BUILD_DIR}

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Build and run the binary
run: build
	@echo "Running ${BINARY_NAME}..."
	@./${BUILD_DIR}/${BINARY_NAME}

# Manage dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download

tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

#-------------------------------------------------------------------------------
# Docker targets
#-------------------------------------------------------------------------------

# Build Docker image
docker:
	@echo "Building Docker image..."
	@docker build -t ${BINARY_NAME}:${VERSION} .

# Start with Docker Compose
docker-compose:
	@echo "Starting with Docker Compose..."
	@docker-compose -f docker/docker-compose.yaml up -d

# Stop Docker Compose services
docker-compose-down:
	@echo "Stopping Docker Compose services..."
	@docker-compose -f docker/docker-compose.yaml down

#-------------------------------------------------------------------------------
# Code quality targets
#-------------------------------------------------------------------------------

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...

#-------------------------------------------------------------------------------
# Documentation targets
#-------------------------------------------------------------------------------

# Build documentation image
docs-image:
	@echo "Building docs image..."
	@docker build -t scansca-docs:latest -f docs-config/Dockerfile docs-config

# Stop documentation server
docs-stop:
	@echo "Stopping any running docs containers..."
	@docker stop scansca-docs-container 2>/dev/null || true
	@docker rm scansca-docs-container 2>/dev/null || true

# Start documentation server
docs: docs-image docs-stop
	@echo "Starting documentation server on port ${DOCS_PORT}..."
	@docker run --rm -d --name scansca-docs-container -p ${DOCS_PORT}:8000 -v $(PWD)/docs:/content -v $(PWD)/docs-config:/config scansca-docs:latest
	@echo "Documentation server running at http://localhost:${DOCS_PORT}"
	@echo "Press Ctrl+C to stop viewing this message (docs will continue running in background)"
	@echo "Run 'make docs-stop' to stop the documentation server"