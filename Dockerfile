FROM golang:1.21-alpine3.20 as build

WORKDIR /app

RUN go install github.com/mitranim/gow@latest

CMD ["gow", "run", "main.go"]