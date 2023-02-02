FROM golang:1.18 AS builder
WORKDIR /usr/src/app
COPY . .
ENV CGO_ENABLED=0
RUN go build -mod=vendor -o pika-viewer ./pikaviewer/main.go
FROM alpine:latest

RUN apk --no-cache add tzdata  && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

ENV CONF_PATH="newhttp/"
ENV PIKA_VIEWER_ADDR=":9991"
WORKDIR /app
COPY --from=builder /usr/src/app/pika-viewer .
ENTRYPOINT ["./pika-viewer"]
