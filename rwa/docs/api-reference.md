# Anchored Finance API 接口文档

## 1. API 概述

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| 基础 URL | `http://{host}:{port}{basePath}`，basePath 在配置文件中定义 |
| 协议 | HTTP REST API + WebSocket |
| 数据格式 | JSON |
| 时间戳 | 所有时间字段均为 Unix 秒级时间戳（int64） |
| Swagger | 非生产环境可访问 `{basePath}/swagger-ui/index.html` |

### 1.2 认证方式

#### API Key 认证

在请求头中携带 `X-API-Key`，仅在非 dev 环境且配置了 ApiKeys 时生效。

```
X-API-Key: your-api-key-here
```

#### API 签名认证（ApiSignMiddleware）

所有 API 请求均需通过签名中间件验证（可通过配置关闭）。签名所需请求头：

| Header | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `X-Api-Nonce` | string | 是 | 随机字符串，用于生成签名密钥 |
| `X-Api-Sign` | string | 是 | 请求签名（Keccak256 哈希） |
| `X-Api-Ts` | string | 是 | 当前时间戳（毫秒），有效窗口 45 秒 |
| `Authorization` | string | 否 | 可选的授权令牌（参与签名计算） |

**签名流程**：
1. 使用 nonce 的 Keccak256 哈希派生 AES-256-CBC 密钥和 IV
2. 构造 JSON payload：`{"uri": "<排序后的请求路径>", "nonce": "<nonce>", "ts": <timestamp>, "body": "<canonical body>", "authorization": "<auth>"}`
3. AES-256-CBC 加密 payload，Base64 编码
4. 对密文再做一次 Base64 编码
5. 对双重 Base64 结果取 Keccak256 哈希，得到最终签名（hex 字符串）

### 1.3 统一响应格式

```json
{
  "data": {},
  "msg": "",
  "code": 0,
  "requestId": "uuid-v4"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `data` | any | 业务数据，成功时返回 |
| `msg` | string | 错误信息，成功时为空字符串 |
| `code` | int | 状态码，`0` 表示成功，其他为错误码 |
| `requestId` | string | 请求唯一 ID（UUID v4） |

---

## 2. 接口列表

### 2.1 通用接口（Common）

#### 健康检查

检查服务是否正常运行。

- **URL**: `GET /common/health`
- **请求参数**: 无
- **响应示例**:

```json
{
  "data": "ok",
  "msg": "",
  "code": 0,
  "requestId": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### 2.2 行情接口（Trade）

#### 获取当前价格

获取指定股票的当前价格。

- **URL**: `GET /trade/currentPrice`
- **请求参数**:

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `symbol` | string | 是 | 股票代码 | `AAPL` |

- **响应示例**:

```json
{
  "data": {
    "symbol": "AAPL",
    "price": 178.52,
    "volume": 52341200.0,
    "timestamp": 1704067200
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

#### 获取最新报价

获取指定股票的最新买卖报价（Bid/Ask）。

- **URL**: `GET /trade/latestQuote`
- **请求参数**:

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `symbol` | string | 是 | 股票代码 | `AAPL` |

- **响应示例**:

```json
{
  "data": {
    "quote": {
      "timestamp": 1704067200,
      "bid_price": 178.50,
      "bid_size": 100,
      "ask_price": 178.55,
      "ask_size": 200,
      "bid_exchange": "Q",
      "ask_exchange": "Q",
      "conditions": ["R"],
      "tape": "C"
    }
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

#### 获取市场快照

获取指定股票的综合市场快照，包括最新交易、报价和 K 线数据。

- **URL**: `GET /trade/snapshot`
- **请求参数**:

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `symbol` | string | 是 | 股票代码 | `AAPL` |

- **响应示例**:

```json
{
  "data": {
    "snapshot": {
      "symbol": "AAPL",
      "latest_trade": {
        "timestamp": 1704067200,
        "price": 178.52,
        "size": 100,
        "exchange": "V",
        "id": 12345,
        "conditions": ["@"],
        "tape": "C"
      },
      "latest_quote": {
        "timestamp": 1704067200,
        "bid_price": 178.50,
        "bid_size": 100,
        "ask_price": 178.55,
        "ask_size": 200
      },
      "minute_bar": {
        "timestamp": 1704067200,
        "open": 178.40,
        "high": 178.60,
        "low": 178.35,
        "close": 178.52,
        "volume": 15234,
        "trade_count": 120,
        "vwap": 178.48
      },
      "daily_bar": {
        "timestamp": 1704067200,
        "open": 177.50,
        "high": 179.20,
        "low": 177.10,
        "close": 178.52,
        "volume": 52341200,
        "trade_count": 432100,
        "vwap": 178.35
      },
      "prev_daily_bar": {
        "timestamp": 1703980800,
        "open": 176.80,
        "high": 178.00,
        "low": 176.50,
        "close": 177.50,
        "volume": 48200300,
        "trade_count": 398000,
        "vwap": 177.25
      }
    }
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

#### 获取历史 K 线数据

获取指定股票的历史价格数据。

- **URL**: `GET /trade/historicalData`
- **请求参数**:

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `symbol` | string | 是 | 股票代码 | `AAPL` |
| `start_time` | int | 是 | 开始时间（Unix 秒级时间戳） | `1704067200` |
| `end_time` | int | 是 | 结束时间（Unix 秒级时间戳） | `1706745599` |
| `interval` | string | 是 | 时间间隔 | `1d` |
| `limit` | int | 否 | 返回最大条数 | `100` |

**interval 支持的格式**：

| 短格式 | 完整格式 | 说明 |
|--------|----------|------|
| `1m` | `1Min` | 1 分钟 |
| `5m` | `5Min` | 5 分钟 |
| `15m` | `15Min` | 15 分钟 |
| `1h` | `1Hour` | 1 小时 |
| `1d` | `1Day` | 1 天 |
| `1w` | `1Week` | 1 周 |

- **响应示例**:

```json
{
  "data": {
    "symbol": "AAPL",
    "data": [
      {
        "open": 177.50,
        "high": 179.20,
        "low": 177.10,
        "close": 178.52,
        "volume": 52341200,
        "timestamp": 1704067200
      }
    ]
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

#### 获取市场时钟

获取当前市场状态（开盘/闭盘），以及下一次开盘和闭盘时间。

- **URL**: `GET /trade/marketClock`
- **请求参数**: 无
- **响应示例**:

```json
{
  "data": {
    "timestamp": 1704067200,
    "is_open": true,
    "next_open": 1704153600,
    "next_close": 1704110400
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

#### 获取资产列表

获取可交易资产列表，支持按状态、类型、交易所筛选。

- **URL**: `GET /trade/assets`
- **请求参数**:

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `status` | string | 否 | 资产状态（`active` / `inactive`） | `active` |
| `asset_class` | string | 否 | 资产类型（`us_equity` / `crypto`） | `us_equity` |
| `exchange` | string | 否 | 交易所名称 | `NASDAQ` |

- **响应示例**:

```json
{
  "data": {
    "assets": [
      {
        "id": "b0b6dd9d-8b9b-48a9-ba46-b9d54906e415",
        "class": "us_equity",
        "exchange": "NASDAQ",
        "symbol": "AAPL",
        "name": "Apple Inc.",
        "status": "active",
        "tradable": true,
        "marginable": true,
        "maintenance_margin_requirement": 25,
        "shortable": true,
        "easy_to_borrow": true,
        "fractionable": true,
        "attributes": []
      }
    ]
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

#### 获取单个资产信息

按股票代码获取单个资产的详细信息。

- **URL**: `GET /trade/asset`
- **请求参数**:

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `symbol` | string | 是 | 股票代码 | `AAPL` |

- **响应示例**:

```json
{
  "data": {
    "asset": {
      "id": "b0b6dd9d-8b9b-48a9-ba46-b9d54906e415",
      "class": "us_equity",
      "exchange": "NASDAQ",
      "symbol": "AAPL",
      "name": "Apple Inc.",
      "status": "active",
      "tradable": true,
      "marginable": true,
      "maintenance_margin_requirement": 25,
      "shortable": true,
      "easy_to_borrow": true,
      "fractionable": true,
      "attributes": []
    }
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

### 2.3 股票接口（Stock）

#### 获取股票列表

获取系统中已上架的活跃股票列表。

- **URL**: `GET /stock/list`
- **请求参数**: 无
- **响应示例**:

```json
{
  "data": {
    "list": [
      {
        "id": 1,
        "symbol": "AAPL",
        "name": "Apple Inc.",
        "exchange": "NASDAQ",
        "about": "Apple Inc. designs, manufactures...",
        "status": "active",
        "contract": "0x1234...abcd",
        "createdAt": 1704067200,
        "updatedAt": 1704067200
      }
    ]
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

#### 获取股票详情

按股票代码获取股票的详细信息。

- **URL**: `GET /stock/detail`
- **请求参数**:

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `symbol` | string | 是 | 股票代码 | `AAPL` |

- **响应示例**:

```json
{
  "data": {
    "id": 1,
    "symbol": "AAPL",
    "name": "Apple Inc.",
    "exchange": "NASDAQ",
    "about": "Apple Inc. designs, manufactures...",
    "status": "active",
    "contract": "0x1234...abcd",
    "createdAt": 1704067200,
    "updatedAt": 1704067200
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

### 2.4 订单接口（Order）

#### 获取订单列表

获取订单列表，支持筛选和分页。

- **URL**: `GET /order/list`
- **请求参数**:

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `account_id` | int | 否 | 账户 ID | `1` |
| `symbol` | string | 否 | 股票代码 | `AAPL` |
| `side` | string | 否 | 订单方向（`buy` / `sell`） | `buy` |
| `status` | string | 否 | 订单状态 | `filled` |
| `page` | int | 否 | 页码（默认 1） | `1` |
| `page_size` | int | 否 | 每页条数（默认 20，最大 100） | `20` |

- **响应示例**:

```json
{
  "data": {
    "list": [
      {
        "id": 1,
        "clientOrderId": "order-001",
        "accountId": 1,
        "symbol": "AAPL",
        "assetType": "us_equity",
        "side": "buy",
        "type": "market",
        "quantity": "10.000000000000000000",
        "price": "178.520000000000000000",
        "stopPrice": "0",
        "status": "filled",
        "filledQuantity": "10.000000000000000000",
        "filledPrice": "178.520000000000000000",
        "remainingQuantity": "0",
        "contractTxHash": "0xabc...123",
        "externalOrderId": "alpaca-order-id",
        "provider": "alpaca",
        "commission": "0.01",
        "commissionAsset": "USD",
        "createdAt": 1704067200,
        "updatedAt": 1704067200,
        "submittedAt": 1704067200,
        "filledAt": 1704067210
      }
    ],
    "total": 1
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

#### 获取订单详情

按订单 ID 或客户端订单 ID 获取单个订单详情。

- **URL**: `GET /order/detail`
- **请求参数**:

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `id` | int | 否* | 订单 ID | `1` |
| `client_order_id` | string | 否* | 客户端订单 ID | `order-001` |

> *`id` 和 `client_order_id` 至少提供一个。

- **响应示例**:

```json
{
  "data": {
    "order": {
      "id": 1,
      "clientOrderId": "order-001",
      "accountId": 1,
      "symbol": "AAPL",
      "assetType": "us_equity",
      "side": "buy",
      "type": "market",
      "quantity": "10.000000000000000000",
      "price": "178.520000000000000000",
      "stopPrice": "0",
      "status": "filled",
      "filledQuantity": "10.000000000000000000",
      "filledPrice": "178.520000000000000000",
      "remainingQuantity": "0",
      "contractTxHash": "0xabc...123",
      "externalOrderId": "alpaca-order-id",
      "provider": "alpaca",
      "commission": "0.01",
      "commissionAsset": "USD",
      "createdAt": 1704067200,
      "updatedAt": 1704067200,
      "submittedAt": 1704067200,
      "filledAt": 1704067210
    }
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

#### 获取订单成交记录

获取指定订单的成交执行记录。

- **URL**: `GET /order/executions`
- **请求参数**:

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `order_id` | int | 是 | 订单 ID | `1` |

- **响应示例**:

```json
{
  "data": {
    "list": [
      {
        "id": 1,
        "orderId": 1,
        "executionId": "exec-001",
        "quantity": "5.000000000000000000",
        "price": "178.520000000000000000",
        "commission": "0.005",
        "commissionAsset": "USD",
        "provider": "alpaca",
        "externalId": "alpaca-exec-id",
        "executedAt": 1704067205,
        "createdAt": 1704067205
      }
    ]
  },
  "code": 0,
  "msg": "",
  "requestId": "..."
}
```

---

## 3. WebSocket 协议

### 3.1 连接方式

```
ws://{host}:{port}{basePath}
```

WebSocket 服务基于 [melody](https://github.com/olahol/melody) 库实现，使用标准 WebSocket 协议。

### 3.2 心跳保活

客户端发送文本消息 `ping`，服务端回复 `pong`。

```
--> ping
<-- pong
```

### 3.3 消息格式

#### 客户端发送消息（请求）

```json
{
  "id": 1,
  "method": "SUBSCRIBE",
  "params": {
    "type": "bar",
    "symbols": ["AAPL", "GOOGL"]
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint64 | 消息 ID，服务端原样返回 |
| `method` | string | 操作类型：`SUBSCRIBE` 或 `UNSUBSCRIBE` |
| `params` | object | 参数对象 |
| `params.type` | string | 订阅类型，目前支持 `bar`（K 线数据） |
| `params.symbols` | string[] | 股票代码列表 |

#### 服务端响应（操作结果）

```json
{
  "id": 1,
  "result": "success"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint64 | 对应请求的消息 ID |
| `result` | any | 操作结果 |

#### 服务端推送（数据流）

```json
{
  "stream": "bar",
  "data": {
    "symbol": "AAPL",
    "open": 178.40,
    "high": 178.60,
    "low": 178.35,
    "close": 178.52,
    "volume": 15234,
    "timestamp": 1704067200,
    "tradeCount": 120,
    "vwap": 178.48
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `stream` | string | 数据流类型，目前为 `bar` |
| `data` | object | 推送数据 |

**BarData 字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `symbol` | string | 股票代码 |
| `open` | float64 | 开盘价 |
| `high` | float64 | 最高价 |
| `low` | float64 | 最低价 |
| `close` | float64 | 收盘价 |
| `volume` | int64 | 成交量 |
| `timestamp` | int64 | 时间戳 |
| `tradeCount` | int64 | 成交笔数（可选） |
| `vwap` | float64 | 成交量加权平均价（可选） |

### 3.4 订阅/取消订阅

#### 订阅 K 线数据

```json
{
  "id": 1,
  "method": "SUBSCRIBE",
  "params": {
    "type": "bar",
    "symbols": ["AAPL", "GOOGL", "MSFT"]
  }
}
```

订阅后，服务端会将对应股票的实时 K 线数据通过 `stream: "bar"` 推送给该客户端。数据来源于 Alpaca Market Data WebSocket。

#### 取消订阅

```json
{
  "id": 2,
  "method": "UNSUBSCRIBE",
  "params": {
    "type": "bar",
    "symbols": ["GOOGL"]
  }
}
```

取消订阅后，该客户端不再接收对应股票的 K 线推送。

### 3.5 数据流架构

```mermaid
sequenceDiagram
    participant Client as 前端客户端
    participant WS as WebSocket Server
    participant Alpaca as Alpaca Market Data WS

    Client->>WS: 建立 WebSocket 连接
    WS-->>Client: 连接成功

    Client->>WS: {"method":"SUBSCRIBE","params":{"type":"bar","symbols":["AAPL"]}}
    WS->>Alpaca: subscribe bars ["AAPL"]
    WS-->>Client: {"id":1,"result":"success"}

    loop 实时数据推送
        Alpaca->>WS: bar data (AAPL)
        WS->>WS: 过滤已订阅的客户端
        WS-->>Client: {"stream":"bar","data":{...}}
    end

    Client->>WS: {"method":"UNSUBSCRIBE","params":{"type":"bar","symbols":["AAPL"]}}
    WS-->>Client: {"id":2,"result":"success"}
```

---

## 4. 错误码

### 4.1 通用错误码

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 0 | - | 成功 |
| 401 | - | 未授权（API Key 无效或签名校验失败） |
| 1000 | `ErrInvalidRequestParams` | 请求参数无效 |
| 1001 | `ErrInternalServerError` | 服务器内部错误 |
| 1002 | `ErrNotFound` | 数据未找到 |

### 4.2 股票模块错误码（2xxx）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 2000 | `ErrFailedToGetStockList` | 获取股票列表失败 |
| 2001 | `ErrFailedToGetStockDetail` | 获取股票详情失败 |
| 2002 | `ErrStockNotFound` | 股票未找到 |

### 4.3 行情模块错误码（3xxx）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 3000 | `ErrFailedToGetCurrentPrice` | 获取当前价格失败 |
| 3001 | `ErrFailedToGetHistoricalData` | 获取历史数据失败 |
| 3002 | `ErrFailedToGetMarketClock` | 获取市场时钟失败 |
| 3003 | `ErrFailedToGetLatestQuote` | 获取最新报价失败 |
| 3004 | `ErrInvalidTimestampFormat` | 时间戳格式无效 |
| 3005 | `ErrFailedToGetSnapshot` | 获取市场快照失败 |
| 3006 | `ErrFailedToGetAssets` | 获取资产列表失败 |
| 3007 | `ErrSymbolRequired` | 股票代码为必填项 |
| 3008 | `ErrFailedToGetAsset` | 获取资产信息失败 |

### 4.4 订单模块错误码（4xxx）

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 4000 | `ErrFailedToGetOrders` | 获取订单列表失败 |
| 4001 | `ErrFailedToGetOrderDetail` | 获取订单详情失败 |
| 4002 | `ErrOrderNotFound` | 订单未找到 |
| 4003 | `ErrFailedToGetOrderExecutions` | 获取订单成交记录失败 |

---

## 5. 第三方 API 文档链接

本系统的行情数据和交易执行通过 Alpaca 平台实现。以下为相关 API 文档：

### 5.1 Alpaca Trading API

- **官方文档**: [https://docs.alpaca.markets/docs/trading-api](https://docs.alpaca.markets/docs/trading-api)
- **订单 API**: [https://docs.alpaca.markets/reference/postorder](https://docs.alpaca.markets/reference/postorder)
- **账户 API**: [https://docs.alpaca.markets/reference/getaccount-1](https://docs.alpaca.markets/reference/getaccount-1)
- **资产 API**: [https://docs.alpaca.markets/reference/get-v2-assets](https://docs.alpaca.markets/reference/get-v2-assets)

### 5.2 Alpaca Market Data API

- **官方文档**: [https://docs.alpaca.markets/docs/market-data-api](https://docs.alpaca.markets/docs/market-data-api)
- **实时行情 WebSocket**: [https://docs.alpaca.markets/docs/real-time-stock-pricing-data](https://docs.alpaca.markets/docs/real-time-stock-pricing-data)
- **WebSocket 端点**: `wss://stream.data.alpaca.markets/v2/iex`（IEX 数据源）
- **历史 K 线**: [https://docs.alpaca.markets/reference/stockbars](https://docs.alpaca.markets/reference/stockbars)
- **快照数据**: [https://docs.alpaca.markets/reference/stocksnapshot](https://docs.alpaca.markets/reference/stocksnapshot)
- **最新报价**: [https://docs.alpaca.markets/reference/stocklatestquote](https://docs.alpaca.markets/reference/stocklatestquote)

### 5.3 Alpaca 认证方式

Alpaca API 使用 API Key + Secret 认证：

```
APCA-API-KEY-ID: <your-api-key>
APCA-API-SECRET-KEY: <your-api-secret>
```

WebSocket 认证消息：
```json
{
  "action": "auth",
  "key": "<your-api-key>",
  "secret": "<your-api-secret>"
}
```
