package redis_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/AnchoredLabs/rwa-backend/libs/grpc/cache_definition"
	"github.com/redis/go-redis/v9"
)

type OrderBookCacheService struct {
	redisClient redis.UniversalClient
}

func NewOrderBookCacheService(redisClient redis.UniversalClient) *OrderBookCacheService {
	return &OrderBookCacheService{
		redisClient: redisClient,
	}
}

// SetCache sets the order book cache for a specific chain ID and instrument.
func (o *OrderBookCacheService) SetCache(ctx context.Context, chainId uint64, instrument string, expiry uint32, orderBook map[int32]*cache_definition.OrderBookCacheDto) error {
	key := buildOrderBookKey(chainId, instrument, expiry)

	// Convert map to Redis hash format
	hashData := make(map[string]interface{})
	for tick, dto := range orderBook {
		if dto != nil {
			jsonData, err := json.Marshal(dto)
			if err != nil {
				return fmt.Errorf("failed to marshal orderbook dto: %w", err)
			}
			hashData[strconv.Itoa(int(tick))] = string(jsonData)
		}
	}

	if len(hashData) > 0 {
		return o.redisClient.HSet(ctx, key, hashData).Err()
	}
	return nil
}

// GetCache gets the order book cache for a specific chain ID and instrument.
func (o *OrderBookCacheService) GetCache(ctx context.Context, chainId uint64, instrument string, expiry uint32) (map[int32]*cache_definition.OrderBookCacheDto, error) {
	result := o.redisClient.HGetAll(ctx, buildOrderBookKey(chainId, instrument, expiry))
	if result.Err() != nil {
		return nil, result.Err()
	}

	if len(result.Val()) == 0 {
		return nil, nil
	}

	// Parse the JSON data
	orderBook := make(map[int32]*cache_definition.OrderBookCacheDto)
	for key, value := range result.Val() {
		var tick int32
		if _, err := fmt.Sscanf(key, "%d", &tick); err != nil {
			continue
		}

		var orderBookDto cache_definition.OrderBookCacheDto
		if err := json.Unmarshal([]byte(value), &orderBookDto); err != nil {
			continue
		}

		orderBook[tick] = &orderBookDto
	}

	return orderBook, nil
}

func buildOrderBookKey(chainId uint64, instrument string, expiry uint32) string {
	return OrderBookKeyPrefix + strconv.FormatUint(chainId, 10) + ":" + instrument + ":" + strconv.Itoa(int(expiry))
}
