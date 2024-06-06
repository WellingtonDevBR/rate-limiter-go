package middleware

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"rate-limiter/persistence"

	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	PrimaryStorage   persistence.Storage
	SecondaryStorage persistence.Storage
	LimitIP          int
	TTLIP            time.Duration
	TokenLimits      map[string]int
	TokenTTLs        map[string]time.Duration
	useSecondary     bool
	mu               sync.Mutex
}

func NewRateLimiter(primary, secondary persistence.Storage, limitIP int, ttlIP time.Duration, tokenLimits map[string]int, tokenTTLs map[string]time.Duration) *RateLimiter {
	return &RateLimiter{
		PrimaryStorage:   primary,
		SecondaryStorage: secondary,
		LimitIP:          limitIP,
		TTLIP:            ttlIP,
		TokenLimits:      tokenLimits,
		TokenTTLs:        tokenTTLs,
		useSecondary:     false,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, ttl time.Duration) (bool, error) {
	log.Printf("Allow called with key=%s, limit=%d, ttl=%v", key, limit, ttl)
	rl.mu.Lock()
	useSecondary := rl.useSecondary
	rl.mu.Unlock()

	if !useSecondary {
		allowed, err := rl.tryAllow(ctx, rl.PrimaryStorage, key, limit, ttl)
		if err == nil {
			log.Printf("Primary storage allowed: %v", allowed)
			return allowed, nil
		}
		log.Printf("Primary storage failed: %v", err)
		rl.switchToSecondary()
	}

	// If primary storage fails or useSecondary is true, use secondary storage
	allowed, err := rl.tryAllow(ctx, rl.SecondaryStorage, key, limit, ttl)
	log.Printf("Secondary storage allowed: %v", allowed)
	return allowed, err
}

func (rl *RateLimiter) tryAllow(ctx context.Context, storage persistence.Storage, key string, limit int, ttl time.Duration) (bool, error) {
	result, err := storage.Get(ctx, key)
	log.Printf("tryAllow called with key=%s, result=%d, err=%v", key, result, err)
	if err != nil && err != redis.Nil {
		log.Printf("tryAllow Get error: %v", err)
		return false, err
	}

	if err == redis.Nil || result == 0 {
		err = storage.Set(ctx, key, 1, ttl)
		log.Printf("tryAllow Set called with key=%s, value=1, ttl=%v, err=%v", key, ttl, err)
		if err != nil {
			log.Printf("tryAllow Set error: %v", err)
			return false, err
		}
		return true, nil
	}

	if result >= limit {
		log.Printf("tryAllow limit exceeded for key=%s", key)
		return false, nil
	}

	err = storage.Incr(ctx, key)
	log.Printf("tryAllow Incr called with key=%s, err=%v", key, err)
	if err != nil {
		log.Printf("tryAllow Incr error: %v", err)
		return false, err
	}

	return true, nil
}

func (rl *RateLimiter) switchToSecondary() {
	rl.mu.Lock()
	rl.useSecondary = true
	rl.mu.Unlock()
}

func RateLimiterMiddleware(primary, secondary persistence.Storage, limitIP int, ttlIP time.Duration, tokenLimits map[string]int, tokenTTLs map[string]time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(primary, secondary, limitIP, ttlIP, tokenLimits, tokenTTLs)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("API_KEY")
			var key string
			var limit int
			var ttl time.Duration

			if token != "" && limiter.TokenLimits != nil {
				limit = limiter.TokenLimits[token]
				ttl = limiter.TokenTTLs[token]
				key = "token_" + token
			} else {
				key = r.RemoteAddr
				limit = limiter.LimitIP
				ttl = limiter.TTLIP
			}

			allowed, err := limiter.Allow(r.Context(), key, limit, ttl)
			if err != nil || !allowed {
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
