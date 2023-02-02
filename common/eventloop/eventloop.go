package eventloop

import (
	"runtime"
	"sync/atomic"
	"time"

	"coin-server/common/logger"
	"coin-server/common/safego"
	"coin-server/common/timer"

	"go.uber.org/zap"
)

type EventLoop struct {
	queue   *EsQueue
	stopped int32
	over    chan struct{}
	log     *logger.Logger
}

func NewEventLoop(log *logger.Logger) *EventLoop {
	queue := NewQueue(100000)
	return &EventLoop{
		queue: queue,
		log:   log,
		over:  make(chan struct{}),
	}
}

func NewEventLoopWithQueueSize(log *logger.Logger, size uint32) *EventLoop {
	queue := NewQueue(size)
	return &EventLoop{
		queue: queue,
		log:   log,
		over:  make(chan struct{}),
	}
}

func (this_ *EventLoop) PostEventQueue(e interface{}) {
	ok, _ := this_.queue.Put(e)
	for !ok {
		runtime.Gosched()
		ok, _ = this_.queue.Put(e)
	}
}

func (this_ *EventLoop) PostFuncQueue(f func()) {
	this_.PostEventQueue(f)
}

func (this_ *EventLoop) AfterFuncQueue(d time.Duration, f func()) {
	timer.AfterFunc(d, func() {
		this_.PostFuncQueue(f)
	})
}

func (this_ *EventLoop) UntilFuncQueue(t time.Time, f func()) {
	timer.UntilFunc(t, func() {
		this_.PostFuncQueue(f)
	})
}

func (this_ *EventLoop) TickQueue(d time.Duration, f func() bool) {
	timer.AfterFunc(d, func() {
		this_.PostFuncQueue(func() {
			ok := f()
			if ok {
				this_.TickQueue(d, f)
			}
		})
	})
}

func (this_ *EventLoop) Start(f func(event interface{}), endF ...func()) {
	xf := func(event interface{}) {
		defer safego.Recover(func(e interface{}) {
			this_.log.Error("event func panic", zap.Any("panic info", e))
		})
		switch fx := event.(type) {
		case func():
			fx()
		default:
			f(event)
		}
	}
	var rF []func()
	if len(endF) > 0 {
		for _, v := range endF {
			rF = append(rF, func() {
				defer safego.Recover(func(e interface{}) {
					this_.log.Error("end func panic", zap.Any("panic info", e))
				})
				v()
			})
		}
	}
	go this_.start(xf, rF...)
}

func (this_ *EventLoop) start(f func(event interface{}), endF ...func()) {
	defer func() {
		this_.log.Warn("eventloop closed")
		close(this_.over)
	}()
	events := make([]interface{}, 4096)
	for atomic.LoadInt32(&this_.stopped) == 0 {
		gets, _ := this_.queue.Gets(events)
		if gets > 0 {
			es := events[:gets]
			for _, v := range es {
				f(v)
			}
		} else {
			time.Sleep(time.Millisecond * 1)
		}
	}
	for {
		v, _, _ := this_.queue.Get()
		if v == nil {
			break
		} else {
			f(v)
		}
	}
	for _, v := range endF {
		v()
	}

}

func (this_ *EventLoop) Stop() {
	atomic.StoreInt32(&this_.stopped, 1)
	<-this_.over
}

func (this_ *EventLoop) Stopped() bool {
	return atomic.LoadInt32(&this_.stopped) == 1
}
