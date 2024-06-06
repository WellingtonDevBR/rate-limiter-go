package handler

import (
	"net/http"
	"rate-limiter/middleware"
	"rate-limiter/persistence"
)

func NewHandler(primary, secondary persistence.Storage) http.Handler {
	return middleware.RateLimiterMiddleware(primary, secondary)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
}
