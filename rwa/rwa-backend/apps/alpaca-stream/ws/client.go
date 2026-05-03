package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/alpaca-stream/constants"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Client represents an Alpaca WebSocket client
type Client struct {
	conn                 *websocket.Conn
	apiKey               string
	apiSecret            string
	url                  string
	mu                   sync.RWMutex // protects conn, isAuthenticated, messageHandlers, onError, onReconnect
	writeMu              sync.Mutex   // protects all WebSocket write operations
	isAuthenticated      bool
	messageHandlers      map[string][]MessageHandler
	onError              func(error)
	onReconnect          func(ctx context.Context) // 重连成功后的回调，用于重新订阅
	reconnectDelay       time.Duration
	maxReconnectDelay    time.Duration
	reconnectAttempts    int
	maxReconnectAttempts int
	ctx                  context.Context
	cancel               context.CancelFunc
	connCtx              context.Context    // per-connection context, cancelled on disconnect
	connCancel           context.CancelFunc // cancels connCtx
	reconnecting         int32              // atomic flag to prevent concurrent reconnects
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, message json.RawMessage) error

// messageFormat 表示检测到的 WebSocket 消息格式
type messageFormat int

const (
	formatUnknown        messageFormat = iota
	formatArrayControl                 // [{"T":"...",...}] 控制消息或市场数据消息
	formatArrayStream                  // ["stream_name", {...}] 流消息
	formatObjectStandard               // {"stream":"...",...} 标准对象消息
)

// NewClient creates a new WebSocket client
func NewClient(apiKey, apiSecret, url string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		apiKey:               apiKey,
		apiSecret:            apiSecret,
		url:                  url,
		messageHandlers:      make(map[string][]MessageHandler),
		reconnectDelay:       time.Duration(constants.DefaultReconnectDelay) * time.Second,
		maxReconnectDelay:    time.Duration(constants.DefaultMaxReconnectDelay) * time.Second,
		maxReconnectAttempts: constants.DefaultMaxReconnectAttempts,
		ctx:                  ctx,
		cancel:               cancel,
	}
}

// Connect connects to the WebSocket server
func (c *Client) Connect(ctx context.Context) error {
	// Check if already connected (short lock)
	c.mu.RLock()
	if c.conn != nil {
		c.mu.RUnlock()
		return errors.New("client already connected")
	}
	c.mu.RUnlock()

	dialer := websocket.Dialer{
		HandshakeTimeout: time.Duration(constants.WriteDeadline) * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to dial websocket: %w", err)
	}

	// Create per-connection context
	connCtx, connCancel := context.WithCancel(c.ctx)

	// Set state under lock
	c.mu.Lock()
	c.conn = conn
	c.isAuthenticated = false
	c.connCtx = connCtx
	c.connCancel = connCancel
	c.mu.Unlock()

	// 设置 Pong handler：收到 pong 时延长 ReadDeadline，保持连接活跃
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(time.Duration(constants.ReadDeadline) * time.Second))
		return nil
	})

	// 启动 ping 定时器，定期发送 ping 保活
	go c.pingLoop(connCtx)

	// Start reading messages
	go c.readMessages(connCtx)

	// Authenticate (no lock held, writeJSON uses writeMu internally)
	if err := c.authenticate(ctx); err != nil {
		c.mu.Lock()
		c.conn = nil
		c.isAuthenticated = false
		c.mu.Unlock()
		connCancel()
		conn.Close()
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.mu.Lock()
	c.reconnectAttempts = 0
	c.mu.Unlock()

	log.InfoZ(ctx, "WebSocket connected and authenticated", zap.String("url", c.url))
	return nil
}

// authenticate sends authentication message
func (c *Client) authenticate(ctx context.Context) error {
	authMsg := map[string]string{
		"action": "auth",
		"key":    c.apiKey,
		"secret": c.apiSecret,
	}

	if err := c.writeJSON(ctx, authMsg); err != nil {
		return err
	}

	// Wait for authentication response (with timeout)
	// The authentication response will be handled in handleMessage
	timeout := time.NewTimer(time.Duration(constants.AuthTimeout) * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer timeout.Stop()
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			c.mu.RLock()
			authenticated := c.isAuthenticated
			c.mu.RUnlock()
			if !authenticated {
				return fmt.Errorf("authentication timeout")
			}
			return nil
		case <-ticker.C:
			c.mu.RLock()
			authenticated := c.isAuthenticated
			c.mu.RUnlock()
			if authenticated {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Subscribe subscribes to a stream
func (c *Client) Subscribe(ctx context.Context, streams []string) error {
	c.mu.RLock()
	authenticated := c.isAuthenticated
	c.mu.RUnlock()

	if !authenticated {
		return errors.New("client not authenticated")
	}

	msg := map[string]any{
		"action": "listen",
		"data": map[string]any{
			"streams": streams,
		},
	}

	log.InfoZ(ctx, "Sending subscription message", zap.Any("streams", streams), zap.Any("message", msg))
	if err := c.writeJSON(ctx, msg); err != nil {
		log.ErrorZ(ctx, "Failed to send subscription message", zap.Error(err))
		return err
	}
	log.InfoZ(ctx, "Subscription message sent successfully")
	return nil
}

// Unsubscribe unsubscribes from a stream
func (c *Client) Unsubscribe(ctx context.Context, streams []string) error {
	c.mu.RLock()
	authenticated := c.isAuthenticated
	c.mu.RUnlock()

	if !authenticated {
		return errors.New("client not authenticated")
	}

	// To unsubscribe, send an empty list for those streams
	// First, get current subscriptions and remove the ones we want to unsubscribe from
	msg := map[string]any{
		"action": "listen",
		"data": map[string]any{
			"streams": []string{}, // Empty list means unsubscribe from all
		},
	}

	return c.writeJSON(ctx, msg)
}

// RegisterHandler registers a message handler for a specific stream
func (c *Client) RegisterHandler(stream string, handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.messageHandlers[stream] = append(c.messageHandlers[stream], handler)
}

// SetErrorHandler sets the error handler
func (c *Client) SetErrorHandler(handler func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onError = handler
}

// SetReconnectHandler 设置重连成功后的回调，用于重新订阅等操作
func (c *Client) SetReconnectHandler(handler func(ctx context.Context)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onReconnect = handler
}

// writeJSON writes a JSON message
func (c *Client) writeJSON(ctx context.Context, v any) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return errors.New("connection not established")
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	conn.SetWriteDeadline(time.Now().Add(time.Duration(constants.WriteDeadline) * time.Second))
	return conn.WriteJSON(v)
}

// readMessages reads messages from the WebSocket
func (c *Client) readMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			conn.SetReadDeadline(time.Now().Add(time.Duration(constants.ReadDeadline) * time.Second))
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.ErrorZ(ctx, "WebSocket read error", zap.Error(err))
					c.handleError(err)
				}
				// Try to reconnect
				c.reconnect(ctx)
				return
			}

			if err := c.handleMessage(ctx, message); err != nil {
				log.ErrorZ(ctx, "Failed to handle message", zap.Error(err))
			}
		}
	}
}

// detectMessageFormat 通过轻量级检查确定 WebSocket 消息的格式
//
// Alpaca WebSocket API 消息格式说明：
//
//  1. formatArrayControl - 市场数据流消息格式（Market Data API）
//     [{"T":"b","S":"AAPL","o":150.25,"c":150.30,...}]           // K线数据
//     [{"T":"q","S":"AAPL","bp":150.25,"ap":150.30,...}]         // 报价数据
//     [{"T":"t","S":"AAPL","p":150.25,"v":100,...}]              // 交易数据
//     [{"T":"success","message":"authenticated"}]                // 认证成功（市场数据API）
//     [{"T":"error","message":"..."}]                            // 错误消息（市场数据API）
//     [{"T":"subscription","bars":["AAPL"],"trades":["TSLA"]}]   // 订阅确认
//
//  2. formatArrayStream - 流消息格式（部分API响应）
//     ["trade_updates", {"id":"123","event":"new","status":"new"}]
//
//  3. formatObjectStandard - 标准对象格式（Trading API）
//     {"stream":"authorization","data":{"status":"authorized"}}  // 认证响应（交易API）
//     {"stream":"listening","data":{"streams":["trade_updates"]}} // 订阅确认（交易API）
//     {"stream":"trade_updates","data":{"event":"fill",...}}     // 订单更新
//
// 检测策略：
// - 首字符 '[' → 尝试区分数组格式类型
// - 首字符 '{' → 标准对象格式
// - 其他 → formatUnknown
func detectMessageFormat(data []byte) messageFormat {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return formatUnknown
	}

	// 检查首字符快速判断数组或对象
	if data[0] == '[' {
		// 尝试区分数组格式类型
		var arrMsg []any
		if err := json.Unmarshal(data, &arrMsg); err != nil {
			return formatUnknown
		}

		if len(arrMsg) == 0 {
			return formatUnknown
		}

		// 检查是否为 [{"T":"...",...}] 格式（控制消息或市场数据）
		if firstItem, ok := arrMsg[0].(map[string]any); ok {
			if _, hasType := firstItem[constants.FieldType]; hasType {
				return formatArrayControl
			}
		}

		// 检查是否为 ["stream_name", {...}] 格式
		if len(arrMsg) >= 2 {
			if _, isString := arrMsg[0].(string); isString {
				return formatArrayStream
			}
		}

		return formatUnknown
	}

	if data[0] == '{' {
		return formatObjectStandard
	}

	return formatUnknown
}

// mapMessageTypeToStream 将市场数据消息类型（T 字段）转换为流名称
func mapMessageTypeToStream(msgType string) string {
	switch msgType {
	case constants.MessageTypeBars:
		return constants.StreamTypeBars
	case constants.MessageTypeQuotes:
		return constants.StreamTypeQuotes
	case constants.MessageTypeTrades:
		return constants.StreamTypeTrades
	default:
		return ""
	}
}

// extractStringSlice 从 []any 类型中提取字符串切片
func extractStringSlice(arr []any) []string {
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// handleControlMessage 处理数组格式的 Alpaca 控制消息
// 返回 true 表示消息是控制消息（已处理），false 表示不是控制消息
func (c *Client) handleControlMessage(ctx context.Context, arrMsg []any) (bool, error) {
	if len(arrMsg) == 0 {
		return false, nil
	}

	firstItem, ok := arrMsg[0].(map[string]any)
	if !ok {
		return false, nil
	}

	msgType, ok := firstItem[constants.FieldType].(string)
	if !ok {
		return false, nil
	}

	// 处理成功控制消息
	if msgType == constants.MessageTypeSuccess {
		msg, _ := firstItem[constants.FieldMessage].(string)
		if msg == constants.MessageAuthenticated {
			c.mu.Lock()
			wasAuthenticated := c.isAuthenticated
			c.isAuthenticated = true
			c.mu.Unlock()
			if !wasAuthenticated {
				log.InfoZ(ctx, "WebSocket authenticated successfully (market data API format)")
			} else {
				log.DebugZ(ctx, "Received authenticated confirmation (already authenticated)")
			}
		} else {
			log.DebugZ(ctx, "Received Alpaca control message",
				zap.String("type", msgType),
				zap.String("message", msg),
			)
		}
		return true, nil
	}

	// 处理错误控制消息
	if msgType == constants.MessageTypeError {
		errMsg, _ := firstItem[constants.FieldMessage].(string)
		err := fmt.Errorf("authentication error: %s", errMsg)
		c.handleError(err)
		return true, err
	}

	// 检查是否为市场数据消息
	streamType := mapMessageTypeToStream(msgType)
	if streamType != "" {
		log.DebugZ(ctx, "Received market data message",
			zap.String("stream_type", streamType),
		)
		return false, nil // 不是控制消息，由市场数据处理器处理
	}

	return false, nil
}

// handleSubscriptionMessage 处理订阅确认消息 [{"T":"subscription","bars":["AAPL","TSLA"]}]
func (c *Client) handleSubscriptionMessage(ctx context.Context, data map[string]any) {
	if bars, ok := data["bars"].([]any); ok {
		symbols := extractStringSlice(bars)
		log.InfoZ(ctx, "Market data subscription confirmed",
			zap.Strings("bars", symbols),
		)
	}
	if trades, ok := data["trades"].([]any); ok {
		symbols := extractStringSlice(trades)
		log.InfoZ(ctx, "Market data subscription confirmed",
			zap.Strings("trades", symbols),
		)
	}
	if quotes, ok := data["quotes"].([]any); ok {
		symbols := extractStringSlice(quotes)
		log.InfoZ(ctx, "Market data subscription confirmed",
			zap.Strings("quotes", symbols),
		)
	}
}

// dispatchToHandlers 将原始消息分发到流的所有已注册处理器
func (c *Client) dispatchToHandlers(ctx context.Context, stream string, message json.RawMessage, logPreview ...string) error {
	c.mu.RLock()
	handlers := c.messageHandlers[stream]
	c.mu.RUnlock()

	if len(handlers) == 0 {
		if len(logPreview) > 0 {
			log.DebugZ(ctx, "Received message but no handler registered",
				zap.String("stream", stream),
				zap.String("raw_message_preview", logPreview[0]),
			)
		} else {
			previewLen := min(len(message), 200)
			log.DebugZ(ctx, "Received message but no handler registered",
				zap.String("stream", stream),
				zap.String("raw_message_preview", string(message[:previewLen])),
			)
		}
		return nil
	}

	log.DebugZ(ctx, "Dispatching message to handlers",
		zap.String("stream", stream),
		zap.Int("handler_count", len(handlers)),
	)

	for _, handler := range handlers {
		if err := handler(ctx, message); err != nil {
			log.ErrorZ(ctx, "Handler execution error",
				zap.String("stream", stream),
				zap.Error(err))
		}
	}

	return nil
}

// handleMessage handles incoming messages
func (c *Client) handleMessage(ctx context.Context, message []byte) error {
	format := detectMessageFormat(message)

	switch format {
	case formatArrayControl:
		return c.handleArrayControlFormat(ctx, message)

	case formatArrayStream:
		return c.handleArrayStreamFormat(ctx, message)

	case formatObjectStandard:
		return c.handleObjectFormat(ctx, message)

	default:
		log.DebugZ(ctx, "Received message with unknown format",
			zap.String("raw_message", string(message)),
		)
		return nil
	}
}

// handleArrayControlFormat 处理数组格式的控制消息和市场数据消息
func (c *Client) handleArrayControlFormat(ctx context.Context, message []byte) error {
	var arrMsg []any
	if err := json.Unmarshal(message, &arrMsg); err != nil {
		return fmt.Errorf("failed to unmarshal array control message: %w", err)
	}

	// 尝试作为控制消息处理
	handled, err := c.handleControlMessage(ctx, arrMsg)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	// 处理market data
	if len(arrMsg) > 0 {
		if firstItem, ok := arrMsg[0].(map[string]any); ok {
			if msgType, ok := firstItem[constants.FieldType].(string); ok {
				// 处理订阅确认消息
				if msgType == constants.MessageTypeSubscription {
					c.handleSubscriptionMessage(ctx, firstItem)
					return nil
				}
				// 处理市场数据消息 - 转换 Alpaca 格式为 Handler 期望的格式
				streamType := mapMessageTypeToStream(msgType)
				if streamType != "" {
					// 将 [{"T":"b","S":"AAPL",...}] 转换为 {"stream":"bars","data":[{"S":"AAPL",...}]}
					wrappedMsg := map[string]any{
						constants.FieldStream: streamType,
						constants.FieldData:   arrMsg,
					}
					rawMessage, _ := json.Marshal(wrappedMsg)
					return c.dispatchToHandlers(ctx, streamType, json.RawMessage(rawMessage))
				}
			}
		}
	}

	log.DebugZ(ctx, "Received array message with unexpected structure",
		zap.String("raw_message", string(message)),
		zap.Any("data", arrMsg),
	)
	return nil
}

// handleArrayStreamFormat 处理 ["stream_name", {...}] 格式的消息
func (c *Client) handleArrayStreamFormat(ctx context.Context, message []byte) error {
	var arrMsg []any
	if err := json.Unmarshal(message, &arrMsg); err != nil {
		return fmt.Errorf("failed to unmarshal array stream message: %w", err)
	}

	if len(arrMsg) >= 2 {
		if stream, ok := arrMsg[0].(string); ok {
			// 包装为对象格式以保持与现有处理器的兼容性
			msgObj := map[string]any{
				constants.FieldStream: stream,
				constants.FieldData:   arrMsg[1],
			}
			rawMessage, _ := json.Marshal(msgObj)
			return c.dispatchToHandlers(ctx, stream, json.RawMessage(rawMessage))
		}
	}

	return nil
}

// handleObjectFormat 处理标准对象格式消息
func (c *Client) handleObjectFormat(ctx context.Context, message []byte) error {
	var msg map[string]any
	if err := json.Unmarshal(message, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal object message: %w", err)
	}

	stream, ok := msg[constants.FieldStream].(string)
	if !ok {
		// 处理错误消息
		if action, ok := msg[constants.FieldAction].(string); ok && action == constants.ActionError {
			if data, ok := msg[constants.FieldData].(map[string]any); ok {
				if errMsg, ok := data[constants.FieldErrorMessage].(string); ok {
					err := fmt.Errorf("server error: %s", errMsg)
					c.handleError(err)
					return err
				}
			}
		}
		return nil
	}

	// 处理授权流
	if stream == constants.StreamNameAuthorization {
		return c.handleAuthorizationStream(ctx, msg)
	}

	// 处理订阅确认流
	if stream == constants.StreamNameListening {
		return c.handleListeningStream(ctx, msg)
	}

	// 分发到其他流的处理器
	return c.dispatchToHandlers(ctx, stream, json.RawMessage(message))
}

// handleAuthorizationStream 处理授权流消息
func (c *Client) handleAuthorizationStream(ctx context.Context, msg map[string]any) error {
	if data, ok := msg[constants.FieldData].(map[string]any); ok {
		if status, ok := data[constants.FieldStatus].(string); ok {
			c.mu.Lock()
			c.isAuthenticated = (status == constants.StatusAuthorized)
			authenticated := c.isAuthenticated
			c.mu.Unlock()

			if authenticated {
				log.InfoZ(ctx, "WebSocket authenticated successfully")
			} else {
				err := errors.New("authentication failed: unauthorized")
				c.handleError(err)
				return err
			}
		}
	}
	return nil
}

// handleListeningStream 处理订阅确认流消息
func (c *Client) handleListeningStream(ctx context.Context, msg map[string]any) error {
	if data, ok := msg[constants.FieldData].(map[string]any); ok {
		if streams, ok := data[constants.FieldStreams].([]any); ok {
			log.InfoZ(ctx, "Subscription confirmed: successfully subscribed to streams", zap.Any("streams", streams))
			for _, s := range streams {
				if sStr, ok := s.(string); ok && sStr == constants.StreamTypeTradeUpdates {
					log.InfoZ(ctx, "Trade updates stream subscribed successfully! Now listening to all order events: new, fill, partial_fill, canceled, etc.")
				}
			}
		}
	}
	return nil
}

// handleStreamMessage handles stream messages in array format [stream, data]
func (c *Client) handleStreamMessage(ctx context.Context, stream string, data any) error {
	// 包装为对象格式以保持与现有处理器的兼容性
	msgObj := map[string]any{
		constants.FieldStream: stream,
		constants.FieldData:   data,
	}
	rawMessage, _ := json.Marshal(msgObj)
	return c.dispatchToHandlers(ctx, stream, json.RawMessage(rawMessage))
}

// handleMarketDataMessage handles Alpaca market data messages in array format [{"T":"b","S":"AAPL",...}]
func (c *Client) handleMarketDataMessage(ctx context.Context, streamType string, message []byte) error {
	return c.dispatchToHandlers(ctx, streamType, json.RawMessage(message))
}

// reconnect attempts to reconnect to the WebSocket
func (c *Client) reconnect(ctx context.Context) {
	// Prevent concurrent reconnects using atomic flag
	if !atomic.CompareAndSwapInt32(&c.reconnecting, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&c.reconnecting, 0)

	c.mu.RLock()
	maxAttempts := c.maxReconnectAttempts
	attempts := c.reconnectAttempts
	c.mu.RUnlock()

	if maxAttempts >= 0 && attempts >= maxAttempts {
		log.ErrorZ(ctx, "Max reconnect attempts reached")
		return
	}

	// Cancel old per-connection context to stop old pingLoop/readMessages
	c.mu.Lock()
	if c.connCancel != nil {
		c.connCancel()
	}
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.isAuthenticated = false
	delay := min(c.reconnectDelay*time.Duration(1<<uint(c.reconnectAttempts)), c.maxReconnectDelay)
	c.reconnectAttempts++
	attempt := c.reconnectAttempts
	c.mu.Unlock()

	log.InfoZ(ctx, "Reconnecting to WebSocket",
		zap.Int("attempt", attempt),
		zap.Duration("delay", delay))

	time.Sleep(delay)

	if err := c.Connect(c.ctx); err != nil {
		log.ErrorZ(c.ctx, "Reconnection failed", zap.Error(err))
		return
	}

	// 重连成功后调用回调，用于重新订阅
	c.mu.RLock()
	reconnectHandler := c.onReconnect
	c.mu.RUnlock()

	if reconnectHandler != nil {
		log.InfoZ(c.ctx, "Reconnection successful, invoking reconnect handler to resubscribe")
		reconnectHandler(c.ctx)
	}
}

// pingLoop 定期发送 ping 帧保持连接活跃，防止非交易时段因无消息导致超时断连
func (c *Client) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(constants.PingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			c.writeMu.Lock()
			conn.SetWriteDeadline(time.Now().Add(time.Duration(constants.WriteDeadline) * time.Second))
			err := conn.WriteMessage(websocket.PingMessage, nil)
			c.writeMu.Unlock()

			if err != nil {
				log.ErrorZ(ctx, "Failed to send ping", zap.Error(err))
				return
			}
		}
	}
}

// handleError handles errors
func (c *Client) handleError(err error) {
	c.mu.RLock()
	handler := c.onError
	c.mu.RUnlock()

	if handler != nil {
		handler(err)
	}
}

// Close closes the WebSocket connection
func (c *Client) Close() error {
	c.cancel()
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connCancel != nil {
		c.connCancel()
	}

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.isAuthenticated = false
		return err
	}
	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && c.isAuthenticated
}

// GetConnection returns the underlying WebSocket connection
func (c *Client) GetConnection() *websocket.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// WriteJSON writes a JSON message (public method for custom messages)
func (c *Client) WriteJSON(ctx context.Context, v any) error {
	return c.writeJSON(ctx, v)
}
