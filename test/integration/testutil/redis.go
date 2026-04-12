package testutil

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

const defaultTestRedisAddr = "localhost:6380"

type TestRedis struct {
	Client *redis.Client
}

func StartTestRedis(ctx context.Context) (*TestRedis, error) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = defaultTestRedisAddr
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
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
