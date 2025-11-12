.PHONY: help test test-unit test-integration test-e2e test-load test-coverage lint lint-fix fmt clean update-deps security-scan

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

# Code Quality
lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run ./...
	@cd frontend && npm run lint

lint-fix: ## Fix linting issues
	@echo "Fixing linting issues..."
	@golangci-lint run --fix ./...
	@cd frontend && npm run lint:fix
	@cd frontend && npm run format

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .
	@cd frontend && npm run format

# Dependencies
update-deps: ## Update all dependencies
	@echo "Updating dependencies..."
	@./scripts/update-dependencies.sh

security-scan: ## Run security scans
	@echo "Running security scans..."
	@govulncheck ./... || echo "govulncheck not installed"
	@cd frontend && npm audit

check: lint test ## Run linters and tests

ci: lint test security-scan ## Run CI checks

# Migrations
migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@./scripts/run-migrations.sh

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users)
	@test -n "$(NAME)" || (echo "Please provide NAME=migration_name" && exit 1)
	@NUM=$$(ls migrations/*.sql 2>/dev/null | wc -l | xargs expr 1 +); \
	FILE=$$(printf "migrations/%03d_$(NAME).sql" $$NUM); \
	echo "-- Up" > $$FILE; \
	echo "" >> $$FILE; \
	echo "-- Down" >> $$FILE; \
	echo "Created $$FILE"

# Backup & Restore
backup: ## Create database backup
	@echo "Creating backup..."
	@./scripts/backup.sh

restore: ## Restore from backup (usage: make restore BACKUP=file.sql.gz)
	@test -n "$(BACKUP)" || (echo "Please provide BACKUP=file.sql.gz" && exit 1)
	@./scripts/restore.sh $(BACKUP)

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f shared/coverage.out shared/coverage.html
	@find . -name "*.test" -delete
	@find . -name "coverage.out" -delete
	@rm -rf frontend/build frontend/dist frontend/node_modules/.cache
