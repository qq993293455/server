package handler

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"coin-server/common/bytespool"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventloop"
	"coin-server/common/msgcreate"
	"coin-server/common/proto/models"
	"coin-server/common/protocol"
	"coin-server/common/timer"

	"github.com/gogo/protobuf/types"

	"go.uber.org/zap"

	"github.com/gogo/protobuf/proto"
)

type MiddleWare = func(HandleFunc) HandleFunc

type HandleFunc = func(ctx *ctx.Context) *errmsg.ErrMsg

type Elem struct {
	h                HandleFunc
	desc             string
	midS             []MiddleWare
	isEvent          bool
	url              string
	funcType         string
	isLogMiddleError bool
}

func (this_ *Elem) IsEvent() bool {
	return this_.isEvent
}

func (this_ *Elem) String() string {
	if this_.isEvent {
		return fmt.Sprintf("[%s] event: %s", this_.desc, this_.funcType)
	}
	return fmt.Sprintf("[%s] handler: %s", this_.desc, this_.funcType)
}

func (this_ *Elem) Call(c *ctx.Context) *errmsg.ErrMsg {
	hf := this_.call(0, c)
	//if this_.isLogMiddleError {
	//	newHf := func(ctx *ctx.Context) *errmsg.ErrMsg {
	//		var errMsg *errmsg.ErrMsg
	//		errMsg = hf(ctx)
	//		if errMsg != nil {
	//			fn := runtime.FuncForPC(reflect.ValueOf(hf).Pointer()).Name()
	//			c.TraceLogger.Debug("call middleware error", zap.String("func", fn), zap.Error(errMsg))
	//		}
	//		return errMsg
	//	}
	//	return newHf(c)
	//}
	return hf(c)
}

func (this_ *Elem) call(i int, c *ctx.Context) HandleFunc {
	if i == len(this_.midS) {
		return this_.h
	}
	hf := this_.call(i+1, c)
	//if this_.isLogMiddleError {
	//	newHf := func(ctx *ctx.Context) *errmsg.ErrMsg {
	//		var errMsg *errmsg.ErrMsg
	//		errMsg = hf(ctx)
	//		if errMsg != nil {
	//			fn := runtime.FuncForPC(reflect.ValueOf(hf).Pointer()).Name()
	//			c.TraceLogger.Debug("call middleware error", zap.String("func", fn), zap.Error(errMsg))
	//		}
	//		return errMsg
	//	}
	//
	//	return this_.midS[i](newHf)
	//}
	return this_.midS[i](hf)
}

type Handler struct {
	m                map[string]*Elem
	subjMap          map[string]bool
	groupHandlers    []MiddleWare
	el               *eventloop.EventLoop
	isDebug          bool
	isLogMiddleError bool
}

func NewHandler(el *eventloop.EventLoop, isDebug bool, isLogMiddleError bool, ware ...MiddleWare) *Handler {
	h := &Handler{
		m:                map[string]*Elem{},
		subjMap:          map[string]bool{},
		el:               el,
		isDebug:          isDebug,
		isLogMiddleError: isLogMiddleError,
	}
	h.groupHandlers = make([]MiddleWare, 0, len(ware))
	h.groupHandlers = append(h.groupHandlers, ware...)
	return h
}

func (this_ *Handler) Group(mid MiddleWare, midS ...MiddleWare) *Handler {
	h := &Handler{
		m:                this_.m,
		subjMap:          this_.subjMap,
		el:               this_.el,
		isDebug:          this_.isDebug,
		isLogMiddleError: this_.isLogMiddleError,
	}
	h.groupHandlers = make([]MiddleWare, 0, len(this_.groupHandlers)+len(midS)+1)
	h.groupHandlers = append(h.groupHandlers, this_.groupHandlers...)
	h.groupHandlers = append(h.groupHandlers, mid)
	h.groupHandlers = append(h.groupHandlers, midS...)
	return h
}

func (this_ *Handler) AddHandler(ware MiddleWare) {
	this_.groupHandlers = append(this_.groupHandlers, ware)
}

var (
	expected      = ",expected: func(*Context,proto.Message)(proto.Message,*errmsg.ErrMsg)"
	expectedEvent = ",expected: func(*Context,proto.Message)"
)

func (this_ *Handler) GetHandlers() []*Elem {
	es := make([]*Elem, 0, len(this_.m))
	for _, v := range this_.m {
		e := *v
		es = append(es, &e)
	}
	sort.Slice(es, func(i, j int) bool {
		return es[i].url < es[j].url
	})
	return es
}

var handleRespName = (&models.Resp{}).XXX_MessageName()

func (this_ *Handler) Handle(c *ctx.Context) {
	defer c.Release()
	var elem *Elem
	var err *errmsg.ErrMsg
	if c.Req != nil {
		msgName := proto.MessageName(c.Req)
		var ok bool
		elem, ok = this_.m[msgName]
		if !ok {
			c.Error("handle not found", zap.String("req", msgName))
			RespErr(c, errmsg.NewProtocolErrorInfo(fmt.Sprintf("not found url:%s", msgName)))
			return
		}
		err = elem.Call(c)
	} else if c.F != nil {
		c.StartTime = timer.Now().UnixNano()
		elem = &Elem{
			h: func(ctx *ctx.Context) *errmsg.ErrMsg {
				ctx.F(ctx)
				return nil
			},
			midS:             this_.groupHandlers,
			isEvent:          true,
			isLogMiddleError: this_.isLogMiddleError,
		}
		err = elem.Call(c)
	} else {
		return
	}

	if !elem.isEvent {
		var out *models.Resp
		if err != nil {
			out = (*models.Resp)(err)
			out.StackStace = nil
			out.ErrInternalInfo = ""
			out.OtherMsg = nil
			out.Resp = nil
			out.OtherRequest = nil
		} else {
			out = &models.Resp{}
			if c.Resp != nil {
				out.Resp = &types.Any{}
				msgToAny(c.Resp, out.Resp)
			}
			out.OtherRequest = c.OtherRequest
			if len(c.OtherMsg) > 0 {
				if c.ServerType == models.ServerType_GatewayStdTcp {
					out.OtherMsg = make([]*types.Any, len(c.OtherMsg))
					for i, v := range c.OtherMsg {
						tv := new(types.Any)
						msgToAny(v, tv)
						out.OtherMsg[i] = tv
					}
				} else {
					c.PushMessages = append(c.PushMessages, &ctx.PushMessage{
						Messages: c.OtherMsg,
						Roles:    []string{c.RoleId},
					})
				}
			}
			if len(c.PushMessages) > 0 {
				if this_.isDebug {
					c.Debug("push messages", zap.Any("push", c.PushMessages))
				}
				this_.el.PostEventQueue(c.MarshalPush())
			}
			if len(c.EventRemote) > 0 {
				if this_.isDebug {
					c.Debug("push EventRemote to other role", zap.Any("push", c.EventRemote))
				}
				this_.el.PostEventQueue(c.MarshalEventRemote())
			}
		}
		size := protocol.GetEncodeInternalToSize(nil, out)
		dh := bytespool.GetSample(size)
		defer bytespool.PutSample(dh)
		err = protocol.EncodeInternalFrom(dh, nil, out)
		if err != nil {
			errOut := (*models.Resp)(err)
			size := protocol.GetEncodeInternalToSize(nil, errOut)
			dh1 := bytespool.GetSample(size)
			dh = dh1
			defer bytespool.PutSample(dh1)
			err = protocol.EncodeInternalFrom(dh, nil, errOut)
			if err != nil {
				c.Error("protocol.EncodeInternalFrom (errOut) error", zap.Error(err), zap.Any("errOut", errOut))
				return
			}
		}
		writeErr := c.Respond(handleRespName, dh)
		if writeErr != nil {
			c.Error("nats write error", zap.Error(writeErr))
		}
	} else {
		c.MergePushMessage()
		if len(c.PushMessages) > 0 {
			if this_.isDebug {
				c.Debug(fmt.Sprintf("push messages"), zap.Any("push", c.PushMessages))
			}
			this_.el.PostEventQueue(c.MarshalPush())
		}
		if len(c.EventRemote) > 0 {
			if this_.isDebug {
				c.Debug("push EventRemote to other role", zap.Any("push", c.PushMessages))
			}
			this_.el.PostEventQueue(c.MarshalEventRemote())
		}
	}
}

var zapFiledPool = sync.Pool{
	New: func() interface{} {
		return make([]zap.Field, 0, 16)
	},
}

func (this_ *Handler) GetSubjArray() []string {
	out := make([]string, 0, len(this_.subjMap))
	for k := range this_.subjMap {
		out = append(out, k)
	}
	return out
}

func RespErr(ctx *ctx.Context, err error) {
	e, ok := err.(*errmsg.ErrMsg)
	if !ok {
		e = errmsg.NewNormalErr(errmsg.InternalErrMsg, err.Error())
	}
	h := (*models.Resp)(e)
	size := protocol.GetEncodeInternalToSize(nil, h)
	dh := bytespool.GetSample(size)
	defer bytespool.PutSample(dh)
	err = protocol.EncodeInternalFrom(dh, nil, h)
	if err != nil {
		ctx.Error("nats respond failed", zap.Error(err), zap.Any("h", h))
		return
	}
	err1 := ctx.Respond(handleRespName, dh)
	if err1 != nil {
		ctx.Error("nats respond failed", zap.Error(err1), zap.ByteString("data", dh))
	}
}

func msgToAny(msg proto.Message, any *types.Any) {
	any.TypeUrl = proto.MessageName(msg)
	any.Value, _ = proto.Marshal(msg)
}

func (this_ *Handler) RegisterFunc(desc string, v interface{}, midWares ...MiddleWare) {
	tf := reflect.TypeOf(v)
	if tf.Kind() != reflect.Func {
		panic("Handler.RegisterFunc: param v must be a func" + expected)
	}
	if tf.NumIn() != 2 {
		panic("Handler.RegisterFunc: in params num must is 2" + expected)
	}
	if tf.NumOut() != 2 {
		panic("Handler.RegisterFunc: out params num must is 2" + expected)
	}
	in0t := tf.In(0)
	if in0t.Kind() != reflect.TypeOf((*ctx.Context)(nil)).Kind() {
		panic("Handler.RegisterFunc: in params 0 not is *ctx.Context" + expected)
	}
	in1t := tf.In(1)
	in1tInterface := reflect.New(in1t.Elem()).Interface()
	if _, ok := in1tInterface.(proto.Message); !ok {
		panic("Handler.RegisterFunc: in params 1 not is proto.Message" + expected)
	}

	out0t := tf.Out(0)
	out0Interface := reflect.New(out0t.Elem()).Interface()
	if _, ok := out0Interface.(proto.Message); !ok {
		panic("Handler.RegisterFunc: out params 0 not is proto.Message" + expected)
	}
	out1t := tf.Out(1)
	if out1t.Kind() != reflect.TypeOf((*errmsg.ErrMsg)(nil)).Kind() {
		panic("Handler.RegisterFunc: out params 1 not is error" + expected)
	}
	tv := reflect.ValueOf(v)
	in1, ok := in1tInterface.(proto.Message)
	if !ok {
		panic("Handler.RegisterFunc: in params 1 not is proto.Message(github.com/gogo/protobuf/proto)" + expected)
	}
	messageName := proto.MessageName(in1)
	if _, ok := this_.m[messageName]; ok {
		panic(fmt.Sprintf("Handler %s already registered!", messageName))
	}
	out0MessageName := proto.MessageName(out0Interface.(proto.Message))
	moduleName := strings.Split(messageName, ".")
	if len(moduleName) < 3 {
		panic(fmt.Sprintf("in param 1 must use child message:%s", messageName))
	}
	subj := moduleName[0]
	if _, ok := this_.subjMap[subj]; !ok {
		this_.subjMap[subj] = true
	}
	midS := make([]MiddleWare, 0, 16)
	//midS = append(midS, setModuleName(module))
	midS = append(midS, this_.groupHandlers...)
	midS = append(midS, midWares...)
	call := getCall(tv, out0MessageName)
	this_.m[messageName] = &Elem{
		h:                call,
		desc:             desc,
		midS:             midS,
		url:              messageName,
		funcType:         fmt.Sprintf("func(c *ctx.Context, %s) (%s, *errmsg.ErrMsg)", reflect.TypeOf(in1tInterface).String(), reflect.TypeOf(out0Interface).String()),
		isLogMiddleError: this_.isLogMiddleError,
	}
}

func getCall(tv reflect.Value, out0MessageName string) HandleFunc {
	return func(c *ctx.Context) *errmsg.ErrMsg {
		out := tv.Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(c.Req)})
		if !out[1].IsNil() {
			return out[1].Interface().(*errmsg.ErrMsg)
		}
		respI := out[0].Interface()
		if out[0].IsNil() {
			respI = msgcreate.NewMessage(out0MessageName)
		}
		c.Resp = respI.(proto.Message)
		return nil
	}
}

func (this_ *Handler) RegisterEvent(desc string, v interface{}, midWares ...MiddleWare) {
	tf := reflect.TypeOf(v)
	if tf.Kind() != reflect.Func {
		panic("Handler.RegisterFunc: param v must be a func" + expectedEvent)
	}
	if tf.NumIn() != 2 {
		panic("Handler.RegisterFunc: in params num must is 2" + expectedEvent)
	}
	if tf.NumOut() != 0 {
		panic("Handler.RegisterFunc: out params num must is 0" + expectedEvent)
	}
	in0t := tf.In(0)
	if in0t.Kind() != reflect.TypeOf((*ctx.Context)(nil)).Kind() {
		panic("Handler.RegisterFunc: in params 0 not is *Context" + expectedEvent)
	}
	in1t := tf.In(1)
	in1tInterface := reflect.New(in1t.Elem()).Interface()
	if _, ok := in1tInterface.(proto.Message); !ok {
		panic("Handler.RegisterFunc: in params 1 not is proto.Message" + expectedEvent)
	}

	tv := reflect.ValueOf(v)
	in1, ok := in1tInterface.(proto.Message)
	if !ok {
		panic("Handler.RegisterFunc: in params 1 not is proto.Message(github.com/gogo/protobuf/proto)" + expectedEvent)
	}
	messageName := proto.MessageName(in1)
	if _, ok := this_.m[messageName]; ok {
		panic(fmt.Sprintf("Handler %s already registered!", messageName))
	}
	moduleName := strings.Split(messageName, ".")
	if len(moduleName) < 3 {
		panic(fmt.Sprintf("in param 1 must use child message:%s", messageName))
	}
	subj := moduleName[0]
	if _, ok := this_.subjMap[subj]; !ok {
		this_.subjMap[subj] = true
	}
	midS := make([]MiddleWare, 0, 16)
	midS = append(midS, this_.groupHandlers...)
	midS = append(midS, midWares...)
	call := getCallEvent(tv)
	this_.m[messageName] = &Elem{
		h:                call,
		desc:             desc,
		midS:             midS,
		isEvent:          true,
		url:              messageName,
		funcType:         fmt.Sprintf("func(c *ctx.Context, %s) ", reflect.TypeOf(in1tInterface).String()),
		isLogMiddleError: this_.isLogMiddleError,
	}
}

func getCallEvent(tv reflect.Value) HandleFunc {
	return func(c *ctx.Context) *errmsg.ErrMsg {
		tv.Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(c.Req)})
		return nil
	}
}
