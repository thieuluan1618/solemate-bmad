.PHONY: help build run test clean docker-build docker-up docker-down migrate

# Variables
SERVICES = user-service product-service
GO = go
GOFLAGS = -v
DOCKER_COMPOSE = docker-compose

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		cd services/$$service && $(GO) build $(GOFLAGS) -o bin/$$service ./cmd/main.go && cd ../..; \
	done

run-%: ## Run a specific service (e.g., make run-user-service)
	@echo "Running $*..."
	@cd services/$* && $(GO) run ./cmd/main.go

test: ## Run all tests
	@echo "Running tests..."
	@$(GO) test -v -cover -race ./...

test-service: ## Test a specific service (e.g., make test-service SERVICE=user-service)
	@echo "Testing $(SERVICE)..."
	@cd services/$(SERVICE) && $(GO) test -v -cover -race ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@goimports -w .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf services/*/bin
	@$(GO) clean -cache

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@$(DOCKER_COMPOSE) build

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	@$(DOCKER_COMPOSE) up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) down

docker-logs: ## View logs for all services
	@$(DOCKER_COMPOSE) logs -f

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@migrate -path ./migrations -database "postgresql://solemate:password@localhost:5432/solemate_db?sslmode=disable" up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@migrate -path ./migrations -database "postgresql://solemate:password@localhost:5432/solemate_db?sslmode=disable" down

proto: ## Generate protobuf files
	@echo "Generating proto files..."
	@protoc --go_out=. --go-grpc_out=. proto/*.proto

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

.DEFAULT_GOAL := help