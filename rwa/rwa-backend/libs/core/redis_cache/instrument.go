package redis_cache

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/AnchoredLabs/rwa-backend/libs/errors"
	"github.com/AnchoredLabs/rwa-backend/libs/grpc/cache_definition"
	"github.com/redis/go-redis/v9"
)

type InstrumentCacheService struct {
	redisClient redis.UniversalClient
}

func NewInstrumentCacheService(redisClient redis.UniversalClient) *InstrumentCacheService {
	return &InstrumentCacheService{
		redisClient: redisClient,
	}
}

// SetCache sets the instrument cache for a specific chain ID and instrument.
func (o *InstrumentCacheService) SetCache(ctx context.Context, chainId uint64, instrument string, instrumentData *cache_definition.InstrumentStateDetailsDot) error {

	jsonData, err := json.Marshal(instrumentData)
	if err != nil {
		return errors.Annotate(err, "failed to marshal instrument data")
	}
	return o.redisClient.HSet(ctx, InstrumentKeyPrefix+strconv.FormatUint(chainId, 10), instrument, jsonData).Err()
}

// GetCache gets the instrument cache for a specific chain ID and instrument.
func (o *InstrumentCacheService) GetCache(ctx context.Context, chainId uint64, instrument string) (*cache_definition.InstrumentStateDetailsDot, error) {
	result := o.redisClient.HGet(ctx, InstrumentKeyPrefix+strconv.FormatUint(chainId, 10), instrument)
	if result.Err() != nil {
		if errors.Is(result.Err(), redis.Nil) {
			return nil, errors.NotFoundf("instrument not found") // Key doesn't exist
		}
		return nil, result.Err()
	}
	jsonData := result.Val()
	if jsonData == "" {
		return nil, nil
	}
	var instrumentData cache_definition.InstrumentStateDetailsDot
	if err := json.Unmarshal([]byte(jsonData), &instrumentData); err != nil {
		return nil, errors.Annotate(err, "failed to unmarshal instrument data")
	}

	return &instrumentData, nil
}
