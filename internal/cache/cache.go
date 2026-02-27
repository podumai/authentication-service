package cache

import (
	"authentication_service/internal/config"
	"authentication_service/internal/logger"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
	Close() error
}

type RedisCache struct {
	client *redis.Client
}

func (rc *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}

func (rc *RedisCache) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

type Opts struct {
	Config *config.Redis
	Logger logger.Logger
}

func NewRedisCache(ctx context.Context, opts *Opts) (CacheService, error) {
	cfg := opts.Config
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	r := &RedisCache{client: rdb}

	if err := r.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	opts.Logger.Info("Redis connected")

	return r, nil
}
