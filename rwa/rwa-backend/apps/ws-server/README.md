# WS Server

WebSocket server for RWA backend that provides real-time bar (candle) data from Alpaca to frontend clients.

## Features

- WebSocket server using Melody
- Real-time bar data streaming from Alpaca
- Client subscription/unsubscription support
- Automatic Alpaca subscription management

## Configuration

Edit `config/config.yaml`:

```yaml
appName: "WS Server"
server:
  port: 8082
  basePath: /api/v1/ws

redis:
  hosts:
    - "127.0.0.1:6379"
  password: ""
  db: 0

logger:
  level: "debug"
  encoderType: "console"
  outputType: "all"
  maxAge: 30
  enableColor: true

alpaca:
  api_key: "your_api_key"
  api_secret: "your_api_secret"
  ws_url: "wss://stream.data.alpaca.markets/v2/iex"
  ws_data_url: "wss://stream.data.alpaca.markets/v2/iex"
```

## Usage

### Start the server

```bash
go run main.go -a ws -c ./config/config.yaml
```

### WebSocket API

#### Connect

Connect to: `ws://localhost:8082/api/v1/ws`

#### Subscribe to Bar Data

Send a subscription message:

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

#### Unsubscribe from Bar Data

Send an unsubscription message:

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

#### Receive Bar Data

You will receive bar data in the following format:

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

## Architecture

- `main.go`: Application entry point
- `config/`: Configuration management
- `ws/`: WebSocket server implementation
  - `ws_server.go`: Melody-based WebSocket server
  - `ws_sub_unsub_service.go`: Subscription/unsubscription handling
- `service/`: Business logic
  - `alpaca_service.go`: Alpaca WebSocket client and bar data handling
- `types/`: Type definitions

## Integration with Alpaca

The service automatically:
1. Connects to Alpaca WebSocket when configured
2. Subscribes to bar data when clients request it
3. Broadcasts received bar data to subscribed clients
4. Manages symbol subscriptions efficiently (avoids duplicate subscriptions)


