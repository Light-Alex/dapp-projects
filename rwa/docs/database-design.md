# RWA 平台数据库设计文档

## 1. 数据库概述

本项目是一个 **RWA（Real World Assets，真实世界资产）** 交易平台的后端系统，负责将链上事件与传统金融交易（如美股交易）打通。

### 技术栈

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 数据库 | **PostgreSQL** | 关系型数据库，支持 NUMERIC 高精度计算、TIMESTAMPTZ 时区感知时间戳 |
| ORM | **GORM** | Go 语言最流行的 ORM 框架，提供模型映射、自动迁移等能力 |
| ORM 增强 | **gorm-plus** | GORM 的增强库，提供更便捷的查询构建器 |
| 精度计算 | **shopspring/decimal** | Go 高精度十进制库，避免浮点数精度丢失（金融场景必需） |
| 迁移工具 | **golang-migrate** | 数据库版本化迁移管理工具 |

### 设计原则

1. **幂等性** — 关键表通过唯一约束保证事件不被重复处理（如 `orders.client_order_id`、`event_logs(tx_hash, log_index)`）
2. **可追溯性** — 每条订单都记录了链上交易哈希、事件日志 ID，便于链上链下数据对账
3. **高精度** — 金额字段统一使用 `NUMERIC(38,18)` 类型，与 ERC-20 代币的 18 位小数精度一致
4. **时区安全** — 迁移 000004 将所有时间戳字段从 `TIMESTAMP` 升级为 `TIMESTAMPTZ`，避免时区转换问题

---

## 2. ER 关系图

```mermaid
erDiagram
    account ||--o{ orders : "一个账户拥有多个订单"
    account ||--o{ positions : "一个账户拥有多个持仓"
    account ||--o{ deposit_operations : "一个账户拥有多笔充值"
    account ||--o{ withdrawal_operations : "一个账户拥有多笔提现"
    account ||--o{ event_logs : "一个账户关联多条事件日志"

    orders ||--o{ order_executions : "一个订单包含多条成交明细"
    orders ||--o| event_logs : "一个订单可关联一条事件日志"
    orders ||--o{ contract_transactions : "一个订单关联多条合约交易"

    stocks ||--o{ orders : "一只股票被多个订单交易（通过 symbol）"
    stocks ||--o{ positions : "一只股票对应多个持仓（通过 symbol）"

    account {
        BIGSERIAL id PK
        VARCHAR address UK
        BOOLEAN is_active
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    stocks {
        BIGSERIAL id PK
        VARCHAR symbol UK
        VARCHAR name
        VARCHAR exchange
        TEXT about
        VARCHAR status
        VARCHAR contract
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    trading_accounts {
        VARCHAR id PK
        VARCHAR external_account_id
        VARCHAR provider
        VARCHAR account_type
        VARCHAR status
        BOOLEAN is_active
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    positions {
        BIGSERIAL id PK
        BIGINT account_id FK
        VARCHAR symbol
        VARCHAR asset_type
        NUMERIC quantity
        NUMERIC average_price
        NUMERIC market_value
        NUMERIC unrealized_pnl
        NUMERIC realized_pnl
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    orders {
        BIGSERIAL id PK
        VARCHAR client_order_id UK
        BIGINT account_id FK
        VARCHAR symbol
        VARCHAR asset_type
        VARCHAR side
        VARCHAR type
        NUMERIC quantity
        NUMERIC price
        NUMERIC stop_price
        VARCHAR status
        NUMERIC filled_quantity
        NUMERIC filled_price
        NUMERIC remaining_quantity
        NUMERIC escrow_amount
        VARCHAR escrow_asset
        NUMERIC refund_amount
        VARCHAR contract_tx_hash
        VARCHAR execute_tx_hash
        VARCHAR cancel_tx_hash
        BIGINT event_log_id FK
        VARCHAR external_order_id
        VARCHAR time_in_force
        VARCHAR provider
        VARCHAR commission
        VARCHAR commission_asset
        TEXT metadata
        TEXT notes
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
        TIMESTAMPTZ submitted_at
        TIMESTAMPTZ accepted_at
        TIMESTAMPTZ filled_at
        TIMESTAMPTZ cancelled_at
        TIMESTAMPTZ expired_at
    }

    order_executions {
        BIGSERIAL id PK
        BIGINT order_id FK
        VARCHAR execution_id UK
        NUMERIC quantity
        NUMERIC price
        VARCHAR commission
        VARCHAR commission_asset
        VARCHAR provider
        VARCHAR external_id
        TIMESTAMP executed_at
        TIMESTAMP created_at
    }

    event_logs {
        BIGSERIAL id PK
        VARCHAR tx_hash
        BIGINT block_number
        VARCHAR block_hash
        INTEGER log_index
        VARCHAR contract_address
        VARCHAR event_name
        TEXT event_data
        BIGINT account_id FK
        BIGINT order_id FK
        TIMESTAMP processed_at
        VARCHAR status
        TEXT error_message
        INTEGER retry_count
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    event_client_record {
        BIGINT chain_id PK
        BIGINT last_block
        BIGINT last_event_id
        TIMESTAMP update_at
    }

    failed_events {
        BIGSERIAL id PK
        VARCHAR client_order_id
        VARCHAR event_type
        VARCHAR execution_id
        TEXT event_data
        TEXT error_message
        VARCHAR source
        INTEGER retry_count
        VARCHAR status
        TIMESTAMP resolved_at
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    deposit_operations {
        BIGSERIAL id PK
        BIGINT account_id FK
        VARCHAR operation_id
        NUMERIC usdc_amount
        NUMERIC usdm_amount
        VARCHAR status
        VARCHAR tx_hash
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    withdrawal_operations {
        BIGSERIAL id PK
        BIGINT account_id FK
        VARCHAR operation_id
        NUMERIC usdm_amount
        NUMERIC usdc_amount
        VARCHAR status
        VARCHAR tx_hash
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    contract_transactions {
        BIGSERIAL id PK
        VARCHAR tx_hash UK
        VARCHAR method_name
        VARCHAR contract_address
        VARCHAR status
        BIGINT gas_used
        TEXT error_message
        BIGINT related_order_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ confirmed_at
    }
```

---

## 3. 表结构详细说明

### 3.1 account — 用户账户表

> 存储链上用户的钱包地址，是整个系统的用户身份基础。当链上事件（如充值、下单）被捕获时，系统根据钱包地址查找或创建对应的 account 记录。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| address | VARCHAR(255) | NOT NULL, UNIQUE | 用户钱包地址（如 `0xAbC...123`） |
| is_active | BOOLEAN | NOT NULL, DEFAULT true | 账户是否激活 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| account_pkey | id | 主键索引 | 主键查询 |
| idx_account_address | address | 唯一索引 | 根据钱包地址快速查找账户 |

**关联关系：**
- `account.id` 被 `orders.account_id`、`positions.account_id`、`event_logs.account_id`、`deposit_operations.account_id`、`withdrawal_operations.account_id` 引用

**Go 模型：** `libs/core/models/rwa/account.go` -> `Account`

---

### 3.2 stocks — 股票信息表

> 存储平台支持交易的股票（RWA 资产）元信息。每只股票对应一个链上合约地址（`contract` 字段），用于代币化交易。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| symbol | VARCHAR(255) | NOT NULL, UNIQUE | 股票代码（如 `AAPL`、`TSLA`） |
| name | VARCHAR(255) | NOT NULL | 股票全称（如 `Apple Inc.`） |
| exchange | VARCHAR(255) | NOT NULL | 交易所（如 `NASDAQ`、`NYSE`） |
| about | TEXT | 可选 | 股票简介 |
| status | VARCHAR(50) | NOT NULL, DEFAULT 'active' | 状态：`active` / `inactive` |
| contract | VARCHAR(255) | 可选 | 对应的链上合约地址 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| stocks_pkey | id | 主键索引 | 主键查询 |
| idx_stocks_symbol | symbol | 唯一索引 | 根据股票代码快速查找 |

**关联关系：**
- 通过 `symbol` 字段与 `orders.symbol`、`positions.symbol` 逻辑关联（非外键约束）

**Go 模型：** `libs/core/models/rwa/stock.go` -> `Stock`

**枚举值（StockStatus）：**
- `active` — 可交易
- `inactive` — 已下架

---

### 3.3 trading_accounts — Alpaca 交易账户表

> 存储对接外部交易提供商（如 Alpaca）的账户信息。系统通过此账户向 Alpaca 提交真实美股订单。整个平台可能只有一个或少量交易账户。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(255) | PRIMARY KEY | 主键（非自增，使用外部分配的 ID） |
| external_account_id | VARCHAR(255) | 可选 | 外部提供商的账户 ID |
| provider | VARCHAR(255) | 可选 | 提供商名称（如 `alpaca`） |
| account_type | VARCHAR(255) | 可选 | 账户类型（如 `live`、`paper`） |
| status | VARCHAR(255) | 可选 | 账户状态 |
| is_active | BOOLEAN | NOT NULL | 是否启用 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| trading_accounts_pkey | id | 主键索引 | 主键查询 |

**关联关系：**
- 独立表，作为系统级配置使用，不直接与其他表建立外键关系

**Go 模型：** `libs/core/models/rwa/trading.go` -> `TradingAccount`

---

### 3.4 positions — 持仓表

> 记录每个账户在每只股票上的持仓情况。当订单成交后，系统更新对应的持仓记录（数量、均价、盈亏等）。每个账户对每只股票只有一条持仓记录（通过唯一约束保证）。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| account_id | BIGINT | NOT NULL | 所属账户 ID |
| symbol | VARCHAR(255) | NOT NULL | 股票代码 |
| asset_type | VARCHAR(255) | 可选 | 资产类型（`stock` / `crypto` 等） |
| quantity | NUMERIC | NOT NULL | 持有数量 |
| average_price | NUMERIC | NOT NULL | 持仓均价 |
| market_value | NUMERIC | NOT NULL | 当前市值 |
| unrealized_pnl | NUMERIC | NOT NULL | 未实现盈亏（浮盈浮亏） |
| realized_pnl | NUMERIC | NOT NULL | 已实现盈亏 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| positions_pkey | id | 主键索引 | 主键查询 |
| idx_positions_account_id | account_id | 普通索引 | 查询某账户的所有持仓 |
| idx_positions_symbol | symbol | 普通索引 | 查询某只股票的所有持仓 |
| uq_positions_account_symbol | (account_id, symbol) | 唯一约束 | 保证每个账户每只股票只有一条记录 |

**关联关系：**
- `positions.account_id` -> `account.id`（逻辑外键）
- `positions.symbol` -> `stocks.symbol`（逻辑关联）

**Go 模型：** `libs/core/models/rwa/trading.go` -> `Position`

---

### 3.5 orders — 订单表

> **核心表**。记录用户的交易订单，从链上事件触发创建，到 Alpaca 执行，再到最终成交/取消/过期的完整生命周期。每个订单记录了链上交易哈希（溯源）、外部订单 ID（对账）、托管金额（资金安全）等关键信息。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| client_order_id | VARCHAR(255) | NOT NULL, UNIQUE | 客户端订单 ID（链上事件生成，用于幂等去重） |
| account_id | BIGINT | NOT NULL | 所属账户 ID |
| symbol | VARCHAR(255) | NOT NULL | 交易股票代码 |
| asset_type | VARCHAR(255) | NOT NULL | 资产类型 |
| side | VARCHAR(50) | NOT NULL | 买卖方向：`buy` / `sell` |
| type | VARCHAR(50) | NOT NULL | 订单类型：`market` / `limit` / `stop` / `stop_limit` |
| quantity | NUMERIC | NOT NULL | 下单数量 |
| price | NUMERIC | NOT NULL | 下单价格（市价单时为 0） |
| stop_price | NUMERIC | NOT NULL | 止损价格（仅 stop/stop_limit 类型使用） |
| status | VARCHAR(50) | NOT NULL | 订单状态（见状态机章节） |
| filled_quantity | NUMERIC | NOT NULL | 已成交数量 |
| filled_price | NUMERIC | NOT NULL | 成交均价 |
| remaining_quantity | NUMERIC | NOT NULL | 剩余未成交数量 |
| escrow_amount | NUMERIC(38,18) | 可选 | 托管金额（买入时冻结的 USDC/USDm 数量） |
| escrow_asset | VARCHAR(42) | 可选 | 托管资产地址（代币合约地址） |
| refund_amount | NUMERIC(38,18) | 可选 | 退还金额（订单取消/部分成交后退还的金额） |
| contract_tx_hash | VARCHAR(255) | 可选 | 链上下单交易哈希（用户发起下单的 tx） |
| execute_tx_hash | VARCHAR(66) | 可选 | 执行交易哈希（后端调用合约执行成交的 tx） |
| cancel_tx_hash | VARCHAR(66) | 可选 | 取消交易哈希（后端调用合约取消订单的 tx） |
| event_log_id | BIGINT | 可选 | 关联的事件日志 ID |
| external_order_id | VARCHAR(255) | 可选 | Alpaca 返回的订单 ID |
| time_in_force | VARCHAR(10) | DEFAULT 'DAY' | 有效期类型（`DAY` / `GTC` / `IOC` 等） |
| provider | VARCHAR(255) | 可选 | 交易提供商（如 `alpaca`） |
| commission | VARCHAR(255) | 可选 | 手续费金额 |
| commission_asset | VARCHAR(255) | 可选 | 手续费币种 |
| metadata | TEXT | 可选 | 扩展元数据（JSON 格式） |
| notes | TEXT | 可选 | 备注信息 |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 更新时间 |
| submitted_at | TIMESTAMPTZ | 可选 | 提交到 Alpaca 的时间 |
| accepted_at | TIMESTAMPTZ | 可选 | Alpaca 接受订单的时间 |
| filled_at | TIMESTAMPTZ | 可选 | 完全成交时间 |
| cancelled_at | TIMESTAMPTZ | 可选 | 取消时间 |
| expired_at | TIMESTAMPTZ | 可选 | 过期时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| orders_pkey | id | 主键索引 | 主键查询 |
| uq_orders_client_order_id | client_order_id | 唯一约束 | 幂等去重，防止同一链上事件重复创建订单 |
| idx_orders_account_id | account_id | 普通索引 | 查询某账户的所有订单 |
| idx_orders_symbol | symbol | 普通索引 | 按股票代码筛选订单 |
| idx_orders_contract_tx_hash | contract_tx_hash | 普通索引 | 根据链上 tx hash 查找订单 |
| idx_orders_event_log_id | event_log_id | 普通索引 | 根据事件日志查找关联订单 |

**关联关系：**
- `orders.account_id` -> `account.id`
- `orders.event_log_id` -> `event_logs.id`
- `orders.id` <- `order_executions.order_id`（一对多）
- `orders.id` <- `contract_transactions.related_order_id`

**Go 模型：** `libs/core/models/rwa/order.go` -> `Order`

---

### 3.6 order_executions — 成交明细表

> 记录每笔订单的逐笔成交记录。一个订单可能有多次部分成交（partial fill），每次成交都会生成一条 execution 记录。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| order_id | BIGINT | NOT NULL | 所属订单 ID |
| execution_id | VARCHAR(255) | UNIQUE | 成交 ID（来自 Alpaca，用于幂等去重） |
| quantity | NUMERIC(38,18) | — | 本次成交数量 |
| price | NUMERIC(38,18) | — | 本次成交价格 |
| commission | VARCHAR(255) | 可选 | 本次成交的手续费 |
| commission_asset | VARCHAR(255) | 可选 | 手续费币种 |
| provider | VARCHAR(255) | 可选 | 交易提供商 |
| external_id | VARCHAR(255) | 可选 | 外部系统 ID |
| executed_at | TIMESTAMP | 可选 | 成交时间（Alpaca 返回的时间） |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 记录创建时间 |

> **注意：** `quantity` 和 `price` 初始迁移为 `VARCHAR(255)` 类型，在迁移 000004 中修正为 `NUMERIC(38,18)`。

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| order_executions_pkey | id | 主键索引 | 主键查询 |
| idx_order_executions_order_id | order_id | 普通索引 | 查询某订单的所有成交记录 |
| uq_order_executions_execution_id | execution_id | 唯一约束 | 防止同一笔成交被重复入库 |

**关联关系：**
- `order_executions.order_id` -> `orders.id`

**Go 模型：** `libs/core/models/rwa/order.go` -> `OrderExecution`

---

### 3.7 event_logs — 链上事件日志表

> 记录从区块链上捕获的所有合约事件。每个事件对应一笔链上交易中的一条 log。系统解析这些事件来驱动订单创建、充值确认等业务逻辑。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| tx_hash | VARCHAR(255) | NOT NULL | 交易哈希 |
| block_number | BIGINT | NOT NULL | 区块高度 |
| block_hash | VARCHAR(255) | 可选 | 区块哈希 |
| log_index | INTEGER | NOT NULL | 事件在交易中的索引位置 |
| contract_address | VARCHAR(255) | NOT NULL | 触发事件的合约地址 |
| event_name | VARCHAR(255) | 可选 | 事件名称（如 `OrderPlaced`、`Deposit`） |
| event_data | TEXT | 可选 | 事件原始数据（JSON 格式） |
| account_id | BIGINT | 可选 | 关联的账户 ID |
| order_id | BIGINT | 可选 | 关联的订单 ID |
| processed_at | TIMESTAMP | 可选 | 处理完成时间（NULL 表示未处理） |
| status | VARCHAR(50) | NOT NULL, DEFAULT 'pending' | 处理状态：`pending` / `processed` / `failed` |
| error_message | TEXT | 可选 | 处理失败时的错误信息 |
| retry_count | INTEGER | NOT NULL, DEFAULT 0 | 重试次数 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| event_logs_pkey | id | 主键索引 | 主键查询 |
| idx_event_logs_tx_hash | tx_hash | 普通索引 | 根据交易哈希查找事件 |
| idx_event_logs_block_number | block_number | 普通索引 | 按区块高度范围查询 |
| idx_event_logs_contract_address | contract_address | 普通索引 | 按合约地址筛选 |
| idx_event_logs_account_id | account_id | 普通索引 | 查询某账户相关的事件 |
| idx_event_logs_order_id | order_id | 普通索引 | 查询某订单相关的事件 |
| uq_event_logs_tx_log | (tx_hash, log_index) | 唯一约束 | 保证同一交易中的同一事件不会被重复记录 |

**关联关系：**
- `event_logs.account_id` -> `account.id`
- `event_logs.order_id` -> `orders.id`
- `event_logs.id` <- `orders.event_log_id`

**Go 模型：** `libs/core/models/rwa/event_log.go` -> `EventLog`

---

### 3.8 event_client_record — 区块同步进度表

> 记录事件监听客户端的同步进度。每条链（通过 `chain_id` 区分）有一条记录，记录已处理到的最新区块和事件 ID。系统重启时从此记录恢复，避免重复扫描已处理的区块。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| chain_id | BIGINT | PRIMARY KEY | 链 ID（如 Ethereum Mainnet = 1, Sepolia = 11155111） |
| last_block | BIGINT | NOT NULL, DEFAULT 0 | 最后处理的区块高度 |
| last_event_id | BIGINT | NOT NULL, DEFAULT 0 | 最后处理的事件 ID |
| update_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | 最后更新时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| event_client_record_pkey | chain_id | 主键索引 | 主键查询（每条链一条记录） |

**关联关系：**
- 独立表，不与其他表建立外键关系

**Go 模型：** `libs/core/models/rwa/event_client_record.go` -> `EventClientRecord`

---

### 3.9 failed_events — 失败事件记录表

> 存储 WebSocket 交易更新事件（来自 Alpaca）处理失败的记录。当 Alpaca 推送的 trade update 事件处理异常（如对应订单找不到）时，将原始事件数据保存到此表，供人工排查或后续重试。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| client_order_id | VARCHAR(255) | 可选 | 关联的客户端订单 ID |
| event_type | VARCHAR(50) | NOT NULL | 事件类型（如 `fill`、`partial_fill`、`canceled`） |
| execution_id | VARCHAR(255) | 可选 | 成交 ID |
| event_data | TEXT | NOT NULL | 原始事件数据（JSON 格式，完整保留以便排查） |
| error_message | TEXT | 可选 | 错误信息 |
| source | VARCHAR(50) | NOT NULL, DEFAULT 'alpaca' | 事件来源 |
| retry_count | INTEGER | NOT NULL, DEFAULT 0 | 重试次数 |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | 处理状态：`pending` / `resolved` |
| resolved_at | TIMESTAMP | 可选 | 解决时间 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 更新时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| failed_events_pkey | id | 主键索引 | 主键查询 |
| idx_failed_events_client_order_id | client_order_id | 普通索引 | 根据订单 ID 查找失败事件 |
| idx_failed_events_status | status | 普通索引 | 筛选待处理的失败事件 |
| idx_failed_events_created_at | created_at | 普通索引 | 按时间范围查询 |

**关联关系：**
- `failed_events.client_order_id` -> `orders.client_order_id`（逻辑关联）

**Go 模型：** `libs/core/models/rwa/failed_event.go` -> `FailedEvent`

---

### 3.10 deposit_operations — 充值操作表

> 记录用户从链上充值 USDC 并兑换为 USDm（平台稳定币）的操作。每笔充值绑定一个账户和链上交易。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| account_id | BIGINT | NOT NULL | 所属账户 ID |
| operation_id | VARCHAR(66) | 可选 | 操作 ID（链上事件生成） |
| usdc_amount | NUMERIC(38,18) | NOT NULL | 充值的 USDC 数量 |
| usdm_amount | NUMERIC(38,18) | NOT NULL | 兑换得到的 USDm 数量 |
| status | VARCHAR(50) | NOT NULL, DEFAULT 'pending' | 状态：`pending` / `completed` / `failed` |
| tx_hash | VARCHAR(66) | 可选 | 链上交易哈希 |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | 更新时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| deposit_operations_pkey | id | 主键索引 | 主键查询 |
| idx_deposit_ops_account | account_id | 普通索引 | 查询某账户的充值记录 |
| idx_deposit_ops_status | status | 普通索引 | 筛选特定状态的充值操作 |

**关联关系：**
- `deposit_operations.account_id` -> `account.id`

---

### 3.11 withdrawal_operations — 提现操作表

> 记录用户将 USDm 兑换回 USDC 并提现到链上钱包的操作。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| account_id | BIGINT | NOT NULL | 所属账户 ID |
| operation_id | VARCHAR(66) | 可选 | 操作 ID |
| usdm_amount | NUMERIC(38,18) | NOT NULL | 赎回的 USDm 数量 |
| usdc_amount | NUMERIC(38,18) | NOT NULL | 获得的 USDC 数量 |
| status | VARCHAR(50) | NOT NULL, DEFAULT 'pending' | 状态：`pending` / `completed` / `failed` |
| tx_hash | VARCHAR(66) | 可选 | 链上交易哈希 |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | 更新时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| withdrawal_operations_pkey | id | 主键索引 | 主键查询 |
| idx_withdrawal_ops_account | account_id | 普通索引 | 查询某账户的提现记录 |
| idx_withdrawal_ops_status | status | 普通索引 | 筛选特定状态的提现操作 |

**关联关系：**
- `withdrawal_operations.account_id` -> `account.id`

---

### 3.12 contract_transactions — 合约交易记录表

> 记录后端主动发起的链上合约调用（如执行成交、取消订单）。用于追踪每笔链上交易的确认状态和 gas 消耗。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 自增主键 |
| tx_hash | VARCHAR(66) | NOT NULL, UNIQUE | 交易哈希 |
| method_name | VARCHAR(100) | NOT NULL | 调用的合约方法名（如 `executeBuy`、`cancelOrder`） |
| contract_address | VARCHAR(42) | NOT NULL | 目标合约地址 |
| status | VARCHAR(50) | NOT NULL, DEFAULT 'pending' | 状态：`pending` / `confirmed` / `failed` |
| gas_used | BIGINT | 可选 | 实际消耗的 gas 量 |
| error_message | TEXT | 可选 | 交易失败时的错误信息 |
| related_order_id | BIGINT | 可选 | 关联的订单 ID |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | 创建时间 |
| confirmed_at | TIMESTAMPTZ | 可选 | 确认上链时间 |

**索引：**

| 索引名 | 字段 | 类型 | 用途 |
|--------|------|------|------|
| contract_transactions_pkey | id | 主键索引 | 主键查询 |
| contract_transactions_tx_hash_key | tx_hash | 唯一索引 | 根据 tx hash 快速查找 |
| idx_contract_tx_status | status | 普通索引 | 筛选待确认的交易 |
| idx_contract_tx_order | related_order_id | 普通索引 | 查询某订单关联的合约交易 |

**关联关系：**
- `contract_transactions.related_order_id` -> `orders.id`

---

## 4. 枚举/常量定义

所有枚举类型定义在 `libs/core/models/rwa/trading.go` 中。

### 4.1 AssetType — 资产类型

| 值 | Go 常量 | 说明 |
|----|---------|------|
| `stock` | `AssetTypeStock` | 股票 |
| `crypto` | `AssetTypeCrypto` | 加密货币 |
| `forex` | `AssetTypeForex` | 外汇 |
| `bond` | `AssetTypeBond` | 债券 |
| `option` | `AssetTypeOption` | 期权 |

### 4.2 OrderSide — 订单方向

| 值 | Go 常量 | 说明 |
|----|---------|------|
| `buy` | `OrderSideBuy` | 买入 |
| `sell` | `OrderSideSell` | 卖出 |

### 4.3 OrderType — 订单类型

| 值 | Go 常量 | 说明 |
|----|---------|------|
| `market` | `OrderTypeMarket` | 市价单 — 以当前市场价格立即成交 |
| `limit` | `OrderTypeLimit` | 限价单 — 指定价格，到达时成交 |
| `stop` | `OrderTypeStop` | 止损单 — 到达止损价时变为市价单 |
| `stop_limit` | `OrderTypeStopLimit` | 止损限价单 — 到达止损价时变为限价单 |

### 4.4 OrderStatus — 订单状态

| 值 | Go 常量 | 说明 |
|----|---------|------|
| `new` | `OrderStatusNew` | 新建 — 刚从链上事件创建，尚未提交到 Alpaca |
| `pending` | `OrderStatusPending` | 待处理 — 已提交到 Alpaca，等待确认 |
| `accepted` | `OrderStatusAccepted` | 已接受 — Alpaca 已接受，等待市场成交 |
| `rejected` | `OrderStatusRejected` | 已拒绝 — Alpaca 拒绝了订单（如余额不足） |
| `filled` | `OrderStatusFilled` | 已成交 — 完全成交 |
| `partially_filled` | `OrderStatusPartiallyFilled` | 部分成交 — 仅部分数量已成交 |
| `cancel_requested` | `OrderStatusCancelRequested` | 取消请求中 — 已发起取消请求，等待确认 |
| `cancelled` | `OrderStatusCancelled` | 已取消 — 订单被成功取消 |
| `expired` | `OrderStatusExpired` | 已过期 — 订单到期未成交自动过期 |

### 4.5 StockStatus — 股票状态

| 值 | Go 常量 | 说明 |
|----|---------|------|
| `active` | `StockStatusActive` | 可交易 |
| `inactive` | `StockStatusInactive` | 已下架 |

---

## 5. 订单状态机

订单从创建到结束的完整状态流转图：

```mermaid
stateDiagram-v2
    [*] --> new : 链上事件触发创建订单

    new --> pending : 提交到 Alpaca
    new --> rejected : 参数校验失败

    pending --> accepted : Alpaca 接受订单
    pending --> rejected : Alpaca 拒绝订单

    accepted --> filled : 完全成交
    accepted --> partially_filled : 部分成交
    accepted --> cancel_requested : 用户请求取消
    accepted --> expired : 订单过期（如 DAY 单收盘未成交）
    accepted --> cancelled : 直接取消（Alpaca 侧取消）

    partially_filled --> filled : 剩余全部成交
    partially_filled --> cancel_requested : 用户请求取消剩余部分
    partially_filled --> cancelled : Alpaca 侧取消剩余部分
    partially_filled --> expired : 订单过期

    cancel_requested --> cancelled : 取消确认
    cancel_requested --> filled : 取消前已全部成交（竞态）

    rejected --> [*]
    filled --> [*]
    cancelled --> [*]
    expired --> [*]
```

### 状态流转说明

| 状态变迁 | 触发条件 | 系统行为 |
|---------|---------|---------|
| `[*]` -> `new` | 链上 `OrderPlaced` 事件被捕获 | 解析事件数据，创建订单记录，冻结托管金额 |
| `new` -> `pending` | 系统调用 Alpaca API 提交订单 | 记录 `submitted_at`，设置 `external_order_id` |
| `pending` -> `accepted` | 收到 Alpaca WebSocket `accepted` 事件 | 记录 `accepted_at` |
| `accepted` -> `partially_filled` | 收到 Alpaca `partial_fill` 事件 | 更新 `filled_quantity`、`filled_price`，创建 `order_executions` 记录 |
| `accepted`/`partially_filled` -> `filled` | 收到 Alpaca `fill` 事件 | 更新 `filled_at`，更新持仓，调用合约执行链上成交 |
| `*` -> `cancelled` | 收到 Alpaca `canceled` 事件 | 记录 `cancelled_at`，调用合约退还托管金额 |
| `*` -> `expired` | 收到 Alpaca `expired` 事件 | 记录 `expired_at`，调用合约退还托管金额 |
| `*` -> `rejected` | 订单被拒绝 | 调用合约退还托管金额 |

---

## 6. 数据流

### 6.1 链上事件 -> 数据库 写入路径

```
区块链节点
  │
  ▼
Event Listener（事件监听器）
  │  - 从 event_client_record 获取上次同步的区块高度
  │  - 从 last_block+1 开始扫描新区块
  │
  ▼
解析合约事件 Log
  │  - 匹配 event signature（如 OrderPlaced、Deposit）
  │  - 解码 event data
  │
  ▼
写入 event_logs 表
  │  - status = 'pending'
  │  - 唯一约束 (tx_hash, log_index) 保证幂等
  │
  ▼
事件处理器（Event Processor）
  │  - 根据 event_name 分发到对应处理逻辑
  │
  ├──▶ OrderPlaced 事件
  │     1. 查找或创建 account（根据 address）
  │     2. 创建 orders 记录（status = 'new'）
  │     3. 调用 Alpaca API 下单
  │     4. 更新 orders（status = 'pending'，填入 external_order_id）
  │     5. 更新 event_logs（status = 'processed'）
  │
  ├──▶ Deposit 事件
  │     1. 查找或创建 account
  │     2. 创建 deposit_operations 记录
  │     3. 更新 event_logs（status = 'processed'）
  │
  ├──▶ Withdrawal 事件
  │     1. 查找 account
  │     2. 创建 withdrawal_operations 记录
  │     3. 更新 event_logs（status = 'processed'）
  │
  ▼
更新 event_client_record
  │  - 更新 last_block、last_event_id
  │  - 下次启动从此处继续
```

### 6.2 Alpaca WebSocket -> 数据库 写入路径

```
Alpaca WebSocket
  │  - 推送 trade update 事件（fill、partial_fill、canceled 等）
  │
  ▼
Trade Update Handler
  │  - 根据 client_order_id 查找 orders 记录
  │  - 如果找不到 -> 写入 failed_events 表
  │
  ├──▶ fill / partial_fill
  │     1. 创建 order_executions 记录
  │     2. 更新 orders（filled_quantity, filled_price, status）
  │     3. 更新 positions（quantity, average_price）
  │     4. 如果完全成交：调用合约 executeBuy/executeSell
  │     5. 记录 contract_transactions
  │
  ├──▶ canceled
  │     1. 更新 orders（status = 'cancelled', cancelled_at）
  │     2. 调用合约退还托管金额
  │     3. 记录 contract_transactions
  │
  ├──▶ expired
  │     1. 更新 orders（status = 'expired', expired_at）
  │     2. 调用合约退还托管金额
  │
  ▼
处理失败时
  └──▶ 写入 failed_events 表（保留原始 event_data 供排查）
```

---

## 7. 迁移管理

### 7.1 迁移文件结构

迁移文件位于 `migrations/rwa/` 目录，使用 [golang-migrate](https://github.com/golang-migrate/migrate) 工具管理。

```
migrations/rwa/
├── 000001_rwa.up.sql                      # 基础表：account, stocks, trading_accounts, positions, orders, order_executions, event_logs
├── 000001_rwa.down.sql                    # 回滚：删除基础表
├── 000002_event_client_record.up.sql      # 新增 event_client_record 表
├── 000002_event_client_record.down.sql    # 回滚：删除 event_client_record
├── 000003_add_deposit_withdrawal.up.sql   # 新增 deposit_operations, withdrawal_operations, contract_transactions
├── 000003_add_deposit_withdrawal.down.sql # 回滚：删除充提和合约交易表
├── 000004_fix_schema.up.sql               # 修复：补充字段、修正类型、添加唯一约束、TIMESTAMPTZ
├── 000004_fix_schema.down.sql             # 回滚：撤销修复
├── 000005_add_failed_events.up.sql        # 新增 failed_events 表
└── 000005_add_failed_events.down.sql      # 回滚：删除 failed_events
```

### 7.2 迁移版本演进

| 版本 | 文件 | 主要变更 |
|------|------|---------|
| 000001 | `rwa.up.sql` | 创建 7 张基础表：account、stocks、trading_accounts、positions、orders、order_executions、event_logs |
| 000002 | `event_client_record.up.sql` | 新增 event_client_record 表，用于记录区块链事件的同步进度 |
| 000003 | `add_deposit_withdrawal.up.sql` | 新增 deposit_operations、withdrawal_operations、contract_transactions 三张表 |
| 000004 | `fix_schema.up.sql` | 修复和完善：为 orders 表补充 `time_in_force`、`escrow_amount`、`escrow_asset`、`refund_amount`、`execute_tx_hash`、`cancel_tx_hash`、`accepted_at` 字段；将 order_executions 的 quantity/price 从 VARCHAR 修正为 NUMERIC；添加多个唯一约束；时间戳统一为 TIMESTAMPTZ |
| 000005 | `add_failed_events.up.sql` | 新增 failed_events 表，存储处理失败的交易更新事件 |

### 7.3 常用命令

```bash
# 安装 golang-migrate CLI（macOS）
brew install golang-migrate

# 执行所有待执行的迁移（向上迁移到最新版本）
migrate -path migrations/rwa -database "postgres://user:pass@localhost:5432/rwa?sslmode=disable" up

# 回滚最近一次迁移
migrate -path migrations/rwa -database "postgres://user:pass@localhost:5432/rwa?sslmode=disable" down 1

# 回滚所有迁移
migrate -path migrations/rwa -database "postgres://user:pass@localhost:5432/rwa?sslmode=disable" down

# 迁移到指定版本
migrate -path migrations/rwa -database "postgres://user:pass@localhost:5432/rwa?sslmode=disable" goto 3

# 查看当前迁移版本
migrate -path migrations/rwa -database "postgres://user:pass@localhost:5432/rwa?sslmode=disable" version

# 强制设置版本（修复 dirty 状态时使用，谨慎操作）
migrate -path migrations/rwa -database "postgres://user:pass@localhost:5432/rwa?sslmode=disable" force 4

# 创建新的迁移文件
migrate create -ext sql -dir migrations/rwa -seq add_new_feature
```

### 7.4 注意事项

1. **迁移顺序不可跳过** — 每次 `up` 会按序号从当前版本依次执行到最新版本
2. **down 文件必须与 up 对应** — 每个 up 迁移都要有对应的 down 回滚脚本
3. **使用 `IF NOT EXISTS` / `IF EXISTS`** — 保证迁移脚本可重复执行不报错
4. **ALTER 操作用 `ADD COLUMN IF NOT EXISTS`** — 避免重复添加字段导致迁移失败
5. **dirty 状态修复** — 如果迁移中途失败导致 `dirty` 标记，需要手动修复数据库后用 `force` 命令重置版本号
