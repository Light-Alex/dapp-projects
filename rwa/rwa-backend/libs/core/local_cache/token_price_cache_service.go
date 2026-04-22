package local_cache

import (
	"context"
	"sync"
	"time"

	"github.com/AnchoredLabs/rwa-backend/libs/core/redis_cache"
	"github.com/AnchoredLabs/rwa-backend/libs/grpc/cache_definition"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/go-co-op/gocron/v2"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type TokenPriceCacheService struct {
	tokenPriceCacheService *redis_cache.TokenPriceCacheService
	scheduler              gocron.Scheduler
	cacheBySymbol          map[string]*cache_definition.TokenPriceCacheBySymbol             // symbol -> cache
	cacheByAddress         map[uint64]map[string]*cache_definition.TokenPriceCacheByAddress // chainId -> address -> cache
	mu                     sync.Mutex
}

func NewTokenPriceCacheService(tokenPriceCacheService *redis_cache.TokenPriceCacheService) (*TokenPriceCacheService, error) {
	ctx := context.Background()
	s, err := gocron.NewScheduler()
	if err != nil {
		log.ErrorZ(ctx, "failed to create scheduler", zap.Error(err))
		return nil, err
	}
	t := &TokenPriceCacheService{
		tokenPriceCacheService: tokenPriceCacheService,
		scheduler:              s,
		mu:                     sync.Mutex{},
	}
	err = t.updateCache(ctx)
	if err != nil {
		log.ErrorZ(ctx, "failed to update token price cache on startup", zap.Error(err))
		return nil, err
	}
	err = t.initJobs(ctx)
	if err != nil {
		log.ErrorZ(ctx, "failed to initialize background jobs", zap.Error(err))
		return nil, err
	}
	return t, nil
}

// GetAllTokenPriceBySymbol returns the cached token prices by symbol.
func (s *TokenPriceCacheService) GetAllTokenPriceBySymbol() map[string]*cache_definition.TokenPriceCacheBySymbol {
	return s.cacheBySymbol
}

// GetTokenBySymbol returns the cached token price by symbol.
func (s *TokenPriceCacheService) GetTokenBySymbol(symbol string) (*cache_definition.TokenPriceCacheBySymbol, bool) {
	token, exists := s.cacheBySymbol[symbol]
	return token, exists
}

// GetAllTokenPriceByAddress returns the cached token prices by address for a given chain ID.
func (s *TokenPriceCacheService) GetAllTokenPriceByAddress(chainId uint64) (map[string]*cache_definition.TokenPriceCacheByAddress, bool) {
	chainCache, exists := s.cacheByAddress[chainId]
	return chainCache, exists
}

// GetTokenByAddressAndChainId returns the cached token price by address for a given chain ID.
func (s *TokenPriceCacheService) GetTokenByAddressAndChainId(chainId uint64, address string) (*cache_definition.TokenPriceCacheByAddress, bool) {
	chainCache, exists := s.cacheByAddress[chainId]
	if !exists {
		return nil, false
	}
	token, exists := chainCache[address]
	return token, exists
}

// GetPriceBySymbol returns the current price of a token by its symbol.
func (s *TokenPriceCacheService) GetPriceBySymbol(symbol string) decimal.Decimal {
	token, exists := s.cacheBySymbol[symbol]
	if !exists {
		return decimal.Zero
	}
	return decimal.NewFromFloat32(token.CurrentPrice)
}

// GetPriceByAddressAndChainId returns the current price of a token by its address and chain ID.
func (s *TokenPriceCacheService) GetPriceByAddressAndChainId(chainId uint64, address string) decimal.Decimal {
	chainCache, exists := s.cacheByAddress[chainId]
	if !exists {
		return decimal.Zero
	}
	token, exists := chainCache[address]
	if !exists {
		return decimal.Zero
	}
	return decimal.NewFromFloat32(token.CurrentPrice)
}

// initJobs initializes and starts the scheduled background jobs.
func (s *TokenPriceCacheService) initJobs(ctx context.Context) error {
	err := s.addUpdateCacheJob(ctx)
	if err != nil {
		return err
	}
	s.scheduler.Start()
	return nil
}

// addUpdateCacheJob adds a job to update the S3 config every 30 seconds.
func (s *TokenPriceCacheService) addUpdateCacheJob(ctx context.Context) error {
	ctx = context.WithValue(context.Background(), log.TraceID, "update_token_price_cache_job")
	job, err := s.scheduler.NewJob(
		gocron.DurationJob(
			30*time.Second,
		),
		gocron.NewTask(
			s.updateCache,
			ctx,
		),
	)
	if err != nil {
		return err
	}
	log.InfoZ(ctx, "update_token_price_cache_job job created", zap.Any("job", job))
	return nil
}

// updateCache updates the token price cache.
func (s *TokenPriceCacheService) updateCache(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokenPriceBySymbol, err := s.tokenPriceCacheService.GetAllCoinCache(ctx)
	if err != nil {
		log.ErrorZ(ctx, "failed to get token price by symbol", zap.Error(err))
		return err
	}
	s.cacheBySymbol = tokenPriceBySymbol
	allTokenCache, err := s.tokenPriceCacheService.GetAllTokenCache(ctx)
	if err != nil {
		log.ErrorZ(ctx, "failed to get all token cache", zap.Error(err))
		return err
	}
	s.cacheByAddress = allTokenCache
	log.InfoZ(ctx, "successfully updated token price cache")
	return nil
}
