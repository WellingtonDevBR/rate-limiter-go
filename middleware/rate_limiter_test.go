package middleware

import (
	"context"
	"testing"
	"time"

	"rate-limiter/persistence"

	"github.com/go-redis/redis/v8" // Correct import path
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func TestRateLimiterWithRedisAndInMemory(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.FlushDB(ctx)

	primaryStorage := persistence.NewRedisStorage(rdb)
	secondaryStorage := persistence.NewInMemoryStorage()

	limiter := NewRateLimiter(primaryStorage, secondaryStorage, 1, time.Minute)

	key := "test_key"

	// First request should be allowed
	allowed, err := limiter.Allow(ctx, key)
	assert.NoError(t, err, "expected no error on first allow")
	assert.True(t, allowed, "expected first request to be allowed")

	// Second request should be disallowed
	allowed, err = limiter.Allow(ctx, key)
	assert.NoError(t, err, "expected no error on second allow")
	assert.False(t, allowed, "expected second request to be disallowed")

	// Simulate Redis failure by disconnecting
	rdb.Close()

	// Third request should fallback to in-memory storage and be allowed (since it's first in-memory request)
	allowed, err = limiter.Allow(ctx, key)
	assert.NoError(t, err, "expected no error on third allow after Redis failure")
	assert.True(t, allowed, "expected third request to be allowed after Redis failure")

	// Fourth request should be disallowed in in-memory storage
	allowed, err = limiter.Allow(ctx, key)
	assert.NoError(t, err, "expected no error on fourth allow after Redis failure")
	assert.False(t, allowed, "expected fourth request to be disallowed after Redis failure")
}

func TestRateLimiterWithMockStorage(t *testing.T) {
	storage := persistence.NewInMemoryStorage()
	limiter := NewRateLimiter(storage, storage, 1, time.Minute)

	key := "test_key"
	allowed, err := limiter.Allow(ctx, key)
	assert.NoError(t, err, "expected no error on first allow")
	assert.True(t, allowed, "expected first request to be allowed")

	allowed, err = limiter.Allow(ctx, key)
	assert.NoError(t, err, "expected no error on second allow")
	assert.False(t, allowed, "expected second request to be disallowed")
}
