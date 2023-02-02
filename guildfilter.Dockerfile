FROM golang:1.18 AS builder
WORKDIR /usr/src/app
COPY . .
ENV CGO_ENABLED=0
RUN go build -mod=vendor -o guild-filter ./guild-filter-server/main.go
FROM alpine:latest

WORKDIR /app
COPY --from=builder /usr/src/app/guild-filter .
ENTRYPOINT ["./guild-filter"]
