# gitbitex-new WebSocket 实时推送系统详解

## 1. 概述

gitbitex-new 使用 Spring WebSocket 实现实时行情和订单推送。系统通过 Redis Pub/Sub 接收撮合引擎的处理结果，再通过 WebSocket 推送给订阅了相应频道的客户端。

WebSocket 服务端点：`ws://host:8080/ws`

## 2. 整体架构

```mermaid
graph TB
    subgraph "撮合引擎消费者线程"
        T3[AccountPersistenceThread]
        T1[OrderPersistenceThread]
        T2[TradePersistenceThread]
        T6[OrderBookSnapshotThread]
    end

    subgraph "Redis"
        PUB_L2[l2_batch 频道]
        PUB_ORDER[order 频道]
        PUB_TRADE[trade 频道]
        PUB_ACCOUNT[account 频道]
    end

    subgraph "WebSocket 服务"
        FML[FeedMessageListener<br/>Redis Pub/Sub 订阅]
        SM[SessionManager<br/>会话管理]
        SE[StripedExecutorService<br/>条纹化执行器]
        HANDLER[FeedTextWebSocketHandler<br/>WebSocket 端点 /ws]
    end

    subgraph "客户端"
        C1[客户端 1]
        C2[客户端 2]
        C3[客户端 N]
    end

    T6 -->|publish| PUB_L2
    T1 -->|publish| PUB_ORDER
    T2 -->|publish| PUB_TRADE
    T3 -->|publish| PUB_ACCOUNT

    PUB_L2 & PUB_ORDER & PUB_TRADE & PUB_ACCOUNT --> FML
    FML --> SM
    SM --> SE
    SE --> HANDLER
    HANDLER --> C1 & C2 & C3

    style FML fill:#e1f5fe
    style SM fill:#fff3e0
    style SE fill:#e8f5e9
```

## 3. 频道列表

客户端通过订阅特定频道来接收相应数据。

| 频道格式 | 示例 | 认证 | 说明 |
|---------|------|------|------|
| `{productId}.level2` | `BTC-USDT.level2` | 否 | L2 盘口深度（25 档） |
| `{productId}.ticker` | `BTC-USDT.ticker` | 否 | 行情统计（24h/30d） |
| `{productId}.match` | `BTC-USDT.match` | 否 | 最新成交（逐笔） |
| `{userId}.{productId}.order` | `user1.BTC-USDT.order` | 是 | 用户订单状态变更 |
| `{userId}.{currency}.funds` | `user1.USDT.funds` | 是 | 用户资金余额变更 |

### 频道分类

```mermaid
graph TB
    subgraph "公开频道 (无需认证)"
        L2["{productId}.level2<br/>盘口深度"]
        TICKER["{productId}.ticker<br/>行情统计"]
        MATCH["{productId}.match<br/>逐笔成交"]
    end

    subgraph "私有频道 (需要认证)"
        ORDER["{userId}.{productId}.order<br/>订单状态"]
        FUNDS["{userId}.{currency}.funds<br/>资金余额"]
    end

    style L2 fill:#e1f5fe
    style TICKER fill:#e1f5fe
    style MATCH fill:#e1f5fe
    style ORDER fill:#fff3e0
    style FUNDS fill:#fff3e0
```

## 4. 消息类型

### 4.1 L2SnapshotFeedMessage — 盘口完整快照

客户端首次订阅 `level2` 频道时收到。

| 字段 | 类型 | 说明 |
|------|------|------|
| `type` | String | `l2_snapshot` |
| `productId` | String | 交易对 |
| `asks` | List\<List\> | 卖盘，每行 [price, size, count]，最多 25 档 |
| `bids` | List\<List\> | 买盘，每行 [price, size, count]，最多 25 档 |

### 4.2 L2UpdateFeedMessage — 盘口增量更新

后续的盘口变更以增量方式推送。

| 字段 | 类型 | 说明 |
|------|------|------|
| `type` | String | `l2_update` |
| `productId` | String | 交易对 |
| `changes` | List\<List\> | 变更列表，每行 [side, price, size]。size=0 表示删除该价位 |

### 4.3 TickerFeedMessage — 行情统计

| 字段 | 类型 | 说明 |
|------|------|------|
| `type` | String | `ticker` |
| `productId` | String | 交易对 |
| `price` | String | 最新价 |
| `open24h` | String | 24h 开盘价 |
| `high24h` | String | 24h 最高价 |
| `low24h` | String | 24h 最低价 |
| `volume24h` | String | 24h 成交量 |
| `volume30d` | String | 30d 成交量 |

### 4.4 OrderFeedMessage — 订单状态变更

| 字段 | 类型 | 说明 |
|------|------|------|
| `type` | String | `order` |
| `orderId` | String | 订单 ID |
| `productId` | String | 交易对 |
| `side` | String | BUY / SELL |
| `price` | String | 价格 |
| `size` | String | 数量 |
| `remainingSize` | String | 剩余量 |
| `filledSize` | String | 已成交量 |
| `status` | String | 订单状态 |

### 4.5 OrderMatchFeedMessage — 逐笔成交

| 字段 | 类型 | 说明 |
|------|------|------|
| `type` | String | `match` |
| `tradeId` | String | 成交 ID |
| `productId` | String | 交易对 |
| `price` | String | 成交价格 |
| `size` | String | 成交数量 |
| `side` | String | Taker 方向 |
| `time` | String | 成交时间 |

### 4.6 PongFeedMessage — 心跳响应

| 字段 | 类型 | 说明 |
|------|------|------|
| `type` | String | `pong` |

## 5. FeedMessageListener — Redis 订阅监听

`FeedMessageListener` 订阅 Redis Pub/Sub 的 4 个频道，将收到的消息转发给 WebSocket 客户端。

```mermaid
flowchart TD
    subgraph "Redis Pub/Sub 频道"
        R1[l2_batch]
        R2[order]
        R3[trade]
        R4[account]
    end

    R1 & R2 & R3 & R4 --> FML[FeedMessageListener]

    FML --> D{消息类型路由}

    D -->|l2_batch| L2[构建 L2UpdateFeedMessage<br/>或 L2SnapshotFeedMessage]
    D -->|order| ORD[构建 OrderFeedMessage]
    D -->|trade| TRD[构建 OrderMatchFeedMessage]
    D -->|account| ACC[构建 FundsFeedMessage]

    L2 --> SM1["SessionManager<br/>查找订阅 {productId}.level2<br/>的所有会话"]
    ORD --> SM2["SessionManager<br/>查找订阅 {userId}.{productId}.order<br/>的所有会话"]
    TRD --> SM3["SessionManager<br/>查找订阅 {productId}.match<br/>的所有会话"]
    ACC --> SM4["SessionManager<br/>查找订阅 {userId}.{currency}.funds<br/>的所有会话"]

    SM1 & SM2 & SM3 & SM4 --> SE[StripedExecutorService<br/>按 session 条纹化发送]

    style FML fill:#e1f5fe
    style SE fill:#e8f5e9
```

## 6. SessionManager — 会话管理

`SessionManager` 维护客户端会话与订阅频道之间的映射关系。

### 核心数据结构

```mermaid
graph TB
    subgraph "SessionManager"
        CS["channelSessions<br/>Map&lt;String(channel),<br/>ConcurrentSkipListSet&lt;Session&gt;&gt;<br/>频道 → 订阅该频道的会话集合"]
        SC["sessionChannels<br/>Map&lt;String(sessionId),<br/>Set&lt;String(channel)&gt;&gt;<br/>会话 → 该会话订阅的频道集合"]
    end

    subgraph "示例"
        CH1["BTC-USDT.level2"] --> S1["Session A"]
        CH1 --> S2["Session B"]
        CH1 --> S3["Session C"]

        CH2["BTC-USDT.match"] --> S1
        CH2 --> S4["Session D"]

        CH3["user1.BTC-USDT.order"] --> S1
    end

    style CS fill:#e1f5fe
    style SC fill:#fff3e0
```

### ConcurrentSkipListSet

选择 `ConcurrentSkipListSet` 而非 `ConcurrentHashSet` 的原因：
- **线程安全**：支持高并发的读写操作
- **有序性**：可按会话 ID 排序，便于调试和管理
- **高效遍历**：推送消息时需要遍历所有订阅会话，SkipList 的遍历性能优秀

## 7. StripedExecutorService — 条纹化执行器

`StripedExecutorService` 是系统最精巧的并发组件之一，保证同一客户端的消息严格有序，同时不同客户端之间可以并行推送。

### 设计原理

```mermaid
graph TB
    subgraph "StripedExecutorService"
        direction TB
        SUBMIT["submit(task, stripe)"]
        SUBMIT --> HASH["hash(stripe) % executorCount"]

        subgraph "SerialExecutor 0"
            Q0["LinkedBlockingQueue"]
            W0["Worker Thread 0"]
            Q0 --> W0
        end

        subgraph "SerialExecutor 1"
            Q1["LinkedBlockingQueue"]
            W1["Worker Thread 1"]
            Q1 --> W1
        end

        subgraph "SerialExecutor 2"
            Q2["LinkedBlockingQueue"]
            W2["Worker Thread 2"]
            Q2 --> W2
        end

        HASH -->|stripe=SessionA → 0| Q0
        HASH -->|stripe=SessionB → 1| Q1
        HASH -->|stripe=SessionC → 2| Q2
    end

    style SUBMIT fill:#e1f5fe
    style HASH fill:#fff3e0
```

### 保序机制

```mermaid
sequenceDiagram
    participant FML as FeedMessageListener
    participant SE as StripedExecutorService
    participant Q as SerialExecutor 队列<br/>(Session A 的条纹)
    participant WS as WebSocket Session A

    Note over FML,WS: 同一 Session 的消息严格串行

    FML->>SE: submit(msg1, sessionA)
    SE->>Q: 入队 msg1
    Q->>WS: 发送 msg1

    FML->>SE: submit(msg2, sessionA)
    SE->>Q: 入队 msg2
    Note over Q: msg1 发送完毕后才发送 msg2
    Q->>WS: 发送 msg2

    FML->>SE: submit(msg3, sessionA)
    SE->>Q: 入队 msg3
    Q->>WS: 发送 msg3

    Note over FML,WS: 保证顺序: msg1 → msg2 → msg3
```

### 并行与串行的平衡

| 维度 | 行为 |
|------|------|
| 同一 Session（同一条纹） | **严格串行** — 消息按入队顺序依次发送 |
| 不同 Session（不同条纹） | **完全并行** — 互不影响，充分利用多核 |

### SerialExecutor 内部结构

每个条纹对应一个 `SerialExecutor`：

```mermaid
graph LR
    subgraph "SerialExecutor"
        Q[LinkedBlockingQueue<br/>无界队列] --> ACTIVE{有正在执行的任务?}
        ACTIVE -->|否| EXEC[从队列取出并执行<br/>标记为 active]
        ACTIVE -->|是| WAIT[等待当前任务完成<br/>自动执行下一个]
        EXEC --> DONE[任务完成]
        DONE --> NEXT{队列还有任务?}
        NEXT -->|是| EXEC
        NEXT -->|否| IDLE[空闲等待]
    end
```

**关键实现细节**：
- `LinkedBlockingQueue` 是无界的，不会因队列满而阻塞生产者
- 每个 SerialExecutor 同一时刻只有一个任务在执行
- 任务完成后自动调度下一个任务（递归触发）

## 8. L2 增量更新机制

### L2OrderBook.diff() 算法

```mermaid
flowchart TD
    A[新快照 T2] --> C[L2OrderBook.diff]
    B[旧快照 T1] --> C
    C --> D{遍历所有价格层级}

    D --> E{该价格在 T1 中存在?}
    E -->|不存在| F["新增: [side, price, newSize]"]
    E -->|存在| G{size 是否变化?}
    G -->|变化| H["更新: [side, price, newSize]"]
    G -->|未变化| I[跳过]

    D --> J{T1 中有 T2 没有的价格?}
    J -->|有| K["删除: [side, price, 0]<br/>size=0 表示移除"]

    F & H & K --> L[汇总为 L2UpdateFeedMessage]
    L --> M[推送给订阅 level2 的客户端]

    style L fill:#e1f5fe
```

### 完整的 L2 推送流程

```mermaid
sequenceDiagram
    participant OBS as OrderBookSnapshotThread
    participant REDIS as Redis
    participant FML as FeedMessageListener
    participant L2OB as L2OrderBook
    participant SM as SessionManager
    participant WS as WebSocket 客户端

    Note over OBS,WS: 首次连接
    WS->>SM: subscribe("BTC-USDT.level2")
    SM->>REDIS: 读取当前 L2 快照
    REDIS-->>SM: 完整快照数据
    SM->>WS: L2SnapshotFeedMessage<br/>(完整 25 档盘口)

    Note over OBS,WS: 后续增量更新
    OBS->>REDIS: 存储新快照 + publish l2_batch
    REDIS->>FML: 收到 l2_batch 消息
    FML->>L2OB: diff(旧快照, 新快照)
    L2OB-->>FML: 差异数据 changes
    FML->>SM: 查找订阅 BTC-USDT.level2 的会话
    SM-->>FML: [Session1, Session2, ...]
    FML->>WS: L2UpdateFeedMessage<br/>(只有变化的价位)
```

## 9. 订阅与取消订阅

### 订阅流程

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant WH as FeedTextWebSocketHandler
    participant SM as SessionManager

    Client->>WH: 发送订阅请求<br/>{"type":"subscribe",<br/>"channels":["BTC-USDT.level2",<br/>"BTC-USDT.match"]}

    WH->>WH: 解析频道列表
    WH->>SM: subscribe(session, "BTC-USDT.level2")
    SM->>SM: channelSessions["BTC-USDT.level2"].add(session)
    SM->>SM: sessionChannels[sessionId].add("BTC-USDT.level2")

    WH->>SM: subscribe(session, "BTC-USDT.match")
    SM->>SM: channelSessions["BTC-USDT.match"].add(session)
    SM->>SM: sessionChannels[sessionId].add("BTC-USDT.match")

    Note over WH,SM: 如果是 level2 频道<br/>立即发送完整快照
    SM-->>Client: L2SnapshotFeedMessage

    Note over Client,SM: 此后该 session 会收到<br/>这两个频道的所有更新
```

### 取消订阅 / 断开连接

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant WH as FeedTextWebSocketHandler
    participant SM as SessionManager

    alt 主动取消订阅
        Client->>WH: {"type":"unsubscribe",<br/>"channels":["BTC-USDT.level2"]}
        WH->>SM: unsubscribe(session, "BTC-USDT.level2")
        SM->>SM: channelSessions["BTC-USDT.level2"].remove(session)
        SM->>SM: sessionChannels[sessionId].remove("BTC-USDT.level2")
    else 连接断开
        Client--xWH: 连接关闭
        WH->>SM: onClose(session)
        SM->>SM: 获取 sessionChannels[sessionId] 的所有频道
        SM->>SM: 从每个频道的 channelSessions 中移除该 session
        SM->>SM: 删除 sessionChannels[sessionId]
    end
```

## 10. Ping/Pong 心跳机制

WebSocket 使用 Ping/Pong 机制保持连接活跃：

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant WH as FeedTextWebSocketHandler

    loop 心跳周期
        Client->>WH: {"type":"ping"}
        WH-->>Client: {"type":"pong"}<br/>PongFeedMessage
    end

    Note over Client,WH: 如果客户端长时间<br/>未发送 ping<br/>服务端可关闭连接
```

## 11. 线程模型总结

```mermaid
graph TB
    subgraph "Redis Pub/Sub 监听线程"
        FML[FeedMessageListener<br/>单线程接收 Redis 消息]
    end

    subgraph "StripedExecutorService 线程池"
        SE0[SerialExecutor 0<br/>处理 Session A, D, G...]
        SE1[SerialExecutor 1<br/>处理 Session B, E, H...]
        SE2[SerialExecutor 2<br/>处理 Session C, F, I...]
    end

    subgraph "WebSocket I/O"
        WS1[Session A]
        WS2[Session B]
        WS3[Session C]
    end

    FML -->|submit task| SE0 & SE1 & SE2
    SE0 --> WS1
    SE1 --> WS2
    SE2 --> WS3

    style FML fill:#e1f5fe
    style SE0 fill:#e8f5e9
    style SE1 fill:#e8f5e9
    style SE2 fill:#e8f5e9
```

**线程分工**：
1. **FeedMessageListener** 线程：从 Redis 接收消息，分发到 StripedExecutorService
2. **StripedExecutorService** 线程池：按条纹分配任务，保证同一 Session 的消息有序
3. **WebSocket I/O**：实际的网络发送

## 12. 关键源文件

| 文件 | 路径 | 说明 |
|------|------|------|
| FeedTextWebSocketHandler | `feed/FeedTextWebSocketHandler.java` | WebSocket 端点处理器 |
| FeedMessageListener | `feed/FeedMessageListener.java` | Redis Pub/Sub 监听器 |
| SessionManager | `feed/SessionManager.java` | 会话与频道管理 |
| StripedExecutorService | `stripedexecutor/StripedExecutorService.java` | 条纹化执行器 |
| L2SnapshotFeedMessage | `feed/message/L2SnapshotFeedMessage.java` | L2 快照消息 |
| L2UpdateFeedMessage | `feed/message/L2UpdateFeedMessage.java` | L2 增量更新消息 |
| TickerFeedMessage | `feed/message/TickerFeedMessage.java` | 行情统计消息 |
| OrderFeedMessage | `feed/message/OrderFeedMessage.java` | 订单状态消息 |
| OrderMatchFeedMessage | `feed/message/OrderMatchFeedMessage.java` | 逐笔成交消息 |
| PongFeedMessage | `feed/message/PongFeedMessage.java` | 心跳响应消息 |
