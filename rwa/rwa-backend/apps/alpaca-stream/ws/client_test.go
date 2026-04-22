package ws

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name                     string
		apiKey                   string
		apiSecret                string
		url                      string
		wantReconnectDelay       time.Duration
		wantMaxReconnectDelay    time.Duration
		wantMaxReconnectAttempts int
	}{
		{
			name:                     "default values with typical inputs",
			apiKey:                   "test-key",
			apiSecret:                "test-secret",
			url:                      "wss://example.com/stream",
			wantReconnectDelay:       1 * time.Second,
			wantMaxReconnectDelay:    30 * time.Second,
			wantMaxReconnectAttempts: -1,
		},
		{
			name:                     "empty credentials",
			apiKey:                   "",
			apiSecret:                "",
			url:                      "",
			wantReconnectDelay:       1 * time.Second,
			wantMaxReconnectDelay:    30 * time.Second,
			wantMaxReconnectAttempts: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(tt.apiKey, tt.apiSecret, tt.url)

			if c == nil {
				t.Fatal("NewClient() returned nil")
			}
			if c.apiKey != tt.apiKey {
				t.Errorf("apiKey = %q, want %q", c.apiKey, tt.apiKey)
			}
			if c.apiSecret != tt.apiSecret {
				t.Errorf("apiSecret = %q, want %q", c.apiSecret, tt.apiSecret)
			}
			if c.url != tt.url {
				t.Errorf("url = %q, want %q", c.url, tt.url)
			}
			if c.reconnectDelay != tt.wantReconnectDelay {
				t.Errorf("reconnectDelay = %v, want %v", c.reconnectDelay, tt.wantReconnectDelay)
			}
			if c.maxReconnectDelay != tt.wantMaxReconnectDelay {
				t.Errorf("maxReconnectDelay = %v, want %v", c.maxReconnectDelay, tt.wantMaxReconnectDelay)
			}
			if c.maxReconnectAttempts != tt.wantMaxReconnectAttempts {
				t.Errorf("maxReconnectAttempts = %d, want %d", c.maxReconnectAttempts, tt.wantMaxReconnectAttempts)
			}
			if c.messageHandlers == nil {
				t.Error("messageHandlers should be initialized, got nil")
			}
			if c.conn != nil {
				t.Error("conn should be nil on new client")
			}
			if c.isAuthenticated {
				t.Error("isAuthenticated should be false on new client")
			}
			if c.ctx == nil {
				t.Error("ctx should not be nil")
			}
			if c.cancel == nil {
				t.Error("cancel should not be nil")
			}

			// Clean up
			c.Close()
		})
	}
}

func TestIsConnected(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{
			name: "initial state is not connected",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("key", "secret", "wss://example.com")
			defer c.Close()

			got := c.IsConnected()
			if got != tt.want {
				t.Errorf("IsConnected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "close without connection returns nil",
			wantErr: false,
		},
		{
			name:    "close twice without connection returns nil",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("key", "secret", "wss://example.com")

			err := c.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}

			// After close, IsConnected should return false
			if c.IsConnected() {
				t.Error("IsConnected() should be false after Close()")
			}

			// conn should be nil
			if c.conn != nil {
				t.Error("conn should be nil after Close()")
			}

			// isAuthenticated should be false
			if c.isAuthenticated {
				t.Error("isAuthenticated should be false after Close()")
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name         string
		stream       string
		handlerCount int
	}{
		{
			name:         "register single handler",
			stream:       "trade_updates",
			handlerCount: 1,
		},
		{
			name:         "register multiple handlers for same stream",
			stream:       "trade_updates",
			handlerCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("key", "secret", "wss://example.com")
			defer c.Close()

			for i := 0; i < tt.handlerCount; i++ {
				c.RegisterHandler(tt.stream, func(_ context.Context, _ json.RawMessage) error {
					return nil
				})
			}

			c.mu.RLock()
			handlers := c.messageHandlers[tt.stream]
			c.mu.RUnlock()

			if len(handlers) != tt.handlerCount {
				t.Errorf("handler count = %d, want %d", len(handlers), tt.handlerCount)
			}
		})
	}
}

func TestSetErrorHandler(t *testing.T) {
	c := NewClient("key", "secret", "wss://example.com")
	defer c.Close()

	if c.onError != nil {
		t.Error("onError should be nil initially")
	}

	c.SetErrorHandler(func(err error) {})

	c.mu.RLock()
	hasHandler := c.onError != nil
	c.mu.RUnlock()

	if !hasHandler {
		t.Error("onError should be set after SetErrorHandler()")
	}
}

func TestSetReconnectHandler(t *testing.T) {
	c := NewClient("key", "secret", "wss://example.com")
	defer c.Close()

	if c.onReconnect != nil {
		t.Error("onReconnect should be nil initially")
	}

	c.SetReconnectHandler(func(_ context.Context) {})

	c.mu.RLock()
	hasHandler := c.onReconnect != nil
	c.mu.RUnlock()

	if !hasHandler {
		t.Error("onReconnect should be set after SetReconnectHandler()")
	}
}
