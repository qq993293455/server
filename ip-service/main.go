package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"coin-server/common/gopool"
)

func main() {
	addr := ":25000"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	fmt.Println("listen on", addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		gopool.Submit(func() {
			defer conn.Close()
			ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()
			b := make([]byte, 1+len(ip))
			b[0] = byte(len(b))
			copy(b[1:], ip)
			_, err = conn.Write(b)
			if err != nil {
				return
			}
			io.ReadFull(conn, b[:1])
		})
	}
}
