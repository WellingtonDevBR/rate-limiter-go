package middleware

import (
	"context"
	"net/http"
	"time"

	"rate-limiter/persistence"

	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	PrimaryStorage   persistence.Storage
	SecondaryStorage persistence.Storage
	Limit            int
	TTL              time.Duration
}

func NewRateLimiter(primary, secondary persistence.Storage, limit int, ttl time.Duration) *RateLimiter {
	return &RateLimiter{
		PrimaryStorage:   primary,
		SecondaryStorage: secondary,
		Limit:            limit,
		TTL:              ttl,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	allowed, err := rl.tryAllow(ctx, rl.PrimaryStorage, key)
	if err == nil {
		return allowed, nil
	}
	// If primary storage fails, use secondary storage
	return rl.tryAllow(ctx, rl.SecondaryStorage, key)
}

func (rl *RateLimiter) tryAllow(ctx context.Context, storage persistence.Storage, key string) (bool, error) {
	result, err := storage.Get(ctx, key)
	if err != nil && err != redis.Nil {
		return false, err
	}

	if err == redis.Nil {
		err = storage.Set(ctx, key, 1, rl.TTL)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	if result >= rl.Limit {
		return false, nil
	}

	err = storage.Incr(ctx, key)
	if err != nil {
		return false, err
	}

	return true, nil
}

func RateLimiterMiddleware(primary, secondary persistence.Storage) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(primary, secondary, 100, time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			allowed, err := limiter.Allow(r.Context(), key)
			if err != nil || !allowed {
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
