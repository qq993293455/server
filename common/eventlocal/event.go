package eventlocal

import (
	"fmt"
	"reflect"
	"runtime"
	"sort"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"

	"go.uber.org/zap"
)

type funcInfo struct {
	Func reflect.Value
	Name string
}

type Events struct {
	eMap    map[reflect.Type][]*funcInfo
	isDebug bool
}

var expected = ",expected: func(*Context,[任何类型])*errmsg.ErrMsg"

func (this_ *Events) SubscribeEventLocal(v interface{}) {
	tf := reflect.TypeOf(v)
	if tf.Kind() != reflect.Func {
		panic("eventlocal.SubscribeEventLocal: param v must be a func" + expected)
	}
	if tf.NumIn() != 2 {
		panic("eventlocal.SubscribeEventLocal: in params num must is 2" + expected)
	}
	in0t := tf.In(0)
	if in0t.Kind() != reflect.TypeOf((*ctx.Context)(nil)).Kind() {
		panic("eventlocal.SubscribeEventLocal: in params 0 not is *ctx.Context" + expected)
	}
	in1t := tf.In(1)
	vf := reflect.ValueOf(v)
	if es, ok := this_.eMap[in1t]; ok {
		for _, v := range es {
			if reflect.DeepEqual(vf, v) {
				panic(fmt.Sprintf("eventlocal.SubscribeEventLocal: duplicate '%s' , had registered", tf.String()))
			}
		}
	}
	fi := &funcInfo{Func: vf, Name: runtime.FuncForPC(vf.Pointer()).Name()}
	this_.eMap[in1t] = append(this_.eMap[in1t], fi)
}

func (this_ *Events) GetAllEventLocal() []string {
	out := make([]string, 0, len(this_.eMap))
	for k := range this_.eMap {
		out = append(out, k.String())
	}
	sort.Strings(out)
	return out
}

func (this_ *Events) Exec(c *ctx.Context, i interface{}) *errmsg.ErrMsg {
	key := reflect.TypeOf(i)
	es, ok := this_.eMap[key]
	if !ok {
		c.Warn("eventlocal.Exec: not found event " + key.String())
		return nil
	}
	keyStr := key.String()
	if this_.isDebug {
		c.Debug("exec eventlocal", zap.String("event", keyStr), zap.Any("data", i))
	}
	for _, v := range es {
		if this_.isDebug {
			c.Debug("exec local event", zap.String("event", keyStr), zap.String("func", v.Name))
		}
		out := v.Func.Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(i)})
		re := out[0]
		if !re.IsNil() {
			return re.Interface().(*errmsg.ErrMsg)
		}
	}
	return nil
}

var events = &Events{eMap: map[reflect.Type][]*funcInfo{}}

func SubscribeEventLocal(v interface{}) {
	events.SubscribeEventLocal(v)
}

func GetAllEventLocal() []string {
	return events.GetAllEventLocal()
}

type LocalEvents []interface{}

func CreateEventLocal(isDebug bool) handler.MiddleWare {
	events.isDebug = isDebug
	return func(next handler.HandleFunc) handler.HandleFunc {
		return func(ctx *ctx.Context) *errmsg.ErrMsg {
			e := next(ctx)
			if e != nil {
				return e
			}
			for len(ctx.EventLocal) > 0 {
				el := make([]interface{}, len(ctx.EventLocal))
				copy(el, ctx.EventLocal)
				ctx.EventLocal = ctx.EventLocal[:0]
				ell := len(el)
				for i := 0; i < ell; i++ {
					e := doEvent(ctx, el[i])
					if e != nil {
						return e
					}
				}
			}
			return nil
		}
	}
}

func doEvent(c *ctx.Context, i interface{}) (err *errmsg.ErrMsg) {
	defer func() {
		if e := recover(); e != nil {
			err = errmsg.NewInternalErr(fmt.Sprintf("do event panic:%v", e))
			c.TraceLogger.Error("do event panic", zap.Any("panic info", e), zap.String("eventName", reflect.TypeOf(i).String()), zap.String("event", fmt.Sprintf("%v", i)))
		}
	}()
	err = events.Exec(c, i)
	return
}
