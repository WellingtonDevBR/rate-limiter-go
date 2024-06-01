FROM golang:1.16-alpine

WORKDIR /go/src/app

COPY . .

RUN go mod tidy

CMD ["go", "run", "main.go"]
