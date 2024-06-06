package limiter

import (
	"context"
	"time"

	"rate-limiter/persistence"
)

type RateLimiter struct {
	Storage persistence.Storage
	Limit   int
	TTL     time.Duration
}

func NewRateLimiter(storage persistence.Storage, limit int, ttl time.Duration) *RateLimiter {
	return &RateLimiter{
		Storage: storage,
		Limit:   limit,
		TTL:     ttl,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	result, err := rl.Storage.Get(ctx, key)
	if err != nil {
		return false, err
	}

	if result == 0 {
		err = rl.Storage.Set(ctx, key, 1, rl.TTL)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	if result >= rl.Limit {
		return false, nil
	}

	err = rl.Storage.Incr(ctx, key)
	if err != nil {
		return false, err
	}

	return true, nil
}
