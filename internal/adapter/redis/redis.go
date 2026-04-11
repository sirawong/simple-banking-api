package adapterredis

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirawong/simple-banking-api/internal/config"
)

func ProvideRedisClient(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}
