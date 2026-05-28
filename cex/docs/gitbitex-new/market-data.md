# gitbitex-new 行情数据管道详解

## 1. 概述

行情数据管道是连接撮合引擎与外部世界的桥梁。撮合引擎产生的消息通过 Kafka Message Topic 分发给 **7 个消费者线程**，每个线程负责特定的数据处理任务：持久化、K线生成、行情统计、盘口快照等。

## 2. 整体管道架构

```mermaid
graph TB
    subgraph "撮合引擎"
        ME[MatchingEngine] --> MS[MessageSender]
    end

    MS --> KAFKA[Kafka Message Topic<br/>Matching-Engine-Message]

    KAFKA --> T1[OrderPersistenceThread<br/>订单持久化]
    KAFKA --> T2[TradePersistenceThread<br/>成交持久化]
    KAFKA --> T3[AccountPersistenceThread<br/>账户持久化]
    KAFKA --> T4[CandleMakerThread<br/>K线生成]
    KAFKA --> T5[TickerThread<br/>行情统计]
    KAFKA --> T6[OrderBookSnapshotThread<br/>盘口快照]
    KAFKA --> T7[MatchingEngineSnapshotThread<br/>引擎快照]

    T1 -->|ORDER 消息| MONGO_O[(MongoDB<br/>orders)]
    T2 -->|TRADE 消息| MONGO_T[(MongoDB<br/>trades)]
    T3 -->|ACCOUNT 消息| MONGO_A[(MongoDB<br/>accounts)]
    T3 -->|Pub/Sub| REDIS_A[(Redis<br/>account 频道)]
    T4 -->|TRADE 消息| MONGO_C[(MongoDB<br/>candles)]
    T5 -->|TRADE 消息| MONGO_TK[(MongoDB<br/>tickers)]
    T6 -->|ORDER 消息| REDIS_L2[(Redis<br/>l2 快照 + Pub/Sub)]
    T7 -->|全部消息| MONGO_S[(MongoDB<br/>snapshot_*)]

    style KAFKA fill:#e1f5fe
    style T1 fill:#c8e6c9
    style T2 fill:#c8e6c9
    style T3 fill:#c8e6c9
    style T4 fill:#fff3e0
    style T5 fill:#fff3e0
    style T6 fill:#fff3e0
    style T7 fill:#ffcdd2
```

## 3. OrderPersistenceThread — 订单持久化

### 职责

消费 Kafka Message Topic 中的 **ORDER** 类型消息，将订单数据持久化到 MongoDB `orders` 集合，并通过 Redis Pub/Sub 通知 WebSocket 层。

### 处理流程

```mermaid
flowchart TD
    A[从 Kafka 消费消息] --> B{消息类型?}
    B -->|ORDER| C[解析 OrderMessage]
    B -->|其他| A
    C --> D[通过 OrderManager<br/>写入 MongoDB orders 集合]
    D --> E{订单已存在?}
    E -->|是 - UPSERT| F[更新订单状态<br/>filledSize, remainingSize,<br/>status 等字段]
    E -->|否 - INSERT| G[插入新订单记录]
    F & G --> H[Redis publish<br/>order 频道]
    H --> A

    style C fill:#e1f5fe
    style H fill:#e8f5e9
```

### 关键点

- 使用 MongoDB **upsert** 操作：如果订单已存在则更新，否则插入
- 一个订单在其生命周期中会收到多次 ORDER 消息（OPEN → 部分成交 → FILLED）
- 每次状态变更都会覆盖更新 MongoDB 中的记录
- 同时发布到 Redis `order` 频道，供 WebSocket `FeedMessageListener` 消费

## 4. TradePersistenceThread — 成交持久化

### 职责

消费 **TRADE** 类型消息，将成交记录持久化到 MongoDB `trades` 集合，并通知 WebSocket。

### 处理流程

```mermaid
flowchart TD
    A[从 Kafka 消费消息] --> B{消息类型?}
    B -->|TRADE| C[解析 TradeMessage]
    B -->|其他| A
    C --> D[通过 TradeManager<br/>写入 MongoDB trades 集合]
    D --> E[Redis publish<br/>trade 频道]
    E --> A

    style C fill:#e1f5fe
    style E fill:#e8f5e9
```

### 成交记录字段

| 字段 | 说明 |
|------|------|
| `tradeId` | 成交 ID（引擎分配） |
| `productId` | 交易对 |
| `takerOrderId` | Taker 订单 ID |
| `makerOrderId` | Maker 订单 ID |
| `price` | 成交价格（以 Maker 价格成交） |
| `size` | 成交数量 |
| `side` | Taker 方向 |
| `time` | 成交时间 |

## 5. AccountPersistenceThread — 账户持久化

### 职责

消费 **ACCOUNT** 类型消息，持久化到 MongoDB 并通过 Redis Pub/Sub 推送实时余额变更。

### 处理流程

```mermaid
flowchart TD
    A[从 Kafka 消费消息] --> B{消息类型?}
    B -->|ACCOUNT| C[解析 AccountMessage]
    B -->|其他| A
    C --> D[通过 AccountManager<br/>写入 MongoDB]
    D --> E[Redis publish<br/>account 频道]
    E --> F[WebSocket FeedMessageListener<br/>推送给订阅了<br/>userId.currency.funds 的客户端]
    F --> A

    style C fill:#e1f5fe
    style E fill:#e8f5e9
    style F fill:#fff3e0
```

### 双写策略

AccountPersistenceThread 同时写入两个存储：
1. **MongoDB**：持久化存储，供 REST API 查询
2. **Redis Pub/Sub**：实时通知，供 WebSocket 推送

## 6. CandleMakerThread — K线生成

### 职责

消费 **TRADE** 类型消息，聚合为 OHLCV K线数据，支持 7 种时间粒度。

### 7 种时间粒度

| 粒度（秒） | 含义 | 说明 |
|-----------|------|------|
| 60 | 1 分钟 | 最小粒度 |
| 300 | 5 分钟 | |
| 900 | 15 分钟 | |
| 1800 | 30 分钟 | |
| 3600 | 1 小时 | |
| 21600 | 6 小时 | |
| 86400 | 1 天 | 最大粒度 |

### K线生成流程

```mermaid
flowchart TD
    A[从 Kafka 消费 TRADE 消息] --> B[解析 TradeMessage<br/>获取 price, size, time]

    B --> C["对每种粒度 (7种) 执行:"]

    C --> D["计算 K线起始时间<br/>candleTime = tradeTime -<br/>(tradeTime % granularity)"]

    D --> E{MongoDB 中已有<br/>该时间段的 K线?}

    E -->|是 - 更新| F["更新已有 K线<br/>high = max(high, price)<br/>low = min(low, price)<br/>close = price<br/>volume += size"]

    E -->|否 - 创建| G["创建新 K线<br/>open = price<br/>high = price<br/>low = price<br/>close = price<br/>volume = size"]

    F & G --> H[写入 MongoDB candles 集合]
    H --> I{还有更多粒度?}
    I -->|是| C
    I -->|否| A

    style D fill:#fff3e0
    style F fill:#e8f5e9
    style G fill:#e1f5fe
```

### 时间对齐公式

```
candleTime = tradeTime - (tradeTime % granularity)
```

**示例**：
- 成交时间：`10:37:42`（Unix timestamp: 1705312662）
- 1 分钟 K线时间：`1705312662 - (1705312662 % 60) = 1705312620` → `10:37:00`
- 5 分钟 K线时间：`1705312662 - (1705312662 % 300) = 1705312500` → `10:35:00`
- 1 小时 K线时间：`1705312662 - (1705312662 % 3600) = 1705312800` → `10:00:00`

### K线更新逻辑

```mermaid
graph LR
    subgraph "K线更新规则"
        O["open<br/>不变<br/>(第一笔成交价)"]
        H["high<br/>max(当前high, 新price)"]
        L["low<br/>min(当前low, 新price)"]
        C["close<br/>= 新price<br/>(最新成交价)"]
        V["volume<br/>+= 新size"]
    end

    style O fill:#e1f5fe
    style H fill:#ffcdd2
    style L fill:#c8e6c9
    style C fill:#fff3e0
    style V fill:#f3e5f5
```

## 7. TickerThread — 行情统计

### 职责

消费 **TRADE** 类型消息，维护每个交易对的滚动窗口行情统计。

### 统计维度

| 指标 | 窗口 | 说明 |
|------|------|------|
| `price` | 实时 | 最新成交价 |
| `open24h` | 24 小时 | 24 小时前的开盘价 |
| `high24h` | 24 小时 | 24 小时内最高价 |
| `low24h` | 24 小时 | 24 小时内最低价 |
| `volume24h` | 24 小时 | 24 小时成交量 |
| `volume30d` | 30 天 | 30 天累计成交量 |

### 滚动窗口机制

```mermaid
flowchart TD
    A[从 Kafka 消费 TRADE 消息] --> B[获取当前时间 now]
    B --> C{是否跨越 24h 边界?<br/>now - last24hReset > 24h}

    C -->|是| D["重置 24h 统计<br/>open24h = 当前 price<br/>high24h = 当前 price<br/>low24h = 当前 price<br/>volume24h = 0"]
    C -->|否| E["更新 24h 统计<br/>high24h = max(high24h, price)<br/>low24h = min(low24h, price)<br/>volume24h += size"]

    D --> F{是否跨越 30d 边界?}
    E --> F

    F -->|是| G["重置 30d 统计<br/>volume30d = 0"]
    F -->|否| H["更新 30d 统计<br/>volume30d += size"]

    G & H --> I["price = 最新成交价"]
    I --> J[写入 MongoDB tickers 集合]

    style D fill:#ffcdd2
    style E fill:#c8e6c9
```

### 边界跨越判断

- **24h 边界**：当前时间 - 上次重置时间 > 86400 秒时触发重置
- **30d 边界**：当前时间 - 上次重置时间 > 2592000 秒时触发重置
- 重置后以当前成交价作为新周期的开盘价

## 8. OrderBookSnapshotThread — 盘口快照

### 职责

消费 **ORDER** 类型消息，在内存中重建 L2 订单簿，定期生成快照存储到 Redis。

### 快照触发条件

| 条件 | 说明 |
|------|------|
| 每 1000 个 sequence | 累计 1000 次订单簿变更后触发 |
| 每 1 秒 | 距上次快照超过 1 秒触发 |

两个条件满足任一即触发快照生成。

### 快照流程

```mermaid
flowchart TD
    A[从 Kafka 消费 ORDER 消息] --> B[在内存中更新 L2OrderBook<br/>根据订单变更调整价格层级]

    B --> C{触发快照?<br/>sequence差值 >= 1000<br/>OR 时间差 >= 1秒}

    C -->|否| A
    C -->|是| D[生成 L2 快照<br/>买卖各 25 档]

    D --> E[计算增量更新<br/>L2OrderBook.diff<br/>新快照 vs 旧快照]

    E --> F["存储到 Redis<br/>{productId}.l2_batch_order_book"]
    F --> G["Redis publish l2_batch<br/>通知 WebSocket 层"]
    G --> H[更新旧快照 = 当前快照]
    H --> A

    style D fill:#e1f5fe
    style E fill:#fff3e0
    style G fill:#e8f5e9
```

### L2 快照格式

```
{productId}.l2_batch_order_book = {
    "asks": [
        ["50000.00", "1.5", 3],    // [价格, 总量, 订单数]
        ["50001.00", "0.8", 1],
        ... (最多 25 档)
    ],
    "bids": [
        ["49999.00", "2.0", 5],
        ["49998.00", "1.2", 2],
        ... (最多 25 档)
    ]
}
```

### L3 快照

系统同时维护 L3（逐笔）快照，存储在 Redis `{productId}.l3_order_book` 中，包含每个价格层级下的所有订单详情。

## 9. MatchingEngineSnapshotThread — 引擎快照

### 职责

定期将撮合引擎的完整状态保存到 MongoDB，用于崩溃恢复。这是最关键的持久化线程。

### 快照内容

```mermaid
graph TB
    subgraph "MatchingEngine 状态快照"
        SE["snapshot_engine<br/>commandOffset<br/>messageOffset<br/>messageSequence<br/>orderSequences<br/>tradeSequences<br/>orderBookSequences"]
        SA["snapshot_account<br/>所有用户的所有币种<br/>available + hold"]
        SO["snapshot_order<br/>所有活跃订单<br/>(OPEN 状态)"]
        SP["snapshot_product<br/>所有交易对配置"]
    end

    SE & SA & SO & SP --> TXN[MongoDB 事务<br/>Snapshot 隔离级别<br/>原子写入 4 个集合]

    style TXN fill:#ffcdd2
```

### 快照写入流程

```mermaid
sequenceDiagram
    participant EST as EngineSnapshotThread
    participant ME as MatchingEngine
    participant MONGO as MongoDB

    loop 定期执行
        EST->>ME: 获取当前状态快照<br/>(线程安全读取)
        ME-->>EST: 引擎状态 + 所有 Account<br/>+ 所有活跃 Order + 所有 Product

        EST->>MONGO: startTransaction<br/>(Snapshot 隔离级别)

        EST->>MONGO: DROP snapshot_engine 旧数据
        EST->>MONGO: INSERT snapshot_engine 新数据

        EST->>MONGO: DROP snapshot_account 旧数据
        EST->>MONGO: BATCH INSERT 所有 Account

        EST->>MONGO: DROP snapshot_order 旧数据
        EST->>MONGO: BATCH INSERT 所有活跃 Order

        EST->>MONGO: DROP snapshot_product 旧数据
        EST->>MONGO: BATCH INSERT 所有 Product

        EST->>MONGO: commitTransaction
    end
```

## 10. 线程模型

### KafkaConsumerThread 基类

所有 7 个消费者线程都继承自 `KafkaConsumerThread`，共享以下行为：

```mermaid
flowchart TD
    A[线程启动] --> B[创建独立的 Kafka Consumer<br/>使用独立的 Consumer Group]
    B --> C[订阅 Message Topic]
    C --> D[主循环: poll 消息]
    D --> E[逐条处理消息<br/>调用子类的处理方法]
    E --> F[手动 commit offset]
    F --> D

    D -->|异常| G[UncaughtExceptionHandler]
    G --> H[等待 3 秒]
    H --> I[重新启动线程]
    I --> B

    style G fill:#ffcdd2
    style I fill:#c8e6c9
```

### 每个线程独立的 Kafka Consumer

```mermaid
graph TB
    TOPIC[Kafka Message Topic<br/>Matching-Engine-Message]

    subgraph "7 个独立 Consumer (独立 Consumer Group)"
        C1["Consumer Group: order-persistence<br/>OrderPersistenceThread"]
        C2["Consumer Group: trade-persistence<br/>TradePersistenceThread"]
        C3["Consumer Group: account-persistence<br/>AccountPersistenceThread"]
        C4["Consumer Group: candle-maker<br/>CandleMakerThread"]
        C5["Consumer Group: ticker<br/>TickerThread"]
        C6["Consumer Group: orderbook-snapshot<br/>OrderBookSnapshotThread"]
        C7["Consumer Group: engine-snapshot<br/>EngineSnapshotThread"]
    end

    TOPIC --> C1 & C2 & C3 & C4 & C5 & C6 & C7

    Note1["每个 Consumer Group 独立消费<br/>Message Topic 的全部消息"]

    style TOPIC fill:#e1f5fe
```

### UncaughtExceptionHandler — 线程自动恢复

当任何消费者线程因未捕获异常而终止时：

1. `UncaughtExceptionHandler` 捕获异常
2. 记录错误日志
3. 等待 **3 秒**（避免快速重启循环）
4. 创建新线程实例并启动

```mermaid
flowchart LR
    A[线程崩溃] --> B[UncaughtExceptionHandler<br/>捕获异常]
    B --> C[记录错误日志]
    C --> D[sleep 3 秒]
    D --> E[创建新线程实例]
    E --> F[启动新线程<br/>从上次 committed offset 恢复]

    style A fill:#ffcdd2
    style F fill:#c8e6c9
```

**恢复机制**：
- 新线程从最后一次 committed offset 开始消费
- 可能会重复处理一些消息（at-least-once 语义）
- 持久化操作需要是幂等的（upsert）以容忍重复处理

## 11. 消息处理总结

| 线程 | 消费消息类型 | 输出目标 | 处理频率 |
|------|------------|---------|---------|
| OrderPersistenceThread | ORDER | MongoDB orders + Redis order | 每条消息 |
| TradePersistenceThread | TRADE | MongoDB trades + Redis trade | 每条消息 |
| AccountPersistenceThread | ACCOUNT | MongoDB + Redis account | 每条消息 |
| CandleMakerThread | TRADE | MongoDB candles | 每条消息 x 7 粒度 |
| TickerThread | TRADE | MongoDB tickers | 每条消息 |
| OrderBookSnapshotThread | ORDER | Redis L2/L3 | 每 1000 seq 或 1 秒 |
| EngineSnapshotThread | 全部 | MongoDB snapshot_* | 定期（事务） |

## 12. 关键源文件

| 文件 | 路径 | 说明 |
|------|------|------|
| KafkaConsumerThread | `kafka/KafkaConsumerThread.java` | 消费者线程基类 |
| OrderPersistenceThread | `marketdata/OrderPersistenceThread.java` | 订单持久化 |
| TradePersistenceThread | `marketdata/TradePersistenceThread.java` | 成交持久化 |
| AccountPersistenceThread | `marketdata/AccountPersistenceThread.java` | 账户持久化 |
| CandleMakerThread | `marketdata/CandleMakerThread.java` | K线生成 |
| TickerThread | `marketdata/TickerThread.java` | 行情统计 |
| OrderBookSnapshotThread | `marketdata/OrderBookSnapshotThread.java` | 盘口快照 |
| MatchingEngineSnapshotThread | `matchingengine/MatchingEngineSnapshotThread.java` | 引擎快照 |
| OrderManager | `manager/OrderManager.java` | 订单 MongoDB 操作 |
| TradeManager | `manager/TradeManager.java` | 成交 MongoDB 操作 |
| AccountManager | `manager/AccountManager.java` | 账户 MongoDB/Redis 操作 |
| L2OrderBook | `matchingengine/L2OrderBook.java` | L2 行情维护与 diff |
| Bootstrap | `Bootstrap.java` | 所有线程的启动入口 |
