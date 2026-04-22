# RWA Platform Test Cases

> Generated: 2026-03-08
> Author: QA Test Engineer
> Scope: Stock Pagination, Market Data Cache, Alpaca WS Reconnect, Order Status Push, Kafka Migration

---

## 1. Stock List Pagination

### 1.1 Functional Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| STK-F-001 | Default pagination | DB has >20 active stocks | GET `/stock/list` without page/pageSize params | Returns page=1, pageSize=20, response contains `list` array (<=20 items) and `total` count | P0 |
| STK-F-002 | Specify page and pageSize | DB has 50 active stocks | GET `/stock/list?page=2&page_size=10` | Returns 10 items from offset 10, total=50 | P0 |
| STK-F-003 | Last page with partial results | DB has 25 active stocks | GET `/stock/list?page=2&page_size=20` | Returns 5 items, total=25 | P0 |
| STK-F-004 | Response structure correctness | DB has active stocks | GET `/stock/list?page=1&page_size=5` | Each item contains: id, symbol, name, exchange, contract, status, createdAt, updatedAt | P1 |
| STK-F-005 | Stock detail query | DB has AAPL stock record | GET `/stock/detail?symbol=AAPL` | Returns StockInfo with correct symbol, name, exchange | P0 |

### 1.2 Boundary Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| STK-B-001 | page=0 defaults to 1 | DB has stocks | GET `/stock/list?page=0` | Treated as page=1, returns first page | P0 |
| STK-B-002 | Negative page defaults to 1 | DB has stocks | GET `/stock/list?page=-1` | Treated as page=1, returns first page | P1 |
| STK-B-003 | pageSize=0 defaults to 20 | DB has stocks | GET `/stock/list?page_size=0` | Treated as pageSize=20 | P0 |
| STK-B-004 | Negative pageSize defaults to 20 | DB has stocks | GET `/stock/list?page_size=-5` | Treated as pageSize=20 | P1 |
| STK-B-005 | pageSize exceeds max (100) | DB has >100 stocks | GET `/stock/list?page_size=200` | Capped at pageSize=100, returns at most 100 items | P0 |
| STK-B-006 | pageSize exactly 100 | DB has >100 stocks | GET `/stock/list?page_size=100` | Returns up to 100 items, not capped | P1 |
| STK-B-007 | pageSize=1 | DB has stocks | GET `/stock/list?page_size=1` | Returns exactly 1 item | P1 |
| STK-B-008 | Page beyond data range | DB has 10 stocks | GET `/stock/list?page=999&page_size=20` | Returns empty list `[]`, total=10 | P1 |
| STK-B-009 | Empty database | DB has 0 active stocks | GET `/stock/list` | Returns empty list `[]`, total=0 | P1 |

### 1.3 Exception Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| STK-E-001 | Non-numeric page param | N/A | GET `/stock/list?page=abc` | Returns 400 with ErrInvalidRequestParams | P1 |
| STK-E-002 | Non-numeric pageSize param | N/A | GET `/stock/list?page_size=xyz` | Returns 400 with ErrInvalidRequestParams | P1 |
| STK-E-003 | Stock detail with empty symbol | N/A | GET `/stock/detail?symbol=` | Returns 400 with ErrInvalidRequestParams (binding:"required") | P0 |
| STK-E-004 | Stock detail with nonexistent symbol | DB has no ZZZZ stock | GET `/stock/detail?symbol=ZZZZ` | Returns 404 with ErrNotFound | P0 |
| STK-E-005 | Stock detail missing symbol param | N/A | GET `/stock/detail` | Returns 400 with ErrInvalidRequestParams | P1 |
| STK-E-006 | DB connection failure | DB is down | GET `/stock/list` | Returns 500 with ErrFailedToGetStockList | P2 |

---

## 2. Market Data Redis Cache

### 2.1 Functional Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| MDC-F-001 | currentPrice cache miss then hit | Redis empty, Alpaca available | 1. GET `/trade/currentPrice?symbol=AAPL` (miss) 2. GET same within 5s (hit) | First call fetches from Alpaca and caches; second call returns cached data without calling Alpaca | P0 |
| MDC-F-002 | latestQuote cache miss then hit | Redis empty | 1. GET `/trade/latestQuote?symbol=AAPL` (miss) 2. GET same within 5s (hit) | Same pattern: upstream on miss, cache on hit | P0 |
| MDC-F-003 | snapshot cache miss then hit | Redis empty | 1. GET `/trade/snapshot?symbol=AAPL` (miss) 2. GET same within 10s (hit) | Snapshot cached for 10s TTL | P0 |
| MDC-F-004 | historicalData cache miss then hit | Redis empty | 1. GET `/trade/historicalData?symbol=AAPL&start_time=1704067200&end_time=1706745599&interval=1d` (miss) 2. Same request within 60s (hit) | Historical data cached for 60s TTL | P0 |
| MDC-F-005 | marketClock cache miss then hit | Redis empty | 1. GET `/trade/marketClock` (miss) 2. GET same within 30s (hit) | Market clock cached for 30s TTL with key "global" | P0 |
| MDC-F-006 | Cache key format correctness | N/A | Call currentPrice for AAPL | Redis key is `rwa:marketdata:currentPrice:AAPL` | P1 |
| MDC-F-007 | historicalData cache key includes all params | N/A | GET historicalData with symbol=AAPL, start=1704067200, end=1706745599, interval=1d, limit=100 | Cache key is `rwa:marketdata:historicalData:AAPL_1704067200_1706745599_1d_100` | P1 |
| MDC-F-008 | Different symbols have different cache entries | Redis empty | 1. GET currentPrice for AAPL 2. GET currentPrice for GOOGL | Each symbol has its own cache entry; GOOGL request hits upstream | P1 |

### 2.2 Boundary Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| MDC-B-001 | currentPrice TTL expiry (5s) | Cache populated for AAPL | Wait 6s, then GET `/trade/currentPrice?symbol=AAPL` | Cache expired, fetches fresh data from Alpaca | P0 |
| MDC-B-002 | latestQuote TTL expiry (5s) | Cache populated | Wait 6s, re-request | Cache miss, fetches from upstream | P1 |
| MDC-B-003 | snapshot TTL expiry (10s) | Cache populated | Wait 11s, re-request | Cache miss | P1 |
| MDC-B-004 | historicalData TTL expiry (60s) | Cache populated | Wait 61s, re-request | Cache miss | P1 |
| MDC-B-005 | marketClock TTL expiry (30s) | Cache populated | Wait 31s, re-request | Cache miss | P1 |
| MDC-B-006 | historicalData with different limit | Same symbol/start/end/interval but limit=50 vs limit=100 | Make two requests with different limits | Different cache keys, no collision | P1 |
| MDC-B-007 | Snapshot with nil sub-fields | Alpaca returns snapshot with nil LatestTrade | GET `/trade/snapshot?symbol=XYZ` | Response has null/missing LatestTrade, no panic; data still cached | P2 |

### 2.3 Exception Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| MDC-E-001 | Redis down, upstream available | Redis unreachable | GET `/trade/currentPrice?symbol=AAPL` | Falls back to Alpaca upstream, returns data; logs warning about Redis error | P0 |
| MDC-E-002 | Redis down, cache Set fails silently | Redis unreachable | GET currentPrice (upstream succeeds, cache.Set fails) | Response returned to client; Set error ignored (fire-and-forget with `_ =`) | P0 |
| MDC-E-003 | Upstream Alpaca fails, cache empty | Redis empty, Alpaca returns error | GET `/trade/currentPrice?symbol=AAPL` | Returns 500 with ErrFailedToGetCurrentPrice | P0 |
| MDC-E-004 | Redis returns corrupted data | Redis has non-JSON data for key | GET `/trade/currentPrice?symbol=AAPL` | json.Unmarshal fails, treated as cache miss, falls back to upstream | P1 |
| MDC-E-005 | Missing symbol param for currentPrice | N/A | GET `/trade/currentPrice` | Returns 400 with ErrInvalidRequestParams | P1 |
| MDC-E-006 | Missing symbol param for latestQuote | N/A | GET `/trade/latestQuote` | Returns 400 with ErrInvalidRequestParams | P1 |
| MDC-E-007 | Missing symbol param for snapshot | N/A | GET `/trade/snapshot` | Returns 400 with ErrInvalidRequestParams | P1 |
| MDC-E-008 | Invalid timestamp format from Alpaca quote | Alpaca returns non-RFC3339 timestamp | GET `/trade/latestQuote?symbol=AAPL` | Returns 500 with ErrInvalidTimestampFormat | P2 |
| MDC-E-009 | historicalData missing required params | N/A | GET `/trade/historicalData?symbol=AAPL` (no start_time/end_time/interval) | Returns 400 with ErrInvalidRequestParams | P1 |

---

## 3. Alpaca WebSocket Auto-Reconnect

### 3.1 Functional Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| AWS-F-001 | Initial connection and auth | Valid Alpaca API credentials | Start AlpacaService | Connects to `wss://stream.data.alpaca.markets/v2/iex`, authenticates, logs success | P0 |
| AWS-F-002 | Bar data reception and broadcast | Connected, client subscribed to AAPL bars | Alpaca sends bar message for AAPL | onBar callback invoked, BroadcastFilter sends to sessions with key `bar_AAPL` | P0 |
| AWS-F-003 | Auto-reconnect on disconnect | Connected, then server closes connection | Connection drops (ReadMessage returns error) | onDisconnect triggers reconnect(); new connection established; logs reconnection | P0 |
| AWS-F-004 | Resubscribe after reconnect | Subscribed to [AAPL, GOOGL], then disconnect | Reconnect succeeds | SubscribeBars called with [AAPL, GOOGL] on new connection | P0 |
| AWS-F-005 | Subscribe to new symbols | Connected and authenticated | Call SubscribeBars(ctx, ["AAPL", "TSLA"]) | Sends subscribe message with bars=["AAPL","TSLA"]; symbols tracked in subscribedSymbols | P0 |
| AWS-F-006 | Skip already-subscribed symbols | Already subscribed to AAPL | Call SubscribeBars(ctx, ["AAPL", "MSFT"]) | Only sends subscribe for MSFT; AAPL skipped | P1 |
| AWS-F-007 | All symbols already subscribed | Subscribed to [AAPL, GOOGL] | Call SubscribeBars(ctx, ["AAPL", "GOOGL"]) | No subscribe message sent, returns nil | P1 |
| AWS-F-008 | Graceful stop | Service running with active connection | Call Stop(ctx) | stopCh closed, client.Close() called, client set to nil | P0 |

### 3.2 Boundary Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| AWS-B-001 | Exponential backoff: attempt 0 | Disconnected | First reconnect attempt | Backoff = 1s (2^0) | P0 |
| AWS-B-002 | Exponential backoff: attempt 1 | First attempt failed | Second reconnect attempt | Backoff = 2s (2^1) | P1 |
| AWS-B-003 | Exponential backoff: attempt 5 | 5 attempts failed | Sixth reconnect attempt | Backoff = 32s (2^5) | P1 |
| AWS-B-004 | Backoff capped at 30s | 5+ attempts failed | Sixth attempt (2^5 = 32s) | Backoff capped at 30s (DefaultMaxReconnectDelay) | P0 |
| AWS-B-005 | Backoff cap at attempt 10 | 10 attempts failed | Attempt 11 (2^10 = 1024s) | Backoff capped at 30s | P1 |
| AWS-B-006 | No concurrent reconnection | Two simultaneous disconnects | Two goroutines call reconnect() | Only one proceeds (CompareAndSwap); second returns immediately | P0 |
| AWS-B-007 | Auth timeout (10s) | Server does not respond to auth | Connect() called | Returns "authentication timeout" error after 10s | P1 |
| AWS-B-008 | Handshake timeout (10s) | Server unreachable | Connect() called | Dialer times out after 10s | P1 |
| AWS-B-009 | No Alpaca config | config.Alpaca is nil or APIKey empty | Start() called | Logs warning, returns nil (no error), skips connection | P1 |
| AWS-B-010 | Default WSDataURL | config.Alpaca.WSDataURL is empty | Start() called | Uses `wss://stream.data.alpaca.markets/v2/iex` | P2 |

### 3.3 Exception Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| AWS-E-001 | Auth rejected by server | Invalid API credentials | Connect() called | authenticated stays false, returns "authentication rejected by server" | P0 |
| AWS-E-002 | Reconnect fails multiple times then succeeds | Server down for 3 attempts | Disconnect triggers reconnect | Retries with backoff, succeeds on 4th attempt, resubscribes symbols | P0 |
| AWS-E-003 | Resubscribe fails after reconnect | Connection restored but subscribe fails | Reconnect succeeds, SubscribeBars errors | Closes new client, increments attempt, continues retry loop | P1 |
| AWS-E-004 | Stop during reconnect backoff | Reconnecting with pending backoff timer | Call Stop() (closes stopCh) | Reconnect loop exits via stopCh select | P1 |
| AWS-E-005 | SubscribeBars on nil client | Client not initialized | Call SubscribeBars(ctx, ["AAPL"]) | Returns error "Alpaca client not initialized" | P1 |
| AWS-E-006 | SubscribeBars on unauthenticated client | Client connected but not authenticated | Call client.SubscribeBars(ctx, ["AAPL"]) | Returns error "client not authenticated" | P1 |
| AWS-E-007 | writeJSON on nil conn | Client conn is nil | Call writeJSON() | Returns error "connection not established" | P2 |
| AWS-E-008 | Malformed message from Alpaca | Alpaca sends invalid JSON | readMessages receives garbage | handleMessage returns unmarshal error, logged; read loop continues | P2 |

---

## 4. Order Status Real-time Push (Kafka Consumer -> WebSocket)

### 4.1 Functional Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| ORP-F-001 | Order update broadcast to subscribed client | WS client subscribed to order updates for account_id=123 | Kafka consumer receives OrderUpdateEvent with AccountID=123 | WS client receives message with stream="order", data containing orderId, status, event | P0 |
| ORP-F-002 | Order update not sent to unsubscribed client | WS client A subscribed to account 123, client B subscribed to account 456 | OrderUpdateEvent for AccountID=123 arrives | Only client A receives the message; client B does not | P0 |
| ORP-F-003 | Multiple events for same order | Order goes through new -> partial_fill -> fill | Three Kafka messages arrive sequentially | WS client receives three separate push messages with correct event types | P0 |
| ORP-F-004 | Event payload completeness | Valid OrderUpdateEvent published | Consumer receives and broadcasts | Payload contains: accountId, orderId, clientOrderId, symbol, side, status, filledQuantity, filledPrice, remainingQuantity, quantity, event, timestamp | P1 |
| ORP-F-005 | Consumer group isolation | Two ws-server instances in same consumer group | OrderUpdateEvent published | Only one instance processes the message (Kafka consumer group semantics) | P1 |

### 4.2 Boundary Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| ORP-B-001 | No WS clients subscribed | No clients connected | OrderUpdateEvent arrives via Kafka | Message consumed, BroadcastFilter sends to 0 sessions, no error | P1 |
| ORP-B-002 | Multiple clients for same account | 3 WS clients subscribed to account 123 | OrderUpdateEvent for account 123 | All 3 clients receive the message | P0 |
| ORP-B-003 | Kafka disabled | KafkaConfig.Enabled = false | Start() called | Logs warning "Kafka not enabled, skipping"; consumer not created; returns nil | P1 |
| ORP-B-004 | Large event payload | OrderUpdateEvent with very long symbol/notes | Consumer processes message | No truncation; entire payload delivered to WS client | P2 |

### 4.3 Exception Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| ORP-E-001 | Malformed Kafka message | Message value is invalid JSON | Consumer handleMessage called | Unmarshal fails, logs error, returns nil (ack to skip) | P0 |
| ORP-E-002 | Kafka broker unavailable at startup | Brokers unreachable | Start() called | NewKafkaConsumer returns error, Start returns error, service fails to start | P0 |
| ORP-E-003 | WS broadcast error | Melody BroadcastFilter fails | Valid OrderUpdateEvent consumed | Error logged "failed to broadcast order update" | P1 |
| ORP-E-004 | Graceful shutdown | Consumer running | Stop(ctx) called | consumer.Close(ctx) called, consumer stops cleanly | P1 |

---

## 5. WebSocket Subscribe/Unsubscribe

### 5.1 Functional Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| WSS-F-001 | Subscribe to bar stream | WS connected | Send `{"method":"SUBSCRIBE","id":1,"params":{"type":"bar","symbols":["AAPL"]}}` | Session key `bar_AAPL` set; response `{"id":1,"result":"success"}`; bar data received via Kafka from alpaca-stream | P0 |
| WSS-F-002 | Subscribe to multiple bar symbols | WS connected | Send SUBSCRIBE with symbols=["AAPL","GOOGL","TSLA"] | Three session keys set: bar_AAPL, bar_GOOGL, bar_TSLA; bar data broadcast from Kafka BarUpdateSubscriber | P0 |
| WSS-F-003 | Subscribe to order stream | WS connected | Send `{"method":"SUBSCRIBE","id":2,"params":{"type":"order","account_id":123}}` | Session key `order_123` set; response `{"id":2,"result":"success"}` | P0 |
| WSS-F-004 | Unsubscribe from bar stream | Subscribed to AAPL bars | Send `{"method":"UNSUBSCRIBE","id":3,"params":{"type":"bar","symbols":["AAPL"]}}` | Session key `bar_AAPL` removed; response `{"id":3,"result":"success"}` | P0 |
| WSS-F-005 | Unsubscribe from order stream | Subscribed to order account 123 | Send `{"method":"UNSUBSCRIBE","id":4,"params":{"type":"order","account_id":123}}` | Session key `order_123` removed; response success | P0 |
| WSS-F-006 | Ping/pong heartbeat | WS connected | Send text message "ping" | Receives "pong" response | P0 |
| WSS-F-007 | Symbol case normalization | WS connected | Send SUBSCRIBE with symbols=["aapl"] | Session key is `bar_AAPL` (uppercased) | P1 |
| WSS-F-008 | Connection/disconnection logging | N/A | Client connects then disconnects | HandleConnect and HandleDisconnect log with client RemoteAddr | P2 |

### 5.2 Boundary Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| WSS-B-001 | Subscribe with empty symbols array | WS connected | Send SUBSCRIBE with type=bar, symbols=[] | No action taken, no error returned (early return) | P1 |
| WSS-B-002 | Subscribe with account_id=0 | WS connected | Send SUBSCRIBE with type=order, account_id=0 | No action (accountID <= 0 check fails) | P1 |
| WSS-B-003 | Subscribe with negative account_id | WS connected | Send SUBSCRIBE with type=order, account_id=-1 | No action (accountID <= 0 check fails) | P1 |
| WSS-B-004 | Unknown subscription type | WS connected | Send SUBSCRIBE with type="unknown" | No action, default case returns silently | P1 |
| WSS-B-005 | Unknown method | WS connected | Send `{"method":"LIST","id":5,"params":{}}` | No action, default case returns silently | P2 |
| WSS-B-006 | Kafka disabled | Kafka not enabled | Send bar subscribe | Session keys still set locally; no bar data pushed (BarUpdateSubscriber skipped) | P2 |

### 5.3 Exception Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| WSS-E-001 | Invalid JSON message | WS connected | Send `{invalid json}` | json.Unmarshal fails, silently returns (no crash) | P1 |
| WSS-E-002 | Missing "type" in params | WS connected | Send `{"method":"SUBSCRIBE","id":1,"params":{}}` | No "type" key found, silently returns | P1 |
| WSS-E-003 | Kafka bar consumer error | Kafka consumer disconnected | Send SUBSCRIBE for bar AAPL | Session keys set locally; bar data unavailable until Kafka reconnects | P1 |
| WSS-E-004 | Missing "account_id" for order sub | WS connected | Send `{"method":"SUBSCRIBE","id":1,"params":{"type":"order"}}` | Type assertion fails for account_id, silently returns | P1 |

---

## 6. Order Sync Service (Alpaca Stream -> Kafka Publish)

### 6.1 Functional Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| OSS-F-001 | HandleNew: accepted status | Order exists in DB with status=pending | Alpaca sends "new" event | Order.Status updated to "accepted", AcceptedAt set, ExternalOrderID saved; Kafka publish with event="new" | P0 |
| OSS-F-002 | HandleFill: full fill | Order exists, partially filled or accepted | Alpaca sends "fill" event with price/qty | OrderExecution created; FilledQuantity/FilledPrice updated; Status="filled"; FilledAt set; Kafka publish event="fill"; callMarkExecuted triggered async | P0 |
| OSS-F-003 | HandlePartialFill | Order exists | Alpaca sends "partial_fill" with price=150.5, qty=5 | OrderExecution created; FilledQuantity incremented; Status="partially_filled"; VWAP computed; Kafka publish event="partial_fill" | P0 |
| OSS-F-004 | HandleCanceled: no fills | Order exists with no fills | Alpaca sends "canceled" | Status="cancelled", CancelledAt set; Kafka publish event="cancelled"; callCancelOrder triggered async | P0 |
| OSS-F-005 | HandleCanceled: with partial fills | Order partially filled then canceled | Alpaca sends "canceled" | Status="cancelled"; callMarkExecuted triggered (not cancelOrder) to settle partial fill + refund | P0 |
| OSS-F-006 | HandleRejected | Order exists | Alpaca sends "rejected" with reason="insufficient buying power" | Status="rejected", Notes contains "Rejected by Alpaca: insufficient buying power"; Kafka publish event="rejected"; callCancelOrder triggered | P0 |
| OSS-F-007 | HandleExpired: no fills | Order exists with no fills | Alpaca sends "expired" | Status="expired", ExpiredAt set; callCancelOrder triggered | P0 |
| OSS-F-008 | HandleExpired: with partial fills | Order partially filled then expired | Alpaca sends "expired" | Status="expired"; callMarkExecuted triggered (same logic as cancel with partial) | P1 |
| OSS-F-009 | HandleDoneForDay | Order exists (GTC) | Alpaca sends "done_for_day" | No status change; log only; order continues next day | P1 |
| OSS-F-010 | Kafka publish payload | Any order event handled | publishOrderUpdate called | OrderUpdateEvent contains correct: AccountID, OrderID, ClientOrderID, Symbol, Side, Status, FilledQuantity, FilledPrice, RemainingQuantity, Quantity, Event, Timestamp | P0 |
| OSS-F-011 | VWAP calculation on multiple partial fills | Order has filledQty=10@$100, new partial fill qty=10@$200 | HandlePartialFill called | FilledPrice = (10*100 + 10*200)/20 = $150 VWAP | P1 |
| OSS-F-012 | Alpaca authoritative fields override | Alpaca provides filled_avg_price and filled_qty | HandleFill processes | order.FilledPrice uses Alpaca's filled_avg_price; FilledQuantity uses Alpaca's filled_qty | P1 |

### 6.2 Boundary Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| OSS-B-001 | Idempotency: HandleNew on already accepted order | Order status is already "accepted" | HandleNew called again | Skips update, logs "order already in accepted or later state" | P0 |
| OSS-B-002 | Idempotency: HandleNew on filled order | Order status is "filled" | HandleNew called | Skips update | P0 |
| OSS-B-003 | Idempotency: duplicate execution_id | OrderExecution with same execution_id exists | HandleFill called with same execution_id | Detects duplicate, skips insert, no Kafka publish, no on-chain call | P0 |
| OSS-B-004 | Idempotency: HandleCanceled on already cancelled | Order already cancelled | HandleCanceled called again | Skips, logs "order already cancelled" | P1 |
| OSS-B-005 | Idempotency: HandleRejected on already rejected | Order already rejected | HandleRejected called | Skips | P1 |
| OSS-B-006 | Idempotency: HandleExpired on already expired | Order already expired | HandleExpired called | Skips | P1 |
| OSS-B-007 | Cancel on filled order | Order fully filled | HandleCanceled called | Logs warning, skips cancel (filled cannot be cancelled) | P0 |
| OSS-B-008 | Reject on filled order | Order fully filled | HandleRejected called | Logs warning, skips reject | P1 |
| OSS-B-009 | Kafka publisher is nil | orderUpdatePub not configured | publishOrderUpdate called | Returns immediately, no panic | P1 |
| OSS-B-010 | Empty timestamp in event | Alpaca sends empty timestamp | parseTimestampOrNow("") | Returns time.Now() | P2 |
| OSS-B-011 | Invalid timestamp format | Alpaca sends non-RFC3339 timestamp | parseTimestampOrNow("bad") | Returns time.Now() | P2 |

### 6.3 Exception Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| OSS-E-001 | Order not found | client_order_id has no matching DB record | Any Handle* called | Error logged, processing stops | P0 |
| OSS-E-002 | Empty client_order_id | Alpaca event has empty client_order_id | Any Handle* called | Error "client_order_id not found or empty", logged and returned | P0 |
| OSS-E-003 | Invalid price string | Alpaca sends price="abc" | HandleFill called | decimal parse fails, logged, returns early | P1 |
| OSS-E-004 | Invalid qty string | Alpaca sends qty="" | HandleFill called | decimal parse fails, logged, returns early | P1 |
| OSS-E-005 | DB transaction failure | DB error during fill update | HandleFill with valid data | Transaction rolled back; failed event persisted to failed_events table | P0 |
| OSS-E-006 | Failed event persistence | DB error during fill, then failed_events insert also fails | HandleFill fails, persistFailedEvent fails | Both errors logged; no crash | P1 |
| OSS-E-007 | Private key not configured | Backend.PrivateKey is empty | callMarkExecuted or callCancelOrder called | Logs warning "backend private key not configured, skipping"; returns | P1 |
| OSS-E-008 | Chain config not set | conf.Chain is nil | callMarkExecuted or callCancelOrder called | Logs warning "chain config not set, skipping"; returns | P1 |
| OSS-E-009 | On-chain markExecuted fails | Contract call returns error | callMarkExecuted called | Error logged with orderId and client_order_id; DB not updated with tx_hash | P1 |
| OSS-E-010 | ClientOrderID not parseable as uint | ClientOrderID = "abc" | callMarkExecuted or callCancelOrder | Parse error logged, returns early | P2 |

---

## 7. Kafka Producer (Order Update Publishing)

### 7.1 Functional Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| KPR-F-001 | Successful publish | Kafka enabled, broker available | Publish(ctx, event) with valid OrderUpdateEvent | Message sent to topic "rwa.order.update" with key=AccountID; logged "published order update" | P0 |
| KPR-F-002 | Message key is AccountID | N/A | Publish event with AccountID=456 | Kafka message key = "456" (ensures ordering per account) | P1 |
| KPR-F-003 | Sync producer acks | RequiredAcks = WaitForAll | Publish called | Blocks until all in-sync replicas acknowledge | P1 |

### 7.2 Boundary Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| KPR-B-001 | Kafka disabled | KafkaConfig.Enabled = false | Publish called | Returns immediately, no message sent | P0 |
| KPR-B-002 | Producer is nil | Kafka disabled, producer not created | Publish called | Returns immediately (nil check) | P1 |
| KPR-B-003 | Initialization with Kafka disabled | KafkaConfig.Enabled = false | NewOrderUpdateKafkaService called | Returns service with nil producer, no error | P1 |

### 7.3 Exception Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| KPR-E-001 | Kafka broker unavailable | Broker down | Publish called | SendMessage fails, error logged with orderId and event type; retries up to 3 times (config.Producer.Retry.Max=3) | P0 |
| KPR-E-002 | JSON marshal failure | Event contains unmarshalable field | Publish called | Marshal error logged, returns early | P2 |
| KPR-E-003 | Producer creation failure | Invalid broker address | NewOrderUpdateKafkaService called | Returns error, logged "failed to create producer" | P1 |

---

## 8. Integration Tests (Cross-Service)

### 8.1 End-to-End Order Flow

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| INT-E2E-001 | Full buy order lifecycle | User places buy order on-chain; WS client subscribed to order updates for the account | 1. Indexer picks up on-chain OrderPlaced event 2. alpaca-stream places order on Alpaca 3. Alpaca sends "new" event 4. HandleNew updates DB, publishes to Kafka 5. ws-server consumer receives Kafka message 6. WS pushes to client | WS client receives order status "accepted" in real-time | P0 |
| INT-E2E-002 | Order fill with token mint | Buy order accepted | 1. Alpaca sends "fill" event 2. HandleFill updates DB, publishes to Kafka 3. ws-server pushes "fill" to WS client 4. callMarkExecuted sends on-chain tx 5. mintTokensAfterFill mints stock tokens to user | WS client receives "fill"; on-chain markExecuted and mint tx succeed | P0 |
| INT-E2E-003 | Order cancel with refund | Buy order pending, no fills | 1. Alpaca sends "canceled" 2. HandleCanceled updates DB, publishes to Kafka 3. ws-server pushes "cancelled" 4. callCancelOrder refunds escrow on-chain | WS client receives "cancelled"; user's USDM refunded | P0 |
| INT-E2E-004 | Partial fill then cancel | Buy order with partial fills | 1. Alpaca sends "partial_fill" -> Kafka -> WS push 2. Alpaca sends "canceled" -> Kafka -> WS push 3. callMarkExecuted settles partial + refunds remainder | WS client receives both events; on-chain settlement handles partial correctly | P0 |

### 8.2 Resilience Tests

| ID | Scenario | Preconditions | Steps | Expected Result | Priority |
|----|----------|---------------|-------|-----------------|----------|
| INT-R-001 | Kafka down then recovers | All services running | 1. Stop Kafka brokers 2. alpaca-stream receives fill event 3. Publish fails 4. Restart Kafka | Publish error logged; on next event, Kafka publish succeeds; WS client receives subsequent updates | P0 |
| INT-R-002 | ws-server restart during order flow | WS client connected, subscribed | 1. ws-server restarts 2. Client reconnects and resubscribes 3. New order event arrives via Kafka | Client receives order update after resubscription | P1 |
| INT-R-003 | Alpaca WS disconnect mid-flow | Alpaca WS connected, bars streaming | 1. Network interruption 2. readMessages gets error 3. Reconnect with exponential backoff 4. Reconnects and resubscribes | Bar data resumes after reconnection; WS clients continue receiving bars | P0 |
| INT-R-004 | Redis cache failure does not block API | Redis down | 1. GET currentPrice (cache miss due to error) 2. Upstream Alpaca returns data 3. Cache Set fails silently | Client gets correct response; performance degraded but not broken | P0 |
| INT-R-005 | Concurrent order updates for same account | Multiple orders for account 123 fill simultaneously | Two fill events arrive close together | Both processed correctly; Kafka messages ordered by AccountID key; WS client receives both | P1 |

---

## Summary

| Module | Functional | Boundary | Exception | Integration | Total |
|--------|-----------|----------|-----------|-------------|-------|
| Stock Pagination | 5 | 9 | 6 | - | 20 |
| Market Data Cache | 8 | 7 | 9 | - | 24 |
| Alpaca WS Reconnect | 8 | 10 | 8 | - | 26 |
| Order Status Push | 5 | 4 | 4 | - | 13 |
| WS Sub/Unsub | 8 | 6 | 4 | - | 18 |
| Order Sync Service | 12 | 11 | 10 | - | 33 |
| Kafka Producer | 3 | 3 | 3 | - | 9 |
| Integration (E2E) | 4 | - | - | 5 | 9 |
| **Total** | **53** | **50** | **44** | **5** | **152** |
