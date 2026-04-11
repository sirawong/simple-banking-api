package adapterredis

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/sirawong/simple-banking-api/internal/config"
)

// @wire:set(name=AdapterSet)
func ProvideRedisClient(cfg *config.Config) (*redis.Client, func(), error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = rdb.Close()
		return nil, func() {}, fmt.Errorf("redis: ping failed: %w", err)
	}

	cleanup := func() {
		if err := rdb.Close(); err != nil {
			log.Printf("failed to close redis connection: %v", err)
		}
	}

	return rdb, cleanup, nil
}
