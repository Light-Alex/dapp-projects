# gitbitex-web 状态管理文档

## 1. 状态管理架构概览

gitbitex-web 使用 Vuex 3.0.1 作为全局状态管理方案，通过 `StoreService` 门面模式提供统一的访问接口。状态层负责接收来自 HTTP API 和 WebSocket 的数据，并通过 Vue 的响应式系统驱动视图更新。

```mermaid
flowchart TB
    subgraph 数据源
        HTTP["HTTP API<br/>/api/*"]
        WS["WebSocket<br/>wss://gitbitex.com:8080/ws"]
    end

    subgraph 服务层
        AccSvc["AccountService<br/>src/script/service/account.ts"]
        TradeSvc["TradeService<br/>src/script/service/trade.ts"]
        OrderSvc["OrderService<br/>src/script/service/order.ts"]
        WSSvc["WebSocketService<br/>src/script/service/websocket.ts"]
    end

    subgraph 状态层
        StoreService["StoreService<br/>src/script/store/service.ts"]
        Buffer["SocketMsgBuffer<br/>src/script/store/buffer.ts"]
        AccStore["AccountStore<br/>src/script/store/account.ts"]
        TradeStore["TradeStore<br/>src/script/store/trade.ts"]
    end

    subgraph 视图层
        Components["Vue 组件<br/>页面 + 功能组件"]
    end

    HTTP --> AccSvc
    HTTP --> TradeSvc
    HTTP --> OrderSvc
    WS --> WSSvc

    AccSvc --> StoreService
    TradeSvc --> StoreService
    OrderSvc --> StoreService
    WSSvc --> Buffer

    StoreService --> AccStore
    StoreService --> TradeStore
    Buffer -->|200-300ms 批量刷新| TradeStore

    AccStore -->|Vue 响应式| Components
    TradeStore -->|Vue 响应式| Components
```

## 2. StoreService 门面模式

`StoreService` 定义在 `src/script/store/service.ts`，是访问所有 Store 的唯一入口。它以静态属性的形式暴露各个子 Store：

```
StoreService
├── .Account  →  AccountStore 实例
└── .Trade    →  TradeStore 实例
```

所有组件和页面通过 `StoreService.Account` 和 `StoreService.Trade` 访问状态，而不是直接引用 Vuex store。这种门面模式提供了以下好处：

- **统一访问接口**: 所有数据操作通过一个入口
- **解耦依赖**: 组件不直接依赖 Vuex store 实例
- **便于测试**: 可以轻松替换 StoreService 实现

## 3. AccountStore 详解

源文件: `src/script/store/account.ts`

### 3.1 状态结构

```typescript
state: {
    userInfo: {
        id: string;
        email: string;
        nickname: string;
        avatar: string;
    } | null;
}
```

### 3.2 方法列表

| 方法 | 类型 | 说明 |
|------|------|------|
| `current()` | 异步 | 获取当前登录用户信息，调用 `GET /api/users/self` |
| `signOut()` | 异步 | 退出登录，清除 token 和用户状态 |
| `saveAvatar(url)` | 异步 | 保存头像，调用 `PUT /api/users/self/avatar` |
| `saveNickname(name)` | 异步 | 保存昵称，调用 `PUT /api/users/self/nickname` |

### 3.3 Token 管理

认证 Token 存储在浏览器 `localStorage` 中，key 为 `access-token`：

```mermaid
flowchart LR
    Login["登录成功"] --> SetToken["localStorage.setItem('access-token', token)"]
    SetToken --> Axios["Axios 请求拦截器<br/>自动附加 Authorization 头"]
    
    Logout["退出登录"] --> RemoveToken["localStorage.removeItem('access-token')"]
    
    WS["WebSocket 订阅"] --> ReadToken["localStorage.getItem('access-token')"]
    ReadToken --> WSAuth["subscribe 消息中携带 token"]
    
    Error401["HTTP 401 响应"] --> Redirect["重定向到 /account/signin"]
```

- **登录**: 成功后将服务端返回的 token 存入 localStorage
- **请求拦截器**: `src/script/service/request.ts` 中 Axios 实例的请求拦截器自动从 localStorage 读取 token 并附加到 HTTP 请求头
- **401 处理**: 响应拦截器检测到 401 状态码时，自动重定向到登录页
- **WebSocket 认证**: 订阅消息中携带 token 字段

## 4. TradeStore 详解

源文件: `src/script/store/trade.ts`

TradeStore 是应用中最复杂的状态模块，管理所有交易相关数据。

### 4.1 状态结构

```mermaid
graph TB
    TradeState["TradeStore.state"]

    subgraph products["products: Product[]"]
        Product1["Product {<br/>id: 'BTC-USDT'<br/>baseCurrency: 'BTC'<br/>quoteCurrency: 'USDT'<br/>baseMinSize: '0.001'<br/>quoteIncrement: '0.01'<br/>price: '50000.00'<br/>change24h: '+2.5%'<br/>volume24h: '1234.56'<br/>high24h: '51000'<br/>low24h: '49000'<br/>}"]
        Product2["Product { ... }"]
        ProductN["..."]
    end

    subgraph objects["objects: { [productId]: TradeObject }"]
        subgraph BTC_USDT["objects['BTC-USDT']"]
            ProductRef["product: Product 引用"]
            subgraph orderBook["orderBook"]
                Bids["bids: [<br/>&nbsp;&nbsp;[price, size],<br/>&nbsp;&nbsp;[price, size],<br/>&nbsp;&nbsp;...<br/>]"]
                Asks["asks: [<br/>&nbsp;&nbsp;[price, size],<br/>&nbsp;&nbsp;[price, size],<br/>&nbsp;&nbsp;...<br/>]"]
            end
            History["history: [<br/>&nbsp;&nbsp;[timestamp, open, high, low, close, volume],<br/>&nbsp;&nbsp;...<br/>]"]
            HistoryType["historyType: '1m' | '5m' | '15m' | ..."]
            TradeHistory["tradeHistory: [<br/>&nbsp;&nbsp;{time, price, size, side},<br/>&nbsp;&nbsp;...<br/>]"]
            OpenOrders["openOrders: [<br/>&nbsp;&nbsp;{id, price, size, side, type, status},<br/>&nbsp;&nbsp;...<br/>]"]
        end
    end

    subgraph funds["funds: { [currency]: Fund }"]
        BTC_Fund["funds['BTC'] {<br/>available: '1.5'<br/>hold: '0.2'<br/>currency: 'BTC'<br/>}"]
        USDT_Fund["funds['USDT'] {<br/>available: '50000'<br/>hold: '10000'<br/>currency: 'USDT'<br/>}"]
    end

    TradeState --> products
    TradeState --> objects
    TradeState --> funds
```

### 4.2 状态字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `products` | `Product[]` | 所有可交易产品（交易对）列表 |
| `objects` | `{ [productId]: TradeObject }` | 按产品 ID 索引的交易数据对象 |
| `objects[id].product` | `Product` | 产品引用 |
| `objects[id].orderBook` | `{ bids, asks }` | 订单簿数据，bids(买盘)和 asks(卖盘) |
| `objects[id].history` | `Array` | K 线历史数据 |
| `objects[id].historyType` | `string` | 当前 K 线时间粒度 |
| `objects[id].tradeHistory` | `Array` | 最新成交历史 |
| `objects[id].openOrders` | `Array` | 用户当前挂单 |
| `funds` | `{ [currency]: Fund }` | 按币种索引的资金数据 |
| `funds[currency].available` | `string` | 可用余额 |
| `funds[currency].hold` | `string` | 冻结金额 |

### 4.3 数据加载方法

| 方法 | HTTP 端点 | 说明 |
|------|-----------|------|
| `loadProducts()` | `GET /api/products` | 加载所有交易对信息，初始化 objects 结构 |
| `loadProductHistory(productId, granularity)` | `GET /api/products/{id}/candles?granularity=` | 加载指定粒度的 K 线历史数据 |
| `loadTradeHistory(productId)` | `GET /api/products/{id}/trades` | 加载最新成交记录 |
| `loadOpenOrders(productId)` | `GET /api/orders?productId=&status=open` | 加载用户当前挂单 |
| `loadFunds()` | `GET /api/accounts` | 加载用户所有币种的资金信息 |

### 4.4 WebSocket 相关方法

| 方法 | 说明 |
|------|------|
| `subscribe(productId)` | 订阅指定产品的 WebSocket 频道 |
| `unsubscribe(productId)` | 取消订阅指定产品的 WebSocket 频道 |
| `updateOrderBook(productId, data)` | 更新订单簿（应用 l2update delta） |
| `updateHistory(productId, data)` | 更新 K 线数据（追加或更新最后一根） |
| `updateTradeHistory(productId, data)` | 追加新的成交记录 |
| `updateProductPrice(productId, data)` | 更新产品的最新价格和涨跌信息 |
| `updateOrder(data)` | 更新用户订单状态 |
| `updateFund(data)` | 更新用户资金余额 |

## 5. WebSocket 消息处理流程

### 5.1 消息路由

WebSocket 收到的消息通过 `parseWebSocketMessage()` 函数按类型分发到对应的 Store 方法。

```mermaid
flowchart TB
    WS["WebSocket 消息到达"]
    Parse["parseWebSocketMessage()"]
    Buffer["SocketMsgBuffer<br/>消息累积"]
    Flush["定时器触发<br/>200-300ms"]

    WS --> Parse
    Parse --> Buffer
    Buffer --> Flush

    subgraph 按消息类型分发
        Snapshot["type: 'snapshot'<br/>完整订单簿快照"]
        L2Update["type: 'l2update'<br/>订单簿增量更新"]
        Candles["type: 'candles'<br/>K 线数据更新"]
        Match["type: 'match'<br/>新成交记录"]
        Ticker["type: 'ticker'<br/>最新行情"]
        OrderMsg["type: 'order'<br/>订单状态变更"]
        FundsMsg["type: 'funds'<br/>资金余额变更"]
    end

    subgraph Store处理
        SetOrderBook["TradeStore<br/>替换整个 orderBook"]
        DeltaOrderBook["TradeStore<br/>updateOrderBook()"]
        UpdateHistory["TradeStore<br/>updateHistory()"]
        AddTrade["TradeStore<br/>updateTradeHistory()"]
        UpdatePrice["TradeStore<br/>updateProductPrice()"]
        UpdateOrder["TradeStore<br/>updateOrder()"]
        UpdateFund["TradeStore<br/>updateFund()"]
    end

    Flush --> Snapshot --> SetOrderBook
    Flush --> L2Update --> DeltaOrderBook
    Flush --> Candles --> UpdateHistory
    Flush --> Match --> AddTrade
    Flush --> Ticker --> UpdatePrice
    Flush --> OrderMsg --> UpdateOrder
    Flush --> FundsMsg --> UpdateFund
```

### 5.2 消息类型与处理

| 消息类型 | 触发条件 | Store 方法 | 处理逻辑 |
|----------|----------|------------|----------|
| `snapshot` | 订阅 level2 后首次推送 | 替换 orderBook | 用完整数据替换当前买卖盘 |
| `l2update` | 订单簿有变化时 | `updateOrderBook()` | 增量更新买卖盘中变化的价位 |
| `candles` | K 线数据变化时 | `updateHistory()` | 更新最后一根 K 线或追加新 K 线 |
| `match` | 新成交产生时 | `updateTradeHistory()` | 在成交列表头部插入新记录 |
| `ticker` | 行情变化时 | `updateProductPrice()` | 更新产品最新价格和 24h 统计 |
| `order` | 用户订单状态变化时 | `updateOrder()` | 更新或移除挂单列表中的订单 |
| `funds` | 用户资金变化时 | `updateFund()` | 更新对应币种的可用和冻结余额 |

### 5.3 updateOrderBook 逻辑详解

订单簿的增量更新是最复杂的状态操作之一：

```mermaid
flowchart TB
    L2["收到 l2update 消息<br/>{changes: [[side, price, size], ...]}"]
    
    Loop["遍历 changes 数组"]
    Check["检查 side 字段"]
    
    subgraph 买盘更新
        FindBid["在 bids 中查找 price"]
        BidExists{price 存在?}
        BidSizeZero{size == 0?}
        RemoveBid["删除该价位"]
        UpdateBid["更新该价位的 size"]
        InsertBid["在正确位置插入<br/>（按价格降序）"]
    end
    
    subgraph 卖盘更新
        FindAsk["在 asks 中查找 price"]
        AskExists{price 存在?}
        AskSizeZero{size == 0?}
        RemoveAsk["删除该价位"]
        UpdateAsk["更新该价位的 size"]
        InsertAsk["在正确位置插入<br/>（按价格升序）"]
    end

    L2 --> Loop
    Loop --> Check
    Check -->|side = buy| FindBid
    Check -->|side = sell| FindAsk
    
    FindBid --> BidExists
    BidExists -->|是| BidSizeZero
    BidSizeZero -->|是| RemoveBid
    BidSizeZero -->|否| UpdateBid
    BidExists -->|否| InsertBid

    FindAsk --> AskExists
    AskExists -->|是| AskSizeZero
    AskSizeZero -->|是| RemoveAsk
    AskSizeZero -->|否| UpdateAsk
    AskExists -->|否| InsertAsk
```

**关键规则：**
- `size == 0` 表示该价位已无挂单，需要从列表中移除
- 买盘 (bids) 按价格**降序**排列（最高买价在前）
- 卖盘 (asks) 按价格**升序**排列（最低卖价在前）
- 新价位需要插入到正确的排序位置

## 6. HTTP 服务层

### 6.1 Request 基础类

源文件: `src/script/service/request.ts`

- 基于 Axios 的单例封装
- Base URL: `/api`
- **请求拦截器**: 自动附加 `Authorization: Bearer {token}` 头
- **响应拦截器**: 401 状态码自动重定向到登录页 `/account/signin`

### 6.2 AccountService

源文件: `src/script/service/account.ts`

| 方法 | HTTP 方法 | 端点 | 说明 |
|------|-----------|------|------|
| `signup(email, password)` | POST | `/api/users` | 用户注册 |
| `signin(email, password)` | POST | `/api/users/sessions` | 用户登录 |
| `signOut()` | DELETE | `/api/users/sessions` | 退出登录 |
| `current()` | GET | `/api/users/self` | 获取当前用户 |
| `getFunds()` | GET | `/api/accounts` | 获取所有币种资金 |
| `getDepositAddress(currency)` | GET | `/api/accounts/{currency}/deposit-address` | 获取充值地址 |
| `postWithdrawal(currency, data)` | POST | `/api/accounts/{currency}/withdrawal` | 提交提现 |
| `getTransactions(currency)` | GET | `/api/accounts/{currency}/transactions` | 获取交易记录 |
| `changePassword(data)` | PUT | `/api/users/self/password` | 修改密码 |
| `sendEmailVerifyCode(email)` | POST | `/api/users/verify-code` | 发送邮箱验证码 |
| `resetPassword(data)` | PUT | `/api/users/password` | 重置密码 |
| `getFileUrl(fileId)` | GET | `/api/files/{id}` | 获取文件 URL |
| `saveNickname(name)` | PUT | `/api/users/self/nickname` | 保存昵称 |
| `saveAvatar(url)` | PUT | `/api/users/self/avatar` | 保存头像 |

### 6.3 TradeService

源文件: `src/script/service/trade.ts`

| 方法 | HTTP 方法 | 端点 | 说明 |
|------|-----------|------|------|
| `getProducts()` | GET | `/api/products` | 获取所有交易对 |
| `getProductHistory(productId, granularity)` | GET | `/api/products/{id}/candles?granularity=` | 获取 K 线历史 |
| `getProductTradeHistory(productId)` | GET | `/api/products/{id}/trades` | 获取成交历史 |

### 6.4 OrderService

源文件: `src/script/service/order.ts`

| 方法 | HTTP 方法 | 端点 | 说明 |
|------|-----------|------|------|
| `createOrder(data)` | POST | `/api/orders` | 创建订单 |
| `getOrders(params)` | GET | `/api/orders?page=&size=&status=` | 分页查询订单 |
| `cancelOrder(orderId)` | DELETE | `/api/orders/{id}` | 取消指定订单 |
| `cancelAll(productId)` | DELETE | `/api/orders?productId=` | 取消某产品所有订单 |

### 6.5 ServerService

源文件: `src/script/service/server.ts`

| 方法 | HTTP 方法 | 端点 | 说明 |
|------|-----------|------|------|
| `getConfig()` | GET | `/api/configs` | 获取服务器配置 |

### 6.6 FileService

源文件: `src/script/service/file.ts`

| 方法 | HTTP 方法 | 端点 | 说明 |
|------|-----------|------|------|
| `getFileUrl(fileId)` | GET | `/api/files/{id}` | 获取文件 URL |
| `upload(file)` | POST | `/api/files` | 上传文件 |

## 7. 数据流完整生命周期

以交易页面为例，展示数据从加载到更新的完整生命周期：

```mermaid
sequenceDiagram
    participant User as 用户
    participant Page as TradePage
    participant Store as StoreService.Trade
    participant HTTP as HTTP API
    participant WS as WebSocket
    participant Buffer as SocketMsgBuffer
    participant Vue as Vue 响应式系统
    participant DOM as 浏览器 DOM

    Note over Page: mounted() 触发
    
    rect rgb(240, 248, 255)
        Note over Page,HTTP: 阶段一：HTTP 加载初始数据
        Page->>Store: loadTradeHistory(productId)
        Store->>HTTP: GET /api/products/{id}/trades
        HTTP-->>Store: 成交历史数据
        Store->>Store: commit('SET_TRADE_HISTORY', data)
        
        Page->>Store: loadFunds()
        Store->>HTTP: GET /api/accounts
        HTTP-->>Store: 资金数据
        Store->>Store: commit('SET_FUNDS', data)
    end

    rect rgb(255, 248, 240)
        Note over Page,WS: 阶段二：WebSocket 订阅实时数据
        Page->>Store: subscribe(productId)
        Store->>WS: subscribe({channels: [LEVEL2, MATCH, CANDLES, ORDER]})
        WS-->>Store: snapshot（完整订单簿）
        Store->>Store: commit('SET_ORDER_BOOK', data)
    end

    rect rgb(240, 255, 240)
        Note over WS,DOM: 阶段三：实时数据持续推送
        loop 持续推送
            WS->>Buffer: l2update / match / ticker / ...
            Buffer->>Buffer: 消息累积
            Note over Buffer: 200-300ms 定时器
            Buffer->>Store: 批量处理消息
            Store->>Store: mutations 更新状态
            Store->>Vue: 状态变更通知
            Vue->>DOM: 响应式重新渲染
        end
    end

    rect rgb(255, 240, 240)
        Note over User,HTTP: 阶段四：用户交互
        User->>Page: 点击"买入"按钮
        Page->>HTTP: POST /api/orders
        HTTP-->>Page: 订单创建成功
        WS->>Buffer: order（订单状态变更）
        WS->>Buffer: funds（资金变更）
        Buffer->>Store: 批量更新
        Store->>Vue: 状态变更
        Vue->>DOM: 更新挂单列表和余额显示
    end
```
