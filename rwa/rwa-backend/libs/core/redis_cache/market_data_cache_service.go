package redis_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	MarketDataKeyPrefix = "rwa:marketdata:"
	CurrentPriceTTL     = 5 * time.Second
	LatestQuoteTTL      = 5 * time.Second
	SnapshotTTL         = 10 * time.Second
	HistoricalDataTTL   = 60 * time.Second
	MarketClockTTL      = 30 * time.Second
)

type MarketDataCacheService struct {
	client redis.UniversalClient
}

func NewMarketDataCacheService(client redis.UniversalClient) *MarketDataCacheService {
	return &MarketDataCacheService{client: client}
}

func (s *MarketDataCacheService) buildKey(dataType, symbol string) string {
	return fmt.Sprintf("%s%s:%s", MarketDataKeyPrefix, dataType, symbol)
}

// Get retrieves cached data. Returns redis.Nil on cache miss, other errors on failure.
func (s *MarketDataCacheService) Get(ctx context.Context, dataType, symbol string, dest any) error {
	key := s.buildKey(dataType, symbol)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Set caches data with the given TTL.
func (s *MarketDataCacheService) Set(ctx context.Context, dataType, symbol string, value any, ttl time.Duration) error {
	key := s.buildKey(dataType, symbol)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, ttl).Err()
}
