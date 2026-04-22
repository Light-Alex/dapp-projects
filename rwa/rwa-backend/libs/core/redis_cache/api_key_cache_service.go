package redis_cache

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const ApiKeyHashKey = "mm_api_keys"

type ApiKeyInfo struct {
	ApiKey     string   `json:"api_key"`
	SecretKey  string   `json:"secret_key"`
	Passphrase string   `json:"passphrase"`
	WhiteList  []string `json:"white_list"`
	Remark     string   `json:"remark"`
}

type ApiKeyCacheService struct {
	redisClient redis.UniversalClient
}

func NewApiKeyCacheService(redisClient redis.UniversalClient) *ApiKeyCacheService {
	return &ApiKeyCacheService{
		redisClient: redisClient,
	}
}

func (s *ApiKeyCacheService) SetCache(ctx context.Context, apiKey string, cacheData *ApiKeyInfo) error {
	data, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}
	return s.redisClient.HSet(ctx, ApiKeyHashKey, apiKey, data).Err()
}

func (s *ApiKeyCacheService) GetCache(ctx context.Context, apiKey string) (*ApiKeyInfo, error) {
	result := s.redisClient.HGet(ctx, ApiKeyHashKey, apiKey)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var apiKeyInfo ApiKeyInfo
	if err := json.Unmarshal([]byte(result.Val()), &apiKeyInfo); err != nil {
		return nil, err
	}
	return &apiKeyInfo, nil
}

func (s *ApiKeyCacheService) GetAllCache(ctx context.Context) ([]*ApiKeyInfo, error) {
	result := s.redisClient.HGetAll(ctx, ApiKeyHashKey)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var apiKeyInfos []*ApiKeyInfo
	for _, data := range result.Val() {
		var apiKeyInfo ApiKeyInfo
		if err := json.Unmarshal([]byte(data), &apiKeyInfo); err != nil {
			return nil, err
		}
		apiKeyInfos = append(apiKeyInfos, &apiKeyInfo)
	}
	return apiKeyInfos, nil
}

func (s *ApiKeyCacheService) DeleteCache(ctx context.Context, apiKey string) error {
	return s.redisClient.HDel(ctx, ApiKeyHashKey, apiKey).Err()
}
