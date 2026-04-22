# RWA Backend Repository Design

This document provides a detailed explanation of the overall architecture design of the RWA Backend repository, as well as the purpose and directory structure logic of each service under the `apps/` directory.

## Repository Overall Architecture

```
rwa-backend/
├── apps/                    # Application service layer
│   ├── alpaca-stream/      # Alpaca WebSocket streaming service
│   ├── api/                # RESTful API service
│   ├── indexer/            # On-chain event indexing service
│   └── ws-server/          # WebSocket server for real-time data streaming
├── libs/                    # Shared library layer
│   ├── contracts/          # Smart contract bindings
│   ├── core/               # Core business logic library
│   ├── database/           # Database client
│   ├── errors/             # Error handling library
│   ├── grpc/               # gRPC definitions
│   ├── kafka/              # Kafka client
│   ├── log/                # Logging library
│   └── oss/                # Object storage service
├── migrations/             # Database migration scripts
├── devops/                 # DevOps configuration
└── docs/                   # Documentation
```

## Service Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         RWA Backend System                       │
└─────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────┼─────────────┼─────────────┼─────────────┐
        │             │             │             │             │
        ▼             ▼             ▼             ▼             ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ alpaca-stream│ │     api      │ │   indexer    │ │  ws-server   │
│   Service    │ │   Service    │ │   Service    │ │   Service    │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
        │                     │                     │
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Alpaca WS   │    │   Database   │    │    Kafka     │
│   Stream     │    │  PostgreSQL  │    │   Consumer   │
└──────────────┘    └──────────────┘    └──────────────┘
        │                     │                     │
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Trade       │    │    Redis     │    │  Blockchain  │
│  Updates     │    │    Cache     │    │    Events    │
│  Market Data │    │              │    │              │
└──────────────┘    └──────────────┘    └──────────────┘
```

## Service Details

### 1. alpaca-stream Service

**Purpose:** Alpaca WebSocket streaming data service for real-time monitoring of order status updates and market data streams from the Alpaca trading platform.

**Core Features:**
- Real-time monitoring of order status updates (new, fill, partial_fill, canceled, expired, rejected, etc.)
- Real-time reception of market data streams (bars/candles, quotes, trades)
- Automatic reconnection mechanism (exponential backoff strategy)
- Structured logging of all events

**Directory Structure:**

```
apps/alpaca-stream/
├── main.go                    # Service entry point, CLI argument parsing and startup logic
├── config/                    # Configuration management
│   ├── config.go             # Configuration struct definitions
│   ├── config.yaml           # Configuration file (Alpaca API keys, WebSocket URLs, etc.)
│   └── fx.go                 # Uber FX dependency injection module
├── constants/                 # Constant definitions
│   └── constants.go          # Stream types, subscription configuration constants
├── ws/                        # WebSocket client implementation
│   ├── client.go             # WebSocket client core logic (connection, authentication, message handling)
│   └── subscription.go       # Subscription manager (manages different types of subscriptions)
├── handlers/                  # Message handlers
│   ├── trade_updates_handler.go  # Trade update message handler
│   └── bars_handler.go           # Bar/candle data handler
├── service/                   # Business logic layer
│   ├── alpaca_ws_service.go  # Main service logic (start/stop streams, event callback handling)
│   └── fx.go                 # Service layer dependency injection
├── types/                     # Type definitions
│   └── cli.go                # CLI argument types
└── logs/                      # Log file directory
```

**Data Flow:**

```
Alpaca WebSocket Server
    │
    │ (WebSocket Connection)
    ▼
ws/client.go (connection, authentication)
    │
    │ (message routing)
    ▼
handlers/ (message parsing and processing)
    │
    │ (event callbacks)
    ▼
service/alpaca_ws_service.go (business logic processing)
    │
    │ (logging)
    ▼
logs/app.log
```

**Key Components:**
- `ws/client.go`: Responsible for WebSocket connection management, automatic reconnection, message read/write
- `ws/subscription.go`: Manages different types of subscriptions (trade_updates, bars, quotes, trades)
- `handlers/`: Parses raw WebSocket messages into structured data
- `service/alpaca_ws_service.go`: Business logic layer that processes parsed data and executes corresponding operations

---

### 2. api Service

**Purpose:** RESTful API service that provides HTTP interfaces for the RWA system, including trading and stock information querying functionality.

**Core Features:**
- Provides RESTful API interfaces (trading, stock queries, etc.)
- API key authentication and signature verification
- Swagger documentation support
- Database operations (PostgreSQL)
- Redis cache support
- Integration with Alpaca trading API

**Directory Structure:**

```
apps/api/
├── main.go                    # Service entry point, initializes dependencies and starts HTTP server
├── config/                    # Configuration management
│   ├── config.go             # Configuration struct (server, database, Redis, Alpaca, etc.)
│   ├── config.yaml           # Configuration file
│   └── fx.go                 # Configuration module dependency injection
├── server/                    # HTTP server layer
│   ├── server.go             # Gin server startup and lifecycle management
│   ├── router.go             # Route definitions and middleware configuration
│   └── middleware/           # Middleware
│       ├── api_sign_middleware.go          # API signature verification middleware
│       └── market_maker_api_sign_middleware.go  # Market maker API signature verification
├── controller/                # Controller layer (handles HTTP requests)
│   ├── common_ctl.go         # Common controller (health checks, etc.)
│   ├── trade_ctl.go          # Trading-related interface controller
│   ├── stock_ctl.go          # Stock information query controller
│   └── fx.go                 # Controller dependency injection
├── service/                   # Business logic layer
│   ├── stock_service.go      # Stock-related business logic
│   └── fx.go                 # Service layer dependency injection
├── dto/                       # Data Transfer Objects
│   ├── common.go             # Common DTOs
│   ├── stock.go              # Stock-related DTOs
│   └── trade.go              # Trading-related DTOs
├── docs/                      # Swagger documentation
│   ├── docs.go               # Swagger annotation-generated code
│   ├── swagger.json          # Swagger JSON definition
│   └── swagger.yaml          # Swagger YAML definition
├── utils/                     # Utility functions
│   └── time.go               # Time handling utilities
├── types/                     # Type definitions
│   └── cli.go                # CLI argument types
├── migrations/                # Database migration scripts
└── logs/                      # Log file directory
```

**Request Processing Flow:**

```
HTTP Request
    │
    ▼
server/router.go (route matching)
    │
    ▼
server/middleware/ (middleware: authentication, signature verification, logging)
    │
    ▼
controller/ (controller: parameter validation, calls service layer)
    │
    ▼
service/ (business logic: database operations, external API calls)
    │
    ▼
libs/core/ (shared libraries: database client, Redis, Alpaca client)
    │
    ▼
Response (JSON)
```

**API Route Structure:**

```
/api/v1/
├── /common/
│   └── GET /health          # Health check
├── /trade/
│   ├── GET /currentPrice    # Get current price
│   ├── GET /latestQuote     # Get latest quote
│   ├── GET /snapshot        # Get snapshot data
│   ├── GET /historicalData  # Get historical data
│   ├── GET /marketClock     # Get market clock
│   ├── GET /assets          # Get asset list
│   └── GET /asset           # Get single asset information
└── /stock/
    ├── GET /list            # Get stock list
    └── GET /detail          # Get stock details
```

**Key Components:**
- `server/router.go`: Defines all API routes, configures middleware (CORS, Gzip, logging, authentication)
- `controller/`: Handles HTTP requests, performs parameter validation, calls service layer
- `service/`: Business logic implementation, interacts with database, Redis, external APIs
- `dto/`: Defines request and response data structures
- `middleware/`: API signature verification, request logging, and other middleware

---

### 3. indexer Service

**Purpose:** On-chain event indexing service that consumes on-chain event data from Kafka, parses smart contract event logs, and processes corresponding business logic.

**Core Features:**
- Consumes on-chain event data from Kafka
- Parses event logs using type-safe contract bindings
- Handles various order and token events (order submission, execution, cancellation, token transfers, etc.)
- Database operations (updates order status, records events, etc.)

**Directory Structure:**

```
apps/indexer/
├── main.go                    # Service entry point, initializes dependencies and starts Kafka consumer
├── config/                    # Configuration management
│   ├── config.go             # Configuration struct (Kafka, database, RPC, etc.)
│   ├── config.yaml           # Configuration file
│   └── fx.go                 # Configuration module dependency injection
├── consumer/                  # Kafka consumer
│   ├── kafka_consumer.go     # Kafka consumer implementation (message consumption, error handling)
│   └── fx.go                 # Consumer module dependency injection
├── service/                   # Business logic layer
│   ├── event_handler.go      # Event handler interface definition
│   ├── process_tx.go         # Transaction processing logic
│   ├── fx.go                 # Service layer dependency injection
│   └── handlers/             # Specific event handler implementations
│       ├── handle_order_submitted.go      # Order submission event handler
│       ├── handle_order_executed.go       # Order execution event handler
│       ├── handle_order_cancelled.go      # Order cancellation event handler
│       ├── handle_order_cancel_requested.go  # Order cancel request event handler
│       ├── handle_poc_token_transfer.go   # POC token transfer event handler
│       ├── handle_poc_token_tokens_minted.go  # POC token minting event handler
│       ├── handle_generic.go              # Generic event handler
│       └── helpers.go                     # Helper functions
├── scripts/                   # Script tools
│   ├── place_order/          # Place order script
│   │   └── main.go
│   ├── cancel_order/         # Cancel order script
│   │   └── main.go
│   └── README.md
├── types/                     # Type definitions
│   └── cli.go                # CLI argument types
├── migrations/                # Database migration scripts
└── logs/                      # Log file directory
```

**Event Processing Flow:**

```
Kafka Topic (snapshot.event.{chainId})
    │
    ▼
consumer/kafka_consumer.go (consume messages)
    │
    ▼
service/process_tx.go (parse event logs)
    │
    ▼
service/handlers/ (route to corresponding handler based on topic0)
    │
    ├── handle_order_submitted.go
    ├── handle_order_executed.go
    ├── handle_order_cancelled.go
    ├── handle_poc_token_transfer.go
    └── ...
    │
    ▼
libs/contracts/rwa/ (parse events using type-safe contract bindings)
    │
    ▼
Database (update order status, record event logs)
```

**Supported Event Types:**

| Event Type | Handler | Description |
|------------|---------|-------------|
| OrderSubmitted | handle_order_submitted.go | Order submission event |
| OrderExecuted | handle_order_executed.go | Order execution event |
| OrderCancelled | handle_order_cancelled.go | Order cancellation event |
| OrderCancelRequested | handle_order_cancel_requested.go | Order cancel request event |
| Transfer (PocToken) | handle_poc_token_transfer.go | POC token transfer event |
| TokensMinted (PocToken) | handle_poc_token_tokens_minted.go | POC token minting event |

**Key Components:**
- `consumer/kafka_consumer.go`: Kafka consumer that consumes messages from specified topics
- `service/process_tx.go`: Parses event logs and routes to corresponding handlers based on topic0
- `service/handlers/`: Specific processing logic for various events, uses type-safe contract bindings to parse events
- `libs/contracts/rwa/`: Auto-generated contract binding code that provides type-safe event parsing methods

---

### 4. ws-server Service

**Purpose:** WebSocket server service that provides real-time market data streaming to frontend clients. It acts as a bridge between Alpaca market data streams and WebSocket clients, enabling real-time bar (candle) data distribution.

**Core Features:**
- WebSocket server using Melody framework for client connections
- Real-time bar data streaming from Alpaca to subscribed clients
- Client subscription/unsubscription management
- Automatic Alpaca WebSocket subscription management (avoids duplicate subscriptions)
- Redis cache support for session management
- Efficient broadcast filtering based on client subscriptions

**Directory Structure:**

```
apps/ws-server/
├── main.go                    # Service entry point, CLI argument parsing and startup logic
├── config/                    # Configuration management
│   ├── config.go             # Configuration struct (server, Redis, Alpaca WebSocket settings)
│   ├── config.yaml           # Configuration file
│   └── fx.go                 # Configuration module dependency injection
├── ws/                        # WebSocket server implementation
│   ├── ws_server.go          # Melody-based WebSocket server (connection management, HTTP binding)
│   ├── ws_sub_unsub_service.go  # Subscription/unsubscription service (handles client messages)
│   └── fx.go                 # WebSocket module dependency injection
├── service/                   # Business logic layer
│   ├── alpaca_service.go     # Alpaca WebSocket client (connection, authentication, bar data handling)
│   └── fx.go                 # Service layer dependency injection
├── types/                     # Type definitions
│   ├── cli.go                # CLI argument types
│   └── ws.go                 # WebSocket message types, interfaces (BarSubscriber, etc.)
├── test/                      # Test files
│   └── index.html            # WebSocket client test page
└── logs/                      # Log file directory
```

**Data Flow:**

```
Frontend WebSocket Client
    │
    │ (WebSocket Connection)
    ▼
ws/ws_server.go (Melody server, connection handling)
    │
    │ (client messages: SUBSCRIBE/UNSUBSCRIBE)
    ▼
ws/ws_sub_unsub_service.go (message parsing, subscription management)
    │
    │ (subscribe to symbols)
    ▼
service/alpaca_service.go (Alpaca WebSocket client)
    │
    │ (WebSocket Connection)
    ▼
Alpaca WebSocket Server (market data stream)
    │
    │ (bar data messages)
    ▼
service/alpaca_service.go (receive and parse bar data)
    │
    │ (broadcast to subscribed clients)
    ▼
ws/ws_server.go (filter and send to clients)
    │
    ▼
Frontend WebSocket Client (receives real-time bar data)
```

**WebSocket Protocol:**

**Client → Server Messages:**

1. **Subscribe to Bar Data:**
```json
{
  "id": 1,
  "method": "SUBSCRIBE",
  "params": {
    "type": "bar",
    "symbols": ["AAPL", "TSLA"]
  }
}
```

2. **Unsubscribe from Bar Data:**
```json
{
  "id": 2,
  "method": "UNSUBSCRIBE",
  "params": {
    "type": "bar",
    "symbols": ["AAPL"]
  }
}
```

**Server → Client Messages:**

1. **Subscription Response:**
```json
{
  "id": 1,
  "result": "success"
}
```

2. **Bar Data Stream:**
```json
{
  "stream": "bar",
  "data": {
    "symbol": "AAPL",
    "open": 150.0,
    "high": 151.0,
    "low": 149.5,
    "close": 150.5,
    "volume": 1000000,
    "timestamp": 1234567890000000000,
    "tradeCount": 1000,
    "vwap": 150.25
  }
}
```

**Key Components:**
- `ws/ws_server.go`: Manages WebSocket server lifecycle, HTTP endpoint binding, Melody instance management
- `ws/ws_sub_unsub_service.go`: Handles client subscription/unsubscription requests, manages session-based subscriptions, coordinates with Alpaca service
- `service/alpaca_service.go`: Manages Alpaca WebSocket connection, authentication, bar data reception, and broadcasting to subscribed clients
- `types/ws.go`: Defines WebSocket message types, interfaces (BarSubscriber), and data structures
- **Dependency Injection Pattern**: Uses `types.BarSubscriber` interface to break circular dependencies between `ws` and `service` packages

**Architecture Highlights:**
- **Separation of Concerns**: WebSocket server logic (`ws/`) is separated from business logic (`service/`)
- **Interface-based Design**: Uses `BarSubscriber` interface in `types` package to enable dependency injection without circular dependencies
- **Efficient Subscription Management**: Tracks subscribed symbols to avoid duplicate Alpaca subscriptions
- **Session-based Filtering**: Uses Melody session storage to filter broadcasts to only subscribed clients

---

## Shared Libraries (libs/) Overview

### core/
Core business logic library, including:
- `evm_helper/`: EVM chain interaction (RPC client, signing, proxy wallet generation)
- `kafka_help/`: Kafka helper utilities (snapshot, transaction service)
- `local_cache/`: Local cache service
- `models/rwa/`: RWA data models (orders, stocks, trades, accounts, event logs)
- `redis_cache/`: Redis cache service (API keys, depth, order book, token prices)
- `trade/`: Trading-related (Alpaca integration, market data, trade execution)

### contracts/
Smart contract bindings, automatically generated type-safe Go code from ABI files.

### database/
Database client providing PostgreSQL connection and migration functionality.

### errors/
Unified error handling library, defining error codes and error response formats.

### grpc/
gRPC service definitions (cache, indexer, transaction service).

### kafka/
Kafka client wrapper (producer, consumer).

### log/
Logging library supporting file rotation, Gin logging, GORM logging, gRPC logging.

---

## Inter-Service Communication

```
┌─────────────┐
│   Frontend  │
└───┬─────┬───┘
    │     │
    │     │ WebSocket
    │     │
    │     ▼
    │ ┌─────────────┐
    │ │ ws-server   │
    │ │  service    │
    │ └──────┬──────┘
    │        │
    │        │ WebSocket
    │        │
    │        ▼
    │ ┌─────────────┐
    │ │   Alpaca    │
    │ │   Platform  │
    │ └─────────────┘
    │
    │ HTTP/REST
    ▼
┌─────────────┐
│  api service│◄────┐
└──────┬──────┘     │
       │            │
       │            │ gRPC
       │            │
       ▼            │
┌─────────────┐     │
│  Database   │     │
│ PostgreSQL  │     │
└─────────────┘     │
                    │
┌─────────────┐     │
│  indexer    │─────┘
│  service    │
└──────┬──────┘
       │
       │ Kafka Consumer
       ▼
┌─────────────┐
│   Kafka     │
│   Topics    │
└─────────────┘

┌─────────────┐
│alpaca-stream│
│  service    │
└──────┬──────┘
       │
       │ WebSocket
       ▼
┌─────────────┐
│   Alpaca    │
│   Platform  │
└─────────────┘
```

## Data Flow Examples

### Order Processing Flow

1. **User places order** → `api` service receives request
2. **Order creation** → `api` service writes to database
3. **On-chain event** → Blockchain generates event logs
4. **Event indexing** → `indexer` service consumes events from Kafka
5. **Event parsing** → `indexer` service parses events and updates database
6. **Order status update** → `alpaca-stream` service monitors Alpaca order status
7. **Status synchronization** → Gets order status updates via WebSocket

### Market Data Flow

1. **Subscribe to market data** → `alpaca-stream` service subscribes to Alpaca market data stream
2. **Receive data** → Real-time reception of bars/quotes/trades data
3. **Data processing** → Process and record market data
4. **API queries** → `api` service provides market data query interfaces

### Real-time WebSocket Data Flow

1. **Client connects** → Frontend client connects to `ws-server` via WebSocket
2. **Client subscribes** → Client sends SUBSCRIBE message with symbols (e.g., ["AAPL", "TSLA"])
3. **Alpaca subscription** → `ws-server` service subscribes to Alpaca WebSocket for requested symbols
4. **Data reception** → `ws-server` receives bar data from Alpaca WebSocket
5. **Broadcast to clients** → `ws-server` filters and broadcasts bar data to subscribed clients
6. **Real-time updates** → Clients receive real-time bar data updates as they occur

## Development Standards

### Directory Structure Standards

Each service follows a unified directory structure:
- `main.go`: Service entry point
- `config/`: Configuration management
- `service/`: Business logic layer
- `types/`: Type definitions
- `logs/`: Log files

### Dependency Injection

All services use Uber FX for dependency injection, with modules defined through `fx.go` files.

### Logging Standards

Uses unified logging library `libs/log`, supporting structured logging and log rotation.

### Error Handling

Uses unified error handling library `libs/errors`, defining standard error codes and error response formats.

## Deployment Instructions

### Local Development

```bash
# Start infrastructure (PostgreSQL, Redis, Kafka)
make install_all

# Run each service
cd apps/api && go run main.go
cd apps/indexer && go run main.go
cd apps/alpaca-stream && go run main.go
cd apps/ws-server && go run main.go -a ws
```

### Configuration

Each service has an independent configuration file `config/config.yaml` that needs to be configured with appropriate parameters based on the environment.

## Summary

RWA Backend adopts a microservices architecture with four core services, each with distinct responsibilities:
- **api**: Provides HTTP API interfaces externally for trading and stock information queries
- **indexer**: Processes on-chain events, maintains synchronization between on-chain and off-chain data
- **alpaca-stream**: Real-time monitoring of Alpaca platform data streams (order updates, market data)
- **ws-server**: WebSocket server that streams real-time market data (bars/candles) to frontend clients

All services share common libraries in `libs/` to ensure code reuse and consistency. Data flow and state synchronization between services are achieved through middleware such as Kafka, databases, and Redis. The architecture supports both RESTful API access and real-time WebSocket streaming for different client needs.
