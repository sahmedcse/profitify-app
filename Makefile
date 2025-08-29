# Profitify Application Makefile
# Provides commands for development, testing, and deployment

.PHONY: help
help: ## Show this help message
	@echo "Profitify Application - Available Commands"
	@echo "==========================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Variables
# =============================================================================
BACKEND_DIR := backend/
FRONTEND_DIR := frontend/
DOCKER_COMPOSE := docker-compose
GO := go
YARN := yarn
NPM := npm

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

# =============================================================================
# Backend Commands
# =============================================================================

.PHONY: backend-run
backend-run: ## Run backend server in development mode
	@echo "$(GREEN)Starting backend server...$(NC)"
	@cd $(BACKEND_DIR) && $(GO) run main.go

.PHONY: backend-build
backend-build: ## Build backend binary
	@echo "$(GREEN)Building backend binary...$(NC)"
	@cd $(BACKEND_DIR) && $(GO) build -o bin/profitify-backend main.go
	@echo "$(GREEN)Binary created at backend/bin/profitify-backend$(NC)"

.PHONY: backend-test
backend-test: ## Run backend tests
	@echo "$(GREEN)Running backend tests...$(NC)"
	@cd $(BACKEND_DIR) && $(GO) test -v ./...

.PHONY: backend-test-coverage
backend-test-coverage: ## Run backend tests with coverage report
	@echo "$(GREEN)Running backend tests with coverage...$(NC)"
	@cd $(BACKEND_DIR) && $(GO) test -coverprofile=coverage.out ./...
	@cd $(BACKEND_DIR) && $(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated at backend/coverage.html$(NC)"

.PHONY: backend-lint
backend-lint: ## Run Go linter
	@echo "$(GREEN)Running Go linter...$(NC)"
	@cd $(BACKEND_DIR) && golangci-lint run ./...

.PHONY: backend-fmt
backend-fmt: ## Format Go code
	@echo "$(GREEN)Formatting Go code...$(NC)"
	@cd $(BACKEND_DIR) && $(GO) fmt ./...
	@cd $(BACKEND_DIR) && goimports -w .

.PHONY: backend-deps
backend-deps: ## Download backend dependencies
	@echo "$(GREEN)Downloading Go dependencies...$(NC)"
	@cd $(BACKEND_DIR) && $(GO) mod download
	@cd $(BACKEND_DIR) && $(GO) mod tidy

.PHONY: backend-benchmark
backend-benchmark: ## Run backend benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@cd $(BACKEND_DIR) && $(GO) test -bench=. -benchmem ./...

# =============================================================================
# Frontend Commands
# =============================================================================

.PHONY: frontend-install
frontend-install: ## Install frontend dependencies
	@echo "$(GREEN)Installing frontend dependencies...$(NC)"
	@cd $(FRONTEND_DIR) && $(YARN) install

.PHONY: frontend-run
frontend-run: ## Run frontend development server
	@echo "$(GREEN)Starting frontend development server...$(NC)"
	@cd $(FRONTEND_DIR) && $(YARN) dev

.PHONY: frontend-build
frontend-build: ## Build frontend for production
	@echo "$(GREEN)Building frontend for production...$(NC)"
	@cd $(FRONTEND_DIR) && $(YARN) build
	@echo "$(GREEN)Build completed at frontend/dist$(NC)"

.PHONY: frontend-preview
frontend-preview: ## Preview production build
	@echo "$(GREEN)Starting preview server...$(NC)"
	@cd $(FRONTEND_DIR) && $(YARN) preview

.PHONY: frontend-lint
frontend-lint: ## Run frontend linter
	@echo "$(GREEN)Running ESLint...$(NC)"
	@cd $(FRONTEND_DIR) && $(YARN) lint

.PHONY: frontend-type-check
frontend-type-check: ## Run TypeScript type checking
	@echo "$(GREEN)Running TypeScript type check...$(NC)"
	@cd $(FRONTEND_DIR) && npx tsc --noEmit

# =============================================================================
# Docker & Infrastructure Commands
# =============================================================================

.PHONY: docker-up
docker-up: ## Start all services with Docker Compose
	@echo "$(GREEN)Starting Docker services...$(NC)"
	@$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Services started. Waiting for initialization...$(NC)"
	@sleep 3
	@echo "$(GREEN)Services ready:$(NC)"
	@echo "  - Frontend: http://localhost:3000"
	@echo "  - Backend:  http://localhost:8080"
	@echo "  - LocalStack: http://localhost:4566"

.PHONY: docker-down
docker-down: ## Stop all Docker services
	@echo "$(YELLOW)Stopping Docker services...$(NC)"
	@$(DOCKER_COMPOSE) down

.PHONY: docker-clean
docker-clean: ## Stop services and remove volumes
	@echo "$(RED)Stopping services and removing volumes...$(NC)"
	@$(DOCKER_COMPOSE) down -v

.PHONY: docker-rebuild
docker-rebuild: ## Rebuild Docker images
	@echo "$(GREEN)Rebuilding Docker images...$(NC)"
	@$(DOCKER_COMPOSE) build --no-cache

.PHONY: docker-logs
docker-logs: ## View Docker logs
	@$(DOCKER_COMPOSE) logs -f

.PHONY: docker-logs-backend
docker-logs-backend: ## View backend logs
	@$(DOCKER_COMPOSE) logs -f backend

.PHONY: docker-logs-frontend
docker-logs-frontend: ## View frontend logs
	@$(DOCKER_COMPOSE) logs -f frontend

# =============================================================================
# Database Commands
# =============================================================================

.PHONY: db-seed
db-seed: ## Seed DynamoDB with sample data
	@echo "$(GREEN)Seeding DynamoDB with sample data...$(NC)"
	@cd $(BACKEND_DIR) && $(GO) run scripts/local/seed_db.go
	@echo "$(GREEN)Database seeded successfully!$(NC)"

.PHONY: db-init
db-init: ## Initialize DynamoDB tables
	@echo "$(GREEN)Initializing DynamoDB tables...$(NC)"
	@$(DOCKER_COMPOSE) --profile tools run --rm dynamodb-init
	@echo "$(GREEN)Tables created successfully!$(NC)"

.PHONY: db-reset
db-reset: docker-clean docker-up db-init db-seed ## Reset database (clean, restart, init, seed)
	@echo "$(GREEN)Database reset complete!$(NC)"

# =============================================================================
# Development Setup Commands
# =============================================================================

.PHONY: setup
setup: ## Complete development setup
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@make docker-up
	@sleep 5
	@make db-init
	@make db-seed
	@make backend-deps
	@make frontend-install
	@echo "$(GREEN)Setup complete! You can now run:$(NC)"
	@echo "  - make dev (to start both frontend and backend)"
	@echo "  - make backend-run (backend only)"
	@echo "  - make frontend-run (frontend only)"

.PHONY: dev
dev: ## Run frontend and backend concurrently (requires concurrently)
	@echo "$(GREEN)Starting development servers...$(NC)"
	@npx concurrently --names "backend,frontend" --prefix-colors "green,blue" \
		"make backend-run" \
		"make frontend-run" \
		|| echo "$(YELLOW)Install concurrently: npm install -g concurrently$(NC)"

.PHONY: dev-backend
dev-backend: docker-up db-init db-seed backend-run ## Start backend with infrastructure

.PHONY: dev-frontend
dev-frontend: frontend-install frontend-run ## Start frontend with dependencies

# =============================================================================
# Testing Commands
# =============================================================================

.PHONY: test
test: backend-test ## Run all tests
	@echo "$(GREEN)All tests completed!$(NC)"

.PHONY: test-watch
test-watch: ## Run backend tests in watch mode
	@echo "$(GREEN)Running tests in watch mode...$(NC)"
	@cd $(BACKEND_DIR) && gotestsum --watch ./... || echo "$(YELLOW)Install gotestsum: go install gotest.tools/gotestsum@latest$(NC)"

# =============================================================================
# Code Quality Commands
# =============================================================================

.PHONY: lint
lint: backend-lint frontend-lint ## Run all linters
	@echo "$(GREEN)Linting completed!$(NC)"

.PHONY: fmt
fmt: backend-fmt ## Format all code
	@echo "$(GREEN)Code formatting completed!$(NC)"

.PHONY: quality
quality: fmt lint test ## Run all quality checks
	@echo "$(GREEN)All quality checks passed!$(NC)"

# =============================================================================
# Utility Commands
# =============================================================================

.PHONY: clean
clean: ## Clean build artifacts and dependencies
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BACKEND_DIR)/bin
	@rm -rf $(BACKEND_DIR)/coverage.*
	@rm -rf $(FRONTEND_DIR)/dist
	@rm -rf $(FRONTEND_DIR)/node_modules
	@echo "$(GREEN)Clean completed!$(NC)"

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install gotest.tools/gotestsum@latest
	@npm install -g concurrently
	@echo "$(GREEN)Tools installed!$(NC)"

.PHONY: check-tools
check-tools: ## Check if required tools are installed
	@echo "$(GREEN)Checking required tools...$(NC)"
	@command -v docker >/dev/null 2>&1 && echo "✓ Docker" || echo "✗ Docker (required)"
	@command -v docker-compose >/dev/null 2>&1 && echo "✓ Docker Compose" || echo "✗ Docker Compose (required)"
	@command -v go >/dev/null 2>&1 && echo "✓ Go" || echo "✗ Go (required)"
	@command -v yarn >/dev/null 2>&1 && echo "✓ Yarn" || echo "✗ Yarn (required)"
	@command -v golangci-lint >/dev/null 2>&1 && echo "✓ golangci-lint" || echo "✗ golangci-lint (optional)"
	@command -v goimports >/dev/null 2>&1 && echo "✓ goimports" || echo "✗ goimports (optional)"
	@command -v gotestsum >/dev/null 2>&1 && echo "✓ gotestsum" || echo "✗ gotestsum (optional)"
	@command -v concurrently >/dev/null 2>&1 && echo "✓ concurrently" || echo "✗ concurrently (optional)"

.PHONY: info
info: ## Show project information
	@echo "$(GREEN)Profitify Application$(NC)"
	@echo "===================="
	@echo "Backend:  Go $(shell cd $(BACKEND_DIR) && go version | cut -d' ' -f3)"
	@echo "Frontend: Node $(shell node --version), Yarn $(shell yarn --version)"
	@echo "Database: DynamoDB (LocalStack)"
	@echo ""
	@echo "Ports:"
	@echo "  - Frontend: 3000"
	@echo "  - Backend:  8080"
	@echo "  - LocalStack: 4566"

# =============================================================================
# CI/CD Commands
# =============================================================================

.PHONY: ci-test
ci-test: ## Run tests for CI
	@echo "$(GREEN)Running CI tests...$(NC)"
	@make backend-test-coverage
	@make frontend-type-check
	@make lint

.PHONY: build
build: backend-build frontend-build ## Build both backend and frontend
	@echo "$(GREEN)Build completed!$(NC)"

# =============================================================================
# Default target
# =============================================================================
.DEFAULT_GOAL := help