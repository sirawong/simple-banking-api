package testutil

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/sirawong/simple-banking-api/internal/config"
)

type TestRedis struct {
	Client *redis.Client
}

func StartTestRedis(ctx context.Context, cfg *config.Config) (*TestRedis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &TestRedis{Client: client}, nil
}

func (r *TestRedis) FlushAll(ctx context.Context) error {
	return r.Client.FlushAll(ctx).Err()
}

func (r *TestRedis) Close() error {
	return r.Client.Close()
}
