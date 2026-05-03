# Anchored Finance RWA - 现实资产代币化平台

## 项目概述

Anchored Finance RWA（Real World Assets）是一个将传统金融资产（美股）代币化的去中心化交易平台。用户通过链上智能合约提交买卖订单并托管资金，后端服务监听链上事件后将订单转发至 Alpaca 券商执行真实股票交易，交易完成后在链上标记订单已执行并铸造/销毁对应的代币资产。

### 核心特性

- **链上提交意图** - 用户通过智能合约提交订单，资金托管在链上
- **链下执行交易** - 后端服务监听事件，通过 Alpaca API 执行真实美股交易
- **链上结算确认** - 交易完成后铸造/销毁代币，实现闭环
- **7x24 交易** - 股票代币可随时转账（非交易时段由做市商提供流动性）
- **资金透明** - 所有操作链上可追溯，资金托管在智能合约中
- **安全可靠** - 基于 OpenZeppelin 合约库，采用 RBAC 权限控制

### 业务流程

```mermaid
graph LR
    U[用户] -->|approve + submitOrder| OC[OrderContract]
    OC -->|emit OrderSubmitted| I[Indexer]
    I -->|PlaceOrder| A[Alpaca API]
    A -->|执行美股交易| AS[Alpaca Stream]
    AS -->|markExecuted + mint| OC
```

---

## 技术栈

### 后端技术

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.25.1 | 后端服务开发语言 |
| Uber Fx | v1.24.0 | 依赖注入和服务生命周期管理 |
| Uber Zap | v1.27.0 | 结构化高性能日志 |
| GORM | v1.31.0 | PostgreSQL ORM 框架 |
| Gin | - | REST API Web 框架 |
| go-ethereum | v1.16.4 | 以太坊交互，链上事件监听 |
| Alpaca SDK | v3.5.0 | 美股交易 API 集成 |
| Redis | - | 行情数据缓存 |
| Kafka | 3.5+ | 订单状态和行情数据消息队列 |

### 智能合约

| 技术 | 版本 | 用途 |
|------|------|------|
| Solidity | ^0.8.20 | 智能合约开发语言 |
| Foundry | - | 合约编译、测试、部署框架 |
| OpenZeppelin | 5.x | 安全合约库（AccessControl、ReentrancyGuard 等） |

### 数据库

- **PostgreSQL 15+** - 业务数据持久化
- **Redis 7.0+** - 行情数据缓存

### 开发工具

- **abigen** - 生成 Go 语言合约绑定代码
- **swag** - Swagger API 文档生成
- **Docker Compose** - 本地开发环境

---

## 项目架构

### 服务架构

参考 [系统架构图](docs/architecture.md#3-系统架构图)

### 核心业务流程

#### 买入流程

1. 用户 approve USDM -> submitOrder（托管 USDM）
2. 链上 emit OrderSubmitted -> Indexer 创建订单记录
3. Indexer 调用 Alpaca PlaceOrder
4. Alpaca Stream 监听成交更新
5. 成交后 markExecuted -> mint 股票代币

#### 卖出流程

1. 用户 approve 股票代币 -> submitOrder（托管股票代币）
2. 链上 emit OrderSubmitted -> Indexer 创建订单
3. Indexer 调用 Alpaca 卖出
4. Alpaca Stream 监听成交
5. 成交后 markExecuted -> mint USDM 给用户

#### 取消流程

1. 用户 cancelOrderIntent（状态变 CancelRequested）
2. Indexer 调用 Alpaca CancelOrder
3. Alpaca Stream 监听取消状态
4. 后端调用 cancelOrder 退款

---

## 目录结构

```
rwa/
├── rwa-backend/                # Go 后端服务
│   ├── apps/                   # 应用程序
│   │   ├── indexer/            # 链上事件监听服务
│   │   ├── alpaca-stream/      # Alpaca WebSocket 监听服务
│   │   ├── api/                # REST API 服务
│   │   └── ws-server/          # WebSocket 实时推送服务
│   ├── libs/                   # 共享库
│   │   ├── core/               # 核心业务逻辑
│   │   │   ├── evm_helper/     # 以太坊交互封装
│   │   │   ├── models/         # 数据模型定义
│   │   │   ├── trade/          # Alpaca 交易封装
│   │   │   ├── redis_cache/    # Redis 缓存封装
│   │   │   └── kafka_help/     # Kafka 封装
│   │   ├── contracts/          # 智能合约 Go 绑定
│   │   ├── database/           # 数据库连接和迁移
│   │   ├── log/                # 日志封装
│   │   └── errors/             # 错误处理
│   ├── migrations/             # 数据库迁移脚本
│   └── devops/                 # 本地开发环境配置
│
├── rwa-contract/              # Solidity 智能合约
│   ├── contracts/             # 合约源码
│   │   ├── poc/               # POC 合约（概念验证版）
│   │   │   ├── Order.sol      # 订单管理合约
│   │   │   ├── PocGate.sol    # 充值/提现网关（简化版）
│   │   │   ├── PocToken.sol   # ERC20 代币（USDM + 股票代币）
│   │   │   └── MockUSDC.sol   # 测试用 USDC
│   │   ├── gate/              # 生产级 Gate 合约
│   │   │   ├── Gate.sol       # 含 pending 状态的充值/提现网关
│   │   │   └── Pending*.sol   # Pending 代币合约
│   │   └── interfaces/        # 合约接口定义
│   ├── script/                # 部署脚本
│   └── test/                  # 合约测试
│
├── docs/                       # 项目文档
│   ├── PRD.md                  # 产品需求文档
│   ├── architecture.md         # 系统架构文档
│   ├── smart-contracts.md      # 智能合约架构
│   ├── api-reference.md        # API 接口文档
│   ├── deployment.md           # 部署文档
│   ├── database-design.md      # 数据库设计
│   └── test-cases.md           # 测试用例
│
├── Makefile                     # 构建和测试命令
├── .env.example                # 环境变量模板
├── CLAUDE.md                   # Claude Code 配置
└── README.md                   # 本文件
```

---

## 部署

参考 [部署文档](docs/deployment.md)

## 文档

| 文档 | 描述 |
|------|------|
| [产品需求文档 (PRD)](docs/PRD.md) | 完整的产品需求和 Phase 规划 |
| [系统架构文档](docs/architecture.md) | 系统架构设计、服务模块说明、业务流程详解 |
| [智能合约架构](docs/smart-contracts.md) | 合约设计、权限模型、事件定义 |
| [API 接口文档](docs/api-reference.md) | REST API 接口规范 |
| [部署文档](docs/deployment.md) | 生产环境部署指南 |
| [数据库设计](docs/database-design.md) | 数据表结构和关系设计 |
| [测试用例](docs/test-cases.md) | 测试场景和用例说明 |

---

## 开发指南

### 构建项目

```bash
# 构建所有后端服务
make build-backend

# 构建所有智能合约
make build-contracts

# 构建全部
make build
```

### 代码规范

```bash
# 后端代码格式化
make lint-backend

# 合约代码格式化
cd rwa-contract && forge fmt
```

### 快速启动命令

```bash
# 1. 克隆项目
git clone <repo-url> rwa && cd rwa

# 2. 安装合约依赖
cd rwa-contract && npm install && forge build && cd ..

# 3. 初始化后端
cd rwa-backend && go work sync

# 4. 启动基础设施
make install_all
# 修改 /etc/hosts 添加 kafka1, kafka2, kafka3

# 5. 创建数据库
docker exec -it postgres psql -U root -d postgres -c "CREATE DATABASE anchored;"

# 6. 运行数据库迁移
migrate -database "postgres://root:root@127.0.0.1:5432/anchored?sslmode=disable" -path migrations/rwa up

# 7. 配置 Alpaca API Key（编辑各服务的 config.yaml）
# cp apps/alpaca-stream/config.example.yaml apps/alpaca-stream/config/config.yaml
# 编辑各配置文件，填入你的 Alpaca API Key 和 Secret

# 8. 启动服务（在不同终端中）
cd apps/indexer && go run main.go
cd apps/alpaca-stream && go run main.go
cd apps/api && go run main.go
cd apps/ws-server && go run main.go
```

---

## 安全性

- **智能合约安全** - 基于 OpenZeppelin 合约库，使用 ReentrancyGuard 防止重入攻击
- **权限控制** - 基于 AccessControl 的 RBAC 权限模型
- **资金安全** - 用户资金托管在智能合约中，后端无法直接动用
- **私钥管理** - 后端私钥通过环境变量管理，不硬编码
- **API 认证** - REST API 使用签名验证中间件
- **幂等性保证** - 所有事件处理幂等，防止重复操作

---

## 项目状态

| Phase | 主题 | 优先级 | 状态 |
|-------|------|--------|------|
| Phase 1 | 合约基础 + 后端框架 | P0 | 已完成 |
| Phase 2 | 核心交易流程 | P0 | 已完成 |
| Phase 3 | 资金管理（充值/提现） | P1 | 合约已完成，后端 Gate handlers 待开发 |
| Phase 4 | API + 实时行情推送 | P2 | 已完成 |
| Phase 5 | 做市商 + 非交易时段 | P3 | 规划中 |

---

## 贡献指南

欢迎贡献代码、报告问题或提出改进建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

---

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件
