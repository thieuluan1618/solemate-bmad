#!/bin/bash
set -e

# Update Dockerfiles for all services with production optimizations
SERVICES=("product-service" "cart-service" "order-service" "payment-service")

for service in "${SERVICES[@]}"; do
    echo "Updating Dockerfile for $service..."

    cat > "services/$service/Dockerfile" << EOF
# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations for production
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \\
    -ldflags='-w -s -extldflags "-static"' \\
    -a -installsuffix cgo \\
    -o $service ./services/$service/cmd/main.go

# Final stage - use distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot

# Copy timezone data and certificates
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder stage
COPY --from=builder /app/$service /app/$service

# Use non-root user for security
USER nonroot:nonroot

# Expose service-specific port
EXPOSE 808\$(echo \$service | cut -d'-' -f1 | wc -c)

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \\
    CMD ["/app/$service", "-health-check"]

# Command to run
ENTRYPOINT ["/app/$service"]
EOF
done

echo "API Gateway Dockerfile..."
cat > "api-gateway/Dockerfile" << EOF
# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations for production
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \\
    -ldflags='-w -s -extldflags "-static"' \\
    -a -installsuffix cgo \\
    -o api-gateway ./api-gateway/cmd/main.go

# Final stage - use distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot

# Copy timezone data and certificates
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder stage
COPY --from=builder /app/api-gateway /app/api-gateway

# Use non-root user for security
USER nonroot:nonroot

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \\
    CMD ["/app/api-gateway", "-health-check"]

# Command to run
ENTRYPOINT ["/app/api-gateway"]
EOF

echo "All Dockerfiles updated successfully!"