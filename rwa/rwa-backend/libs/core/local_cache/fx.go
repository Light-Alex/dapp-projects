package local_cache

import (
	"go.uber.org/fx"
)

func LoadModule() fx.Option {
	return fx.Module("localCache",
		fx.Provide(
			NewTokenPriceCacheService,
		),
		fx.Invoke(func(_ *TokenPriceCacheService) {}),
	)
}
