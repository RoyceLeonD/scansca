FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o scansca-server ./cmd/server

# Use a smaller image for the final build
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/scansca-server .

# Copy config files if needed
COPY --from=builder /app/config /app/config

# Expose necessary ports
EXPOSE 8080
EXPOSE 9000

# Set environment variables
ENV GIN_MODE=release

# Create a non-root user and switch to it
RUN addgroup -S scansca && adduser -S scansca -G scansca
USER scansca

# Run the binary
CMD ["./scansca-server"]