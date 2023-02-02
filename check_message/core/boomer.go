package core

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.cdl5.org/chengdu-l5/boomer"
)

var GlobalBoomer *boomer.Boomer

func RunBoomer(host string, port int, task *boomer.Task) {
	boomer.SpawnWithInitFunc = true
	GlobalBoomer = boomer.NewBoomer(host, port)
	GlobalBoomer.Run(task)
	waitForQuit()
}

func waitForQuit() {
	wg := sync.WaitGroup{}
	wg.Add(1)

	quitByMe := false
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		quitByMe = true
		GlobalBoomer.Quit()
		wg.Done()
	}()

	boomer.Events.Subscribe("boomer:quit", func() {
		if !quitByMe {
			wg.Done()
		}
	})
	boomer.Events.Subscribe("boomer:stop", func() {
		if !quitByMe {
			wg.Done()
		}
	})

	wg.Wait()
}
