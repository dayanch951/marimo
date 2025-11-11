.PHONY: help test test-unit test-integration test-e2e test-load test-coverage clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Testing
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@cd shared && go test -v -cover ./...
	@cd services/users && go test -v -cover ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@cd tests/integration && go test -v -cover ./...

test-e2e: ## Run E2E tests
	@echo "Running E2E tests..."
	@./tests/e2e/auth_flow_test.sh

test-load: ## Run load tests with k6
	@echo "Running load tests..."
	@k6 run tests/load/auth_load_test.js

test-stress: ## Run stress tests with k6
	@echo "Running stress tests..."
	@k6 run tests/load/stress_test.js

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@cd shared && go test -coverprofile=coverage.out ./...
	@cd shared && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: shared/coverage.html"

# Development
build: ## Build all services
	@echo "Building all services..."
	@cd services/gateway && go build -o ../../bin/gateway cmd/server/main.go
	@cd services/users && go build -o ../../bin/users cmd/server/main.go
	@cd services/config && go build -o ../../bin/config cmd/server/main.go
	@cd services/accounting && go build -o ../../bin/accounting cmd/server/main.go
	@cd services/factory && go build -o ../../bin/factory cmd/server/main.go
	@cd services/shop && go build -o ../../bin/shop cmd/server/main.go
	@cd services/main && go build -o ../../bin/main cmd/server/main.go

run-gateway: ## Run API Gateway
	@cd services/gateway && go run cmd/server/main.go

run-users: ## Run Users service
	@cd services/users && go run cmd/server/main.go

# Docker
docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## Show logs from all services
	docker-compose logs -f

docker-build: ## Build all Docker images
	docker-compose build

# Database
db-init: ## Initialize database
	@./scripts/init-db.sh

db-reset: ## Reset database (WARNING: deletes all data)
	@./scripts/reset-db.sh

# SSL
ssl-generate: ## Generate self-signed SSL certificates
	@./scripts/generate-ssl-certs.sh

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f shared/coverage.out shared/coverage.html
	@find . -name "*.test" -delete
	@find . -name "coverage.out" -delete
