package core

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"coin-server/common/logger"
)

var CloserBus *Closer

type Closer struct {
	signalChan chan os.Signal
	closeFunc  []func()
	wg         *sync.WaitGroup
}

func InitCloser() {
	if CloserBus != nil {
		return
	}
	closer := &Closer{
		signalChan: make(chan os.Signal, 1),
		closeFunc:  make([]func(), 0),
		wg:         &sync.WaitGroup{},
	}
	CloserBus = closer
	closer.listenSignal()
	return
}

func (c *Closer) listenSignal() {
	signal.Notify(c.signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for k := range c.signalChan {
			logger.DefaultLogger.Warn(fmt.Sprintf("caught %s signal, the server is now shutting down......", k.String()))
			c.closeAll()
		}
	}()
}

func (c *Closer) RegisterCloser(f func()) {
	if f == nil {
		panic("closer func nil")
	}
	c.wg.Add(1)
	c.closeFunc = append(c.closeFunc, f)
}

func (c *Closer) Done() {
	c.wg.Done()
}

func (c *Closer) closeAll() {
	close(c.signalChan)
	for _, v := range c.closeFunc {
		v()
	}
	c.wg.Wait()
	logger.DefaultLogger.Info("all tasks completed, the server has exited")
	os.Exit(0)
}
