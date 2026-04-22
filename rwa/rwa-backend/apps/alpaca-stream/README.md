# Alpaca WebSocket Service

Alpaca WebSocket streaming service for listening to trade updates and market data in real-time.

## Overview

This service connects to Alpaca's WebSocket API to stream trade updates, bars (candles), quotes, and trades in real-time. It follows the same code structure and patterns as other services in the rwa-backend project.

## Features

- ✅ Real-time trade update streaming
- ✅ Market data streaming (bars, quotes, trades)
- ✅ Automatic reconnection with exponential backoff
- ✅ Structured logging of all events
- ✅ Modular, extensible architecture
- ✅ Support for all trade update event types
- ✅ Configurable subscriptions via YAML

## Architecture

The service is structured following the same patterns as `go-backend/apps/indexer`:

```
alpaca-stream/
├── main.go              # Entry point with CLI argument parsing
├── config/              # Configuration management
│   ├── config.go
│   ├── config.yaml
│   └── fx.go
├── ws/                  # WebSocket client and subscription management
│   ├── client.go        # Core WebSocket client with auto-reconnect
│   └── subscription.go  # Subscription manager
├── handlers/            # Message handlers for different stream types
│   ├── trade_updates_handler.go
│   └── bars_handler.go
├── service/             # Core business logic
│   ├── alpaca_ws_service.go
│   └── fx.go
└── types/               # CLI and type definitions
```

## Configuration

The service uses a YAML configuration file (default: `./config/config.yaml`). Example configuration:

```yaml
appName: "Alpaca WebSocket Service"

alpaca:
  api_key: "your-api-key"
  api_secret: "your-api-secret"
  base_url: "https://paper-api.alpaca.markets"
  data_url: "https://data.alpaca.markets"
  ws_url: "wss://paper-api.alpaca.markets/stream"  # For trade updates
  ws_data_url: "wss://stream.data.alpaca.markets/v2/iex"  # For market data

# Subscription configuration
subscriptions:
  # Enable trade updates subscription
  trade_updates:
    enabled: true
  
  # Market data subscriptions (bars, quotes, trades)
  market_data:
    enabled: true
    symbols:
      - "AAPL"
      - "GOOGL"
    feed: "iex"  # Available: "iex" or "sip"
    bars:
      enabled: true
    quotes:
      enabled: false
    trades:
      enabled: false

logger:
  level: "debug"
  encoderType: "console"
  outputType: "all"
  maxAge: 30
  enableColor: true
```

## Running the Service

```bash
# Run with default config
go run main.go

# Run with custom config file
go run main.go -c /path/to/config.yaml

# Build and run
go build -o alpaca-stream
./alpaca-stream
```

## Stream Types

### Trade Updates

Trade updates are streamed from `wss://paper-api.alpaca.markets/stream` or `wss://api.alpaca.markets/stream`. Event types:

- `new` - New order placed
- `fill` - Order completely filled
- `partial_fill` - Order partially filled
- `canceled` - Order canceled
- `expired` - Order expired
- `rejected` - Order rejected
- `replaced` - Order replaced
- `pending_new` - Order pending new
- `pending_cancel` - Order pending cancel
- `pending_replace` - Order pending replace
- `cancel_rejected` - Cancel rejected
- `replace_rejected` - Replace rejected

### Market Data

Market data streams are available from `wss://stream.data.alpaca.markets/v2/{feed}` where `feed` can be:
- `iex` - IEX feed (free for paper trading)
- `sip` - SIP feed (requires subscription)

Available market data types:
- **Bars** (Candles) - OHLCV data for specified symbols
- **Quotes** - Bid/ask quotes
- **Trades** - Trade executions

## Extending the Service

To add support for new stream types:

1. **Create a handler** in `handlers/`:
```go
type NewDataHandler struct {
    onData func(ctx context.Context, data NewDataType)
}

func (h *NewDataHandler) Handle(ctx context.Context, message json.RawMessage) error {
    // Parse and handle message
    return nil
}
```

2. **Register the handler** in `service/alpaca_ws_service.go`:
```go
handler := handlers.NewNewDataHandler()
handler.SetHandler(s.onNewData)
client.RegisterHandler("new_stream", handler.Handle)
```

3. **Update subscription config** if needed.

## Error Handling

- Automatic reconnection with exponential backoff (max 30s delay)
- Connection errors are logged with structured logging
- Authentication failures are handled gracefully
- Subscription errors are reported but don't stop the service

## WebSocket Protocol

This service implements the Alpaca WebSocket protocol as documented at:
- [WebSocket Streaming Documentation](https://docs.alpaca.markets/docs/websocket-streaming)
- [Real-time Stock Pricing Data](https://docs.alpaca.markets/docs/real-time-stock-pricing-data)

### Authentication

The client authenticates using API key and secret:
```json
{
  "action": "auth",
  "key": "{API_KEY}",
  "secret": "{API_SECRET}"
}
```

### Subscription

For trade updates:
```json
{
  "action": "listen",
  "data": {
    "streams": ["trade_updates"]
  }
}
```

For market data:
```json
{
  "action": "subscribe",
  "bars": ["AAPL", "GOOGL"]
}
```

## Dependencies

- `github.com/gorilla/websocket` - WebSocket client
- `go.uber.org/fx` - Dependency injection framework
- `go.uber.org/zap` - Structured logging
- `github.com/urfave/cli/v3` - CLI framework
- `gopkg.in/yaml.v3` - YAML configuration

## Code Style

The code follows the same patterns as `go-backend/apps/indexer`:
- Uses `fx` for dependency injection
- Structured logging with `zap`
- Context-based cancellation
- Graceful shutdown handling
