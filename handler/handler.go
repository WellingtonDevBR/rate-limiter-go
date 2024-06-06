package handler

import (
	"net/http"
	"rate-limiter/middleware"
	"rate-limiter/persistence"
	"time"
)

func NewHandler(primary, secondary persistence.Storage, limitIP int, ttlIP time.Duration, tokenLimits map[string]int, tokenTTLs map[string]time.Duration) http.Handler {
	return middleware.RateLimiterMiddleware(primary, secondary, limitIP, ttlIP, tokenLimits, tokenTTLs)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
}
