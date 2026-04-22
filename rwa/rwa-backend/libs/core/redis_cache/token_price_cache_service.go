package redis_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/AnchoredLabs/rwa-backend/libs/grpc/cache_definition"
	"github.com/redis/go-redis/v9"
)

type TokenPriceCacheService struct {
	client redis.UniversalClient
}

func NewTokenPriceCacheService(client redis.UniversalClient) *TokenPriceCacheService {
	return &TokenPriceCacheService{
		client: client,
	}
}

// SetAllCoinCache sets the cache for all coin prices.
func (s *TokenPriceCacheService) SetAllCoinCache(ctx context.Context, coinMap map[string]*cache_definition.TokenPriceCacheBySymbol) error {
	data := make(map[string]interface{})
	for key, value := range coinMap {
		jsonData, err := json.Marshal(value)
		if err != nil {
			return err
		}
		data[key] = string(jsonData)
	}
	return s.client.HMSet(ctx, TokenPriceBySymbolKeyPreFix, data).Err()
}

// GetAllCoinCache gets the cache for all coin prices.
func (s *TokenPriceCacheService) GetAllCoinCache(ctx context.Context) (map[string]*cache_definition.TokenPriceCacheBySymbol, error) {
	result := s.client.HGetAll(ctx, TokenPriceBySymbolKeyPreFix)
	if result.Err() != nil {
		return nil, result.Err()
	}
	if len(result.Val()) == 0 {
		return make(map[string]*cache_definition.TokenPriceCacheBySymbol), nil
	}
	tokenMap := make(map[string]*cache_definition.TokenPriceCacheBySymbol)
	for key, value := range result.Val() {
		// json unmarshal
		var tokenPriceCacheBySymbol *cache_definition.TokenPriceCacheBySymbol
		err := json.Unmarshal([]byte(value), &tokenPriceCacheBySymbol)
		if err != nil {
			return nil, err
		}
		tokenMap[key] = tokenPriceCacheBySymbol
	}
	return tokenMap, nil
}

// SetAllTokenCache sets the cache for all token prices.
func (s *TokenPriceCacheService) SetAllTokenCache(ctx context.Context, chainId uint64, tokenMap map[string]*cache_definition.TokenPriceCacheByAddress) error {
	key := fmt.Sprintf("%s%d", TokenPriceByAddressKeyPreFix, chainId)
	data := make(map[string]interface{})
	for tokenAddress, value := range tokenMap {
		jsonData, err := json.Marshal(value)
		if err != nil {
			return err
		}
		data[tokenAddress] = string(jsonData)
	}
	return s.client.HMSet(ctx, key, data).Err()
}

// GetAllTokenCacheByChainId gets the cache for all token prices.
func (s *TokenPriceCacheService) GetAllTokenCacheByChainId(ctx context.Context, chainId uint64) (map[string]*cache_definition.TokenPriceCacheByAddress, error) {
	key := fmt.Sprintf("%s%d", TokenPriceByAddressKeyPreFix, chainId)
	return s.GetAllTokenCacheByKey(ctx, key)
}

// GetAllTokenCacheByKey gets the cache for all token prices by key.
func (s *TokenPriceCacheService) GetAllTokenCacheByKey(ctx context.Context, key string) (map[string]*cache_definition.TokenPriceCacheByAddress, error) {
	result := s.client.HGetAll(ctx, key)
	if result.Err() != nil {
		return nil, result.Err()
	}
	if len(result.Val()) == 0 {
		return make(map[string]*cache_definition.TokenPriceCacheByAddress), nil
	}
	tokenMap := make(map[string]*cache_definition.TokenPriceCacheByAddress)
	for tokenAddress, value := range result.Val() {
		// json unmarshal
		var tokenPriceCacheByAddress *cache_definition.TokenPriceCacheByAddress
		err := json.Unmarshal([]byte(value), &tokenPriceCacheByAddress)
		if err != nil {
			return nil, err
		}
		tokenMap[tokenAddress] = tokenPriceCacheByAddress
	}
	return tokenMap, nil
}

// GetAllTokenCache gets the cache for all token prices.
func (s *TokenPriceCacheService) GetAllTokenCache(ctx context.Context) (map[uint64]map[string]*cache_definition.TokenPriceCacheByAddress, error) {
	// get all key with prefix TokenPriceByAddressKeyPreFix
	var (
		cursor uint64
		keys   []string
	)
	for {
		var batch []string
		var err error
		batch, cursor, err = s.client.Scan(ctx, cursor, TokenPriceByAddressKeyPreFix+"*", 1000).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)

		if cursor == 0 {
			break
		}
	}
	result := make(map[uint64]map[string]*cache_definition.TokenPriceCacheByAddress)
	for _, key := range keys {
		chainId, err := strconv.Atoi(strings.Split(key, ":")[3])
		if err != nil {
			continue
		}
		res, err := s.GetAllTokenCacheByKey(ctx, key)
		if err != nil {
			continue
		}
		if len(res) == 0 {
			continue
		}
		result[uint64(chainId)] = res
	}
	return result, nil
}
