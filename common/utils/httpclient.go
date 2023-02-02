package utils

import (
	"context"
	"net"
	"net/http"
	"time"
)

func NewHttpClient() *http.Client {
	c := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, time.Second*3)
			},
			MaxConnsPerHost:     200,
			MaxIdleConnsPerHost: 200,
			IdleConnTimeout:     30 * time.Second,
		},
		Timeout: time.Second * 5,
	}

	return c
}
