# 撮合引擎详解

## 1. 概述

撮合引擎是 GitBitEx-Spot 交易系统的核心组件，负责按照价格-时间优先原则对买卖订单进行匹配。系统为每个交易对（如 BTC-USDT）创建一个独立的撮合引擎实例，实现产品级别的隔离与并行处理。

## 2. 引擎架构

每个撮合引擎由 4 个协程（goroutine）协同工作，通过 Go channel 串联形成流水线：

```mermaid
graph LR
    subgraph Kafka
        OrderTopic["matching_order_{productId}"]
        LogTopic["matching_message_{productId}"]
    end

    subgraph 撮合引擎 Engine
        Fetcher["Fetcher<br/>订单获取协程"]
        Applier["Applier<br/>撮合执行协程"]
        Committer["Committer<br/>日志提交协程"]
        Snapshotter["Snapshotter<br/>快照协程"]

        OrderCh["orderCh<br/>(channel)"]
        LogCh["logCh<br/>(channel)"]
        SnapshotCh["snapshotCh<br/>(channel)"]
    end

    subgraph Redis
        Snapshot["matching_snapshot_{productId}"]
    end

    OrderTopic -->|消费| Fetcher
    Fetcher -->|Order| OrderCh
    OrderCh -->|Order| Applier
    Applier -->|[]Log| LogCh
    LogCh -->|[]Log| Committer
    Committer -->|写入| LogTopic
    Committer -->|触发| SnapshotCh
    SnapshotCh --> Snapshotter
    Snapshotter -->|保存| Snapshot

    style Fetcher fill:#bbdefb
    style Applier fill:#c8e6c9
    style Committer fill:#fff9c4
    style Snapshotter fill:#f8bbd0
```

### 协程职责

| 协程 | 职责 | 说明 |
|------|------|------|
| **Fetcher** | 订单获取 | 从 Kafka `matching_order_{productId}` 主题消费订单，通过 `orderCh` 传递给 Applier |
| **Applier** | 撮合执行 | 从 `orderCh` 接收订单，执行撮合逻辑，生成匹配日志，通过 `logCh` 传递给 Committer |
| **Committer** | 日志提交 | 从 `logCh` 接收日志批次，写入 Kafka `matching_message_{productId}` 主题，并触发快照 |
| **Snapshotter** | 快照保存 | 定期将引擎状态（订单簿、序列号等）保存到 Redis，用于故障恢复 |

## 3. 价格-时间优先算法

撮合引擎采用 **价格-时间优先 (Price-Time Priority)** 原则：

1. **价格优先**: 买单中出价最高的优先成交，卖单中要价最低的优先成交
2. **时间优先**: 相同价格的订单按照到达时间排序，先到先成交

### 数据结构选择

使用 **TreeMap（红黑树）** 实现有序存储，保证关键操作的时间复杂度：

| 操作 | 时间复杂度 | 说明 |
|------|-----------|------|
| 插入价格层级 | O(log n) | 新价格加入订单簿 |
| 删除价格层级 | O(log n) | 价格层级清空后移除 |
| 查找最优价格 | O(log n) | 获取最高买价/最低卖价 |
| 遍历匹配 | O(k log n) | k 为匹配的价格层级数 |

### 排序规则

- **买单 (Bids)**: 按价格**降序**排列，最高价在前（`TreeMap` 降序比较器）
- **卖单 (Asks)**: 按价格**升序**排列，最低价在前（`TreeMap` 升序比较器）

## 4. 订单撮合流程

```mermaid
flowchart TD
    A[新订单到达] --> B{订单类型?}

    B -->|限价单 Limit| C[使用订单指定价格]
    B -->|市价买单 Market Buy| D["设置价格 = MaxFloat32<br/>按 funds 计算可买数量"]
    B -->|市价卖单 Market Sell| E["设置价格 = 0"]

    C --> F[获取对手方深度]
    D --> F
    E --> F

    F --> G{对手方是否有订单?}
    G -->|否| H{是限价单?}
    H -->|是| I["生成 OpenLog<br/>挂入订单簿"]
    H -->|否| J["生成 DoneLog<br/>市价单取消"]

    G -->|是| K[获取最优价格层级]
    K --> L{价格是否交叉?<br/>买价 >= 卖价}
    L -->|否| H

    L -->|是| M[遍历该层级订单]
    M --> N{当前订单剩余量?}

    N -->|>= 新订单剩余量| O["完全成交<br/>生成 MatchLog"]
    N -->|< 新订单剩余量| P["部分成交<br/>生成 MatchLog"]

    O --> Q["成交价 = Maker 价格"]
    P --> Q

    Q --> R[更新双方数量]
    R --> S{Maker 订单是否成交完?}
    S -->|是| T["生成 DoneLog (Maker)<br/>从簿中移除"]
    S -->|否| U[更新 Maker 剩余量]

    T --> V{新订单是否成交完?}
    U --> V
    V -->|否| W{该层级还有订单?}
    W -->|是| M
    W -->|否| X[移至下一价格层级]
    X --> G

    V -->|是| Y["生成 DoneLog (Taker)"]
    Y --> Z[返回所有日志]
    I --> Z
    J --> Z

    style A fill:#e3f2fd
    style Q fill:#fff3e0
    style Z fill:#e8f5e9
```

### 成交价规则

**所有成交均以 Maker（挂单方）的价格成交**。例如：
- 订单簿中有一笔卖单挂价 100 USDT
- 新买单以 105 USDT 限价进入
- 成交价为 100 USDT（Maker 的挂单价）

## 5. 市价单处理

### 市价卖单 (Market Sell)

- 设置价格为 `0`，确保能与所有买单匹配
- 按指定的 `size`（数量）进行匹配
- 如果买方深度不足以完全成交，剩余部分取消

### 市价买单 (Market Buy)

- 设置价格为 `MaxFloat32`，确保能与所有卖单匹配
- 使用 `funds`（资金额度）而非 `size` 控制买入量
- 在每个价格层级动态计算可买数量：`可买数量 = 剩余资金 / 当前层级价格`
- 逐层消耗资金直至用尽或卖方深度耗尽

## 6. 撮合日志类型

撮合过程产生三种日志，每条日志携带严格递增的序列号 `logSeq`：

| 日志类型 | 触发时机 | 关键字段 | 说明 |
|----------|----------|----------|------|
| **OpenLog** | 限价单挂入订单簿 | orderId, remainingSize, price, side | 订单未能完全成交，剩余部分挂入簿中 |
| **MatchLog** | 订单成交 | takerOrderId, makerOrderId, size, price, side | 记录一次成交的双方信息、成交量与成交价 |
| **DoneLog** | 订单完成 | orderId, remainingSize, reason | 订单完全成交或被取消，从簿中移除 |

### 日志序列号

- 每条日志分配一个严格递增的 `logSeq`
- 用于消费者的顺序保证和断点续传
- 所有日志写入 Kafka 主题：`matching_message_{productId}`

### 批量写入

- 日志先写入缓冲区，积累到 **100 条** 后批量刷写到 Kafka
- 单次撮合产生的所有日志作为一个批次提交，保证原子性

## 7. 去重窗口机制

为防止引擎重启后的重复处理，系统实现了基于位图的滑动窗口去重机制：

```mermaid
graph LR
    subgraph Window 滑动窗口
        Cap["容量: 10000"]
        Bitmap["Bitmap 位图"]
        Start["起始 OrderId"]
    end

    Order["新订单"] -->|检查 OrderId| Window
    Window -->|已存在| Reject["跳过重复订单"]
    Window -->|不存在| Accept["标记并处理"]
    Accept -->|窗口满| Slide["滑动窗口前移"]

    style Reject fill:#ffcdd2
    style Accept fill:#c8e6c9
```

### 工作原理

1. 维护一个容量为 10000 的位图窗口
2. 每处理一个订单，在位图中标记对应位置
3. 当新订单的 ID 落在窗口范围内，检查位图判断是否已处理
4. 窗口随着处理进度向前滑动，释放旧的位图空间
5. 快照时保存窗口状态，恢复时可直接还原

## 8. 快照与恢复

### 快照策略

| 触发条件 | 说明 |
|----------|------|
| 每 30 秒 | 定时快照 |
| 每处理 1000 个订单 | 增量快照 |
| TTL | 快照在 Redis 中保留 7 天 |

### 快照内容

| 数据项 | 说明 |
|--------|------|
| 所有挂单 | 订单簿中的所有未成交订单 |
| tradeSeq | 当前交易序列号 |
| logSeq | 最后日志序列号 |
| orderIdWindow | 去重窗口的完整状态 |
| Kafka offset | 消费位移，用于断点续传 |

### 恢复流程

```mermaid
sequenceDiagram
    participant Engine as 撮合引擎
    participant Redis as Redis
    participant Kafka as Kafka

    Note over Engine: 引擎启动
    Engine->>Redis: 加载快照 matching_snapshot_{productId}

    alt 快照存在
        Redis-->>Engine: 返回快照数据
        Engine->>Engine: 恢复订单簿状态
        Engine->>Engine: 恢复去重窗口
        Engine->>Engine: 恢复序列号
        Engine->>Kafka: 从快照记录的 offset 继续消费
    else 快照不存在
        Redis-->>Engine: 空
        Engine->>Engine: 初始化空订单簿
        Engine->>Kafka: 从最早 offset 开始消费
    end

    Note over Engine: 开始正常撮合

    loop 运行时快照
        Engine->>Engine: 检查快照触发条件
        alt 满足条件（30s/1000单）
            Engine->>Redis: 保存完整快照
            Note over Redis: TTL = 7 天
        end
    end
```

## 9. 关键源文件

| 文件路径 | 说明 |
|----------|------|
| `matching/engine.go` | 引擎主体，4 个协程的实现 |
| `matching/order_book.go` | 订单簿数据结构与撮合逻辑 |
| `matching/log.go` | OpenLog / MatchLog / DoneLog 定义 |
| `matching/window.go` | 位图滑动窗口去重 |
| `matching/api.go` | 引擎对外 API 接口 |
| `matching/bootstrap.go` | 引擎启动、恢复逻辑 |
| `matching/kafka_order_reader.go` | Kafka 订单消费者 |
| `matching/kafka_log_store.go` | Kafka 日志写入（批量缓冲 100） |
| `matching/kafka_log_reader.go` | Kafka 日志消费者（观察者模式） |
| `matching/redis_snapshot_store.go` | Redis 快照读写 |

## 10. 相关文档

- [订单簿数据结构](order-book.md)
- [Kafka 消息系统](kafka-messaging.md)
- [系统架构概览](architecture.md)
