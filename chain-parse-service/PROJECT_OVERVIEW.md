# Chain Parse Service 项目说明文档

## 1. 项目概述

Chain Parse Service 是一个企业级多链 DEX（去中心化交易所）数据解析服务，主要功能包括：

- **多链支持**：实时监控以太坊、BSC、Solana、Sui 等多条区块链网络
- **DEX 事件提取**：从链上交易中提取 Swap、流动性变化、池子创建等事件
- **数据转换**：将原始区块链数据转换为结构化信息
- **REST API**：提供数据查询和分析接口
- **多协议支持**：支持 AMM（PancakeSwap、Uniswap）、Bonding Curve（PumpFun、FourMeme）、Move-based DEX（Bluefin）等协议

### 1.1 技术栈

| 分类      | 技术                                 | 版本     |
| ------- | ---------------------------------- | ------ |
| 语言      | Go                                 | 1.21+  |
| Web 框架  | Gin                                | v1.9.1 |
| 日志      | Logrus                             | v1.9.0 |
| 数据库     | PostgreSQL / MySQL / InfluxDB      | -      |
| 缓存      | Redis                              | v7     |
| 区块链 SDK | go-ethereum, solana-go, sui-go-sdk | -      |

## 2. 系统架构

### 2.1 整体架构图

```mermaid
graph TB
    subgraph blockchain["区块链网络"]
        ETH["以太坊"]
        BSC["BSC"]
        SOL["Solana"]
        SUI["Sui"]
    end

    subgraph service["Chain Parse Service"]
        subgraph parser["Parser 服务"]
            ENGINE["解析引擎"]
            CP1["ETH 处理器"]
            CP2["BSC 处理器"]
            CP3["Solana 处理器"]
            CP4["Sui 处理器"]
            DE1["Uniswap 提取器"]
            DE2["PancakeSwap 提取器"]
            DE3["PumpFun 提取器"]
            DE4["Bluefin 提取器"]
        end

        subgraph api["API 服务"]
            ROUTER["路由层"]
            CTRL["控制器层"]
            SVC["服务层"]
        end

        subgraph storage["存储层"]
            PG[("PostgreSQL")]
            MYSQL[("MySQL")]
            INFLUX[("InfluxDB")]
            REDIS[("Redis")]
        end
    end

    CLIENT["客户端"]

    ETH --> CP1
    BSC --> CP2
    SOL --> CP3
    SUI --> CP4

    CP1 --> ENGINE
    CP2 --> ENGINE
    CP3 --> ENGINE
    CP4 --> ENGINE

    ENGINE --> DE1
    ENGINE --> DE2
    ENGINE --> DE3
    ENGINE --> DE4

    ENGINE ==> PG
    ENGINE ==> MYSQL
    ENGINE ==> INFLUX
    ENGINE ==> REDIS

    CLIENT --> ROUTER
    ROUTER --> CTRL
    CTRL --> SVC
    SVC -.-> PG
    SVC -.-> MYSQL
    SVC -.-> REDIS

    style blockchain fill:#e7f5ff,stroke:#1971c2,color:#000
    style service fill:#f8f9fa,stroke:#868e96,color:#000
    style parser fill:#e5dbff,stroke:#5f3dc4,color:#000
    style api fill:#c5f6fa,stroke:#0c8599,color:#000
    style storage fill:#fff4e6,stroke:#e67700,color:#000
    style ENGINE fill:#ffe8cc,stroke:#d9480f,color:#000
    style CLIENT fill:#d3f9d8,stroke:#2f9e44,color:#000
    style PG fill:#fff4e6,stroke:#e67700,color:#000
    style MYSQL fill:#fff4e6,stroke:#e67700,color:#000
    style INFLUX fill:#fff4e6,stroke:#e67700,color:#000
    style REDIS fill:#fff4e6,stroke:#e67700,color:#000
```

### 2.2 数据流程图

```mermaid
sequenceDiagram
    participant BC as 区块链
    participant CP as 链处理器
    participant ENG as 解析引擎
    participant DEX as DEX 提取器
    participant SE as 存储引擎
    participant DB as 数据库
    participant Redis as Redis

    BC->>CP: 1. 获取最新区块
    CP->>CP: 2. 转换为 UnifiedBlock
    CP->>ENG: 3. 发送区块数据
    ENG->>DEX: 4. 分发到 DEX 提取器
    DEX->>DEX: 5. 解析 DEX 事件
    DEX->>ENG: 6. 返回 DexData

    ENG->>SE: 7. 存储区块数据
    SE->>DB: 8. 写入数据库

    ENG->>SE: 9. 存储 DEX 数据
    SE->>DB: 10. 写入 DEX 表

    ENG->>Redis: 11. 更新处理进度
    Redis--xENG: 12. 确认更新

    ENG->>ENG: 13. 继续下一批次
```

### 2.3 组件交互图

```mermaid
graph LR
    subgraph "核心接口"
        I1[ChainProcessor 接口]
        I2[DexExtractors 接口]
        I3[StorageEngine 接口]
        I4[ProgressTracker 接口]
    end

    subgraph "实现类"
        C1[ETH 处理器]
        C2[BSC 处理器]
        C3[Solana 处理器]
        C4[Sui 处理器]
        E1[Uniswap 提取器]
        E2[PancakeSwap 提取器]
        S1[PostgreSQL 存储]
        S2[MySQL 存储]
    end

    C1 -.->|实现| I1
    C2 -.->|实现| I1
    C3 -.->|实现| I1
    C4 -.->|实现| I1

    E1 -.->|实现| I2
    E2 -.->|实现| I2

    S1 -.->|实现| I3
    S2 -.->|实现| I3
```

### 2.4 DEX Extractor 的组合关系图
```mermaid
classDiagram
    direction TB

    class BaseDexExtractor {
        -protocols []string
        -supportedChains []ChainType
        -quoteAssets map[string]int
        +GetSupportedChains() []ChainType
        +GetSupportedProtocols() []string
        +IsChainSupported(ChainType) bool
    }

    class EVMDexExtractor {
        +ExtractEVMLogs(tx) []*Log
        +FilterLogsByTopics(logs, filter) []*Log
        +IsEVMChainSupported(ChainType) bool
    }

    class SolanaDexExtractor {
        +ExtractDiscriminator(data) []byte
        +MatchDiscriminator(actual, expected) bool
        +IsSolanaChainSupported(ChainType) bool
    }

    class PancakeSwapExtractor {
        支持链: BSC
        协议: pancakeswap-v2/v3
    }

    class UniswapExtractor {
        支持链: Ethereum
        协议: uniswap-v2/v3
    }

    class FourMemeExtractor {
        支持链: BSC
        协议: fourmeme
    }

    class PumpFunExtractor {
        支持链: Solana
        协议: pumpfun
    }

    class PumpSwapExtractor {
        支持链: Solana
        协议: pumpswap
    }

    class CetusExtractor {
        支持链: Sui
        协议: cetus
        -client *sui.SuiProcessor
    }

    class BluefinExtractor {
        支持链: Sui
        协议: bluefin
        -client *sui.SuiProcessor
    }

    BaseDexExtractor <|-- EVMDexExtractor : 嵌入
    BaseDexExtractor <|-- SolanaDexExtractor : 嵌入

    EVMDexExtractor <|-- PancakeSwapExtractor : 嵌入
    EVMDexExtractor <|-- UniswapExtractor : 嵌入
    EVMDexExtractor <|-- FourMemeExtractor : 嵌入

    SolanaDexExtractor <|-- PumpFunExtractor : 嵌入
    SolanaDexExtractor <|-- PumpSwapExtractor : 嵌入

    CetusExtractor ..> SuiProcessorInjectable : 实现
    BluefinExtractor ..> SuiProcessorInjectable : 实现
```

### 2.5 Processor 的组合关系图
```mermaid
classDiagram
    direction TB

    class Processor {
        +ChainType ChainType
        +RPCEndpoint string
        +BatchSize int
        +Log *logrus.Entry
        +Retry RetryConfig
        +GetChainType() ChainType
    }

    class EthereumProcessor {
        -client *ethclient.Client
        -chainID *big.Int
        -config *EthereumConfig
        +FetchBlocks(ctx, start, end) []UnifiedBlock
        +GetLatestBlockNumber(ctx) int64
    }

    class BSCProcessor {
        -client *ethclient.Client
        -chainID *big.Int
        -config *BSCConfig
        +FetchBlocks(ctx, start, end) []UnifiedBlock
        +GetLatestBlockNumber(ctx) int64
    }

    class SolanaProcessor {
        -client *rpc.Client
        -chainID string
        -config *SolanaConfig
        +FetchBlocks(ctx, start, end) []UnifiedBlock
        +GetLatestBlockNumber(ctx) int64
    }

    class SuiProcessor {
        -client sui.ISuiAPI
        -chainID string
        -config *SuiConfig
        +FetchBlocks(ctx, start, end) []UnifiedBlock
        +GetLatestBlockNumber(ctx) int64
        +GetCoinMetadata(coinType) CoinMetadata
        +GetObject(objectID) SuiObjectResponse
    }

    class EthereumConfig {
        +RPCEndpoint string
        +ChainID int64
        +BatchSize int
        +IsTestnet bool
    }

    class BSCConfig {
        +RPCEndpoint string
        +ChainID int64
        +BatchSize int
    }

    class SolanaConfig {
        +RPCEndpoint string
        +ChainID string
        +BatchSize int
        +IsTestnet bool
    }

    class SuiConfig {
        +RPCEndpoint string
        +ChainID string
        +BatchSize int
    }

    Processor <|-- EthereumProcessor : 嵌入
    Processor <|-- BSCProcessor : 嵌入
    Processor <|-- SolanaProcessor : 嵌入
    Processor <|-- SuiProcessor : 嵌入

    EthereumProcessor --> EthereumConfig : 持有
    BSCProcessor --> BSCConfig : 持有
    SolanaProcessor --> SolanaConfig : 持有
    SuiProcessor --> SuiConfig : 持有

    EthereumProcessor ..|> ChainProcessor : 实现
    BSCProcessor ..|> ChainProcessor : 实现
    SolanaProcessor ..|> ChainProcessor : 实现
    SuiProcessor ..|> ChainProcessor : 实现
```



### 2.6 存储层结构体组合关系图

```mermaid
classDiagram
    direction TB

    class StorageEngine {
        <<interface>>
        +StoreDexData(ctx, dexData) error
        +GetStorageStats(ctx) map
        +HealthCheck(ctx) error
        +Close() error
    }

    class MySQLStore {
        -db *sql.DB
        -config *MySQLConfig
    }

    class PgSQLStore {
        -db *sql.DB
        -config *PgSQLConfig
    }

    class SimpleInfluxDBStorage {
        -config *InfluxDBConfig
        -client influxdb2.Client
        -writeAPI api.WriteAPI
        -queryAPI api.QueryAPI
        -ctx context.Context
        -batchCache *DexDataBatch
        -batchMutex sync.RWMutex
    }

    class InfluxDBConfig {
        +URL string
        +Token string
        +Org string
        +Bucket string
        +BatchSize int
        +FlushTime int
        +Precision string
    }

    class DexDataBatch {
        +Pools []Pool
        +Tokens []Token
        +Reserves []Reserve
        +Transactions []Transaction
        +Liquidities []Liquidity
        +TotalCount int
        +LastUpdated time.Time
    }

    MySQLStore ..|> StorageEngine : 实现
    PgSQLStore ..|> StorageEngine : 实现
    SimpleInfluxDBStorage ..|> StorageEngine : 实现

    SimpleInfluxDBStorage --> InfluxDBConfig : 持有
    SimpleInfluxDBStorage --> DexDataBatch : 持有

```





### 2.7 项目核心结构体组合关系图

```mermaid
classDiagram
    direction TB

    %% ========== 应用入口层 ==========
    class application {
        +engine *Engine
        +redisClient *redis.Client
    }

    %% ========== 引擎层 ==========
    class Engine {
        -chainProcessors map~ChainType~ChainProcessor~~
        -dexExtractors []DexExtractors
        -storage StorageEngine
        -progressTracker ProgressTracker
        -config *EngineConfig
        -running bool
        -mu sync.RWMutex
        -ctx context.Context
        -cancel context.CancelFunc
        +Start() error
        +Stop()
        +RegisterChainProcessor(processor)
        +RegisterDexExtractor(extractor)
    }

    class EngineConfig {
        +BatchSize int
        +ProcessInterval time.Duration
        +MaxRetries int
        +ConcurrentChains int
        +RealTimeMode bool
        +ChainConfigs map~ChainType~ChainConfig~~
    }

    %% ========== 接口层 ==========
    class ChainProcessor {
        <<interface>>
        +GetChainType() ChainType
        +GetLatestBlockNumber(ctx) *big.Int
        +GetBlocksByRange(ctx, start, end) []UnifiedBlock
        +GetBlock(ctx, blockNumber) UnifiedBlock
        +HealthCheck(ctx) error
    }

    class DexExtractors {
        <<interface>>
        +GetSupportedProtocols() []string
        +GetSupportedChains() []ChainType
        +ExtractDexData(ctx, blocks) *DexData
        +SupportsBlock(block) bool
    }

    class StorageEngine {
        <<interface>>
        +StoreDexData(ctx, dexData) error
        +GetStorageStats(ctx) map
        +HealthCheck(ctx) error
        +Close() error
    }

    class ProgressTracker {
        <<interface>>
        +GetProgress(chainType) ProcessProgress
        +UpdateProgress(chainType, progress) error
        +SetProcessingStatus(chainType, status) error
        +RecordError(chainType, err) error
        +HealthCheck() error
    }

    class SuiProcessorInjectable {
        <<interface>>
        +SetSuiProcessor(processor)
    }

    %% ========== 链处理器基类 ==========
    class Processor {
        +ChainType ChainType
        +RPCEndpoint string
        +BatchSize int
        +Log *logrus.Entry
        +Retry RetryConfig
    }

    class BSCProcessor {
        -client *ethclient.Client
        -chainID *big.Int
        -config *BSCConfig
    }

    class EthereumProcessor {
        -client *ethclient.Client
        -chainID *big.Int
        -config *EthereumConfig
    }

    class SolanaProcessor {
        -client *rpc.Client
        -chainID string
        -config *SolanaConfig
    }

    class SuiProcessor {
        -client sui.ISuiAPI
        -chainID string
        -config *SuiConfig
    }

    %% ========== DEX 提取器基类 ==========
    class BaseDexExtractor {
        -protocols []string
        -supportedChains []ChainType
        -quoteAssets map~string~int~
        -log *logrus.Entry
        -cacheMutex sync.RWMutex
    }

    class EVMDexExtractor {
        +ExtractEVMLogs(tx) []*Log
        +FilterLogsByTopics(logs, filter) []*Log
    }

    class SolanaDexExtractor {
        +ExtractDiscriminator(data) []byte
        +MatchDiscriminator(actual, expected) bool
    }

    %% ========== DEX 提取器 - EVM ==========
    class PancakeSwapExtractor {
        支持链: BSC
        协议: pancakeswap-v2/v3
    }

    class UniswapExtractor {
        支持链: Ethereum
        协议: uniswap-v2/v3
    }

    class FourMemeExtractor {
        支持链: BSC
        协议: fourmeme
    }

    %% ========== DEX 提取器 - Solana ==========
    class PumpFunExtractor {
        支持链: Solana
        协议: pumpfun
    }

    class PumpSwapExtractor {
        支持链: Solana
        协议: pumpswap
    }

    %% ========== DEX 提取器 - Sui ==========
    class CetusExtractor {
        支持链: Sui
        协议: cetus
        -client *sui.SuiProcessor
        -tokenCache map
        -config *CetusConfig
    }

    class BluefinExtractor {
        支持链: Sui
        协议: bluefin
        -client *sui.SuiProcessor
        -tokenCache map
        -poolCache map
    }

    %% ========== DEX 事件提取器（组合入口） ==========
    class DEXExtractor {
        -supportedChains []ChainType
        -factory *ExtractorFactory
        +GetSupportedProtocols() []string
        +GetSupportedChains() []ChainType
        +ExtractDexData(ctx, blocks) *DexData
        +SupportsBlock(block) bool
    }

    %% ========== 提取器工厂 ==========
    class ExtractorFactory {
        -extractors map~string~DexExtractors~
        +RegisterExtractor(name, extractor)
        +GetAllExtractors() []DexExtractors
    }

    %% ========== 存储层 ==========
    class MySQLStore {
        -db *sql.DB
        -config *MySQLConfig
    }

    class PgSQLStore {
        -db *sql.DB
        -config *PgSQLConfig
    }

    class SimpleInfluxDBStorage {
        -config *InfluxDBConfig
        -client influxdb2.Client
        -writeAPI api.WriteAPI
        -queryAPI api.QueryAPI
        -ctx context.Context
        -batchCache *DexDataBatch
        -batchMutex sync.RWMutex
    }

    class RedisProgressTracker {
        -client *redis.Client
        -keyPrefix string
        -maxErrorHistory int
    }

    class MemoryProgressTracker {
        -progresses map~ChainType~ProcessProgress~
        -stats map~ChainType~ProcessingStats~
        -errors map~ChainType~~ProcessingError~
    }

    %% ========== 数据模型 ==========
    class DexData {
        +Pools []Pool
        +Transactions []Transaction
        +Liquidities []Liquidity
        +Reserves []Reserve
        +Tokens []Token
    }

    %% ========== 组合关系 ==========

    application *-- Engine : 持有
    application *-- RedisClient : 持有

    Engine o-- ChainProcessor : 持有多个
    Engine o-- DexExtractors : 持有多个
    Engine o-- StorageEngine : 持有
    Engine o-- ProgressTracker : 持有
    Engine *-- EngineConfig : 持有

    Processor <|-- BSCProcessor : 嵌入
    Processor <|-- EthereumProcessor : 嵌入
    Processor <|-- SolanaProcessor : 嵌入
    Processor <|-- SuiProcessor : 嵌入

    BSCProcessor ..|> ChainProcessor : 实现
    EthereumProcessor ..|> ChainProcessor : 实现
    SolanaProcessor ..|> ChainProcessor : 实现
    SuiProcessor ..|> ChainProcessor : 实现

    BaseDexExtractor <|-- EVMDexExtractor : 嵌入
    BaseDexExtractor <|-- SolanaDexExtractor : 嵌入

    EVMDexExtractor <|-- PancakeSwapExtractor : 嵌入
    EVMDexExtractor <|-- UniswapExtractor : 嵌入
    EVMDexExtractor <|-- FourMemeExtractor : 嵌入

    SolanaDexExtractor <|-- PumpFunExtractor : 嵌入
    SolanaDexExtractor <|-- PumpSwapExtractor : 嵌入

    CetusExtractor ..|> SuiProcessorInjectable : 实现
    BluefinExtractor ..|> SuiProcessorInjectable : 实现

    PancakeSwapExtractor ..|> DexExtractors : 实现
    UniswapExtractor ..|> DexExtractors : 实现
    FourMemeExtractor ..|> DexExtractors : 实现
    PumpFunExtractor ..|> DexExtractors : 实现
    PumpSwapExtractor ..|> DexExtractors : 实现
    CetusExtractor ..|> DexExtractors : 实现
    BluefinExtractor ..|> DexExtractors : 实现

    ExtractorFactory o-- DexExtractors : 持有多个

    DEXExtractor *-- ExtractorFactory : 持有
    DEXExtractor ..|> DexExtractors : 实现

    MySQLStore ..|> StorageEngine : 实现
    PgSQLStore ..|> StorageEngine : 实现
    SimpleInfluxDBStorage ..|> StorageEngine : 实现

    RedisProgressTracker ..|> ProgressTracker : 实现
    MemoryProgressTracker ..|> ProgressTracker : 实现

    DexData *-- Pool : 包含
    DexData *-- Transaction : 包含
    DexData *-- Liquidity : 包含
    DexData *-- Reserve : 包含
    DexData *-- Token : 包含
```

## 3. 目录结构

```
chain-parse-service/
├── cmd/                        # 应用程序入口
│   ├── parser/                 # 解析服务入口
│   │   └── main.go            # Parser 服务启动文件
│   └── api/                    # API 服务入口
│       └── main.go            # API 服务启动文件
│
├── configs/                    # 配置文件
│   ├── base.yaml              # 共享基础配置
│   ├── api.yaml               # API 服务配置
│   ├── bsc.yaml               # BSC 链配置
│   ├── ethereum.yaml          # 以太坊链配置
│   ├── solana.yaml            # Solana 链配置
│   └── sui.yaml               # Sui 链配置
│
├── database/                   # 数据库脚本
│   ├── mysql/                 # MySQL 建表脚本
│   │   └── schema.sql
│   └── pgsql/                 # PostgreSQL 建表脚本
│       └── schema.sql
│
├── docker/                     # Docker 部署
│   ├── docker-compose.yml     # 服务编排
│   ├── Dockerfile             # 镜像构建
│   └── .env.example           # 环境变量模板
│
├── internal/                   # 内部代码
│   ├── api/                   # API 层
│   │   ├── controller/        # 请求处理器
│   │   ├── middleware/        # HTTP 中间件
│   │   ├── router/            # 路由定义
│   │   └── service/           # 业务逻辑
│   │
│   ├── config/                # 配置管理
│   │   └── config.go
│   │
│   ├── model/                 # 数据模型
│   │   ├── token.go           # Token 模型
│   │   ├── pool.go            # Pool 模型
│   │   ├── reserve.go         # Reserve 模型
│   │   ├── liquidity.go       # Liquidity 模型
│   │   └── transaction.go     # Transaction 模型
│   │
│   ├── parser/                # 解析引擎
│   │   ├── chains/            # 链处理器
│   │   │   ├── base/          # 基础处理器
│   │   │   ├── bsc/           # BSC 处理器
│   │   │   ├── ethereum/      # 以太坊处理器
│   │   │   ├── solana/        # Solana 处理器
│   │   │   └── sui/           # Sui 处理器
│   │   │
│   │   ├── dexs/              # DEX 提取器
│   │   │   ├── bsc/           # BSC DEX
│   │   │   │   ├── pancakeswap/
│   │   │   │   └── fourmeme/
│   │   │   ├── eth/           # 以太坊 DEX
│   │   │   │   └── uniswap/
│   │   │   ├── solanadex/     # Solana DEX
│   │   │   │   ├── pumpfun/
│   │   │   │   └── pumpswap/
│   │   │   └── suidex/        # Sui DEX
│   │   │       └── bluefin/
│   │   │
│   │   └── engine/            # 解析引擎
│   │       └── engine.go
│   │
│   ├── storage/               # 存储层
│   │   ├── mysql/             # MySQL 实现
│   │   ├── pgsql/             # PostgreSQL 实现
│   │   └── influxdb/          # InfluxDB 实现
│   │
│   ├── types/                 # 类型定义
│   │   └── interfaces.go      # 核心接口
│   │
│   └── app/                   # 应用初始化
│
├── docs/                      # 文档
├── Makefile                   # 构建自动化
├── go.mod                     # Go 模块依赖
├── go.sum                     # 依赖校验和
└── README.md                  # 项目说明
```

## 4. 支持的链和协议

### 4.1 支持的区块链

```mermaid
graph TD
    ROOT[Chain Parse Service]

    ROOT --> EVM[EVM 兼容链]
    ROOT --> SOL[Solana]
    ROOT --> SUI[Sui]

    EVM --> ETH[以太坊]
    EVM --> BSC[BSC]

    ETH --> UNI[Uniswap V2/V3]
    BSC --> PAN[PancakeSwap V2/V3]
    BSC --> FOUR[FourMeme V1/V2]

    SOL --> PUMP[PumpFun]
    SOL --> PS[PumpSwap]

    SUI --> BLUE[Bluefin]
    SUI --> CET[Cetus]
```

### 4.2 协议支持详情

| 链 | 协议 | 类型 | 状态 |
|---|------|------|------|
| BSC | PancakeSwap V2/V3 | AMM | ✅ |
| BSC | FourMeme V1/V2 | Bonding Curve | ✅ |
| Ethereum | Uniswap V2/V3 | AMM | ✅ |
| Solana | PumpFun | Bonding Curve | ✅ |
| Solana | PumpSwap | Bonding Curve | ✅ |
| Sui | Bluefin | Move DEX | ✅ |
| Sui | Cetus | Move DEX | ✅ |

## 5. 数据模型

### 5.1 核心数据模型

```mermaid
classDiagram
    class UnifiedBlock {
        +BigInt BlockNumber
        +string BlockHash
        +ChainType ChainType
        +string ChainID
        +string ParentHash
        +Time Timestamp
        +BigInt GasLimit
        +BigInt GasUsed
        +int TxCount
        +List~Transaction~ Transactions
        +List~Event~ Events
    }

    class UnifiedTransaction {
        +string TxHash
        +ChainType ChainType
        +string ChainID
        +BigInt BlockNumber
        +string FromAddress
        +string ToAddress
        +BigInt Value
        +BigInt GasLimit
        +BigInt GasUsed
        +TxStatus Status
    }

    class UnifiedEvent {
        +string EventID
        +ChainType ChainType
        +BigInt BlockNumber
        +string TxHash
        +int EventIndex
        +string EventType
        +string Address
        +List~string~ Topics
        +Any Data
    }

    class Token {
        +string Addr
        +string Name
        +string Symbol
        +int Decimals
        +bool IsStable
        +float64 UsdPrice
    }

    class Pool {
        +string Addr
        +string Factory
        +string Protocol
        +Map Tokens
        +int Fee
        +PoolExtra Extra
    }

    class Transaction {
        +string Addr
        +string Router
        +string Factory
        +string Pool
        +string Hash
        +string From
        +string Side
        +BigInt Amount
        +float64 Price
        +float64 Value
        +uint64 Time
    }

    class Liquidity {
        +string Addr
        +string Router
        +string Factory
        +string Pool
        +string Hash
        +string Side
        +BigInt Amount
        +float64 Value
        +uint64 Time
    }

    class Reserve {
        +string Addr
        +Map Amounts
        +uint64 Time
    }

    UnifiedBlock "1" --> "*" UnifiedTransaction : contains
    UnifiedBlock "1" --> "*" UnifiedEvent : contains
    Pool "1" --> "*" Token : contains
    Pool "1" --> "*" Reserve : tracks
    Pool "1" --> "*" Transaction : records
    Pool "1" --> "*" Liquidity : records
```

### 5.2 数据库表结构

```mermaid
erDiagram
    blocks {
        big_int block_number PK
        string block_hash
        string chain_type
        string chain_id
        string parent_hash
        timestamp timestamp
        int tx_count
    }

    transactions {
        string tx_hash PK
        string chain_type
        string chain_id
        big_int block_number FK
        string from_address
        string to_address
        big_int value
        string status
    }

    dex_pools {
        string addr PK
        string factory
        string protocol
        string token0
        string token1
        int fee
        json extra
    }

    dex_tokens {
        string addr PK
        string name
        string symbol
        int decimals
        bool is_stable
    }

    dex_transactions {
        big_int id PK
        string addr
        string pool FK
        string hash
        string from_addr
        string side
        big_int amount
        float64 price
        float64 value
        uint64 time
        json extra
    }

    dex_liquidities {
        big_int id PK
        string addr
        string pool FK
        string hash
        string from_addr
        string side
        big_int amount
        float64 value
        uint64 time
        json extra
    }

    dex_reserves {
        big_int id PK
        string addr FK
        string amount0
        string amount1
        uint64 time
    }

    processing_progress {
        string chain_type PK
        big_int last_processed_block
        big_int total_transactions
        big_int total_events
    }

    blocks ||--o{ transactions : contains
    dex_pools ||--o{ dex_transactions : has
    dex_pools ||--o{ dex_liquidities : has
    dex_pools ||--o{ dex_reserves : tracks
    dex_pools }o--|| dex_tokens : contains
```

## 6. 核心接口

### 6.1 ChainProcessor 接口

区块链处理器接口，所有链实现必须遵循：

```go
type ChainProcessor interface {
    GetChainType() ChainType
    GetChainID() string
    GetLatestBlockNumber(ctx context.Context) (*big.Int, error)
    GetBlocksByRange(ctx context.Context, startBlock, endBlock *big.Int) ([]UnifiedBlock, error)
    GetBlock(ctx context.Context, blockNumber *big.Int) (*UnifiedBlock, error)
    GetTransaction(ctx context.Context, txHash string) (*UnifiedTransaction, error)
    HealthCheck(ctx context.Context) error
}
```

### 6.2 DexExtractors 接口

DEX 事件提取器接口：

```go
type DexExtractors interface {
    GetSupportedProtocols() []string
    GetSupportedChains() []ChainType
    ExtractDexData(ctx context.Context, blocks []UnifiedBlock) (*DexData, error)
    SupportsBlock(block *UnifiedBlock) bool
}
```

### 6.3 StorageEngine 接口

存储引擎接口：

```go
type StorageEngine interface {
    StoreBlocks(ctx context.Context, blocks []UnifiedBlock) error
    StoreTransactions(ctx context.Context, txs []UnifiedTransaction) error
    StoreDexData(ctx context.Context, dexData *DexData) error
    GetTransactionsByHash(ctx context.Context, hashes []string) ([]UnifiedTransaction, error)
    GetStorageStats(ctx context.Context) (map[string]interface{}, error)
    HealthCheck(ctx context.Context) error
    Close() error
}
```

### 6.4 ProgressTracker 接口

进度追踪接口：

```go
type ProgressTracker interface {
    GetProgress(chainType ChainType) (*ProcessProgress, error)
    UpdateProgress(chainType ChainType, progress *ProcessProgress) error
    GetProcessingStats(chainType ChainType) (*ProcessingStats, error)
    GetGlobalStats() (*GlobalProcessingStats, error)
    SetProcessingStatus(chainType ChainType, status ProcessingStatus) error
    RecordError(chainType ChainType, err error) error
    HealthCheck() error
}
```

## 7. API 接口

### 7.1 API 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /health | 健康检查 |
| GET | /api/v1/transactions/:hash | 根据哈希获取交易 |
| GET | /api/v1/storage/stats | 存储统计信息 |
| GET | /api/v1/progress | 链处理进度 |
| GET | /api/v1/progress/stats | 全局处理统计 |

### 7.2 API 架构图

```mermaid
graph TD
    CLIENT[客户端] -->|HTTP| ROUTER[Gin 路由]

    ROUTER -->|日志| M1[Logger 中间件]
    ROUTER -->|CORS| M2[CORS 中间件]
    ROUTER -->|恢复| M3[Recovery 中间件]

    M1 --> CTRL[控制器层]
    M2 --> CTRL
    M3 --> CTRL

    CTRL --> S1[Health 服务]
    CTRL --> S2[Transaction 服务]
    CTRL --> S3[Stats 服务]

    S1 --> SE[存储引擎]
    S2 --> SE
    S3 --> SE
    S3 --> Redis[Redis]

    SE --> PG[PostgreSQL]
    SE --> MYSQL[MySQL]
    SE --> INFLUX[InfluxDB]
```

## 8. 配置管理

### 8.1 配置层次结构

```mermaid
graph TD
    BASE[base.yaml<br/>基础配置] --> CHAIN[链配置<br/>bsc.yaml/eth.yaml等]
    CHAIN --> ENV[环境变量]

    BASE -->|存储| S1[PostgreSQL/MySQL/InfluxDB]
    BASE -->|Redis| S2[Redis 配置]
    BASE -->|日志| S3[日志配置]
    BASE -->|处理器| S4[处理器配置]

    CHAIN -->|RPC| C1[RPC 端点]
    CHAIN -->|链ID| C2[链 ID]
    CHAIN -->|协议| C3[协议配置]

    ENV -->|覆盖| O1[所有配置]
```

### 8.2 配置文件示例

**base.yaml（基础配置）**
```yaml
api:
  port: 8081
  read_timeout: 30
  write_timeout: 30

redis:
  host: "localhost"
  port: 6379
  db: 0

processor:
  batch_size: 10
  max_concurrent: 10
  retry_delay: 5
  max_retries: 3

storage:
  type: "pgsql"  # pgsql, mysql, influxdb
  pgsql:
    host: "localhost"
    port: 5432
    username: "postgres"
    password: "password"
    database: "unified_tx_parser"
```

**bsc.yaml（链配置）**
```yaml
chains:
  bsc:
    enabled: true
    rpc_endpoint: "https://bsc.publicnode.com"
    chain_id: "bsc-mainnet"
    batch_size: 10

protocols:
  pancakeswap:
    enabled: true
    chain: "bsc"
    contract_addresses:
      - "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
```

## 9. 部署架构

### 9.1 Docker 部署架构

```mermaid
graph TB
    subgraph "Docker Compose"
        subgraph "应用层"
            PARSER[Parser 容器]
            API[API 容器]
        end

        subgraph "基础设施"
            PG[PostgreSQL]
            MYSQL[MySQL]
            REDIS[Redis]
            INFLUX[InfluxDB]
            GRAFANA[Grafana]
        end
    end

    PARSER --> PG
    PARSER --> MYSQL
    PARSER --> INFLUX
    PARSER --> REDIS

    API --> PG
    API --> MYSQL
    API --> REDIS

    INFLUX --> GRAFANA
```

### 9.2 部署命令

```bash
# 构建镜像
make docker-build

# 启动基础设施
docker compose -f docker/docker-compose.yml --profile base up -d

# 启动完整栈
CHAIN_TYPE=bsc docker compose -f docker/docker-compose.yml --profile app up -d

# 查看日志
make docker-logs

# 停止服务
make docker-down
```

## 10. 设计模式

### 10.1 工厂模式

DEX 提取器使用工厂模式注册和管理：

```mermaid
classDiagram
    class ExtractorFactory {
        +RegisterExtractor(name, extractor)
        +GetExtractor(name) DexExtractors
        +GetExtractorsByChain(chain) []DexExtractors
        -map[string]DexExtractors extractors
    }

    class DexExtractors {
        <<interface>>
        +GetSupportedProtocols() []string
        +GetSupportedChains() []ChainType
        +ExtractDexData(blocks) DexData
    }

    class UniswapExtractor {
        +GetSupportedProtocols() []string
        +GetSupportedChains() []ChainType
        +ExtractDexData(blocks) DexData
    }

    class PancakeSwapExtractor {
        +GetSupportedProtocols() []string
        +GetSupportedChains() []ChainType
        +ExtractDexData(blocks) DexData
    }

    ExtractorFactory --> DexExtractors : manages
    DexExtractors <|.. UniswapExtractor : implements
    DexExtractors <|.. PancakeSwapExtractor : implements
```

### 10.2 策略模式

不同链使用不同的处理策略：

```mermaid
classDiagram
    class ChainProcessor {
        <<interface>>
        +GetChainType() ChainType
        +GetLatestBlockNumber() big.Int
        +GetBlocksByRange() []UnifiedBlock
    }

    class EVMProcessor {
        +GetChainType() ChainType
        +GetLatestBlockNumber() big.Int
        +GetBlocksByRange() []UnifiedBlock
        -ethClient *ethclient.Client
    }

    class SolanaProcessor {
        +GetChainType() ChainType
        +GetLatestBlockNumber() big.Int
        +GetBlocksByRange() []UnifiedBlock
        -rpcClient *rpc.Client
    }

    ChainProcessor <|.. EVMProcessor : implements
    ChainProcessor <|.. SolanaProcessor : implements
```

### 10.3 仓储模式

存储层使用仓储模式抽象数据访问：

```mermaid
classDiagram
    class StorageEngine {
        <<interface>>
        +StoreBlocks(blocks) error
        +StoreTransactions(txs) error
        +StoreDexData(dexData) error
    }

    class PostgreSQLStorage {
        +StoreBlocks(blocks) error
        +StoreTransactions(txs) error
        +StoreDexData(dexData) error
        -db *sql.DB
    }

    class MySQLStorage {
        +StoreBlocks(blocks) error
        +StoreTransactions(txs) error
        +StoreDexData(dexData) error
        -db *sql.DB
    }

    class InfluxDBStorage {
        +StoreBlocks(blocks) error
        +StoreTransactions(txs) error
        +StoreDexData(dexData) error
        -client *api.Client
    }

    StorageEngine <|.. PostgreSQLStorage : implements
    StorageEngine <|.. MySQLStorage : implements
    StorageEngine <|.. InfluxDBStorage : implements
```

## 11. 运行流程

### 11.1 Parser 服务启动流程

```mermaid
flowchart TD
    START[启动 Parser] --> LOAD[加载配置文件]
    LOAD --> CHECK{检查链类型}
    CHECK -->|指定链| CHAIN[加载链配置]
    CHECK -->|未指定| ERROR[错误退出]

    CHAIN --> LOG[设置日志]
    LOG --> STORAGE[初始化存储引擎]
    STORAGE --> REDIS[创建 Redis 客户端]

    REDIS --> REG1[注册链处理器]
    REG1 --> REG2[注册 DEX 提取器]
    REG2 --> ENGINE[启动解析引擎]

    ENGINE --> RUN[开始处理循环]
    RUN --> FETCH[获取最新区块]
    FETCH --> BATCH[计算批次范围]
    BATCH --> PROCESS[处理区块批次]
    PROCESS --> STORE[存储解析数据]
    STORE --> REDIS_UPDATE[更新进度]
    REDIS_UPDATE --> RUN
```

### 11.2 API 服务启动流程

```mermaid
flowchart TD
    START[启动 API 服务] --> LOAD[加载配置]
    LOAD --> LOG[设置日志]
    LOG --> STORAGE[初始化存储引擎]
    STORAGE --> REDIS[创建 Redis 客户端]
    REDIS --> GIN[配置 Gin 路由]
    GIN --> MIDL[注册中间件]
    MIDL --> CTRL[注册控制器]
    CTRL --> SERVER[启动 HTTP 服务]
    SERVER --> LISTEN[监听请求]
```

## 12. 监控与可观测性

### 12.1 监控架构

```mermaid
graph LR
    subgraph "应用层"
        PARSER[Parser]
        API[API]
    end

    subgraph "监控层"
        METRICS[指标采集]
        LOGS[日志收集]
        HEALTH[健康检查]
    end

    subgraph "存储层"
        INFLUX[InfluxDB]
        REDIS[Redis]
    end

    subgraph "展示层"
        GRAFANA[Grafana]
    end

    PARSER --> METRICS
    API --> METRICS
    PARSER --> LOGS
    API --> LOGS
    API --> HEALTH

    METRICS --> INFLUX
    LOGS --> INFLUX
    HEALTH --> REDIS

    INFLUX --> GRAFANA
    REDIS --> GRAFANA
```

### 12.2 健康检查端点

- **GET /health**: 服务健康状态
- **GET /api/v1/storage/stats**: 存储统计
- **GET /api/v1/progress**: 各链处理进度
- **GET /api/v1/progress/stats**: 全局处理统计

## 13. 快速开始

### 13.1 环境要求

- Go 1.21+
- Docker & Docker Compose
- Make

### 13.2 启动步骤

```bash
# 1. 启动基础设施
cd docker
docker compose up -d postgres redis

# 2. 启动 Parser（选择一条链）
make run-parser CHAIN=bsc

# 3. 启动 API 服务
make run-api

# 4. 测试 API
curl http://localhost:8081/health
```

## 14. 开发指南

### 14.1 添加新链支持

1. 在 `internal/parser/chains/` 下创建新目录
2. 实现 `ChainProcessor` 接口
3. 在 `configs/` 下添加链配置文件
4. 注册到解析引擎

### 14.2 添加新 DEX 协议

1. 在 `internal/parser/dexs/{chain}/` 下创建目录
2. 实现 `DexExtractors` 接口
3. 在配置文件中启用协议
4. 注册到提取器工厂

### 14.3 常用命令

```bash
make build-all       # 编译所有服务
make test            # 运行测试
make test-cover      # 测试覆盖率
make vet             # 静态检查
make fmt             # 格式化代码
make clean           # 清理构建产物
```

## 15. 许可证

本项目采用 MIT 许可证。
