package limiter

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	client         *redis.Client
	ipRateLimit    int
	tokenRateLimit int
	blockDuration  time.Duration
}

func NewRateLimiter(client *redis.Client, ipRateLimit, tokenRateLimit, blockDuration string) (*RateLimiter, error) {
	ipRate, err := strconv.Atoi(ipRateLimit)
	if err != nil {
		return nil, err
	}
	tokenRate, err := strconv.Atoi(tokenRateLimit)
	if err != nil {
		return nil, err
	}
	blockDur, err := strconv.Atoi(blockDuration)
	if err != nil {
		return nil, err
	}
	return &RateLimiter{
		client:         client,
		ipRateLimit:    ipRate,
		tokenRateLimit: tokenRate,
		blockDuration:  time.Duration(blockDur) * time.Second,
	}, nil
}

func (rl *RateLimiter) AllowIP(ip string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("ip:%s", ip)
	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("Erro ao incrementar a chave para IP %s: %v", ip, err)
		return false
	}
	log.Printf("Contador para IP %s: %d", ip, count)
	if count == 1 {
		rl.client.Expire(ctx, key, rl.blockDuration)
		log.Printf("Definido expiração para chave IP %s em %v", ip, rl.blockDuration)
	}
	if count > int64(rl.ipRateLimit) {
		log.Printf("IP %s atingiu o limite de requisições: %d", ip, count)
		return false
	}
	return true
}

func (rl *RateLimiter) AllowToken(token string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("token:%s", token)
	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("Erro ao incrementar a chave para token %s: %v", token, err)
		return false
	}
	log.Printf("Contador para token %s: %d", token, count)
	if count == 1 {
		rl.client.Expire(ctx, key, rl.blockDuration)
		log.Printf("Definido expiração para chave token %s em %v", token, rl.blockDuration)
	}
	if count > int64(rl.tokenRateLimit) {
		log.Printf("Token %s atingiu o limite de requisições: %d", token, count)
		return false
	}
	return true
}
