# gitbitex-web 图表系统文档

## 1. 图表系统概览

gitbitex-web 的图表系统采用双引擎架构，同时集成 Highcharts 和 TradingView 两个图表库，分别用于不同场景下的行情展示需求。

| 图表引擎 | 用途 | 源文件路径 |
|----------|------|------------|
| Highcharts 7.2.0 | K 线图（内置版）和深度图 | `src/script/component/chart/candle/candle.ts` <br/> `src/script/component/chart/depth/depth.ts` |
| TradingView Charting Library | 专业级交易图表 | `src/script/component/chart/tradingview/tradingview.ts` <br/> `src/script/chart/datafeed.ts` <br/> `src/script/chart/config.ts` |

```mermaid
flowchart TB
    subgraph 图表切换器
        TradeView["TradeViewChartComponent<br/>src/script/component/chart/trade-view/trade-view.ts"]
        Tabs["标签页切换"]
    end

    subgraph Highcharts引擎
        Candle["CandleChartComponent<br/>K 线图 + 成交量"]
        Depth["DepthChartComponent<br/>深度图"]
    end

    subgraph TradingView引擎
        TV["TradingviewChartComponent<br/>专业交易图表"]
        DataFeed["DataFeed<br/>UDF 数据源适配"]
    end

    TradeView --> Tabs
    Tabs -->|"K 线图"| Candle
    Tabs -->|"深度图"| Depth
    Tabs -->|"TradingView"| TV
    TV --> DataFeed
```

## 2. K 线图 (CandleChartComponent)

源文件: `src/script/component/chart/candle/candle.ts`

### 2.1 Highcharts.stockChart 配置

CandleChartComponent 使用 Highcharts 的 `stockChart` 类型，这是专门为金融时间序列数据设计的图表类型。

**核心配置：**

- **图表类型**: `Highcharts.stockChart`
- **主题**: 深色主题 (背景 #15232c)
- **子图**: 两个 Y 轴区域 -- 价格区域（上方 70%）和成交量区域（下方 30%）
- **交互**: 十字准线 (crosshair)、工具提示 (tooltip)

### 2.2 时间范围

| 范围标签 | 粒度 (秒) | 粒度标识 | K 线类型 |
|----------|-----------|----------|----------|
| 1m (1 分钟) | 60 | `candles_60` | 分钟线 |
| 5m (5 分钟) | 300 | `candles_300` | 分钟线 |
| 15m (15 分钟) | 900 | `candles_900` | 分钟线 |
| 30m (30 分钟) | 1800 | `candles_1800` | 分钟线 |
| 1h (1 小时) | 3600 | `candles_3600` | 小时线 |
| 6h (6 小时) | 21600 | `candles_21600` | 小时线 |
| 1d (1 天) | 86400 | `candles_86400` | 日线 |

用户通过 `ChartSliderComponent` (`src/script/component/chart/slider/slider.ts`) 切换时间范围。切换后触发重新加载对应粒度的历史数据并重新订阅 WebSocket 频道。

### 2.3 图表类型

| 类型 | 说明 | Highcharts 系列类型 |
|------|------|---------------------|
| 蜡烛图 (Candlestick) | 显示开盘/最高/最低/收盘价 | `candlestick` |
| 折线图 (Line) | 仅显示收盘价连线 | `line` |

### 2.4 技术指标

| 指标 | 参数 | 颜色 | 说明 |
|------|------|------|------|
| EMA 12 | 周期 = 12 | 黄色 | 12 周期指数移动平均线，反映短期趋势 |
| EMA 26 | 周期 = 26 | 紫色 | 26 周期指数移动平均线，反映中期趋势 |

EMA (指数移动平均线) 计算公式: `EMA_today = Price_today * k + EMA_yesterday * (1 - k)`，其中 `k = 2 / (N + 1)`。

### 2.5 成交量子图

- 位于价格图下方
- 柱状图 (`column` 系列)
- 颜色: 上涨蜡烛对应绿色柱，下跌蜡烛对应红色柱
- Y 轴独立于价格轴

### 2.6 深色主题配置

| 属性 | 值 | 说明 |
|------|-----|------|
| 背景色 | #15232c | 深蓝黑色 |
| 网格线颜色 | #1c2d3a | 微暗网格 |
| 文字颜色 | #7f8c97 | 灰色文字 |
| 上涨颜色 | #00c087 | 绿色 |
| 下跌颜色 | #ef5454 | 红色 |
| 十字准线 | #4c5c68 | 灰色虚线 |

### 2.7 数据加载流程

```mermaid
sequenceDiagram
    participant User as 用户
    participant Slider as ChartSliderComponent
    participant Candle as CandleChartComponent
    participant Store as StoreService.Trade
    participant HTTP as HTTP API
    participant WS as WebSocket
    participant Chart as Highcharts 实例

    User->>Slider: 选择时间范围（如 1h）
    Slider->>Candle: 触发范围切换事件

    rect rgb(230, 245, 255)
        Note over Candle,HTTP: 阶段一：加载历史数据
        Candle->>Store: loadProductHistory(productId, 3600)
        Store->>HTTP: GET /api/products/{id}/candles?granularity=3600
        HTTP-->>Store: K 线历史数据数组
        Store->>Store: 存入 objects[productId].history
        Store->>Store: 设置 historyType = 3600
    end

    rect rgb(240, 255, 240)
        Note over Candle,Chart: 阶段二：渲染图表
        Store-->>Candle: Vue 响应式通知数据变更
        Candle->>Candle: 解析 history 数据
        Candle->>Candle: 计算 EMA(12) 和 EMA(26)
        Candle->>Candle: 分离 OHLC 数据和 Volume 数据
        Candle->>Chart: chart.series[0].setData(ohlc)
        Candle->>Chart: chart.series[1].setData(volume)
        Candle->>Chart: chart.series[2].setData(ema12)
        Candle->>Chart: chart.series[3].setData(ema26)
    end

    rect rgb(255, 248, 240)
        Note over WS,Chart: 阶段三：实时更新
        loop WebSocket 推送
            WS->>Store: candles 消息 (match 更新最新价格)
            Store->>Store: updateHistory() 更新最后一根 K 线
            Store-->>Candle: Vue 响应式通知
            Candle->>Chart: 更新当前 K 线<br/>或追加新 K 线
        end
    end
```

### 2.8 实时 K 线更新逻辑

当收到 WebSocket 的 `candles` 或 `match` 消息时：

1. **当前 K 线更新**: 如果消息的时间戳在最后一根 K 线的时间区间内
   - 更新 high (如果新价更高)
   - 更新 low (如果新价更低)
   - 更新 close 为最新价
   - 累加 volume
2. **新 K 线追加**: 如果消息的时间戳超过当前 K 线区间
   - 创建新的 K 线数据点: `[timestamp, open, high, low, close, volume]`
   - open = close = high = low = 最新价

## 3. 深度图 (DepthChartComponent)

源文件: `src/script/component/chart/depth/depth.ts`

### 3.1 Highcharts.chart 配置

深度图使用 Highcharts 基础 `chart` 类型，以面积图 (`area`) 形式展示买卖双方的累计挂单深度。

### 3.2 核心特性

| 特性 | 说明 |
|------|------|
| 图表类型 | 面积图 (area) |
| 深度档数 | 最大 200 档 |
| 买盘颜色 | 绿色 (#00c087) 面积 |
| 卖盘颜色 | 红色 (#ef5454) 面积 |
| 中间价标线 | 买一卖一中间的 plotline 虚线 |
| 更新频率 | 每 1 秒检查更新 |
| 聚合精度 | 与订单簿面板共享 8 级聚合 |
| 可见性优化 | 页面不可见时停止更新 |

### 3.3 累计深度计算

深度图的 Y 轴显示的是累计挂单量，而非单档挂单量。计算方式如下：

```mermaid
flowchart TB
    subgraph 原始订单簿
        Bids["买盘 (按价格降序)<br/>50000: 1.0<br/>49999: 0.5<br/>49998: 0.8<br/>49997: 1.2<br/>..."]
        Asks["卖盘 (按价格升序)<br/>50001: 0.8<br/>50002: 1.5<br/>50003: 0.3<br/>50004: 2.0<br/>..."]
    end

    subgraph 累计计算
        BidCum["买盘累计 (从高到低累加)<br/>50000: 1.0<br/>49999: 1.5 (1.0+0.5)<br/>49998: 2.3 (1.5+0.8)<br/>49997: 3.5 (2.3+1.2)<br/>..."]
        AskCum["卖盘累计 (从低到高累加)<br/>50001: 0.8<br/>50002: 2.3 (0.8+1.5)<br/>50003: 2.6 (2.3+0.3)<br/>50004: 4.6 (2.6+2.0)<br/>..."]
    end

    subgraph 面积图
        DepthChart["绿色面积 (买盘累计)<br/>红色面积 (卖盘累计)<br/>中间虚线 (中间价)"]
    end

    Bids --> BidCum
    Asks --> AskCum
    BidCum --> DepthChart
    AskCum --> DepthChart
```

### 3.4 聚合与订单簿同步

深度图支持与订单簿面板相同的 8 级聚合精度：

| 聚合级别 | 精度值 | 效果 |
|----------|--------|------|
| 1 | 0.01 | 精细显示，曲线锯齿多 |
| 2 | 0.1 | |
| 3 | 1 | |
| 4 | 10 | |
| 5 | 50 | |
| 6 | 100 | |
| 7 | 500 | |
| 8 | 1000 | 粗略显示，曲线平滑 |

聚合使用与订单簿相同的 `Trade_margeOrderBook()` 函数 (`src/script/helper.ts`)。

### 3.5 更新策略

```mermaid
flowchart TB
    Timer["1 秒定时器触发"]
    Visible{"页面可见?"}
    DataChanged{"订单簿数据<br/>有变化?"}
    Skip["跳过更新"]
    Update["重新计算累计深度"]
    Render["更新 Highcharts 图表"]

    Timer --> Visible
    Visible -->|否| Skip
    Visible -->|是| DataChanged
    DataChanged -->|否| Skip
    DataChanged -->|是| Update
    Update --> Render
```

**可见性优化**: 使用 Page Visibility API 检测页面是否在前台显示。当用户切换到其他标签页时，停止深度图的定时更新，节省计算资源。当页面重新可见时，立即执行一次更新并恢复定时器。

### 3.6 中间价标线

在买盘最高价和卖盘最低价之间绘制一条垂直虚线 (Highcharts plotLine)：

- **位置**: `(bestBid + bestAsk) / 2`
- **颜色**: 灰色虚线
- **用途**: 直观标识当前市场价差位置

## 4. TradingView 集成

### 4.1 架构概览

TradingView 图表通过其 JavaScript Charting Library 集成，采用 UDF (Universal Data Feed) 协议与应用数据层对接。

```mermaid
flowchart TB
    subgraph TradingView库
        Widget["TradingView Widget<br/>图表控件"]
        API["Widget API<br/>图表操作接口"]
    end

    subgraph 数据源适配
        DataFeed["DataFeed<br/>src/script/chart/datafeed.ts"]
        Config["图表配置<br/>src/script/chart/config.ts"]
    end

    subgraph 实时数据
        DPU["DataPulseUpdater<br/>10 秒轮询"]
        QPU["QuotesPulseUpdater<br/>60s/10s 轮询"]
    end

    subgraph 应用数据层
        Store["StoreService.Trade<br/>TradeStore 状态"]
        HTTP["HTTP API<br/>历史 K 线"]
    end

    Widget --> DataFeed
    DataFeed --> Config
    DataFeed --> DPU
    DataFeed --> QPU
    DPU --> Store
    QPU --> Store
    DataFeed --> Store
    DataFeed --> HTTP
```

源文件:
- `src/script/component/chart/tradingview/tradingview.ts` -- TradingView 组件
- `src/script/chart/datafeed.ts` -- UDF DataFeed 实现
- `src/script/chart/config.ts` -- 图表配置

### 4.2 UDF DataFeed 协议

DataFeed 是 TradingView 与应用数据层之间的适配器，实现了 TradingView 要求的 UDF 接口：

| 方法 | 说明 | 数据来源 |
|------|------|----------|
| `onReady(callback)` | 初始化 DataFeed，返回配置 | 静态配置 |
| `resolveSymbol(symbolName, onResolve)` | 解析交易对符号信息 | `StoreService.Trade.products` |
| `getBars(symbolInfo, resolution, from, to, onResult)` | 获取历史 K 线数据 | `StoreService.Trade.getObject().history` / HTTP |
| `subscribeBars(symbolInfo, resolution, onTick)` | 订阅实时 K 线更新 | DataPulseUpdater |
| `unsubscribeBars(listenerGuid)` | 取消订阅实时更新 | - |
| `searchSymbols(input, exchange, type, onResult)` | 搜索交易对符号 | `StoreService.Trade.products` |

### 4.3 resolveSymbol 元数据

`resolveSymbol()` 方法返回交易对的元数据，供 TradingView 图表使用：

```javascript
{
    name: "BTC-USDT",
    full_name: "BTC-USDT",
    description: "BTC/USDT",
    type: "crypto",
    session: "24x7",          // 加密货币 24 小时交易
    exchange: "GitBitEx",
    timezone: "Asia/Shanghai",
    minmov: 1,
    pricescale: 100,          // 价格精度
    has_intraday: true,
    has_daily: true,
    supported_resolutions: ["1", "5", "15", "30", "60", "360", "D"]
}
```

### 4.4 getBars 数据获取

```mermaid
flowchart TB
    TV["TradingView 请求 getBars()"]
    CheckStore{"Store 中有<br/>对应粒度数据?"}
    UseStore["直接从 Store 读取"]
    LoadHTTP["loadProductHistory()<br/>GET /api/products/{id}/candles"]
    Filter["按 from/to 时间范围过滤"]
    Transform["转换为 TradingView Bar 格式<br/>{time, open, high, low, close, volume}"]
    Return["返回给 TradingView"]

    TV --> CheckStore
    CheckStore -->|是| UseStore
    CheckStore -->|否| LoadHTTP
    LoadHTTP --> UseStore
    UseStore --> Filter
    Filter --> Transform
    Transform --> Return
```

### 4.5 DataPulseUpdater

`DataPulseUpdater` 负责将实时数据推送给 TradingView 图表：

| 属性 | 值 | 说明 |
|------|-----|------|
| 轮询间隔 | 10 秒 | 每 10 秒检查新数据 |
| 数据来源 | StoreService.Trade 中的 history | 由 WebSocket 持续更新 |
| 推送方式 | 调用 `onTick(bar)` 回调 | TradingView 注册的回调函数 |

工作流程：

```mermaid
flowchart TB
    Timer["10 秒定时器"]
    Read["读取 Store 中最新的 K 线"]
    Compare{"与上次推送的<br/>数据不同?"}
    Push["调用 onTick(bar)<br/>推送给 TradingView"]
    Skip["跳过"]

    Timer --> Read
    Read --> Compare
    Compare -->|是| Push
    Compare -->|否| Skip
    Push --> Timer
    Skip --> Timer
```

### 4.6 QuotesPulseUpdater

`QuotesPulseUpdater` 负责更新 TradingView 的报价信息（如当前价格、涨跌等）：

| 属性 | 值 | 说明 |
|------|-----|------|
| 首次间隔 | 10 秒 | 首次加载后较快更新 |
| 后续间隔 | 60 秒 | 稳定后降低更新频率 |
| 数据来源 | StoreService.Trade.products | 由 WebSocket ticker 更新 |

### 4.7 深色主题

TradingView 图表使用与整体应用一致的深色主题：

| 属性 | 值 |
|------|-----|
| 背景色 | #15232c |
| 网格线颜色 | #1c2d3a |
| 文字颜色 | #7f8c97 |
| 蜡烛上涨色 | #00c087 |
| 蜡烛下跌色 | #ef5454 |
| 工具栏背景 | #15232c |

## 5. 图表数据生命周期

展示从数据产生到图表渲染的完整生命周期：

```mermaid
flowchart TB
    subgraph 历史数据加载
        UserSelect["用户选择时间范围<br/>或进入交易页面"]
        HTTPReq["HTTP 请求<br/>GET /api/products/{id}/candles<br/>?granularity=3600"]
        HTTPResp["服务端返回<br/>K 线历史数组"]
        StoreHistory["TradeStore<br/>objects[id].history = data"]
    end

    subgraph 图表初始渲染
        CandleRender["CandleChart<br/>Highcharts stockChart 渲染"]
        DepthRender["DepthChart<br/>Highcharts area 渲染"]
        TVRender["TradingView<br/>getBars() 加载渲染"]
    end

    subgraph 实时数据更新
        WSMsg["WebSocket 消息<br/>candles / match / l2update"]
        Buffer["SocketMsgBuffer<br/>200ms 缓冲"]
        StoreUpdate["TradeStore<br/>updateHistory()<br/>updateOrderBook()"]
    end

    subgraph 图表实时更新
        CandleUpdate["CandleChart<br/>更新最后一根 K 线"]
        DepthUpdate["DepthChart<br/>重新计算累计深度"]
        TVUpdate["TradingView<br/>DataPulseUpdater onTick()"]
    end

    UserSelect --> HTTPReq --> HTTPResp --> StoreHistory
    StoreHistory --> CandleRender
    StoreHistory --> DepthRender
    StoreHistory --> TVRender

    WSMsg --> Buffer --> StoreUpdate
    StoreUpdate -->|Vue 响应式| CandleUpdate
    StoreUpdate -->|1s 定时器| DepthUpdate
    StoreUpdate -->|10s 轮询| TVUpdate
```

### 5.1 各图表的更新机制对比

| 图表 | 更新触发 | 更新延迟 | 更新方式 |
|------|----------|----------|----------|
| CandleChart | Vue 响应式 (watch) | ~200ms (Buffer 延迟) | 直接操作 Highcharts API 更新数据点 |
| DepthChart | 1 秒定时器轮询 | ~1000ms | 重新计算并替换整个数据集 |
| TradingView | 10 秒定时器轮询 | ~10000ms | 通过 `onTick()` 回调推送新数据 |

### 5.2 更新策略选择原因

- **CandleChart (200ms)**: K 线图需要较快反映价格变化，Vue 响应式提供最低延迟
- **DepthChart (1s)**: 深度图计算量较大（200 档累计），适当降低频率平衡性能
- **TradingView (10s)**: TradingView 库自身有渲染优化，较低频率的推送已经足够，且避免与库内部状态冲突

## 6. 图表组件切换器 (TradeViewChartComponent)

源文件: `src/script/component/chart/trade-view/trade-view.ts`

TradeViewChartComponent 作为图表区域的容器，管理三种图表之间的切换：

```mermaid
stateDiagram-v2
    [*] --> K线图: 默认显示
    K线图 --> 深度图: 点击"深度图"标签
    K线图 --> TradingView: 点击"TradingView"标签
    深度图 --> K线图: 点击"K线图"标签
    深度图 --> TradingView: 点击"TradingView"标签
    TradingView --> K线图: 点击"K线图"标签
    TradingView --> 深度图: 点击"深度图"标签
```

**切换行为：**
- 非活跃图表通过 CSS `display:none` 隐藏（不销毁），保留内部状态
- 切换到 CandleChart 时，如果时间范围已更改，触发数据重新加载
- 切换到 DepthChart 时，立即执行一次深度计算更新
- 切换到 TradingView 时，触发 `chart.resize()` 适配容器尺寸

## 7. ChartSliderComponent (时间范围选择器)

源文件: `src/script/component/chart/slider/slider.ts`

时间范围选择器与 K 线图联动，允许用户在 7 个时间粒度之间切换：

```
 [1m] [5m] [15m] [30m] [1h] [6h] [1d]   [蜡烛图/折线图切换]
```

**交互流程：**

1. 用户点击时间范围按钮（如 "1h"）
2. SliderComponent 发出 `change` 事件，携带 granularity 值 (3600)
3. CandleChartComponent 接收事件
4. 调用 `StoreService.Trade.loadProductHistory(productId, 3600)`
5. 取消旧粒度的 WebSocket 频道订阅 (如 `candles_60`)
6. 订阅新粒度的 WebSocket 频道 (`candles_3600`)
7. 历史数据加载完成后重新渲染图表

## 8. 核心源文件索引

| 文件路径 | 职责 |
|----------|------|
| `src/script/component/chart/candle/candle.ts` | K 线图组件，Highcharts stockChart |
| `src/script/component/chart/depth/depth.ts` | 深度图组件，Highcharts area chart |
| `src/script/component/chart/trade-view/trade-view.ts` | 图表切换器容器 |
| `src/script/component/chart/tradingview/tradingview.ts` | TradingView 图表组件 |
| `src/script/component/chart/slider/slider.ts` | 时间范围选择器 |
| `src/script/chart/datafeed.ts` | TradingView UDF DataFeed 实现 |
| `src/script/chart/config.ts` | TradingView 图表配置 |
| `src/script/helper.ts` | 工具函数（深度聚合计算等）|
| `src/script/store/trade.ts` | K 线和订单簿数据存储 |
| `src/script/constant.ts` | 聚合精度常量、时间范围常量 |
| `src/style/component/chart.less` | 图表组件样式 |
