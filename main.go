package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"rate-limiter/middleware"
	"rate-limiter/persistence"

	"github.com/go-redis/redis/v8"
)

func main() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisAddr := redisHost + ":" + redisPort

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	primaryStorage := persistence.NewRedisStorage(rdb)
	secondaryStorage := persistence.NewInMemoryStorage()

	tokenLimits := map[string]int{
		"token123": 10,
	}
	tokenTTLs := map[string]time.Duration{
		"token123": time.Minute,
	}

	rateLimiterMiddleware := middleware.RateLimiterMiddleware(primaryStorage, secondaryStorage, 5, time.Minute, tokenLimits, tokenTTLs)

	http.Handle("/", rateLimiterMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})))

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
