# Include .env file if it exists
-include .env

# Set defaults (overridden by .env if present)
BINARY_NAME?=scansca
BUILD_DIR?=bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS?=-ldflags "-X main.version=${VERSION}"
DOCS_PORT?=8090
DOCS_IMAGE_NAME?=scansca-docs
DOCS_STATIC_IMAGE_NAME?=scansca-docs-static
DOCKER_REGISTRY?=docker.example.com

# Define all targets as phony
.PHONY: all build clean test run deps tidy \
        docker docker-compose docker-compose-down \
        lint fmt vet \
        docs docs-image docs-stop docs-publish

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
	@docker build -t ${DOCS_IMAGE_NAME}:latest -f docs-config/Dockerfile docs-config

# Stop documentation server
docs-stop:
	@echo "Stopping any running docs containers..."
	@docker stop scansca-docs-container 2>/dev/null || true
	@docker rm scansca-docs-container 2>/dev/null || true

# Start documentation server
docs: docs-image docs-stop
	@echo "Starting documentation server on port ${DOCS_PORT}..."
	@docker run --rm -d --name scansca-docs-container -p ${DOCS_PORT}:8000 -v $(PWD)/docs:/content -v $(PWD)/docs-config:/config ${DOCS_IMAGE_NAME}:latest
	@echo "Documentation server running at http://localhost:${DOCS_PORT}"
	@echo "Press Ctrl+C to stop viewing this message (docs will continue running in background)"
	@echo "Run 'make docs-stop' to stop the documentation server"

# Build and publish static documentation
docs-publish:
	@echo "Building static documentation image..."
	@docker build -t ${DOCS_STATIC_IMAGE_NAME}:latest -f docs-config/Dockerfile.static .
	@echo "Tagging image for registry..."
	@docker tag ${DOCS_STATIC_IMAGE_NAME}:latest ${DOCKER_REGISTRY}/${BINARY_NAME}-docs:latest
	@echo "Pushing to registry..."
	@docker push ${DOCKER_REGISTRY}/${BINARY_NAME}-docs:latest
	@echo "Documentation successfully published to ${DOCKER_REGISTRY}/${BINARY_NAME}-docs:latest"
ifdef DEPLOY_ENDPOINT_DOCS
	@echo "Waiting 5 seconds before triggering deployment..."
	@sleep 5
	@echo "Triggering deployment to production..."
	@curl -s -X POST ${DEPLOY_ENDPOINT_DOCS} -H "Content-Type: application/json" -d '{"image":"${DOCKER_REGISTRY}/${BINARY_NAME}-docs:latest","source":"makefile"}' || echo "Deployment trigger failed!"
	@echo "Deployment triggered successfully."
endif