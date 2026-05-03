.PHONY: help build test clean

# ---- Variables ----
BACKEND_DIR := rwa-backend
CONTRACT_DIR := rwa-contract

help: ## Show this help (显示帮助信息)
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ---- Backend ----
build-api: ## Build API server (构建 API 服务器)
	cd $(BACKEND_DIR) && go build -o bin/api ./apps/api

build-indexer: ## Build Indexer service (构建索引器服务)
	cd $(BACKEND_DIR) && go build -o bin/indexer ./apps/indexer

build-stream: ## Build Alpaca Stream service (构建 Alpaca Stream 服务)
	cd $(BACKEND_DIR) && go build -o bin/alpaca-stream ./apps/alpaca-stream

build-ws: ## Build WebSocket server (构建 WebSocket 服务器)
	cd $(BACKEND_DIR) && go build -o bin/ws-server ./apps/ws-server

build-backend: build-api build-indexer build-stream build-ws ## Build all backend services (构建所有后端服务)

test-backend: ## Run backend tests (运行后端测试)
	cd $(BACKEND_DIR) && go test ./... -v

lint-backend: ## Lint backend code (检查后端代码)
	cd $(BACKEND_DIR) && golangci-lint run ./...

# ---- Contracts ----
build-contracts: ## Build Solidity contracts (编译 Solidity 合约)
	cd $(CONTRACT_DIR) && forge build

test-contracts: ## Run contract tests (运行合约测试)
	cd $(CONTRACT_DIR) && forge test -vvv

# ---- Database ----
migrate-up: ## Run database migrations (执行数据库迁移)
	cd $(BACKEND_DIR) && go run ./cmd/migrate up

migrate-down: ## Rollback database migrations (回滚数据库迁移)
	cd $(BACKEND_DIR) && go run ./cmd/migrate down

# ---- All ----
build: build-backend build-contracts ## Build everything (构建所有项目)

test: test-backend test-contracts ## Run all tests (运行所有测试)

clean: ## Clean build artifacts (清理构建产物)
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(CONTRACT_DIR)/out $(CONTRACT_DIR)/cache
