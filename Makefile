# =============================================================================
# DataLens 2.0 — Makefile
# =============================================================================

.PHONY: help build test lint dev clean migrate seed docker-up docker-down fmt vet

# Default target
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- Build ---
build: ## Build all binaries
	go build -o bin/api ./cmd/api
	go build -o bin/agent ./cmd/agent
	go build -o bin/migrate ./cmd/migrate

build-api: ## Build API server only
	go build -o bin/api ./cmd/api

build-agent: ## Build agent only
	go build -o bin/agent ./cmd/agent

# --- Development ---
dev: docker-up ## Start full development stack
	@echo "Starting API server..."
	go run ./cmd/api

dev-agent: ## Start agent in development mode
	go run ./cmd/agent

# --- Testing ---
test: ## Run all tests
	go test ./... -v -race -count=1

test-cover: ## Run tests with coverage
	go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-unit: ## Run unit tests only
	go test ./internal/... ./pkg/... -v -race -short

test-integration: ## Run integration tests
	go test ./test/integration/... -v -race -count=1

# --- Code Quality ---
fmt: ## Format code
	gofmt -s -w .
	goimports -w .

vet: ## Run go vet
	go vet ./...

lint: ## Run linter
	golangci-lint run ./...

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

# --- Database ---
migrate: ## Run database migrations
	go run ./cmd/migrate up

migrate-down: ## Rollback last migration
	go run ./cmd/migrate down

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=migration_name"; exit 1; fi
	@touch migrations/$$(date +%Y%m%d%H%M%S)_$(NAME).up.sql
	@touch migrations/$$(date +%Y%m%d%H%M%S)_$(NAME).down.sql
	@echo "Created migration: $(NAME)"

seed: ## Seed development data
	go run ./scripts/seed.go

# --- Docker ---
docker-up: ## Start infrastructure (Postgres, Redis, NATS)
	docker compose -f docker-compose.dev.yml up -d

docker-down: ## Stop infrastructure
	docker compose -f docker-compose.dev.yml down

docker-clean: ## Stop infrastructure and remove volumes
	docker compose -f docker-compose.dev.yml down -v

docker-build: ## Build production Docker image
	docker build -t datalens:latest .

# --- Utilities ---
clean: ## Clean build artifacts
	rm -rf bin/ dist/ coverage.out coverage.html

deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

generate: ## Run code generation
	go generate ./...

# --- Production Docker ---
build-docker-prod: ## Build production Docker images locally
	docker build -t datalens-api:local -f Dockerfile .
	docker build -t datalens-frontend:local -f frontend/Dockerfile ./frontend

docker-prod-up: ## Start production stack
	docker compose -f docker-compose.prod.yml up -d

docker-prod-down: ## Stop production stack
	docker compose -f docker-compose.prod.yml down

docker-prod-logs: ## View production stack logs
	docker compose -f docker-compose.prod.yml logs -f

