package middleware

import (
	"log"
	"net"
	"net/http"
	"rate-limiter/limiter"
	"strings"
)

func RateLimiterMiddleware(rl *limiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if strings.Contains(ip, ":") {
				var err error
				ip, _, err = net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					log.Printf("Erro ao dividir o host e a porta: %v", err)
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
			}

			token := r.Header.Get("API_KEY")

			log.Printf("Requisição recebida de IP %s com token %s", ip, token)

			if token != "" {
				if !rl.AllowToken(token) {
					log.Printf("Limite de requisições atingido para token %s", token)
					http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
					return
				}
			} else {
				if !rl.AllowIP(ip) {
					log.Printf("Limite de requisições atingido para IP %s", ip)
					http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
