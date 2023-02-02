package eventemitter

import (
	"coin-server/common/eventloop"
	"coin-server/common/logger"
)

type Event struct {
	Name string
	Data interface{}
}

type EventEmitter struct {
	log *logger.Logger
	el  *eventloop.EventLoop
	ef  map[string][]func(interface{})
}

func NewEventEmitter(log *logger.Logger) *EventEmitter {
	ee := &EventEmitter{
		log: log,
		el:  eventloop.NewEventLoop(log),
		ef:  make(map[string][]func(interface{})),
	}
	ee.start()
	return ee
}

func (e *EventEmitter) On(event string, fn func(interface{})) {
	if _, ok := e.ef[event]; !ok {
		e.ef[event] = make([]func(interface{}), 0)
	}
	e.ef[event] = append(e.ef[event], fn)
}

func (e *EventEmitter) Emit(event string, data interface{}) {
	e.el.PostEventQueue(Event{
		Name: event,
		Data: data,
	})
}

func (e *EventEmitter) EmitFunc(fn func()) {
	e.el.PostFuncQueue(fn)
}

func (e *EventEmitter) start() {
	e.el.Start(func(event interface{}) {
		switch ent := event.(type) {
		case Event:
			if fns, ok := e.ef[ent.Name]; ok {
				for _, f := range fns {
					f(ent.Data)
				}
			}
		}
	})
}
