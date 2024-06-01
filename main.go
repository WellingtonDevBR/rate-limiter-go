package main

import (
	"log"
	"net/http"
	"os"
	"rate-limiter/handler"
	"rate-limiter/limiter"
	"rate-limiter/middleware"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}

	// Conexão com o Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "",
		DB:       0,
	})

	// Configuração do Rate Limiter
	ipRateLimit := os.Getenv("IP_RATE_LIMIT")
	tokenRateLimit := os.Getenv("TOKEN_RATE_LIMIT")
	blockDuration := os.Getenv("BLOCK_DURATION")

	rl, err := limiter.NewRateLimiter(rdb, ipRateLimit, tokenRateLimit, blockDuration)
	if err != nil {
		log.Fatalf("Erro ao criar o Rate Limiter: %v", err)
	}

	// Configuração do Middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.HomeHandler)
	log.Println("Servidor iniciado na porta 8080")
	http.ListenAndServe(":8080", middleware.RateLimiterMiddleware(rl)(mux))
}
