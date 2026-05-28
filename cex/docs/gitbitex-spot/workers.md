# Worker 处理管道

## 1. 概述

Worker 处理管道是连接撮合引擎与数据持久化、资金结算的核心环节。系统通过多种 Worker 将 Kafka 中的撮合日志转化为数据库记录，再通过 MySQL Binlog CDC 驱动 Redis 任务队列，最终由 Executor 完成资金结算。

## 2. 整体管道架构

```mermaid
graph TB
    subgraph 撮合引擎输出
        Kafka["Kafka<br/>matching_message_{productId}<br/>撮合日志"]
    end

    subgraph "数据 Maker (Kafka → MySQL)"
        FillMaker["FillMaker<br/>生成成交明细"]
        TradeMaker["TradeMaker<br/>生成交易记录"]
        TickMaker["TickMaker<br/>生成K线数据"]
    end

    subgraph MySQL
        FillTable["g_fill 表"]
        TradeTable["g_trade 表"]
        TickTable["g_tick 表"]
        OrderTable["g_order 表"]
        BillTable["g_bill 表"]
        AccountTable["g_account 表"]
    end

    subgraph "Binlog CDC"
        BinLog["BinLogStream<br/>MySQL → Redis"]
    end

    subgraph "Redis 任务队列"
        FillQueue["g_fill (List)"]
        BillQueue["g_bill (List)"]
    end

    subgraph "Executor (Redis → MySQL)"
        FillExecutor["FillExecutor x10<br/>成交结算"]
        BillExecutor["BillExecutor x10<br/>账单结算"]
    end

    Kafka --> FillMaker & TradeMaker & TickMaker

    FillMaker -->|INSERT IGNORE| FillTable
    TradeMaker -->|INSERT IGNORE| TradeTable
    TickMaker -->|REPLACE INTO| TickTable

    FillTable -->|Binlog| BinLog
    BillTable -->|Binlog| BinLog

    BinLog -->|LPUSH| FillQueue
    BinLog -->|LPUSH| BillQueue

    FillQueue -->|BRPOP| FillExecutor
    BillQueue -->|BRPOP| BillExecutor

    FillExecutor -->|更新| OrderTable
    FillExecutor -->|创建结算| BillTable

    BillExecutor -->|更新| AccountTable

    style Kafka fill:#bbdefb
    style BinLog fill:#e1bee7
    style FillExecutor fill:#c8e6c9
    style BillExecutor fill:#c8e6c9
```

## 3. FillMaker - 成交记录生成器

### 职责

消费 Kafka 撮合日志中的 MatchLog，为 Taker 和 Maker 双方各创建一条 Fill 记录。

### 工作流程

```mermaid
graph TB
    Kafka["Kafka MatchLog"] --> Parse["解析 MatchLog"]
    Parse --> CreateTaker["创建 Taker Fill<br/>liquidity = T"]
    Parse --> CreateMaker["创建 Maker Fill<br/>liquidity = M"]

    CreateTaker --> Buffer["批量缓冲区<br/>容量: 1000"]
    CreateMaker --> Buffer

    Buffer -->|"达到 1000 条<br/>或超时"| BatchInsert["批量 INSERT IGNORE<br/>into g_fill"]

    BatchInsert --> MySQL["MySQL g_fill 表"]

    style Buffer fill:#fff3e0
```

### 关键机制

| 特性 | 说明 |
|------|------|
| 每条 MatchLog 产生 2 条 Fill | Taker (liquidity=T) 和 Maker (liquidity=M) |
| 批量写入 | 缓冲 1000 条后批量 INSERT |
| 幂等去重 | `INSERT IGNORE` + `UNIQUE(order_id, message_seq)` |
| 记录字段 | orderId, tradeId, messageSeq, size, price, funds, liquidity, fee |

### Fill 字段计算

| 字段 | Taker Fill | Maker Fill |
|------|-----------|-----------|
| `order_id` | takerOrderId | makerOrderId |
| `trade_id` | tradeId | tradeId |
| `message_seq` | MatchLog.sequence | MatchLog.sequence |
| `size` | 成交数量 | 成交数量 |
| `price` | Maker 价格 | Maker 价格 |
| `funds` | size * price | size * price |
| `liquidity` | T | M |

## 4. TradeMaker - 交易记录生成器

### 职责

消费 Kafka 撮合日志中的 MatchLog，为每次成交创建一条 Trade 记录。

### 关键机制

| 特性 | 说明 |
|------|------|
| 每条 MatchLog 产生 1 条 Trade | 记录完整的成交信息 |
| 批量写入 | 批量 INSERT IGNORE |
| 幂等去重 | `INSERT IGNORE` 基于主键或唯一索引 |
| 记录字段 | productId, takerOrderId, makerOrderId, price, size, side |

### Trade 与 Fill 的关系

```
1 个 MatchLog → 1 个 Trade + 2 个 Fill (Taker + Maker)
```

## 5. TickMaker - K 线生成器

### 职责

消费 Kafka 撮合日志中的 MatchLog，维护内存中的 K 线桶，计算 OHLCV 数据。

### 支持的粒度

| 粒度 (分钟) | 说明 |
|-------------|------|
| 1 | 1 分钟 K |
| 3 | 3 分钟 K |
| 5 | 5 分钟 K |
| 15 | 15 分钟 K |
| 30 | 30 分钟 K |
| 60 | 1 小时 K |
| 120 | 2 小时 K |
| 240 | 4 小时 K |
| 360 | 6 小时 K |
| 720 | 12 小时 K |
| 1440 | 日 K |

共 **11 种** 粒度同时维护。

### 工作流程

```mermaid
graph TB
    MatchLog["MatchLog<br/>(price, size, time)"] --> Dispatch["分发到 11 个粒度"]

    Dispatch --> G1["1 分钟桶"]
    Dispatch --> G3["3 分钟桶"]
    Dispatch --> G5["5 分钟桶"]
    Dispatch --> GN["... 其他粒度"]
    Dispatch --> G1440["日线桶"]

    subgraph 内存桶 Bucket
        G1 --> Calc["OHLCV 计算"]
        G3 --> Calc
        G5 --> Calc
        GN --> Calc
        G1440 --> Calc
    end

    Calc --> |"REPLACE INTO<br/>g_tick"| MySQL["MySQL g_tick 表"]

    style Dispatch fill:#e1bee7
    style Calc fill:#fff3e0
```

### OHLCV 计算规则

| 字段 | 计算方式 |
|------|----------|
| **Open** | 时间窗口内第一笔成交价 |
| **High** | 时间窗口内最高成交价 |
| **Low** | 时间窗口内最低成交价 |
| **Close** | 时间窗口内最后一笔成交价 |
| **Volume** | 时间窗口内成交量之和 |

### 数据写入

- 使用 `REPLACE INTO` 语句
- 以 `(product_id, granularity, time)` 为唯一键
- 每次成交都可能更新多个粒度的 K 线数据

## 6. FillExecutor - 成交结算执行器

### 职责

消费 Redis `g_fill` 列表中的成交记录，执行订单状态更新和结算账单创建。

### 架构

```mermaid
graph TB
    subgraph "FillExecutor x10"
        direction TB
        Redis["Redis g_fill (List)"]
        Redis -->|BRPOP| Shard0["Worker #0<br/>orderId%10==0"]
        Redis -->|BRPOP| Shard1["Worker #1<br/>orderId%10==1"]
        Redis -->|BRPOP| Shard2["Worker #2<br/>orderId%10==2"]
        Redis -->|BRPOP| ShardN["..."]
        Redis -->|BRPOP| Shard9["Worker #9<br/>orderId%10==9"]
    end

    subgraph 处理逻辑
        LRU["LRU Cache<br/>容量: 1000"]
        UpdateOrder["更新 g_order<br/>filledSize, executedValue, status"]
        CreateBill["创建结算 g_bill<br/>释放冻结/增加可用"]
    end

    Shard0 & Shard1 & Shard2 & ShardN & Shard9 --> LRU
    LRU --> UpdateOrder
    UpdateOrder --> CreateBill

    subgraph 兜底机制
        Poll["DB 轮询<br/>每 1 秒"]
    end

    Poll -->|查询未结算 Fill| UpdateOrder

    style LRU fill:#fff3e0
    style Poll fill:#ffcdd2
```

### 关键机制

| 特性 | 说明 |
|------|------|
| 实例数量 | 10 个并行 Worker |
| 分片策略 | `orderId % 10` 路由到对应 Worker |
| 消费方式 | Redis `BRPOP` 阻塞弹出 |
| LRU 缓存 | 容量 1000，缓存已处理的订单状态 |
| 兜底机制 | 每 1 秒轮询数据库查找未结算 Fill |

### 处理流程

```mermaid
sequenceDiagram
    participant Redis as Redis g_fill
    participant Worker as FillExecutor
    participant LRU as LRU Cache
    participant MySQL as MySQL

    Redis->>Worker: BRPOP 获取 Fill 记录

    Worker->>LRU: 查询订单缓存
    alt 缓存命中
        LRU-->>Worker: 返回订单信息
    else 缓存未命中
        Worker->>MySQL: 查询 g_order
        MySQL-->>Worker: 返回订单
        Worker->>LRU: 更新缓存
    end

    Worker->>Worker: 计算新的 filledSize<br/>filledSize += fill.size

    Worker->>MySQL: 更新 g_order<br/>SET filledSize, executedValue

    alt filledSize == size
        Worker->>MySQL: 更新 status = filled
    end

    Worker->>Worker: 计算结算金额

    alt 买单成交
        Note over Worker: 释放 hold 的报价币<br/>增加 available 的基础币
    else 卖单成交
        Note over Worker: 释放 hold 的基础币<br/>增加 available 的报价币
    end

    Worker->>MySQL: INSERT g_bill<br/>(结算账单)

    Note over Worker: Bill 写入触发<br/>Binlog CDC → BillExecutor
```

### 分片的意义

- **相同订单路由到同一 Worker**: 避免并发更新同一订单的竞态条件
- **不同订单并行处理**: 10 个 Worker 并行执行，提升吞吐量
- **LRU 缓存局部性**: 同一订单的多次 Fill 可命中缓存

## 7. BillExecutor - 账单结算执行器

### 职责

消费 Redis `g_bill` 列表中的账单记录，将账单中的变动金额应用到用户账户。

### 架构

```mermaid
graph TB
    subgraph "BillExecutor x10"
        direction TB
        Redis["Redis g_bill (List)"]
        Redis -->|BRPOP| Shard0["Worker #0<br/>userId%10==0"]
        Redis -->|BRPOP| Shard1["Worker #1<br/>userId%10==1"]
        Redis -->|BRPOP| Shard2["Worker #2<br/>userId%10==2"]
        Redis -->|BRPOP| ShardN["..."]
        Redis -->|BRPOP| Shard9["Worker #9<br/>userId%10==9"]
    end

    subgraph 处理逻辑
        ReadBill["读取 Bill<br/>available 变动, hold 变动"]
        UpdateAccount["更新 g_account<br/>available += bill.available<br/>hold += bill.hold"]
        SettleBill["标记 bill.settled = true"]
    end

    Shard0 & Shard1 & Shard2 & ShardN & Shard9 --> ReadBill
    ReadBill --> UpdateAccount --> SettleBill

    style ReadBill fill:#fff3e0
```

### 关键机制

| 特性 | 说明 |
|------|------|
| 实例数量 | 10 个并行 Worker |
| 分片策略 | `userId % 10` 路由到对应 Worker |
| 消费方式 | Redis `BRPOP` 阻塞弹出 |
| 操作 | 将 bill 中的 available/hold 增量应用到 account |

### 分片的意义

- **相同用户路由到同一 Worker**: 避免并发更新同一用户账户的竞态条件
- **保证余额一致性**: 同一用户的所有账单按序处理
- **不同用户并行处理**: 10 个 Worker 并行执行

## 8. 完整结算流程

以一笔限价买单成交为例，展示从撮合到余额更新的完整结算链路：

```mermaid
sequenceDiagram
    participant Engine as 撮合引擎
    participant Kafka as Kafka
    participant FillMaker as FillMaker
    participant MySQL as MySQL
    participant Binlog as BinlogCDC
    participant Redis as Redis
    participant FillExec as FillExecutor
    participant BillExec as BillExecutor

    Note over Engine: BTC-USDT 买单成交<br/>price=50000, size=0.1

    Engine->>Kafka: MatchLog<br/>(taker, maker, price, size)

    Kafka->>FillMaker: 消费 MatchLog
    FillMaker->>MySQL: INSERT IGNORE g_fill x2<br/>(Taker Fill + Maker Fill)

    MySQL->>Binlog: g_fill 表变更
    Binlog->>Redis: LPUSH g_fill

    Redis->>FillExec: BRPOP g_fill (Taker Fill)
    FillExec->>MySQL: UPDATE g_order<br/>SET filledSize += 0.1
    FillExec->>MySQL: INSERT g_bill<br/>(userId=taker, currency=USDT,<br/>available=0, hold=-5000)
    FillExec->>MySQL: INSERT g_bill<br/>(userId=taker, currency=BTC,<br/>available=+0.1, hold=0)

    MySQL->>Binlog: g_bill 表变更
    Binlog->>Redis: LPUSH g_bill

    Redis->>BillExec: BRPOP g_bill (USDT)
    BillExec->>MySQL: UPDATE g_account<br/>SET hold -= 5000<br/>WHERE userId=taker, currency=USDT

    Redis->>BillExec: BRPOP g_bill (BTC)
    BillExec->>MySQL: UPDATE g_account<br/>SET available += 0.1<br/>WHERE userId=taker, currency=BTC

    Note over MySQL: Taker 的 USDT 冻结释放<br/>Taker 的 BTC 可用增加

    Note over FillExec,BillExec: Maker 侧同理:<br/>释放 BTC 冻结, 增加 USDT 可用
```

## 9. 两阶段结算机制

系统采用两阶段结算（Two-Phase Settlement）设计，将结算过程分为两个独立的阶段：

### 第一阶段: FillExecutor

| 操作 | 说明 |
|------|------|
| 输入 | g_fill 记录（Binlog CDC 触发） |
| 处理 | 更新订单状态（filledSize, executedValue, status） |
| 输出 | 创建结算 g_bill 记录 |

### 第二阶段: BillExecutor

| 操作 | 说明 |
|------|------|
| 输入 | g_bill 记录（Binlog CDC 触发） |
| 处理 | 将 bill 的 available/hold 增量应用到 g_account |
| 输出 | 更新用户账户余额，标记 bill 已结算 |

### 设计优势

```mermaid
graph LR
    subgraph "第一阶段 FillExecutor"
        Fill["g_fill"] --> Order["更新 g_order"]
        Order --> Bill["创建 g_bill"]
    end

    subgraph "Binlog CDC"
        CDC["MySQL Binlog"]
    end

    subgraph "第二阶段 BillExecutor"
        BillIn["g_bill"] --> Account["更新 g_account"]
    end

    Bill --> CDC --> BillIn

    style Fill fill:#bbdefb
    style Bill fill:#fff3e0
    style Account fill:#c8e6c9
```

1. **解耦**: 订单结算和余额结算分离，降低系统复杂度
2. **可靠性**: 每个阶段都有 Binlog CDC 驱动，即使 Worker 崩溃也能恢复
3. **可追溯**: 通过 g_bill 表记录每一笔余额变动，完整审计链
4. **并发安全**: 按 orderId/userId 分片，避免竞态条件
5. **兜底机制**: FillExecutor 每 1 秒轮询数据库，确保不遗漏未处理的 Fill

## 10. 关键源文件

| 文件路径 | 说明 |
|----------|------|
| `worker/fill_maker.go` | FillMaker: Kafka MatchLog → g_fill |
| `worker/trade_maker.go` | TradeMaker: Kafka MatchLog → g_trade |
| `worker/tick_maker.go` | TickMaker: Kafka MatchLog → g_tick (11种粒度) |
| `worker/fill_executor.go` | FillExecutor: g_fill → g_order 更新 + g_bill 创建 |
| `worker/bill_executor.go` | BillExecutor: g_bill → g_account 更新 |
| `models/binlog_stream.go` | MySQL Binlog CDC 到 Redis 的桥接 |

## 11. 相关文档

- [系统架构概览](architecture.md)
- [撮合引擎详解](matching-engine.md)
- [数据模型设计](data-model.md)
- [Kafka 消息系统](kafka-messaging.md)
