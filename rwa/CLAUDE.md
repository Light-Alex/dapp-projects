# 项目：NFTMarketplace

## Communication

- **语言**: 请使用中文进行所有交流、注释和文档。
- **术语**: 保持技术术语英文

## 约定
- 画图工具使用 `mermaid`

## 目录结构
参考 README.md 中的[项目结构](README.md#项目结构)章节。

## 常用命令
```bash
# 构建所有后端服务
make build-backend

# 构建所有智能合约
make build-contracts

# 构建全部
make build

# 后端代码格式化
make lint-backend

# 合约代码格式化
cd rwa-contract && forge fmt

# 启动后端服务（在不同终端中）
cd rwa-backend/apps/indexer && go run main.go && cd ../..
cd apps/alpaca-stream && go run main.go && cd ../..
cd apps/api && go run main.go && cd ../..
cd apps/ws-server && go run main.go && cd ../../..
```