package redis_cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AnchoredLabs/rwa-backend/libs/grpc/cache_definition"
	"github.com/redis/go-redis/v9"
)

type DepthCacheService struct {
	redisClient redis.UniversalClient
}

func NewDepthCacheService(redisClient redis.UniversalClient) *DepthCacheService {
	return &DepthCacheService{
		redisClient: redisClient,
	}
}

// SetCache sets the depth cache for a specific chain ID and instrument.
func (d *DepthCacheService) SetCache(ctx context.Context, chainId uint64, instrument string, expiry uint32, depthData *cache_definition.OrderBookOrDepthDto) error {
	key := buildDepthKey(chainId, instrument, expiry)

	if depthData == nil {
		return fmt.Errorf("depth data is nil")
	}

	jsonData, err := json.Marshal(depthData)
	if err != nil {
		return fmt.Errorf("failed to marshal depth data: %w", err)
	}

	return d.redisClient.HSet(ctx, key, buildDepthFieldKey(instrument, expiry), string(jsonData)).Err()
}

// GetCache gets the depth cache for a specific chain ID and instrument.
func (d *DepthCacheService) GetCache(ctx context.Context, chainId uint64, instrument string, expiry uint32) (*cache_definition.OrderBookOrDepthDto, error) {
	key := buildDepthKey(chainId, instrument, expiry)
	field := buildDepthFieldKey(instrument, expiry)

	result := d.redisClient.HGet(ctx, key, field)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nil, nil // Key doesn't exist
		}
		return nil, result.Err()
	}

	jsonData := result.Val()
	if jsonData == "" {
		return nil, nil
	}

	var depthData cache_definition.OrderBookOrDepthDto
	if err := json.Unmarshal([]byte(jsonData), &depthData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal depth data: %w", err)
	}

	return &depthData, nil
}

func buildDepthKey(chainId uint64, instrument string, expiry uint32) string {
	return fmt.Sprintf("%s:%d", DepthKeyPrefix, chainId)
}

func buildDepthFieldKey(instrument string, expiry uint32) string {
	return fmt.Sprintf("%s_%d", instrument, expiry)
}
