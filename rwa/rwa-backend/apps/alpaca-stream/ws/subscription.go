package ws

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"go.uber.org/zap"
)

// SubscriptionType represents the type of subscription
type SubscriptionType string

// SubscriptionConfig configures a subscription
type SubscriptionConfig struct {
	Type    SubscriptionType
	Symbols []string
	Handler MessageHandler
}

// SubscriptionManager manages WebSocket subscriptions
type SubscriptionManager struct {
	mu            sync.RWMutex
	client        *Client
	subscriptions map[SubscriptionType]*SubscriptionConfig
	streams       []string
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager(client *Client) *SubscriptionManager {
	return &SubscriptionManager{
		client:        client,
		subscriptions: make(map[SubscriptionType]*SubscriptionConfig),
	}
}

// Subscribe subscribes to a stream type
func (sm *SubscriptionManager) Subscribe(ctx context.Context, config *SubscriptionConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Register handler
	sm.client.RegisterHandler(string(config.Type), config.Handler)

	// Store subscription config
	sm.subscriptions[config.Type] = config

	// Add to streams list if not already present
	if !slices.Contains(sm.streams, string(config.Type)) {
		sm.streams = append(sm.streams, string(config.Type))
	}

	// Subscribe to the stream
	if err := sm.client.Subscribe(ctx, sm.streams); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", config.Type, err)
	}

	log.InfoZ(ctx, "Subscribed to stream",
		zap.String("type", string(config.Type)),
		zap.Strings("symbols", config.Symbols))

	return nil
}

// Unsubscribe unsubscribes from a stream type
func (sm *SubscriptionManager) Unsubscribe(ctx context.Context, subscriptionType SubscriptionType) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Remove from subscriptions
	delete(sm.subscriptions, subscriptionType)

	// Remove from streams list
	newStreams := []string{}
	for _, stream := range sm.streams {
		if stream != string(subscriptionType) {
			newStreams = append(newStreams, stream)
		}
	}
	sm.streams = newStreams

	// Update subscription
	if len(sm.streams) > 0 {
		if err := sm.client.Subscribe(ctx, sm.streams); err != nil {
			return fmt.Errorf("failed to update subscriptions: %w", err)
		}
	} else {
		// Unsubscribe from all
		if err := sm.client.Unsubscribe(ctx, []string{}); err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
	}

	log.InfoZ(ctx, "Unsubscribed from stream", zap.String("type", string(subscriptionType)))
	return nil
}

// GetSubscription returns a subscription config
func (sm *SubscriptionManager) GetSubscription(subscriptionType SubscriptionType) *SubscriptionConfig {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.subscriptions[subscriptionType]
}

// GetAllSubscriptions returns all subscriptions
func (sm *SubscriptionManager) GetAllSubscriptions() map[SubscriptionType]*SubscriptionConfig {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	// Return a copy to avoid races on the returned map
	result := make(map[SubscriptionType]*SubscriptionConfig, len(sm.subscriptions))
	for k, v := range sm.subscriptions {
		result[k] = v
	}
	return result
}

// Resubscribe 重新订阅所有已注册的流，用于 WebSocket 重连后恢复订阅
func (sm *SubscriptionManager) Resubscribe(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.streams) == 0 {
		log.InfoZ(ctx, "Resubscribe: no streams to resubscribe")
		return nil
	}

	// 重新注册所有 handler（重连后 client 的 handler map 仍然保留，无需重新注册）
	// 只需要重新发送 listen 请求
	if err := sm.client.Subscribe(ctx, sm.streams); err != nil {
		return fmt.Errorf("failed to resubscribe to streams %v: %w", sm.streams, err)
	}

	log.InfoZ(ctx, "Resubscribe: successfully resubscribed to all streams",
		zap.Strings("streams", sm.streams))
	return nil
}
