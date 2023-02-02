FROM golang:1.18 AS builder
WORKDIR /usr/src/app
COPY . .
ENV CGO_ENABLED=0
RUN go build -mod=vendor -o gameserver ./game-server/main.go
FROM alpine:latest

WORKDIR /app
COPY --from=builder /usr/src/app/gameserver .
ENTRYPOINT ["./gameserver"]
