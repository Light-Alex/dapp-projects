# RWA Indexer (Direct RPC)

An indexer service that listens to on-chain events directly via RPC without relying on Kafka.

## Features

1. **Block Maintenance**: Automatically tracks the latest processed block number with support for resuming from checkpoints
2. **Event Listening**: Listens to on-chain events directly via RPC, no Kafka required
3. **Event Idempotency**: Implements event processing idempotency based on EventId to prevent duplicate processing
4. **Order Processing**: 
   - Listens to `OrderSubmitted` events and automatically creates orders and submits them to Alpaca
   - Listens to `OrderCancelled` events and automatically cancels orders

## Architecture

```
indexer/
├── config/          # Configuration files
├── service/         # Core services
│   ├── block_service.go      # Block maintenance service
│   ├── event_listener.go     # Event listening service
│   ├── process_tx.go         # Event processing service (with idempotency checks)
│   └── handlers/             # Event handlers
│       ├── handle_order_submitted.go   # Order submission handler
│       ├── handle_order_cancelled.go   # Order cancellation handler
│       └── helpers.go                  # Helper functions
├── types/           # Type definitions
└── main.go          # Main entry point
```

## Configuration

Configuration file is located at `config/config.yaml`:

```yaml
chain:
  chainId: 20250903
  pocAddress: "0xae136110e64556bc15df5db254929c3a4a09dece"  # POC contract address

indexer:
  pollInterval: 3              # Polling interval (seconds)
  batchSize: 100               # Number of blocks to process per batch
  startBlock: 0                # Starting block number (0 means start from latest block)
  confirmationBlocks: 0        # Number of confirmation blocks (wait for confirmations before processing)
```

## How It Works

1. **Initialization**: On service startup, loads the last processed block number and event ID from the database
2. **Polling**: Periodically polls the latest block on the chain
3. **Fetch Events**: Uses `eth_getLogs` to fetch events within the specified block range
4. **Idempotency Check**: Checks if events have been processed using EventId
5. **Process Events**: Calls the corresponding handler based on event type
6. **Update Status**: Updates the last processed block number and event ID after processing

## Event Idempotency

Event idempotency is guaranteed through the following mechanisms:

1. **EventId Check**: In `ProcessEvent`, checks if `event.EventId <= latestResolvedEventId`
2. **Database Unique Constraints**: Orders are ensured to be unique through `ClientOrderID`
3. **Status Check**: When canceling orders, checks order status to avoid duplicate cancellations

## Database Tables

### event_client_record

Used to track processing progress:

```sql
CREATE TABLE event_client_record (
    chain_id BIGINT PRIMARY KEY,
    last_block BIGINT NOT NULL DEFAULT 0,
    last_event_id BIGINT NOT NULL DEFAULT 0,
    update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Running

```bash
# Run indexer
go run main.go

# Or with config file
go run main.go -c config/config.yaml
```

## Differences from indexer_kafka

| Feature | indexer_kafka | indexer (this service) |
|---------|--------------|----------------------|
| Event Source | Kafka | Direct RPC |
| Dependencies | Kafka | None |
| Deployment Complexity | Higher | Lower |
| Real-time Performance | High | Medium (polling) |
| Idempotency | EventId | EventId |

## Notes

1. Polling interval should not be too short to avoid putting pressure on RPC nodes
2. Confirmation blocks are recommended to be set to 1-3 to ensure events are not rolled back
3. Batch size should be adjusted according to the chain's block generation speed
4. Ensure stable database connections to avoid losing processing progress
