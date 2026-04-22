package redis_cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type SnapshotPubSub struct {
	client redis.UniversalClient
}

func NewSnapshotPubSub(client redis.UniversalClient) *SnapshotPubSub {
	return &SnapshotPubSub{
		client: client,
	}
}

// Publish publishes a message to a Redis channel
func (r *SnapshotPubSub) Publish(ctx context.Context, data *SnapshotEvent) error {
	channel := ChannelPrefix + PubSubTopicSnapshotChanged
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	err = r.client.Publish(ctx, channel, jsonData).Err()
	if err != nil {
		return fmt.Errorf("failed to publish to redis: %w", err)
	}
	return nil
}
