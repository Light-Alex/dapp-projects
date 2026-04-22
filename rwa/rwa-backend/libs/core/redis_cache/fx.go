package redis_cache

import (
	"go.uber.org/fx"
)

func LoadModule(config *RedisConfig) fx.Option {
	return fx.Module("redis", fx.Supply(config), fx.Provide(
		NewRedisClient,
		NewTokenPriceCacheService,
		NewOrderBookCacheService,
		NewDepthCacheService,
		NewInstrumentCacheService,
		NewOrderBookPubSub,
		NewSnapshotPubSub,
		NewApiKeyCacheService,
		NewMarketDataCacheService,
	))
}
