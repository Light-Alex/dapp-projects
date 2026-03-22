# Blockchain Parser Go

> BSC（币安智能链）区块链交易监听与解析系统 - 自动处理充值和提现的完整解决方案

[![Go Version](https://img.shields.io/badge/Go-1.24.3-blue)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-blue)](https://www.postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7-red)](https://redis.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 📖 项目简介

Blockchain Parser Go 是一个基于 Go 语言开发的 BSC 区块链交易监听与解析系统。系统能够自动监听区块链上的交易，识别用户的充值行为（BNB 和 USDT），并自动处理提现请求。

### 主要特性

- ✅ **实时监听** - 持续监听 BSC 区块链上的最新区块
- ✅ **自动充值识别** - 自动识别并记录 BNB 和 USDT 充值交易
- ✅ **自动提现** - 自动处理待提现请求并广播到区块链
- ✅ **多交易类型支持** - 支持 Legacy、EIP-2930、EIP-1559、EIP-4844 Blob 交易
- ✅ **账户管理** - 自动创建账户并更新余额
- ✅ **高可靠** - 使用 PostgreSQL 持久化数据，Redis 缓存状态
- ✅ **Docker 部署** - 提供完整的 Docker Compose 配置

### 应用场景

- **加密货币交易所** - 自动处理用户充值提现
- **支付网关** - 监听和处理加密货币支付
- **钱包服务** - 管理用户钱包余额
- **DeFi 应用** - 与 DeFi 协议集成

## 🛠 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.24.3 |
| 数据库 | PostgreSQL | 17 |
| 缓存 | Redis | 7 |
| 区块链客户端 | go-ethereum | v1.17.1 |
| 容器化 | Docker Compose | - |

## 📁 项目结构

```
blockchain-parser-go/
├── main.go                 # 程序入口
├── config/                 # 配置管理
│   └── config.go           # 配置加载逻辑
├── parser/                 # 区块链解析器
│   └── blockchain_parser.go # 核心解析逻辑
├── service/                # 提现服务
│   └── withdrawal_service.go # 提现处理逻辑
├── database/               # 数据库操作
│   └── database.go         # 数据库 CRUD 操作
├── redis/                  # Redis 客户端
│   └── redis.go            # Redis 操作封装
├── types/                  # 数据类型定义
│   └── types.go            # 核心数据结构和常量
├── utils/                  # 工具函数
│   └── utils.go            # 通用工具函数
├── model/                  # 数据模型
│   └── model.go            # 数据模型定义
├── abis/                   # 合约 ABI 文件
│   └── MyERC20.json        # ERC20 合约 ABI
├── docker-compose.yml      # Docker 编排文件
├── go.mod                  # Go 模块依赖
└── README.md               # 项目文档
```

## 🚀 快速开始

### 前置要求

- Go 1.24+
- PostgreSQL 17+
- Redis 7+
- Docker & Docker Compose（可选，用于快速部署）

### 1. 克隆项目

```bash
git clone <repository-url>
cd blockchain-parser-go
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境变量

创建 `.env` 文件：

```env
# ==================== Blockchain ====================
# BSC RPC 节点地址
BSC_RPC_URL=https://bsc-dataseed.binance.org/
# 链 ID（56 = BSC 主网，97 = BSC 测试网）
CHAIN_ID=56
# 项目地址（监听与此地址相关的交易）
PROJECT_ADDRESS=0xYourProjectAddress
# USDT 合约地址（BSC 主网）
USDT_ADDRESS=0x55d398326f99059fF775485246999027B3197955
# 私钥（用于发送提现交易）
SEPOLIA_PRIVATE_KEY=your_private_key

# ==================== Database ====================
# PostgreSQL 配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=blockchain_parser

# ==================== Redis ====================
# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# ==================== Application ====================
# 扫描间隔（毫秒）
SCAN_INTERVAL=3000
# 确认区块数（等待多少个区块后才处理交易）
CONFIRMATION_BLOCKS=6
```

### 4. 启动依赖服务（使用 Docker Compose V2）

```bash
# 启动 PostgreSQL 和 Redis
docker compose up -d

# 查看服务状态
docker compose ps
```

### 5. 初始化数据库

```bash
# 连接到 PostgreSQL
psql -h localhost -U postgres -d blockchain_parser

# 创建表（types/types.go 中定义的表会自动创建）
# 或运行程序时会自动创建表
```

### 6. 运行程序

```bash
go run main.go
```

或编译后运行：

```bash
go build -o blockchain-parser
./blockchain-parser
```

## ⚙️ 配置说明

### 环境变量

| 变量名 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `BSC_RPC_URL` | 是 | - | BSC RPC 节点地址 |
| `CHAIN_ID` | 是 | 56 | 链 ID（56=主网，97=测试网） |
| `PROJECT_ADDRESS` | 是 | - | 项目地址，监听与此地址相关的交易 |
| `USDT_ADDRESS` | 是 | - | USDT 合约地址 |
| `SEPOLIA_PRIVATE_KEY` | 是 | - | 用于发送提现交易的私钥 |
| `DB_HOST` | 否 | localhost | 数据库主机 |
| `DB_PORT` | 否 | 5432 | 数据库端口 |
| `DB_USER` | 否 | postgres | 数据库用户名 |
| `DB_PASSWORD` | 否 | - | 数据库密码 |
| `DB_NAME` | 否 | blockchain_parser | 数据库名称 |
| `REDIS_HOST` | 否 | localhost | Redis 主机 |
| `REDIS_PORT` | 否 | 6379 | Redis 端口 |
| `REDIS_PASSWORD` | 否 | - | Redis 密码 |
| `SCAN_INTERVAL` | 否 | 3000 | 扫描间隔（毫秒） |
| `CONFIRMATION_BLOCKS` | 否 | 6 | 确认区块数 |

### RPC 节点推荐

**BSC 主网**:
- https://bsc-dataseed.binance.org/
- https://bsc-dataseed1.defibit.io/
- https://bsc.publicnode.com/

**BSC 测试网**:
- https://data-seed-prebsc-1-s1.binance.org:8545/

## 🔧 核心功能

### 1. 充值监听

系统会持续监听 BSC 区块链上的新区块，当检测到与项目地址相关的交易时：

1. **BNB 充值** - 检测到发送到项目地址的 BNB 转账
2. **USDT 充值** - 解析 USDT 合约的 Transfer 事件，检测到发送到项目地址的转账

对于每笔充值交易：
- 验证交易状态（必须成功）
- 自动创建用户账户（如果不存在）
- 更新用户余额
- 记录交易详情

### 2. 提现处理

系统会定期查询待处理的提现请求：

1. **查询待处理提现** - 从数据库查询状态为 "init" 的提现记录
2. **构建交易** - 根据提现类型（BNB/USDT）构建相应的交易
3. **签名并发送** - 使用配置的私钥签名交易并发送到区块链
4. **监控交易** - 等待交易被打包确认
5. **更新状态** - 根据交易结果更新提现状态和用户余额

### 3. 账户管理

- **自动创建账户** - 检测到新地址时自动创建账户记录
- **余额更新** - 每笔交易后自动更新用户余额
- **地址缓存** - 使用 Redis 缓存地址与账户 ID 的映射关系

### 4. 多交易类型支持

系统支持所有以太坊兼容的交易类型：

- **Legacy (Type 0)** - 传统交易类型
- **EIP-2930 (Type 1)** - 带访问列表的交易
- **EIP-1559 (Type 2)** - 动态手续费交易
- **EIP-4844 (Type 3)** - Blob 交易（Cancun 升级）
- **EIP-7702 (Type 4)** - 账户代理（临时设置账户代码）

## 💻 开发说明

### 添加新的代币支持

1. 在 `config/config.go` 中添加新代币的地址配置
2. 在 `parser/blockchain_parser.go` 中添加代币特定的事件解析逻辑
3. 在 `service/withdrawal_service.go` 中添加提现代币的逻辑

### 修改交易监听逻辑

核心逻辑在 `parser/blockchain_parser.go`：

- `processBlock()` - 处理单个区块
- `processTransaction()` - 处理单笔交易
- `processBNBTransfer()` - 处理 BNB 转账
- `processERC20Transfer()` - 处理 ERC20 转账

### 添加新的提现代币

在 `service/withdrawal_service.go` 中：

1. 创建新的提现函数（如 `withdrawToken()`）
2. 在 `processWithdrawal()` 中添加新的代币类型判断
3. 实现代币特定的转账逻辑

## 🐳 Docker 部署

### 使用 Docker Compose

项目提供了完整的 `docker-compose.yml` 配置文件，可以快速启动 PostgreSQL 和 Redis：

```bash
# 启动服务
docker compose up -d

# 停止服务
docker compose down

# 查看日志
docker compose logs -f

# 重启服务
docker compose restart
```

### 服务说明

- **PostgreSQL** - 端口 5432，用户名/密码: postgres/postgres
- **Redis** - 端口 6379，无密码

## 📊 数据库结构

### account 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| email | VARCHAR(255) | 用户邮箱 |
| address | VARCHAR(42) | 区块链地址 |
| bnb_amount | VARCHAR(100) | BNB 余额（wei） |
| usdt_amount | VARCHAR(100) | USDT 余额（最小单位） |
| created_time | TIMESTAMP | 创建时间 |
| updated_time | TIMESTAMP | 更新时间 |

### transaction 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| tx_hash | VARCHAR(66) | 交易哈希 |
| block_number | BIGINT | 区块号 |
| from_address | VARCHAR(42) | 发送方地址 |
| to_address | VARCHAR(42) | 接收方地址 |
| value | VARCHAR(100) | 金额（wei） |
| token_decimals | INT | 代币精度 |
| token_address | VARCHAR(42) | 代币合约地址 |
| token_symbol | VARCHAR(32) | 代币符号 |
| status | INT | 交易状态（0=pending,1=success,2=failed） |
| created_time | TIMESTAMP | 创建时间 |

### withdrawal 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| account_id | INT | 账户 ID |
| amount | VARCHAR(100) | 提现金额 |
| token_decimals | INT | 代币精度 |
| token_symbol | VARCHAR(32) | 代币符号 |
| to_address | VARCHAR(42) | 接收地址 |
| tx_hash | VARCHAR(66) | 交易哈希 |
| status | VARCHAR(32) | 状态（init/processing/success/failed） |
| created_time | TIMESTAMP | 创建时间 |
| updated_time | TIMESTAMP | 更新时间 |

## ❓ 常见问题

### 1. 数据库连接失败

**错误**: `failed to connect to database: connection refused`

**解决方案**:
- 确认 PostgreSQL 正在运行：`docker-compose ps`
- 检查数据库配置是否正确
- 确认数据库已创建：`psql -h localhost -U postgres -l`

### 2. Redis 连接失败

**错误**: `Failed to connect to Redis`

**解决方案**:
- 确认 Redis 正在运行：`docker-compose ps`
- 检查 Redis 配置是否正确
- 测试连接：`redis-cli ping`

### 3. RPC 连接问题

**错误**: `failed to dial RPC`

**解决方案**:
- 检查 RPC URL 是否正确
- 尝试更换 RPC 节点
- 确认网络连接正常

### 4. 私钥格式错误

**错误**: `invalid private key`

**解决方案**:
- 确认私钥格式正确（0x 开头或去掉 0x）
- 确认私钥对应的地址与配置一致

### 5. 交易类型不支持

**错误**: `transaction type not supported`

**解决方案**:
- 更新 go-ethereum 到最新版本：`go get github.com/ethereum/go-ethereum@latest`
- 确认程序使用正确的 Signer 处理不同交易类型

## 📝 License

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

如有问题或建议，请提交 Issue。

---

**注意**: 本项目仅供学习和研究使用，请在生产环境中谨慎使用。
