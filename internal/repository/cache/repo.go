package cache

import (
	"context"
	"time"
)

// Repository defines cache operations used across the application.
type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
