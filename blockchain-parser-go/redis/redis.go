package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisClient(host string, port int, password string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + strconv.Itoa(port),
		Password: password,
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisClient) GetLastProcessedBlock() (int64, error) {
	val, err := r.client.Get(r.ctx, "last_processed_block").Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (r *RedisClient) SetLastProcessedBlock(blockNumber int64) error {
	return r.client.Set(r.ctx, "last_processed_block", blockNumber, 0).Err()
}

func (r *RedisClient) IsTxProcessing(txHash string) (bool, error) {
	val, err := r.client.Exists(r.ctx, "tx:processing:"+txHash).Result()
	return val > 0, err
}

func (r *RedisClient) SetTxProcessing(txHash string, expire time.Duration) error {
	return r.client.Set(r.ctx, "tx:processing:"+txHash, "1", expire).Err()
}

func (r *RedisClient) RemoveTxProcessing(txHash string) error {
	return r.client.Del(r.ctx, "tx:processing:"+txHash).Err()
}

func (r *RedisClient) CacheUserAddress(address string, userID int) error {
	return r.client.Set(r.ctx, "address:user:"+address, userID, time.Hour).Err()
}

func (r *RedisClient) GetCachedUserByAddress(address string) (int, error) {
	val, err := r.client.Get(r.ctx, "address:user:"+address).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}
