# Rate Limiter em Go

Este rate limiter é configurável para limitar o número máximo de requisições por segundo baseado em endereço IP ou token de acesso.

## Funcionalidades

- Limitação de requisições por IP.
- Limitação de requisições por token de acesso (API_KEY).
- Configuração do número máximo de requisições permitidas por segundo via variáveis de ambiente.
- Opção de escolher o tempo de bloqueio do IP ou do token após exceder o limite de requisições.
- Armazenamento e consulta das informações de limite em um banco de dados Redis.
- Middleware para integração fácil com servidores web.
- Resposta adequada com código HTTP 429 e mensagem explicativa ao exceder o limite de requisições.
- Failover para armazenamento em memória em caso de falha no Redis.


## Variáveis de Ambiente

As seguintes variáveis de ambiente devem ser configuradas:

- `REDIS_HOST`: O host do Redis (ex: `redis`).
- `REDIS_PORT`: A porta do Redis (ex: `6379`).
- `IP_RATE_LIMIT`: O número máximo de requisições permitidas por segundo por IP (ex: `5`).
- `TOKEN_RATE_LIMIT`: O número máximo de requisições permitidas por segundo por token (ex: `10`).
- `BLOCK_DURATION`: O tempo de bloqueio em segundos quando o limite é excedido (ex: `60`).


## Executando a Aplicação

1. Certifique-se de ter Docker e Docker Compose instalados.
2. No diretório raiz do projeto, execute:

```sh
docker-compose up --build
```

## TEST POR IP
#!/bin/bash
for i in {1..10}
do
   curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080
done


## TEST POR TOKEN
#!/bin/bash
TOKEN="token123"
for i in {1..102}
do
   response=$(curl -s -H "API_KEY: $TOKEN" -o /dev/null -w "%{http_code}\n" http://localhost:8080)
   echo "Request $i: $response"
done


## TEST APP
```sh
go test -v ./tests/...
```
