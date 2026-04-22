package ws

import (
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
	conn            *websocket.Conn
	apiKey          string
	apiSecret       string
	url             string
	mu              sync.RWMutex  // protects conn, isAuthenticated, messageHandlers, onError, onReconnect
	writeMu         sync.Mutex    // protects all WebSocket write operations
	isAuthenticated bool
	messageHandlers map[string][]MessageHandler
	onError         func(error)
	onReconnect     func(ctx context.Context) // 重连成功后的回调，用于重新订阅
	reconnectDelay      time.Duration
	maxReconnectDelay   time.Duration
	reconnectAttempts   int
	maxReconnectAttempts int
	ctx        context.Context
	cancel     context.CancelFunc
	connCtx    context.Context    // per-connection context, cancelled on disconnect
	connCancel context.CancelFunc // cancels connCtx
	reconnecting int32            // atomic flag to prevent concurrent reconnects
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, message json.RawMessage) error

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

// handleMessage handles incoming messages
func (c *Client) handleMessage(ctx context.Context, message []byte) error {
	var msg map[string]any
	if err := json.Unmarshal(message, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	stream, ok := msg["stream"].(string)
	if !ok {
		// Handle error messages
		if action, ok := msg["action"].(string); ok && action == "error" {
			if data, ok := msg["data"].(map[string]any); ok {
				if errMsg, ok := data["error_message"].(string); ok {
					err := fmt.Errorf("server error: %s", errMsg)
					c.handleError(err)
					return err
				}
			}
		}
		return nil
	}

	// Handle authorization stream
	if stream == "authorization" {
		if data, ok := msg["data"].(map[string]any); ok {
			if status, ok := data["status"].(string); ok {
				c.mu.Lock()
				c.isAuthenticated = (status == "authorized")
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

	// Handle listening stream (subscription confirmation)
	if stream == "listening" {
		if data, ok := msg["data"].(map[string]any); ok {
			if streams, ok := data["streams"].([]any); ok {
				log.InfoZ(ctx, "Subscription confirmed: successfully subscribed to streams", zap.Any("streams", streams))
				// Special message for trade_updates
				for _, s := range streams {
					if sStr, ok := s.(string); ok && sStr == constants.StreamTypeTradeUpdates {
						log.InfoZ(ctx, "Trade updates stream subscribed successfully! Now listening to all order events: new, fill, partial_fill, canceled, etc.")
					}
				}
			}
		}
		return nil
	}

	// Handle other streams
	c.mu.RLock()
	handlers := c.messageHandlers[stream]
	c.mu.RUnlock()

	if len(handlers) == 0 {
		previewLen := min(len(message), 200)
		log.DebugZ(ctx, "Received message but no handler registered",
			zap.String("stream", stream),
			zap.String("raw_message_preview", string(message[:previewLen])),
		)
		return nil
	}

	log.DebugZ(ctx, "Received message, processing",
		zap.String("stream", stream),
		zap.Int("handler_count", len(handlers)),
	)

	rawMessage := json.RawMessage(message)
	for _, handler := range handlers {
		if err := handler(ctx, rawMessage); err != nil {
			log.ErrorZ(ctx, "Handler execution error", zap.String("stream", stream), zap.Error(err))
		}
	}

	return nil
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
