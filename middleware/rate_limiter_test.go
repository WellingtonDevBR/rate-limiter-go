package middleware

import (
	"context"
	"log"
	"testing"
	"time"

	"rate-limiter/persistence"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func resetState(limiter *RateLimiter) {
	limiter.mu.Lock()
	limiter.useSecondary = false
	limiter.mu.Unlock()
}

func TestRateLimiterWithRedisAndInMemory(t *testing.T) {
	log.Println("TestRateLimiterWithRedisAndInMemory started")
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.FlushDB(ctx)

	primaryStorage := persistence.NewRedisStorage(rdb)
	secondaryStorage := persistence.NewInMemoryStorage()

	tokenLimits := map[string]int{
		"testtoken": 10,
	}
	tokenTTLs := map[string]time.Duration{
		"testtoken": time.Minute,
	}

	limiter := NewRateLimiter(primaryStorage, secondaryStorage, 10, time.Minute, tokenLimits, tokenTTLs)

	key := "test_key"
	token := "testtoken"

	// First request with token should be allowed
	allowed, err := limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
	assert.NoError(t, err, "expected no error on first allow with token")
	assert.True(t, allowed, "expected first request to be allowed with token")

	// Second request with token should be allowed
	allowed, err = limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
	assert.NoError(t, err, "expected no error on second allow with token")
	assert.True(t, allowed, "expected second request to be allowed with token")

	// Continue for all 10 requests
	for i := 3; i <= 10; i++ {
		allowed, err = limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
		assert.NoError(t, err, "expected no error on allow with token")
		assert.True(t, allowed, "expected request to be allowed with token")
	}

	// The 11th request with token should be blocked
	allowed, err = limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
	assert.NoError(t, err, "expected no error on allow with token")
	assert.False(t, allowed, "expected 11th request to be disallowed with token")

	// First request without token should be allowed
	allowed, err = limiter.Allow(ctx, key, 10, time.Minute)
	assert.NoError(t, err, "expected no error on first allow without token")
	assert.True(t, allowed, "expected first request to be allowed without token")

	// Continue for all 10 requests without token
	for i := 2; i <= 10; i++ {
		allowed, err = limiter.Allow(ctx, key, 10, time.Minute)
		assert.NoError(t, err, "expected no error on allow without token")
		assert.True(t, allowed, "expected request to be allowed without token")
	}

	// The 11th request without token should be blocked
	allowed, err = limiter.Allow(ctx, key, 10, time.Minute)
	assert.NoError(t, err, "expected no error on allow without token")
	assert.False(t, allowed, "expected 11th request to be disallowed without token")

	// Simulate Redis failure by disconnecting
	rdb.Close()

	// Reset the state before failover
	resetState(limiter)

	// First request with token after Redis failure should be allowed by in-memory storage
	allowed, err = limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
	assert.NoError(t, err, "expected no error on first request after Redis failure with token")
	assert.True(t, allowed, "expected first request after Redis failure with token to be allowed")

	// Second request with token after Redis failure should be allowed by in-memory storage
	allowed, err = limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
	assert.NoError(t, err, "expected no error on second request after Redis failure with token")
	assert.True(t, allowed, "expected second request after Redis failure with token to be allowed")

	// Third request with token after Redis failure should be disallowed by in-memory storage
	allowed, err = limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
	assert.NoError(t, err, "expected no error on third request after Redis failure with token")
	assert.False(t, allowed, "expected third request after Redis failure with token to be disallowed")
}

func TestRateLimiterWithInMemoryOnly(t *testing.T) {
	log.Println("TestRateLimiterWithInMemoryOnly started")
	storage := persistence.NewInMemoryStorage()
	tokenLimits := map[string]int{
		"testtoken": 10,
	}
	tokenTTLs := map[string]time.Duration{
		"testtoken": time.Minute,
	}

	limiter := NewRateLimiter(storage, storage, 10, time.Minute, tokenLimits, tokenTTLs)

	key := "test_key"
	token := "testtoken"

	// First request with token should be allowed
	allowed, err := limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
	assert.NoError(t, err, "expected no error on first allow with token")
	assert.True(t, allowed, "expected first request to be allowed with token")

	// Second request with token should be allowed
	allowed, err = limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
	assert.NoError(t, err, "expected no error on second allow with token")
	assert.True(t, allowed, "expected second request to be allowed with token")

	// Continue for all 10 requests
	for i := 3; i <= 10; i++ {
		allowed, err = limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
		assert.NoError(t, err, "expected no error on allow with token")
		assert.True(t, allowed, "expected request to be allowed with token")
	}

	// The 11th request with token should be blocked
	allowed, err = limiter.Allow(ctx, "token_"+token, tokenLimits[token], tokenTTLs[token])
	assert.NoError(t, err, "expected no error on allow with token")
	assert.False(t, allowed, "expected 11th request to be disallowed with token")

	// First request without token should be allowed
	allowed, err = limiter.Allow(ctx, key, 10, time.Minute)
	assert.NoError(t, err, "expected no error on first allow without token")
	assert.True(t, allowed, "expected first request to be allowed without token")

	// Continue for all 10 requests without token
	for i := 2; i <= 10; i++ {
		allowed, err = limiter.Allow(ctx, key, 10, time.Minute)
		assert.NoError(t, err, "expected no error on allow without token")
		assert.True(t, allowed, "expected request to be allowed without token")
	}

	// The 11th request without token should be blocked
	allowed, err = limiter.Allow(ctx, key, 10, time.Minute)
	assert.NoError(t, err, "expected no error on allow without token")
	assert.False(t, allowed, "expected 11th request to be disallowed without token")
}
