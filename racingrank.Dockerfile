FROM golang:1.18 AS builder
WORKDIR /usr/src/app
COPY . .
ENV CGO_ENABLED=0
RUN go build -mod=vendor -o racingrank ./racingrank-server/main.go
FROM alpine:latest

WORKDIR /app
COPY --from=builder /usr/src/app/racingrank .
ENTRYPOINT ["./racingrank"]
