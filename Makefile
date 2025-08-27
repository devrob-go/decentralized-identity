# Decentralized Identity & Authentication System Makefile

.PHONY: help build test clean deploy start stop logs
.DEFAULT_GOAL := help

# Variables
DID_MANAGER_DIR = services/did-manager
AUTH_SERVICE_DIR = services/auth-service
CONTRACTS_DIR = contracts
CLI_DIR = cli

# Colors for output
GREEN = \033[0;32m
YELLOW = \033[1;33m
RED = \033[0;31m
NC = \033[0m # No Color

help: ## Show this help message
	@echo "$(GREEN)Decentralized Identity & Authentication System$(NC)"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development Commands
dev-start: ## Start all services for development
	@echo "$(GREEN)Starting development environment...$(NC)"
	cd deployments/local && docker-compose up -d
	@echo "$(GREEN)Development environment started!$(NC)"
	@echo "Services available at:"
	@echo "  - Auth Service: http://localhost:8080"
	@echo "  - DID Manager: http://localhost:8082"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - NATS: localhost:4222"
	@echo "  - Ganache: localhost:8545"

dev-stop: ## Stop all development services
	@echo "$(YELLOW)Stopping development environment...$(NC)"
	cd deployments/local && docker-compose down
	@echo "$(GREEN)Development environment stopped!$(NC)"

dev-logs: ## Show logs from all services
	cd deployments/local && docker-compose logs -f

dev-restart: ## Restart all development services
	@echo "$(YELLOW)Restarting development environment...$(NC)"
	cd deployments/local && docker-compose restart
	@echo "$(GREEN)Development environment restarted!$(NC)"

# Building Commands
build: build-did-manager build-auth-service ## Build all services

build-did-manager: ## Build DID Manager service
	@echo "$(GREEN)Building DID Manager service...$(NC)"
	cd $(DID_MANAGER_DIR) && go build -o bin/did-manager ./cmd/server

build-auth-service: ## Build Auth Service
	@echo "$(GREEN)Building Auth Service...$(NC)"
	cd $(AUTH_SERVICE_DIR) && go build -o bin/auth-service ./cmd/server

build-cli: ## Build CLI client
	@echo "$(GREEN)Building CLI client...$(NC)"
	cd $(CLI_DIR) && go build -o bin/did-cli .

# Testing Commands
test: test-did-manager test-auth-service test-contracts ## Run all tests

test-did-manager: ## Test DID Manager service
	@echo "$(GREEN)Testing DID Manager service...$(NC)"
	cd $(DID_MANAGER_DIR) && go test -v ./...

test-auth-service: ## Test Auth Service
	@echo "$(GREEN)Testing Auth Service...$(NC)"
	cd $(AUTH_SERVICE_DIR) && go test -v ./...

test-contracts: ## Test smart contracts
	@echo "$(GREEN)Testing smart contracts...$(NC)"
	cd $(CONTRACTS_DIR) && npm test

# Smart Contract Commands
contracts-install: ## Install smart contract dependencies
	@echo "$(GREEN)Installing smart contract dependencies...$(NC)"
	cd $(CONTRACTS_DIR) && npm install

contracts-compile: ## Compile smart contracts
	@echo "$(GREEN)Compiling smart contracts...$(NC)"
	cd $(CONTRACTS_DIR) && npx hardhat compile

contracts-deploy-local: ## Deploy smart contracts to local network
	@echo "$(GREEN)Deploying smart contracts to local network...$(NC)"
	cd $(CONTRACTS_DIR) && npx hardhat run scripts/deploy.js --network localhost

contracts-deploy-testnet: ## Deploy smart contracts to testnet
	@echo "$(YELLOW)Deploying smart contracts to testnet...$(NC)"
	cd $(CONTRACTS_DIR) && npx hardhat run scripts/deploy.js --network testnet

contracts-deploy-mainnet: ## Deploy smart contracts to mainnet
	@echo "$(RED)Deploying smart contracts to mainnet...$(NC)"
	@read -p "Are you sure you want to deploy to mainnet? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	cd $(CONTRACTS_DIR) && npx hardhat run scripts/deploy.js --network mainnet

# Database Commands
db-init: ## Initialize database schema
	@echo "$(GREEN)Initializing database schema...$(NC)"
	cd deployments/local && docker-compose exec postgres psql -U postgres -d starter_db -f /docker-entrypoint-initdb.d/init.sql

db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "$(RED)WARNING: This will delete all data!$(NC)"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	cd deployments/local && docker-compose down -v
	cd deployments/local && docker-compose up -d postgres
	@echo "$(GREEN)Database reset complete!$(NC)"

# CLI Commands
cli-demo: ## Run CLI demo workflow
	@echo "$(GREEN)Running CLI demo...$(NC)"
	cd $(CLI_DIR) && go run did-cli.go demo

cli-health: ## Check service health via CLI
	@echo "$(GREEN)Checking service health...$(NC)"
	cd $(CLI_DIR) && go run did-cli.go health

# Utility Commands
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf $(DID_MANAGER_DIR)/bin/
	rm -rf $(AUTH_SERVICE_DIR)/bin/
	rm -rf $(CLI_DIR)/bin/
	rm -rf $(CONTRACTS_DIR)/cache/
	rm -rf $(CONTRACTS_DIR)/artifacts/

deps: ## Download Go dependencies
	@echo "$(GREEN)Downloading Go dependencies...$(NC)"
	cd $(DID_MANAGER_DIR) && go mod download
	cd $(AUTH_SERVICE_DIR) && go mod download
	cd $(CLI_DIR) && go mod download

fmt: ## Format Go code
	@echo "$(GREEN)Formatting Go code...$(NC)"
	cd $(DID_MANAGER_DIR) && go fmt ./...
	cd $(AUTH_SERVICE_DIR) && go fmt ./...
	cd $(CLI_DIR) && go fmt ./...

lint: ## Lint Go code
	@echo "$(GREEN)Linting Go code...$(NC)"
	cd $(DID_MANAGER_DIR) && golangci-lint run
	cd $(AUTH_SERVICE_DIR) && golangci-lint run
	cd $(CLI_DIR) && golangci-lint run

# Monitoring Commands
status: ## Show service status
	@echo "$(GREEN)Service Status:$(NC)"
	cd deployments/local && docker-compose ps

logs-did-manager: ## Show DID Manager logs
	cd deployments/local && docker-compose logs -f did-manager

logs-auth-service: ## Show Auth Service logs
	cd deployments/local && docker-compose logs -f auth-service

logs-postgres: ## Show PostgreSQL logs
	cd deployments/local && docker-compose logs -f postgres

logs-nats: ## Show NATS logs
	cd deployments/local && docker-compose logs -f nats

logs-ganache: ## Show Ganache logs
	cd deployments/local && docker-compose logs -f ganache

# Production Commands
prod-build: ## Build production Docker images
	@echo "$(GREEN)Building production images...$(NC)"
	docker build -f $(DID_MANAGER_DIR)/Dockerfile.prod -t did-manager:prod $(DID_MANAGER_DIR)
	docker build -f $(AUTH_SERVICE_DIR)/Dockerfile.prod -t auth-service:prod $(AUTH_SERVICE_DIR)

prod-deploy: ## Deploy to production (requires proper configuration)
	@echo "$(RED)Production deployment requires proper configuration!$(NC)"
	@echo "Please ensure you have:"
	@echo "  - Production environment variables"
	@echo "  - Production database credentials"
	@echo "  - Production blockchain network configuration"
	@echo "  - Proper security measures in place"

# Help Commands
check-env: ## Check environment configuration
	@echo "$(GREEN)Checking environment configuration...$(NC)"
	@echo "DID Manager environment:"
	@if [ -f $(DID_MANAGER_DIR)/.env ]; then echo "  ✓ .env file exists"; else echo "  ✗ .env file missing"; fi
	@echo "Auth Service environment:"
	@if [ -f $(AUTH_SERVICE_DIR)/.env ]; then echo "  ✓ .env file exists"; else echo "  ✗ .env file missing"; fi
	@echo "Smart Contract environment:"
	@if [ -f $(CONTRACTS_DIR)/.env ]; then echo "  ✓ .env file exists"; else echo "  ✗ .env file missing"; fi

setup: ## Initial setup for new developers
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@echo "1. Installing Go dependencies..."
	$(MAKE) deps
	@echo "2. Installing smart contract dependencies..."
	$(MAKE) contracts-install
	@echo "3. Building services..."
	$(MAKE) build
	@echo "4. Starting development environment..."
	$(MAKE) dev-start
	@echo "$(GREEN)Setup complete!$(NC)"
	@echo "Next steps:"
	@echo "  - Deploy smart contracts: make contracts-deploy-local"
	@echo "  - Run demo: make cli-demo"
	@echo "  - Check status: make status"

# Quick Commands
quick-start: ## Quick start for development
	@echo "$(GREEN)Quick starting development environment...$(NC)"
	$(MAKE) dev-start
	@echo "Waiting for services to be ready..."
	@sleep 10
	$(MAKE) contracts-deploy-local
	@echo "$(GREEN)Ready for development!$(NC)"

quick-stop: ## Quick stop all services
	$(MAKE) dev-stop

# Show help by default
.DEFAULT_GOAL := help
