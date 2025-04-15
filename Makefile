BINARY_NAME=scansca
BUILD_DIR=bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=${VERSION}"

.PHONY: all build clean test run docker docker-compose

all: clean build

build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./cmd

clean:
	@echo "Cleaning..."
	@rm -rf ${BUILD_DIR}

test:
	@echo "Running tests..."
	@go test -v ./...

run: build
	@echo "Running ${BINARY_NAME}..."
	@./${BUILD_DIR}/${BINARY_NAME}

docker:
	@echo "Building Docker image..."
	@docker build -t ${BINARY_NAME}:${VERSION} .

docker-compose:
	@echo "Starting with Docker Compose..."
	@docker-compose -f docker/docker-compose.yaml up -d

docker-compose-down:
	@echo "Stopping Docker Compose services..."
	@docker-compose -f docker/docker-compose.yaml down

lint:
	@echo "Running linters..."
	@golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Vetting code..."
	@go vet ./...

deps:
	@echo "Downloading dependencies..."
	@go mod download

tidy:
	@echo "Tidying dependencies..."
	@go mod tidy