package persistence

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	Client *redis.Client
}

func NewRedisStorage(client *redis.Client) *RedisStorage {
	return &RedisStorage{Client: client}
}

func (r *RedisStorage) Get(ctx context.Context, key string) (int, error) {
	result, err := r.Client.Get(ctx, key).Int()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *RedisStorage) Set(ctx context.Context, key string, value int, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisStorage) Incr(ctx context.Context, key string) error {
	return r.Client.Incr(ctx, key).Err()
}
