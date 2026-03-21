# Blockchain Parser

区块链监听与解析服务，用于监听 BSC 链上的交易事件，自动更新用户余额，并处理提现请求。

## 功能特性

- 🔗 **区块链监听**：实时监听 BSC 链上的区块和交易
- 💰 **余额自动更新**：自动追踪并更新用户的 BNB 和 ERC20 代币余额
- 💸 **提现处理**：自动处理用户的 BNB 和 ERC20 代币提现请求
- 🔄 **交易去重**：使用 Redis 和数据库确保交易不被重复处理
- 📊 **数据持久化**：使用 PostgreSQL 存储交易记录和用户状态

## 技术栈

- **运行时**: Node.js
- **区块链交互**: ethers.js v6
- **ORM**: TypeORM
- **数据库**: PostgreSQL
- **缓存**: Redis (ioredis)
- **转译**: Babel

## 项目结构

```
blockchain-parser/
├── src/
│   ├── app.js                      # 应用入口
│   ├── config.js                   # 配置文件
│   ├── database.js                 # 数据库连接
│   ├── blockchainParser.js         # 区块链解析核心逻辑
│   ├── withdrawalService.js        # 提现服务
│   ├── redis.js                    # Redis 服务
│   ├── entity/                     # 数据库实体
│   │   ├── Account.js              # 用户账户实体
│   │   ├── Transaction.js          # 交易记录实体
│   │   └── Withdrawal.js           # 提现记录实体
│   └── ...
├── package.json
└── .env                            # 环境变量配置
```

## 快速开始

### 1. 环境要求

- Node.js >= 16
- PostgreSQL >= 12
- Redis >= 6

### 2. 安装依赖

```bash
npm install
```

### 3. 配置环境变量

创建 `.env` 文件并配置以下变量：

```bash
# BSC RPC 节点
BSC_RPC_URL=https://bsc-dataseed.bnbchain.org
CHAIN_ID=56

# 项目方地址（用于监听）
PROJECT_ADDRESS=0x3Ca1392e4A95Aa0f83e97458Ab4495a58cA91bd6

# ERC20 代币合约地址（如 USDT）
USDT_ADDRESS=0x55d398326f99059fF775485246999027B3197955

# 提现服务私钥
SEPOLIA_PRIVATE_KEY=your_private_key_here

# PostgreSQL 数据库配置
DB_HOST=172.19.62.197
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_NAME=blockchain_parser

# Redis 配置
REDIS_HOST=172.19.62.197
REDIS_PORT=6379
REDIS_PASSWORD=

# 扫描配置
SCAN_INTERVAL=3000              # 扫描间隔（毫秒）
CONFIRMATION_BLOCKS=6           # 确认区块数
```

### 4. 初始化数据库

PostgreSQL 会自动创建表结构，无需手动执行 SQL。

### 5. 启动服务

```bash
npm start
```

## 数据库表结构

### account（用户账户表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| email | VARCHAR(128) | 邮箱（唯一） |
| address | VARCHAR(128) | 钱包地址（唯一） |
| bnb_amount | DECIMAL(36, 18) | BNB 余额 |
| usdt_amount | DECIMAL(36, 6) | USDT 余额 |
| created_time | TIMESTAMP | 创建时间 |
| updated_time | TIMESTAMP | 更新时间 |

### transaction（交易记录表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| tx_hash | VARCHAR(66) | 交易哈希（唯一） |
| block_number | BIGINT | 区块号 |
| from_address | VARCHAR(128) | 发送方地址 |
| to_address | VARCHAR(128) | 接收方地址 |
| value | DECIMAL(40, 18) | 交易金额（原始值） |
| token_decimals | INTEGER | 代币精度 |
| token_address | VARCHAR(128) | 代币合约地址 |
| token_symbol | VARCHAR(10) | 代币符号 |
| status | SMALLINT | 交易状态（0: 失败, 1: 成功） |
| created_time | TIMESTAMP | 创建时间 |

### withdrawal（提现记录表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| account_id | INTEGER | 关联用户 ID |
| amount | DECIMAL | 提现金额 |
| token_decimals | INTEGER | 代币精度 |
| token_symbol | VARCHAR(10) | 代币符号（BNB/USDT） |
| to_address | VARCHAR(128) | 接收地址 |
| tx_hash | VARCHAR(66) | 交易哈希 |
| status | VARCHAR(20) | 状态（init/processing/success/failed） |
| created_time | TIMESTAMP | 创建时间 |

## 工作原理

### 区块链监听流程

```
1. 定时扫描新区块（每 3 秒）
   ↓
2. 获取区块中的所有交易
   ↓
3. 过滤相关交易（涉及项目方地址或代币合约）
   ↓
4. 解析交易事件（BNB 转账、ERC20 Transfer）
   ↓
5. 更新用户余额和交易记录
```

### 提现处理流程

```
1. 定时查询待处理提现（每 5 秒）
   ↓
2. 原子性更新提现状态（防止重复处理）
   ↓
3. 发起链上交易
   ↓
4. 监听交易确认（等待 3 个确认）
   ↓
5. 更新提现状态（成功/失败）
```

## 核心功能说明

### 交易去重机制

- **Redis 锁**: 防止同一交易被并发处理
- **数据库唯一索引**: `tx_hash` 字段确保交易不重复
- **原子性更新**: 使用 `UPDATE ... WHERE` 防止竞态条件

### 余额更新策略

- **BNB 转账**: 扣除发送方 gas 费和转账金额
- **ERC20 转账**: 扣除/增加对应代币余额
- **提现失败**: 自动回滚用户余额

## 常见问题

### Q: 交易被重复处理如何解决？

A: 检查以下几点：
1. Redis 是否正常运行
2. 数据库唯一索引是否生效
3. 提现服务的原子性更新是否正确执行

### Q: 余额不准确？

A: 可能原因：
1. 数据库字段精度不足（`DECIMAL` 精度配置）
2. BigInt 转换为字符串时丢失精度
3. 提现失败后未正确回滚

### Q: 如何添加新的 ERC20 代币？

A:
1. 在 `.env` 中添加代币合约地址
2. 在 `blockchainParser.js` 中添加对应的合约实例
3. 根据需要调整数据库表结构

## 开发说明

### 代码风格

- 使用装饰器语法（需要 Babel 转译）
- 异步操作使用 `async/await`
- 错误处理使用 `try/catch`

### 调试模式

在代码中添加 `console.log` 输出调试信息：

```javascript
console.log(`Processing transaction: ${tx.hash}`);
console.log(`Updated balance: ${newBalance}`);
```

## 许可证

ISC
