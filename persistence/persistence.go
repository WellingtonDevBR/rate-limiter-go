package persistence

import (
	"context"
	"time"
)

type Storage interface {
	Get(ctx context.Context, key string) (int, error)
	Set(ctx context.Context, key string, value int, ttl time.Duration) error
	Incr(ctx context.Context, key string) error
}
