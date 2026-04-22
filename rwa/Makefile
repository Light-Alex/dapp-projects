.PHONY: help build test clean

# ---- Variables ----
BACKEND_DIR := rwa-backend
CONTRACT_DIR := rwa-contract

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ---- Backend ----
build-api: ## Build API server
	cd $(BACKEND_DIR) && go build -o bin/api ./apps/api

build-indexer: ## Build Indexer service
	cd $(BACKEND_DIR) && go build -o bin/indexer ./apps/indexer

build-stream: ## Build Alpaca Stream service
	cd $(BACKEND_DIR) && go build -o bin/alpaca-stream ./apps/alpaca-stream

build-ws: ## Build WebSocket server
	cd $(BACKEND_DIR) && go build -o bin/ws-server ./apps/ws-server

build-backend: build-api build-indexer build-stream build-ws ## Build all backend services

test-backend: ## Run backend tests
	cd $(BACKEND_DIR) && go test ./... -v

lint-backend: ## Lint backend code
	cd $(BACKEND_DIR) && golangci-lint run ./...

# ---- Contracts ----
build-contracts: ## Build Solidity contracts
	cd $(CONTRACT_DIR) && forge build

test-contracts: ## Run contract tests
	cd $(CONTRACT_DIR) && forge test -vvv

# ---- Database ----
migrate-up: ## Run database migrations
	cd $(BACKEND_DIR) && go run ./cmd/migrate up

migrate-down: ## Rollback database migrations
	cd $(BACKEND_DIR) && go run ./cmd/migrate down

# ---- All ----
build: build-backend build-contracts ## Build everything

test: test-backend test-contracts ## Run all tests

clean: ## Clean build artifacts
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(CONTRACT_DIR)/out $(CONTRACT_DIR)/cache
