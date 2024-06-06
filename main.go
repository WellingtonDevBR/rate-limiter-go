package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"rate-limiter/handler"
	"rate-limiter/persistence"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func main() {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})
	primaryStorage := persistence.NewRedisStorage(rdb)

	// Initialize in-memory storage
	secondaryStorage := persistence.NewInMemoryStorage()

	http.Handle("/", handler.NewHandler(primaryStorage, secondaryStorage))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
