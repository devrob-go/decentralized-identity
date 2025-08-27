# Go Auth Service - Root Makefile
# This Makefile provides common commands for building, testing, and deploying all services

.PHONY: help build test clean generate deploy local staging production lint fmt security

# Default target
.DEFAULT_GOAL := help

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Variables
SERVICES := auth-service
ENVIRONMENTS := local staging production

help: ## Show this help message
	@echo "$(BLUE)Go Auth Service - Available Commands$(NC)"
	@echo ""
	@echo "$(YELLOW)Development Commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)Service Commands:$(NC)"
	@echo "  make service-build SERVICE=<service-name>    # Build specific service"
	@echo "  make service-test SERVICE=<service-name>     # Test specific service"
	@echo "  make service-clean SERVICE=<service-name>    # Clean specific service"
	@echo ""
	@echo "$(YELLOW)Environment Commands:$(NC)"
	@echo "  make deploy-local                            # Deploy to local environment"
	@echo "  make deploy-staging                          # Deploy to staging environment"
	@echo "  make deploy-production                       # Deploy to production environment"
	@echo ""
	@echo "$(YELLOW)Available Services:$(NC)"
	@for service in $(SERVICES); do \
		echo "  - $$service"; \
	done
	@echo ""
	@echo "$(YELLOW)Available Environments:$(NC)"
	@for env in $(ENVIRONMENTS); do \
		echo "  - $$env"; \
	done

# Development Commands

deps: ## Install/update dependencies for all modules
	@echo "$(YELLOW)Installing dependencies for all modules...$(NC)"
	@go work sync
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

build: ## Build all services
	@echo "$(YELLOW)Building all services...$(NC)"
	@./scripts/build.sh

test: ## Run tests for all services
	@echo "$(YELLOW)Running tests for all services...$(NC)"
	@./scripts/test.sh

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@rm -rf services/*/bin/
	@rm -rf coverage.out
	@echo "$(GREEN)✓ Cleaned build artifacts$(NC)"

generate: ## Generate protobuf and mock code
	@echo "$(YELLOW)Generating code...$(NC)"
	@./scripts/generate.sh
	@./tools/mockgen.sh
	@echo "$(GREEN)✓ Code generation completed$(NC)"

lint: ## Run linter for all services
	@echo "$(YELLOW)Running linter for all services...$(NC)"
	@for service in $(SERVICES); do \
		echo "Linting $$service..."; \
		cd services/$$service && golangci-lint run && cd ../..; \
	done
	@echo "$(GREEN)✓ Linting completed$(NC)"

fmt: ## Format code for all services
	@echo "$(YELLOW)Formatting code for all services...$(NC)"
	@go fmt ./...
	@for service in $(SERVICES); do \
		echo "Formatting $$service..."; \
		cd services/$$service && go fmt ./... && cd ../..; \
	done
	@echo "$(GREEN)✓ Code formatting completed$(NC)"

security: ## Run security checks
	@echo "$(YELLOW)Running security checks...$(NC)"
	@go vet ./...
	@for service in $(SERVICES); do \
		echo "Checking $$service..."; \
		cd services/$$service && go vet ./... && cd ../..; \
	done
	@echo "$(GREEN)✓ Security checks completed$(NC)"

# Service-specific Commands

service-build: ## Build a specific service (use SERVICE=<service-name>)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: SERVICE variable not set$(NC)"; \
		echo "Usage: make service-build SERVICE=<service-name>"; \
		echo "Available services: $(SERVICES)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Building $(SERVICE)...$(NC)"
	@cd services/$(SERVICE) && go build -o bin/$(SERVICE) ./cmd/server

service-test: ## Test a specific service (use SERVICE=<service-name>)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: SERVICE variable not set$(NC)"; \
		echo "Usage: make service-test SERVICE=<service-name>"; \
		echo "Available services: $(SERVICES)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Testing $(SERVICE)...$(NC)"
	@cd services/$(SERVICE) && go test ./...

service-clean: ## Clean a specific service (use SERVICE=<service-name>)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: SERVICE variable not set$(NC)"; \
		echo "Usage: make service-clean SERVICE=<service-name>"; \
		echo "Available services: $(SERVICES)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Cleaning $(SERVICE)...$(NC)"
	@rm -rf services/$(SERVICE)/bin/
	@echo "$(GREEN)✓ Cleaned $(SERVICE)$(NC)"

# Deployment Commands

deploy: ## Deploy all services to specified environment (use ENV=<environment>)
	@if [ -z "$(ENV)" ]; then \
		echo "$(RED)Error: ENV variable not set$(NC)"; \
		echo "Usage: make deploy ENV=<environment>"; \
		echo "Available environments: $(ENVIRONMENTS)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Deploying to $(ENV) environment...$(NC)"
	@./scripts/deploy.sh -e $(ENV)

deploy-local: ## Deploy to local environment
	@echo "$(YELLOW)Deploying to local environment...$(NC)"
	@./scripts/deploy.sh -e local

deploy-staging: ## Deploy to staging environment
	@echo "$(YELLOW)Deploying to staging environment...$(NC)"
	@./scripts/deploy.sh -e staging

deploy-production: ## Deploy to production environment
	@echo "$(YELLOW)Deploying to production environment...$(NC)"
	@./scripts/deploy.sh -e production

# Database Commands

db-start: ## Start local database
	@echo "$(YELLOW)Starting local database...$(NC)"
	@cd deployments/local && docker-compose up -d postgres

db-stop: ## Stop local database
	@echo "$(YELLOW)Stopping local database...$(NC)"
	@cd deployments/local && docker-compose stop postgres

db-reset: ## Reset local database
	@echo "$(YELLOW)Resetting local database...$(NC)"
	@cd deployments/local && docker-compose down -v postgres && docker-compose up -d postgres

db-migrate: ## Run database migrations for all services
	@echo "$(YELLOW)Running database migrations...$(NC)"
	@for service in $(SERVICES); do \
		echo "Migrating $$service..."; \
		cd services/$$service && go run cmd/migrate/main.go && cd ../..; \
	done
	@echo "$(GREEN)✓ Database migrations completed$(NC)"

# Utility Commands

proto: ## Generate protobuf code only
	@echo "$(YELLOW)Generating protobuf code...$(NC)"
	@./tools/protoc-gen-go.sh

mocks: ## Generate mock code only
	@echo "$(YELLOW)Generating mock code...$(NC)"
	@./tools/mockgen.sh

coverage: ## Generate test coverage report
	@echo "$(YELLOW)Generating test coverage report...$(NC)"
	@go test -coverpkg=./... -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

# Development workflow commands

dev-setup: ## Complete development environment setup
	@echo "$(YELLOW)Setting up development environment...$(NC)"
	@make deps
	@make generate
	@make db-start
	@echo "$(GREEN)✓ Development environment setup completed$(NC)"

dev-start: ## Start development environment
	@echo "$(YELLOW)Starting development environment...$(NC)"
	@make deploy-local
	@echo "$(GREEN)✓ Development environment started$(NC)"

dev-stop: ## Stop development environment
	@echo "$(YELLOW)Stopping development environment...$(NC)"
	@cd deployments/local && docker-compose down
	@echo "$(GREEN)✓ Development environment stopped$(NC)"

# CI/CD Commands

ci-build: ## CI build command
	@echo "$(YELLOW)Running CI build...$(NC)"
	@make deps
	@make generate
	@make build
	@make test
	@make lint
	@make security
	@echo "$(GREEN)✓ CI build completed successfully$(NC)"

ci-deploy: ## CI deploy command
	@echo "$(YELLOW)Running CI deploy...$(NC)"
	@make deploy-staging
	@echo "$(GREEN)✓ CI deploy completed successfully$(NC)"

# Documentation Commands

docs-serve: ## Serve documentation locally
	@echo "$(YELLOW)Serving documentation...$(NC)"
	@if command -v python3 &> /dev/null; then \
		cd docs && python3 -m http.server 8000; \
	elif command -v python &> /dev/null; then \
		cd docs && python -m SimpleHTTPServer 8000; \
	else \
		echo "$(RED)Python not found. Please install Python to serve documentation.$(NC)"; \
		exit 1; \
	fi

# Cleanup Commands

clean-all: ## Clean everything including Docker
	@echo "$(YELLOW)Cleaning everything...$(NC)"
	@make clean
	@cd deployments/local && docker-compose down -v
	@docker system prune -f
	@echo "$(GREEN)✓ Complete cleanup completed$(NC)"

# Status Commands

status: ## Show status of all services
	@echo "$(YELLOW)Service Status:$(NC)"
	@for service in $(SERVICES); do \
		if [ -d "services/$$service/bin" ]; then \
			echo "  $$service: $(GREEN)✓ Built$(NC)"; \
		else \
			echo "  $$service: $(RED)✗ Not built$(NC)"; \
		fi; \
	done
	@echo ""
	@echo "$(YELLOW)Local Environment Status:$(NC)"
	@cd deployments/local && docker-compose ps

# Quick commands for common tasks

quick-build: ## Quick build without tests
	@echo "$(YELLOW)Quick building all services...$(NC)"
	@make deps
	@make generate
	@make build

quick-test: ## Quick test without building
	@echo "$(YELLOW)Quick testing all services...$(NC)"
	@make test

quick-deploy: ## Quick deploy to local
	@echo "$(YELLOW)Quick deploying to local...$(NC)"
	@make deploy-local
