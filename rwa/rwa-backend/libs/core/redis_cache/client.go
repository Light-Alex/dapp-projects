package redis_cache

import (
	"context"

	"github.com/AnchoredLabs/rwa-backend/libs/errors"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewRedisClient creates and returns a new Redis client based on the provided configuration.
func NewRedisClient(config *RedisConfig, lc fx.Lifecycle) (redis.UniversalClient, error) {
	ctx := context.WithValue(context.Background(), log.TraceID, "redis_init")
	if config == nil || len(config.Hosts) == 0 {
		return nil, errors.New("redis hosts is empty")
	}
	redisCli := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:      config.Hosts,
		Password:   config.Password,
		DB:         config.Db,
		MasterName: config.MasterName,
		Username:   config.UserName,
		Protocol:   2,
	})
	info := redisCli.Info(ctx)
	if info.Err() != nil {
		log.ErrorZ(ctx, "redis connect error", zap.Error(info.Err()))
		return nil, errors.New(info.Err().Error())
	}
	log.InfoZ(ctx, "redis connected", zap.Strings("hosts", config.Hosts))
	log.InfoZ(ctx, "redis "+info.String())
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.InfoZ(ctx, "stoping redis....")
			return redisCli.Close()
		},
	})
	return redisCli, nil
}
