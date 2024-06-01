package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"rate-limiter/handler"
	"rate-limiter/limiter"
	"rate-limiter/middleware"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func setup() *redis.Client {
	err := godotenv.Load("../.env")
	if err != nil {
		panic("Erro ao carregar o arquivo .env")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "",
		DB:       0,
	})

	// Verificar conexão com o Redis
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		panic("Erro ao conectar ao Redis: " + err.Error())
	}

	return rdb
}

func TestRateLimiterIP(t *testing.T) {
	t.Log("Iniciando o teste de limite por IP")
	rdb := setup()
	defer rdb.Close()

	ipRateLimit := os.Getenv("IP_RATE_LIMIT")
	tokenRateLimit := os.Getenv("TOKEN_RATE_LIMIT")
	blockDuration := os.Getenv("BLOCK_DURATION")

	t.Logf("Configurações - IP_RATE_LIMIT: %s, TOKEN_RATE_LIMIT: %s, BLOCK_DURATION: %s", ipRateLimit, tokenRateLimit, blockDuration)

	rl, err := limiter.NewRateLimiter(rdb, ipRateLimit, tokenRateLimit, blockDuration)
	assert.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.HomeHandler)
	handler := middleware.RateLimiterMiddleware(rl)(mux)

	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345" // Adiciona um IP fixo para o teste

	// Testando 5 requisições válidas
	for i := 0; i < 5; i++ {
		t.Logf("Requisição %d", i+1)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		t.Logf("Resposta: %d", rr.Code)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Testando a 6ª requisição que deve ser bloqueada
	t.Log("Requisição 6")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	t.Logf("Resposta: %d", rr.Code)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}

func TestRateLimiterToken(t *testing.T) {
	t.Log("Iniciando o teste de limite por Token")
	rdb := setup()
	defer rdb.Close()

	ipRateLimit := os.Getenv("IP_RATE_LIMIT")
	tokenRateLimit := os.Getenv("TOKEN_RATE_LIMIT")
	blockDuration := os.Getenv("BLOCK_DURATION")

	t.Logf("Configurações - IP_RATE_LIMIT: %s, TOKEN_RATE_LIMIT: %s, BLOCK_DURATION: %s", ipRateLimit, tokenRateLimit, blockDuration)

	rl, err := limiter.NewRateLimiter(rdb, ipRateLimit, tokenRateLimit, blockDuration)
	assert.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.HomeHandler)
	handler := middleware.RateLimiterMiddleware(rl)(mux)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("API_KEY", "testtoken")
	req.RemoteAddr = "127.0.0.1:12345" // Adiciona um IP fixo para o teste

	// Testando 10 requisições válidas
	for i := 0; i < 10; i++ {
		t.Logf("Requisição %d", i+1)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		t.Logf("Resposta: %d", rr.Code)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Testando a 11ª requisição que deve ser bloqueada
	t.Log("Requisição 11")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	t.Logf("Resposta: %d", rr.Code)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}
