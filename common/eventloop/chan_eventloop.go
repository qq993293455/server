package eventloop

import (
	"fmt"
	"reflect"
	"sync"

	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/timer"

	"go.uber.org/zap"
)

type Resp struct {
	Resp interface{}
	Err  *errmsg.ErrMsg
}

type chanEvent struct {
	Resp chan *Resp
	Req  interface{}
	Call reflect.Value
}

type ChanEventLoop struct {
	*EventLoop
	chanEvent map[string]reflect.Value
}

func NewChanEventLoop(log *logger.Logger) *ChanEventLoop {
	return &ChanEventLoop{
		EventLoop: NewEventLoop(log),
		chanEvent: map[string]reflect.Value{},
	}
}

const panicInfo = "eventloop.ChanEventLoop.Register: param f must be a func: func(r req) resp"

func RegisterFuncChanEventLoop[req any, resp any](ce *ChanEventLoop, f func(r req) (resp, *errmsg.ErrMsg)) {
	tf := reflect.TypeOf(f)
	if tf.Kind() != reflect.Func {
		panic(panicInfo)
	}
	if tf.NumIn() != 1 {
		panic(panicInfo)
	}
	if tf.NumOut() != 2 {
		panic(panicInfo)
	}
	in0 := tf.In(0)

	name := in0.String()
	v, ok := ce.chanEvent[name]
	if ok {
		panic(name + " had registered:" + v.String())
	}
	ce.chanEvent[name] = reflect.ValueOf(f)
}

var poolMap = sync.Pool{New: func() interface{} {
	return &chanEvent{Resp: make(chan *Resp, 1)}
}}

var poolResp = sync.Pool{New: func() interface{} {
	return &Resp{}
}}

func getChanEvent() *chanEvent {
	return poolMap.Get().(*chanEvent)
}

func putChanEvent(ce *chanEvent) {
	ce.Req = nil
	ce.Call = reflect.Value{}
	poolMap.Put(ce)
}

func getResp() *Resp {
	return poolResp.Get().(*Resp)
}

func putResp(r *Resp) {
	r.Resp = nil
	r.Err = nil
	poolResp.Put(r)
}

func RequestChanEventLoop[req any, resp any](ce *ChanEventLoop, r req) (resp, *errmsg.ErrMsg) {
	now := timer.Now()
	defer func() {
		ce.log.Debug("RequestChanEventLoop", zap.Duration("cost", timer.Now().Sub(now)), zap.Any("req", r), zap.String("req type", reflect.TypeOf(r).Name()))
	}()

	c := getChanEvent()
	defer putChanEvent(c)
	c.Req = r
	ce.PostEventQueue(c)
	out := <-c.Resp
	defer putResp(out)
	return out.Resp.(resp), out.Err
}

func CallChanEventLoop[req any, resp any](ce *ChanEventLoop, r req, f func(r req) (resp, *errmsg.ErrMsg)) (resp, *errmsg.ErrMsg) {
	now := timer.Now()
	defer func() {
		ce.log.Debug("CallChanEventLoop", zap.Duration("cost", timer.Now().Sub(now)), zap.Any("req", r), zap.String("req type", reflect.TypeOf(r).Name()))
	}()
	c := getChanEvent()
	defer putChanEvent(c)
	c.Req = r
	c.Call = reflect.ValueOf(f)
	ce.PostEventQueue(c)
	out := <-c.Resp
	defer putResp(out)
	if out.Resp == nil {
		out.Resp = reflect.New(reflect.TypeOf(new(resp)).Elem()).Elem().Interface()
	}
	return out.Resp.(resp), out.Err
}

func (this_ *ChanEventLoop) Start(f func(e interface{})) {
	this_.EventLoop.Start(func(event interface{}) {
		switch msg := event.(type) {
		case *chanEvent:
			if msg.Call.Kind() != reflect.Invalid {
				this_.dealChanEvent(reflect.ValueOf(msg), msg.Call)
				return
			}

			name := reflect.TypeOf(msg.Req).String()
			tv, ok := this_.chanEvent[name]
			if ok {
				this_.dealChanEvent(reflect.ValueOf(msg), tv)
				return
			}
		}

		f(event)
	})
}

func (this_ *ChanEventLoop) dealChanEvent(ev reflect.Value, tv reflect.Value) {
	resp := getResp()
	defer func() {
		if e := recover(); e != nil {
			this_.log.Error("dealChanEvent panic", zap.Any("panic info", e))
			resp.Err = errmsg.NewInternalErr(fmt.Sprintf("%v", e))
		}
		ev.Elem().Field(0).Send(reflect.ValueOf(resp))
	}()
	f1 := ev.Elem().Field(1).Elem()
	out := tv.Call([]reflect.Value{f1})
	if !out[1].IsNil() {
		resp.Err = out[1].Interface().(*errmsg.ErrMsg)
	} else if out[0].IsValid() {
		resp.Resp = out[0].Interface()
	}
}
