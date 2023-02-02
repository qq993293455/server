package core

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func WaitClose(f func()) {
	ch := make(chan os.Signal, 10) // 10 为了IDE调试的时候不卡在
	signal.Notify(ch, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	sig := <-ch
	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("Receive Signal: %s, server is shutting down...", sig.String())
	fmt.Println()
	fmt.Println("--------------------------------------------------------------------")
	f()
}
