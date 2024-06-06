package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func main() {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // ajuste o endereço conforme necessário
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Erro ao conectar ao Redis: %v\n", err)
	} else {
		fmt.Println("Conectado ao Redis com sucesso!")
	}
}
