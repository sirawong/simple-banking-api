package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type cacheRepository struct {
	rdb *redis.Client
}

// @wire:set(name=RepositorySet)
func ProvideCacheRepository(rdb *redis.Client) Repository {
	return &cacheRepository{rdb: rdb}
}

func (r *cacheRepository) Get(ctx context.Context, key string) (string, error) {
	return r.rdb.Get(ctx, key).Result()
}

func (r *cacheRepository) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.rdb.Set(ctx, key, value, ttl).Err()
}

func (r *cacheRepository) Delete(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}
