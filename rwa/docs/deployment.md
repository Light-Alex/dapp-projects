# Anchored Finance (RWA) 部署与服务启动文档

本文档面向新人，从零开始指导你完成整个 Anchored Finance RWA 系统的部署和启动。

---

## 目录

1. [环境要求](#1-环境要求)
2. [项目克隆和初始化](#2-项目克隆和初始化)
3. [基础设施启动](#3-基础设施启动)
4. [数据库迁移](#4-数据库迁移)
5. [智能合约](#5-智能合约)
6. [Go 合约绑定生成](#6-go-合约绑定生成)
7. [配置文件说明](#7-配置文件说明)
8. [服务启动顺序](#8-服务启动顺序)
9. [启动命令](#9-启动命令)
10. [健康检查](#10-健康检查)
11. [常见问题排查](#11-常见问题排查)
12. [Alpaca 账户注册](#12-alpaca-账户注册)

---

## 1. 环境要求

### 操作系统

- macOS (推荐) / Linux / Windows (WSL2)

### 必装软件及版本

| 软件 | 最低版本 | 说明 |
|------|---------|------|
| **Go** | 1.25.1+ | 后端服务开发语言，项目 `go.work` 指定了 `go 1.25.1` |
| **Node.js** | 18+ | 合约部署脚本使用 TypeScript |
| **npm / pnpm** | 最新稳定版 | 合约项目的包管理器 |
| **Docker** | 20.10+ | 运行 PostgreSQL、Redis、Kafka 容器 |
| **Docker Compose** | v2+ | 容器编排工具 |
| **Foundry** | 最新版 | Solidity 合约编译、测试、部署 |
| **Git** | 2.30+ | 版本控制 |

### 基础设施组件（通过 Docker 运行，无需手动安装）

| 组件 | 版本 | 默认端口 |
|------|------|---------|
| **PostgreSQL** | 16.9 | 5432 |
| **Redis Stack** | latest | 6379 |
| **Kafka** (3 节点集群) | 7.7.5 (Confluent) | 39092/39093/39094 (外部) |
| **Kafka UI** | latest | 39090 |

### 可选工具

| 工具 | 说明 |
|------|------|
| **golangci-lint** | Go 代码静态检查 |
| **swag** (1.10.3+) | Swagger API 文档生成 |
| **abigen** (1.20+) | 从合约 ABI 生成 Go 绑定代码 |
| **migrate** | 数据库迁移工具 |

### 安装 Go

```bash
# macOS (Homebrew)
brew install go

# Linux
wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

验证安装：

```bash
go version
# 输出: go version go1.25.x ...
```

### 安装 Node.js

```bash
# 推荐使用 nvm 管理 Node.js 版本
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash
nvm install 18
nvm use 18
```

### 安装 Docker 和 Docker Compose

```bash
# macOS: 安装 Docker Desktop
# https://docs.docker.com/desktop/install/mac-install/

# Linux: 安装 Docker Engine
curl -fsSL https://get.docker.com | sh
```

### 安装 Foundry

```bash
curl -L https://foundry.paradigm.xyz | bash
foundryup
```

验证安装：

```bash
forge --version
cast --version
anvil --version
```

---

## 2. 项目克隆和初始化

### 2.1 克隆项目

```bash
# 克隆整个 rwa 项目（包含后端和合约）
git clone <your-repo-url> rwa
cd rwa
```

项目整体目录结构：

```
rwa/
├── rwa-backend/          # Go 后端服务（多个微服务）
│   ├── apps/
│   │   ├── api/          # REST API 服务
│   │   ├── indexer/      # 链上数据索引服务
│   │   ├── alpaca-stream/# Alpaca WebSocket 行情流服务
│   │   └── ws-server/    # WebSocket 推送服务
│   ├── libs/             # 共享库
│   │   ├── contracts/    # 合约 Go 绑定
│   │   ├── core/         # 核心工具库
│   │   ├── database/     # 数据库工具
│   │   ├── kafka/        # Kafka 工具
│   │   ├── log/          # 日志工具
│   │   └── ...
│   ├── migrations/       # 数据库迁移文件
│   ├── devops/           # Docker Compose 配置
│   ├── go.work           # Go workspace 配置
│   └── Makefile
├── rwa-contract/         # Solidity 智能合约
│   ├── contracts/        # 合约源码
│   ├── script/           # 部署脚本 (TypeScript)
│   ├── test/             # 测试
│   ├── foundry.toml      # Foundry 配置
│   └── package.json
└── docs/                 # 文档
```

### 2.2 后端依赖初始化

```bash
cd rwa-backend

# 同步 Go workspace 依赖
# go.work 定义了所有模块：apps/api, apps/indexer, apps/alpaca-stream, apps/ws-server, libs/* 等
go work sync
```

`go.work` 文件包含了以下工作区模块：

```
apps/alpaca-stream    # Alpaca 行情流服务
apps/api              # REST API 服务
apps/indexer          # 链上索引服务
apps/ws-server        # WebSocket 服务
libs/contracts        # 合约 Go 绑定
libs/core             # 核心库
libs/database         # 数据库库
libs/errors           # 错误处理库
libs/grpc/*           # gRPC 定义
libs/kafka            # Kafka 库
libs/log              # 日志库
libs/oss              # 对象存储库
```

### 2.3 合约依赖初始化

```bash
cd rwa-contract

# 安装 npm 依赖（合约使用了 @openzeppelin/contracts、viem 等）
npm install

# 安装 Foundry 依赖（如果有 lib/ 下的 submodule）
forge install
```

---

## 3. 基础设施启动

### 3.1 一键启动所有基础设施

```bash
cd rwa-backend

# 同时启动 PostgreSQL + Redis + Kafka
make install_all
```

或者分开启动：

```bash
# 只启动 PostgreSQL 和 Redis
make install_database

# 只启动 Kafka 集群
make install_kafka
```

### 3.2 PostgreSQL + Redis（database docker-compose）

配置文件位于：`devops/local/database/docker-compose.yml`

- **PostgreSQL 16.9**
  - 端口：`5432`
  - 用户名：`root`
  - 密码：`root`
  - 默认数据库：`postgres`
  - 数据持久化：`devops/local/database/pg/data/`
  - 环境变量文件：`devops/local/database/pg/pg.env`

- **Redis Stack Server**
  - 端口：`6379`
  - 无密码
  - 数据持久化：`devops/local/database/redis-stack/data/`

### 3.3 Kafka 集群（kafka docker-compose）

配置文件位于：`devops/local/kafka/docker-compose.yml`

部署了 3 节点 KRaft 模式的 Kafka 集群（无需 ZooKeeper）：

| 节点 | 外部端口 | Broker 端口 |
|------|---------|------------|
| kafka1 | 39092 | 19092 |
| kafka2 | 39093 | 19093 |
| kafka3 | 39094 | 19094 |

另外还有 **Kafka UI**，端口为 `39090`，可通过浏览器访问 `http://localhost:39090` 来查看 Kafka 集群状态。

### 3.4 修改 hosts 文件（重要）

启动 Kafka 后，**必须**修改 `/etc/hosts` 文件，否则本地客户端无法连接 Kafka：

```bash
sudo vim /etc/hosts
```

添加以下内容：

```
127.0.0.1 kafka1
127.0.0.1 kafka2
127.0.0.1 kafka3
```

### 3.5 创建业务数据库

PostgreSQL 容器默认创建的数据库是 `postgres`，但服务配置中使用的数据库名为 `anchored`，需要手动创建：

```bash
# 连接到 PostgreSQL
docker exec -it postgres psql -U root -d postgres

# 在 psql 中执行
CREATE DATABASE anchored;

# 退出
\q
```

### 3.6 验证基础设施

```bash
# 验证所有容器是否在运行
docker ps

# 预期看到以下容器：
# - postgres
# - redis-stack
# - kafka1, kafka2, kafka3
# - kafka-ui

# 验证 PostgreSQL 连接
docker exec -it postgres psql -U root -d anchored -c "SELECT 1;"

# 验证 Redis 连接
docker exec -it redis-stack redis-cli ping
# 预期输出: PONG

# 验证 Kafka 集群
docker exec -it kafka1 kafka-topics --bootstrap-server kafka1:19092 --list
```

---

## 4. 数据库迁移

### 4.1 安装 migrate 工具

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

确保 `$GOPATH/bin` 在你的 `$PATH` 中：

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### 4.2 创建迁移文件

如果需要创建新的迁移版本：

```bash
cd rwa-backend/migrations

# 创建迁移文件（会生成 up 和 down 两个 SQL 文件）
migrate create -ext sql -dir rwa -seq <migration_name>

# 例如：
migrate create -ext sql -dir rwa -seq create_users_table
# 会生成:
#   rwa/000001_create_users_table.up.sql
#   rwa/000001_create_users_table.down.sql
```

### 4.3 运行迁移

```bash
# 执行所有待运行的迁移（向上迁移）
migrate -database "postgres://root:root@127.0.0.1:5432/anchored?sslmode=disable" -path rwa-backend/migrations/rwa up

# 回滚最后一次迁移
migrate -database "postgres://root:root@127.0.0.1:5432/anchored?sslmode=disable" -path rwa-backend/migrations/rwa down 1

# 查看当前迁移版本
migrate -database "postgres://root:root@127.0.0.1:5432/anchored?sslmode=disable" -path rwa-backend/migrations/rwa version
```

> **注意：** 各服务的 `config.yaml` 中也配置了 `migrationPath: "file://migrations/rwa"`，部分服务可能在启动时自动执行迁移。

---

## 5. 智能合约

### 5.1 项目结构

```
rwa-contract/
├── contracts/              # Solidity 合约源码
│   ├── AnchoredToken.sol          # RWA 代币合约
│   ├── AnchoredTokenFactory.sol   # 代币工厂合约
│   ├── AnchoredTokenManager.sol   # 代币管理合约
│   ├── AnchoredCompliance.sol     # 合规合约
│   ├── AnchoredBlocklist.sol      # 黑名单合约
│   ├── AnchoredSanctionsList.sol  # 制裁名单合约
│   ├── poc/                       # PoC（概念验证）合约
│   │   ├── PocToken.sol           # PoC 代币
│   │   ├── Order.sol              # 订单合约
│   │   ├── MockUSDC.sol           # 模拟 USDC
│   │   └── PocGate.sol            # PoC 网关
│   ├── gate/                      # 网关合约
│   └── interfaces/                # 接口定义
├── script/                 # 部署脚本
│   └── poc/
│       ├── deploy_poc_contracts.ts  # PoC 合约部署脚本
│       ├── constants.ts             # 已部署合约地址常量
│       ├── config_poc.ts            # PoC 配置
│       ├── placeOrder.ts            # 下单脚本
│       └── upgrade_poc.ts           # 升级脚本
├── test/foundry/           # Foundry 测试
├── foundry.toml            # Foundry 配置
└── package.json            # npm 依赖
```

### 5.2 编译合约

```bash
cd rwa-contract

# 编译所有合约
forge build
```

Foundry 配置要点（`foundry.toml`）：
- 合约源码目录：`contracts/`
- 编译输出目录：`out/`
- 测试目录：`test/foundry/`
- 启用优化器，运行次数 200
- 使用 `@openzeppelin/contracts` 通过 remappings 映射到 `node_modules`

### 5.3 运行测试

```bash
# 运行所有测试
forge test

# 运行测试并显示详细输出
forge test -vvv

# 运行指定测试文件
forge test --match-path test/foundry/SomeTest.t.sol

# 查看 Gas 消耗报告
forge snapshot

# 格式化合约代码
forge fmt
```

### 5.4 部署合约到测试网

项目使用 Foundry `forge script` 进行合约部署。当前测试网为 **BSC Testnet**（Chain ID: `97`）。

详细的合约部署流程请参考 [合约部署文档](./contract-deployment.md)。

```bash
cd rwa-contract

# 部署所有 POC 合约到 BSC Testnet
forge script script/poc/DeployAll.s.sol \
  --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545 \
  --broadcast \
  --verify
```

部署完成后，将输出的合约地址更新到各服务的 `config.yaml` 中。

### 5.5 使用 Anvil 进行本地测试

```bash
# 启动本地以太坊节点
anvil

# 使用 cast 与合约交互
cast call <contract_address> "functionName()" --rpc-url http://localhost:8545
```

---

## 6. Go 合约绑定生成

### 6.1 安装 abigen

`abigen` 是 go-ethereum 提供的工具，用于从合约 ABI 生成 Go 语言的绑定代码。

```bash
# 方法一：通过 go install 安装
go install github.com/ethereum/go-ethereum/cmd/abigen@latest

# 方法二：通过 Homebrew 安装（macOS）
brew install ethereum
```

验证安装：

```bash
abigen --version
```

### 6.2 生成 Go 绑定

首先需要编译合约，然后从编译输出中提取 ABI：

```bash
cd rwa-contract

# 确保合约已编译
forge build

# 从编译产物中提取 ABI（以 PocToken 为例）
# 编译产物位于 out/ 目录下

# 生成 Go 绑定代码
abigen \
  --abi out/PocToken.sol/PocToken.json \
  --pkg contracts \
  --type PocToken \
  --out ../rwa-backend/libs/contracts/poc_token.go

# 对 Order 合约同样操作
abigen \
  --abi out/Order.sol/OrderContract.json \
  --pkg contracts \
  --type OrderContract \
  --out ../rwa-backend/libs/contracts/order_contract.go
```

> **注意：** 如果 ABI 文件是完整的 Foundry 输出 JSON（包含 `abi` 字段和其他元数据），你可能需要先用 `jq` 提取纯 ABI 部分：
>
> ```bash
> jq '.abi' out/PocToken.sol/PocToken.json > /tmp/PocToken.abi
> abigen --abi /tmp/PocToken.abi --pkg contracts --type PocToken --out ../rwa-backend/libs/contracts/poc_token.go
> ```

生成的绑定代码位于 `rwa-backend/libs/contracts/` 目录下，被 indexer 等服务引用。

---

## 7. 配置文件说明

所有服务的配置文件均使用 YAML 格式，默认路径为各服务目录下的 `config/config.yaml`。

### 7.1 API 服务配置 (`apps/api/config/config.yaml`)

```yaml
appName: "RWA API"                          # 应用名称，用于日志标识

server:
  port: 8000                                # HTTP 服务监听端口
  env: dev                                  # 运行环境: dev / prod
  basePath: /api/v1                         # API 路由前缀
  apiKeys:                                  # API 访问密钥列表（用于接口鉴权）
    - 099ba6fc-8da2-4d83-b3e9-a43813e471d3
  ginMode: debug                            # Gin 框架模式: debug / release

rpcInfo:                                    # 区块链 RPC 配置（按 chainId 分组）
  97:                                       # BSC Testnet Chain ID
    rpcUrl: "https://data-seed-prebsc-1-s1.binance.org:8545"
    wssUrl: "wss://bsc-testnet-rpc.publicnode.com"
    nativeTokenSymbol: BNB                  # 原生代币符号
    nativeTokenDecimals: 18                 # 原生代币精度

redis:
  hosts:                                    # Redis 地址列表
    - "127.0.0.1:6379"
  password: ""                              # Redis 密码（本地开发为空）
  db: 0                                     # Redis 数据库编号

db:                                         # PostgreSQL 数据库配置
  host: "127.0.0.1"
  port: 5432
  username: "root"
  password: "root"
  database: "anchored"                      # 数据库名
  maxIdleConns: 20                          # 最大空闲连接数
  maxOpenConns: 200                         # 最大打开连接数
  logLevel: 2                               # GORM 日志级别: 1=Silent, 2=Error, 3=Warn, 4=Info
  sqlSlowThresholdMill: 10000               # 慢 SQL 阈值（毫秒）
  migrationPath: "file://migrations/rwa"    # 迁移文件路径
  sslMode: "disable"                        # SSL 模式

alpaca:                                     # Alpaca 交易 API 配置
  api_key: "YOUR_ALPACA_API_KEY"            # Alpaca API Key
  api_secret: "YOUR_ALPACA_API_SECRET"      # Alpaca API Secret
  base_url: "https://paper-api.alpaca.markets"   # Alpaca API 基地址（paper = 模拟交易）
  data_url: "https://data.alpaca.markets"         # Alpaca 数据 API 地址

logger:                                     # 日志配置
  level: "debug"                            # 日志级别: debug/info/warn/error/fatal
  encoderType: "console"                    # 编码格式: json / console
  outputType: "all"                         # 输出方式: file / stdout / all
  maxAge: 30                                # 日志文件最大保留天数
  enableColor: true                         # 是否启用彩色输出

frontendTx:                                 # 前端交易代理配置
  proxyWallet:
    mnemonic: "your mnemonic ..."           # 代理钱包助记词（生产环境必须更换）
    derivationPath: "m/44'/60'/0'/0"        # BIP44 派生路径
  jwt:
    secret: "your-jwt-secret-key"           # JWT 签名密钥（生产环境必须更换）
    expirationHours: 24                     # JWT 过期时间（小时）
```

### 7.2 Indexer 服务配置 (`apps/indexer/config/config.yaml`)

```yaml
appName: "RWA Indexer Direct"

chain:                                      # 链上配置
  chainId: 97                               # BSC Testnet Chain ID
  pocAddress: "0x..."                       # Order 合约（POC）代理地址（部署后替换）
  usdmAddress: "0x..."                      # USDM 代币地址（部署后替换）

rpcInfo:                                    # 同 API 服务
  97:
    rpcUrl: "https://data-seed-prebsc-1-s1.binance.org:8545"
    wssUrl: "wss://bsc-testnet-rpc.publicnode.com"

db:                                         # 同 API 服务
  ...

logger:                                     # 同 API 服务
  ...

backend:                                    # 后端钱包配置（用于 Alpaca PlaceOrder 失败时调用链上 cancelOrder 退款）
  privateKey: ""                            # 后端钱包私钥（hex，需要 BACKEND_ROLE 权限）

alpaca:                                     # Alpaca API 配置（用于 Indexer 下单和取消）
  api_key: "YOUR_ALPACA_API_KEY"
  api_secret: "YOUR_ALPACA_API_SECRET"
  base_url: "https://paper-api.alpaca.markets"
  data_url: "https://data.alpaca.markets"

indexer:                                    # 索引器特有配置
  pollInterval: 3                           # 轮询间隔（秒）
  batchSize: 100                            # 每次批量处理的区块/事件数
  startBlock: 0                             # 起始区块号（0 表示从头开始，生产环境应设为合约部署区块）
  confirmationBlocks: 0                     # 确认区块数（0 表示不等待确认）
```

> **注意**：`backend.privateKey` 配置可选。如果配置了，当 Alpaca PlaceOrder 失败时，Indexer 会异步调用链上 `cancelOrder` 退还用户锁定的资金。如果未配置，则仅将订单标记为 Rejected 状态，需要人工处理退款。

### 7.3 Alpaca Stream 服务配置 (`apps/alpaca-stream/config/config.yaml`)

参考示例文件 `apps/alpaca-stream/config.example.yaml`：

```bash
# 首先从示例文件复制
cp apps/alpaca-stream/config.example.yaml apps/alpaca-stream/config/config.yaml
```

```yaml
appName: "Alpaca WebSocket Service"

alpaca:                                     # Alpaca WebSocket 配置
  api_key: "YOUR_ALPACA_API_KEY"            # Alpaca API Key
  api_secret: "YOUR_ALPACA_API_SECRET"      # Alpaca API Secret
  ws_url: "wss://paper-api.alpaca.markets/stream"          # 交易更新 WebSocket URL
  ws_data_url: "wss://stream.data.alpaca.markets/v2/iex"   # 市场数据 WebSocket URL（iex 或 sip）

chain:                                      # 链上配置（用于 markExecuted/cancelOrder/mint）
  chainId: 97                               # BSC Testnet Chain ID
  pocAddress: "0x..."                       # OrderContract 代理地址
  usdmAddress: "0x..."                      # USDM 代币地址

backend:                                    # 后端钱包配置（用于签署链上交易）
  privateKey: ""                            # 后端钱包私钥（hex，需要 BACKEND_ROLE 权限）

rpcInfo:                                    # 区块链 RPC 配置（用于链上合约调用）
  97:
    rpcUrl: "https://data-seed-prebsc-1-s1.binance.org:8545"
    wssUrl: "wss://bsc-testnet-rpc.publicnode.com"
    nativeTokenSymbol: "BNB"
    nativeTokenDecimals: 18

db:                                         # PostgreSQL 数据库配置（同 API 服务）
  host: "127.0.0.1"
  port: 5432
  username: "root"
  password: "root"
  database: "anchored"
  maxIdleConns: 20
  maxOpenConns: 200
  logLevel: 2
  sqlSlowThresholdMill: 10000
  migrationPath: "file://migrations/rwa"
  sslMode: "disable"

logger:                                     # 同 API 服务
  ...
```

> **重要**：`chain`、`backend`、`rpcInfo` 和 `db` 配置是 Alpaca Stream 服务正常工作的必需项。
> - `chain` + `backend` + `rpcInfo`：用于在订单成交后调用链上合约（`markExecuted`、`cancelOrder`、`PocToken.mint`）
> - `db`：用于更新订单状态和保存链上交易哈希
> - `backend.privateKey` 对应的钱包地址必须在 OrderContract 上拥有 `BACKEND_ROLE`，且在各 PocToken 上拥有 `MINTER_ROLE`

### 7.4 WebSocket Server 配置 (`apps/ws-server/config/config.yaml`)

```yaml
appName: "WS Server"

server:
  port: 8082                                # WebSocket 服务监听端口
  basePath: /api/v1/ws                      # WebSocket 路由前缀

redis:                                      # Redis 配置（用于发布/订阅）
  hosts:
    - "127.0.0.1:6379"
  password: ""
  db: 0

logger:                                     # 同 API 服务
  ...

alpaca:                                     # Alpaca WebSocket 配置
  api_key: "YOUR_ALPACA_API_KEY"
  api_secret: "YOUR_ALPACA_API_SECRET"
  ws_url: "wss://stream.data.alpaca.markets/v2/iex"
  ws_data_url: "wss://stream.data.alpaca.markets/v2/iex"
```

---

## 8. 服务启动顺序

服务之间存在依赖关系，**必须按照以下顺序启动**：

```
第 1 步: 基础设施（PostgreSQL + Redis + Kafka）
   |
第 2 步: 智能合约部署（如果是首次部署或合约有更新）
   |
第 3 步: 数据库迁移
   |
第 4 步: Indexer 服务（依赖：DB + RPC 节点 + 合约已部署）
   |
第 5 步: Alpaca Stream 服务（依赖：DB + Alpaca API）
   |
第 6 步: API 服务（依赖：DB + Redis + RPC 节点 + Alpaca API）
   |
第 7 步: WebSocket Server（依赖：Redis + Alpaca API）
```

### 依赖关系详解

| 服务 | 依赖的基础设施 | 依赖的外部服务 |
|------|---------------|---------------|
| **Indexer** | PostgreSQL | 区块链 RPC 节点, Alpaca REST API |
| **Alpaca Stream** | PostgreSQL | Alpaca WebSocket API, 区块链 RPC 节点（用于 markExecuted/cancelOrder/mint） |
| **API** | PostgreSQL, Redis | 区块链 RPC 节点, Alpaca REST API |
| **WS Server** | Redis | Alpaca WebSocket API |

---

## 9. 启动命令

### 9.1 第 1 步：启动基础设施

```bash
cd rwa-backend

# 启动所有基础设施
make install_all

# 等待容器完全启动（约 10-30 秒）
sleep 15

# 确认 /etc/hosts 已添加 kafka1, kafka2, kafka3 映射
cat /etc/hosts | grep kafka
```

### 9.2 第 2 步：部署合约（仅首次或合约更新时）

```bash
cd rwa-contract

# 编译合约
forge build

# 运行部署脚本
npx ts-node --esm script/poc/deploy_poc_contracts.ts

# 记录输出的合约地址，更新到各服务配置文件中
```

### 9.3 第 3 步：数据库迁移

```bash
# 确保已创建 anchored 数据库（参见 3.5 节）

# 运行迁移
cd rwa-backend
migrate -database "postgres://root:root@127.0.0.1:5432/anchored?sslmode=disable" -path migrations/rwa up
```

### 9.4 第 4 步：启动 Indexer 服务

```bash
cd rwa-backend/apps/indexer

# 使用默认配置启动
go run main.go

# 或指定配置文件
go run main.go -c config/config.yaml
```

### 9.5 第 5 步：启动 Alpaca Stream 服务

```bash
cd rwa-backend/apps/alpaca-stream

# 确保已从 config.example.yaml 复制并配置了 config/config.yaml
# cp config.example.yaml config/config.yaml

# 启动服务
go run main.go

# 或指定配置文件
go run main.go -c config/config.yaml
```

### 9.6 第 6 步：启动 API 服务

```bash
cd rwa-backend/apps/api

# 启动服务
go run main.go

# 或指定配置文件和应用类型
go run main.go -a api -c config/config.yaml
```

API 服务默认监听 `http://localhost:8000`，路由前缀为 `/api/v1`。

### 9.7 第 7 步：启动 WebSocket Server

```bash
cd rwa-backend/apps/ws-server

# 启动服务
go run main.go

# 或指定配置文件和应用类型
go run main.go -a ws -c config/config.yaml
```

WebSocket 服务默认监听 `http://localhost:8082`，路由前缀为 `/api/v1/ws`。

### 9.8 多终端启动汇总

建议使用多个终端窗口（或 tmux/screen），每个窗口启动一个服务：

```bash
# 终端 1：Indexer
cd rwa-backend/apps/indexer && go run main.go

# 终端 2：Alpaca Stream
cd rwa-backend/apps/alpaca-stream && go run main.go

# 终端 3：API
cd rwa-backend/apps/api && go run main.go

# 终端 4：WS Server
cd rwa-backend/apps/ws-server && go run main.go
```

---

## 10. 健康检查

### 10.1 基础设施检查

```bash
# Docker 容器状态
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# PostgreSQL
docker exec -it postgres psql -U root -d anchored -c "SELECT 1 AS health;"

# Redis
docker exec -it redis-stack redis-cli ping
# 预期: PONG

# Kafka（通过 Kafka UI）
# 浏览器访问 http://localhost:39090

# Kafka 命令行检查
docker exec -it kafka1 kafka-topics --bootstrap-server kafka1:19092 --list
```

### 10.2 API 服务检查

```bash
# 测试 API 是否可访问
curl -s http://localhost:8000/api/v1/health
# 或者简单地检查端口是否在监听
lsof -i :8000
```

### 10.3 WebSocket 服务检查

```bash
# 检查端口是否在监听
lsof -i :8082

# 使用 websocat 或其他 WebSocket 客户端测试（需安装 websocat）
# brew install websocat
# websocat ws://localhost:8082/api/v1/ws
```

### 10.4 Indexer 服务检查

Indexer 是后台进程，通过日志确认运行状态：

- 观察终端输出，应能看到类似 `"Begin to start rwa-indexer-direct service"` 的日志
- 没有 `ERROR` 级别日志
- 持续输出新区块的处理日志

### 10.5 Alpaca Stream 服务检查

- 观察终端输出，应能看到 `"Begin to start alpaca websocket service"` 的日志
- 如果 Alpaca API Key 配置正确，能看到 WebSocket 连接成功的日志
- 交易时段内应能看到行情数据推送日志

### 10.6 数据库数据检查

```bash
# 连接数据库查看表是否创建成功
docker exec -it postgres psql -U root -d anchored

# 查看所有表
\dt

# 查看迁移状态表
SELECT * FROM schema_migrations;
```

---

## 11. 常见问题排查

### 11.1 Docker 相关

**问题：Docker Compose 命令找不到**

```
解决方案：
- 新版 Docker Desktop 使用 `docker compose`（无连字符）
- 旧版使用 `docker-compose`
- 如果 Makefile 中使用的是 `docker-compose`，确保安装了独立的 docker-compose 或创建别名：
  alias docker-compose='docker compose'
```

**问题：容器启动失败，端口已被占用**

```bash
# 查看占用端口的进程
lsof -i :5432    # PostgreSQL
lsof -i :6379    # Redis
lsof -i :39092   # Kafka

# 停止占用端口的进程或修改 docker-compose 中的端口映射
```

**问题：Kafka 容器反复重启**

```bash
# 查看 Kafka 日志
docker logs kafka1

# 常见原因：数据目录权限问题
# 解决方案：清除旧数据后重新启动
rm -rf devops/local/kafka/data/
make install_kafka
```

### 11.2 数据库相关

**问题：连接数据库时报 "database anchored does not exist"**

```bash
# 需要手动创建数据库
docker exec -it postgres psql -U root -d postgres -c "CREATE DATABASE anchored;"
```

**问题：迁移失败，报 dirty database version**

```bash
# 强制设置迁移版本
migrate -database "postgres://root:root@127.0.0.1:5432/anchored?sslmode=disable" -path migrations/rwa force <version>

# 然后重新运行迁移
migrate -database "postgres://root:root@127.0.0.1:5432/anchored?sslmode=disable" -path migrations/rwa up
```

**问题：migrate 命令报 "unknown driver postgres"**

```bash
# 需要安装带 postgres 标签的版本
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 11.3 Go 相关

**问题：`go work sync` 失败**

```bash
# 确保 Go 版本 >= 1.25.1
go version

# 清除模块缓存后重试
go clean -modcache
go work sync
```

**问题：服务启动报 "config file not found"**

```bash
# 检查当前工作目录和配置文件路径
# 默认配置文件路径为 config/config.yaml（相对于当前目录）
# 确保在正确目录下运行，或使用 -c 参数指定绝对路径
go run main.go -c /absolute/path/to/config/config.yaml
```

### 11.4 Kafka 相关

**问题：服务无法连接到 Kafka**

```bash
# 检查 /etc/hosts 是否配置正确
cat /etc/hosts | grep kafka

# 应包含以下行：
# 127.0.0.1 kafka1
# 127.0.0.1 kafka2
# 127.0.0.1 kafka3
```

### 11.5 Alpaca 相关

**问题：Alpaca WebSocket 连接失败**

```
排查步骤：
1. 确认 API Key 和 Secret 是否正确
2. 确认使用的是 Paper Trading 的 URL（不是 Live）
3. 检查网络是否可以访问 alpaca.markets（部分地区可能需要代理）
4. 美股非交易时段（北京时间约 21:30 - 次日 4:00 之外），市场数据流可能没有数据
```

### 11.6 合约相关

**问题：forge build 失败**

```bash
# 确保 npm 依赖已安装（合约使用了 @openzeppelin/contracts）
cd rwa-contract
npm install

# 如果有 lib/ 下的 git submodule
forge install

# 清除缓存后重新编译
rm -rf out/ forge-cache/
forge build
```

**问题：abigen 生成的代码编译错误**

```
排查步骤：
1. 确保使用正确的 ABI 文件格式（纯 ABI 数组，不是 Foundry 完整输出）
2. 检查 abigen 版本是否与 go-ethereum 版本兼容
3. 确保 --pkg 参数与目标目录的 package 名称一致
```

---

## 12. Alpaca 账户注册

Alpaca 是美股经纪商，提供 Paper Trading（模拟交易）账户，本项目通过 Alpaca API 获取美股行情和执行交易。

### 12.1 注册账户

1. 访问 Alpaca 官网：[https://app.alpaca.markets/signup](https://app.alpaca.markets/signup)
2. 点击 **Sign Up** 注册账户
3. 填写邮箱地址和密码
4. 验证邮箱
5. 登录后进入 Dashboard

### 12.2 切换到 Paper Trading 模式

1. 登录后，在 Dashboard 页面的左侧栏或顶部切换到 **Paper Trading** 模式
2. Paper Trading 是模拟交易环境，使用虚拟资金，不涉及真实资金
3. Paper Trading 环境会自动分配 $100,000 的模拟资金

### 12.3 获取 API Key

1. 在 Paper Trading Dashboard 中，找到 **API Keys** 部分（通常在首页右侧或导航菜单中）
2. 点击 **Generate New Key** 或 **View** 按钮
3. 记录以下两个值（Secret 只会显示一次，请妥善保存）：
   - **API Key ID** (例如: `PKXXXXXXXXXXXXXXXX`)
   - **API Secret Key** (例如: `H7JJwXXXXXXXXXXXXXXXXXXXXXXXXX`)

### 12.4 配置 API Key 到项目中

将获取的 API Key 和 Secret 填入以下服务的配置文件：

**API 服务** (`apps/api/config/config.yaml`)：

```yaml
alpaca:
  api_key: "你的 API Key"
  api_secret: "你的 API Secret"
  base_url: "https://paper-api.alpaca.markets"
  data_url: "https://data.alpaca.markets"
```

**Alpaca Stream 服务** (`apps/alpaca-stream/config/config.yaml`)：

```yaml
alpaca:
  api_key: "你的 API Key"
  api_secret: "你的 API Secret"
  ws_url: "wss://paper-api.alpaca.markets/stream"
  ws_data_url: "wss://stream.data.alpaca.markets/v2/iex"
```

**WebSocket Server** (`apps/ws-server/config/config.yaml`)：

```yaml
alpaca:
  api_key: "你的 API Key"
  api_secret: "你的 API Secret"
  ws_url: "wss://stream.data.alpaca.markets/v2/iex"
  ws_data_url: "wss://stream.data.alpaca.markets/v2/iex"
```

### 12.5 Alpaca API URL 说明

| 用途 | Paper Trading URL | Live Trading URL |
|------|-------------------|------------------|
| REST API | `https://paper-api.alpaca.markets` | `https://api.alpaca.markets` |
| 市场数据 REST | `https://data.alpaca.markets` | `https://data.alpaca.markets` |
| 交易更新 WebSocket | `wss://paper-api.alpaca.markets/stream` | `wss://api.alpaca.markets/stream` |
| 市场数据 WebSocket (IEX) | `wss://stream.data.alpaca.markets/v2/iex` | `wss://stream.data.alpaca.markets/v2/iex` |
| 市场数据 WebSocket (SIP) | `wss://stream.data.alpaca.markets/v2/sip` | `wss://stream.data.alpaca.markets/v2/sip` |

> **IEX vs SIP：**
> - **IEX** 是免费数据源，来自 IEX 交易所，数据覆盖有限
> - **SIP** 是付费数据源（需要 Alpaca 付费订阅），包含所有美国交易所的数据
> - 开发和测试阶段使用 IEX 即可

### 12.6 注意事项

- Paper Trading 的 API Key 和 Live Trading 的 API Key 是不同的，不能混用
- 美股交易时间为美东时间 9:30-16:00（北京时间约 21:30-次日4:00，夏令时提前1小时）
- 非交易时段，市场数据 WebSocket 可能不会推送数据（盘前盘后有有限数据）
- API 有频率限制（Rate Limit），Paper Trading 的限制为每分钟 200 次请求

---

## 附录：快速启动清单

以下是从零开始启动整个系统的简要步骤清单：

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
