package natsclient

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"coin-server/common/bytespool"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	"coin-server/common/proto/models"
	"coin-server/common/protocol"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values"

	"github.com/gogo/protobuf/proto"

	"github.com/rs/xid"

	"go.uber.org/zap"

	"github.com/nats-io/nats.go"
)

var (
	ErrInvalidRequestLength = errmsg.NewProtocolErrorInfo("invalid request length")
)

type NatsClient struct {
	subs       map[string]*nats.Subscription
	name       string
	urls       string
	log        *logger.Logger
	conn       *nats.Conn
	serverId   int64
	serverType models.ServerType
	closed     int32
}

func NewNatsClient(serverType models.ServerType, serverId int64, urls string, log *logger.Logger) *NatsClient {
	name := serverType.String()
	if serverId > 0 {
		name = fmt.Sprintf("%s:%d", serverType.String(), serverId)
	}
	nc := &NatsClient{
		subs:       map[string]*nats.Subscription{},
		name:       name,
		urls:       urls,
		log:        log,
		serverType: serverType,
		serverId:   serverId,
	}
	c, err := nats.Connect(urls, nats.ReconnectWait(time.Millisecond*10), nats.MaxReconnects(math.MaxInt64),
		nats.PingInterval(time.Second*3), nats.MaxPingsOutstanding(2), nats.Timeout(time.Second),
		nats.DrainTimeout(time.Second*5), nats.Name(name),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			if atomic.LoadInt32(&nc.closed) == 0 {
				log.Error("nats disconnected", zap.Error(err), zap.String("urls", urls), zap.String("nats-server", conn.ConnectedAddr()))
			}
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			log.Warn("nats reconnected", zap.String("urls", urls), zap.String("nats-server", conn.ConnectedAddr()))
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			log.Warn("nats closed", zap.String("urls", urls), zap.String("nats-server", conn.ConnectedAddr()))
		}),
	)
	if err != nil {
		panic(err)
	}
	nc.conn = c
	return nc
}

func (this_ *NatsClient) Close() {
	if this_.subs != nil {
		for _, v := range this_.subs {
			if v.IsValid() {
				_ = v.Drain()
			}
		}
		for _, v := range this_.subs {
			for v.IsValid() {
				time.Sleep(time.Millisecond * 10)
			}
		}
	}
}

func (this_ *NatsClient) PublishCtx(c *ctx.Context, serverId int64, msg proto.Message) *errmsg.ErrMsg {
	return this_.Publish(serverId, c.ServerHeader, msg)
}

func (this_ *NatsClient) Publish(serverId int64, h *models.ServerHeader, msg proto.Message) *errmsg.ErrMsg {
	if h != nil && (h.ServerId != this_.serverId || h.ServerType != this_.serverType) {
		os, ot := h.ServerId, h.ServerType
		h.ServerId, h.ServerType = this_.serverId, this_.serverType
		defer func() {
			h.ServerId, h.ServerType = os, ot
		}()
	}
	n := protocol.GetEncodeInternalToSize(h, msg)
	d := bytespool.GetSample(n)
	defer bytespool.PutSample(d)
	err := protocol.EncodeInternalFrom(d, h, msg)
	if err != nil {
		return err
	}
	msgName := proto.MessageName(msg)
	if serverId != 0 {
		n := strings.IndexByte(msgName, '.')
		if n == -1 {
			return errmsg.NewProtocolErrorInfo(fmt.Sprintf("header.Data.TypeUrl is not a proto.MessageName:%s", msgName))
		}

		b := bytespool.GetSample(len(msgName) + 10)
		defer bytespool.PutSample(b)
		b = b[:0]
		b = append(b, msgName[:n]...)
		b = append(b, '.')
		b = strconv.AppendInt(b, serverId, 10)
		b = append(b, msgName[n:]...)
		msgName = *(*string)(unsafe.Pointer(&b))
	}
	// 	this_.log.Info("publish", zap.String("msgName", msgName), zap.Any("msg", msg))
	return errmsg.NewProtocolError(this_.conn.Publish(msgName, d))
}

func (this_ *NatsClient) PublishRawData(serverId int64, h *models.ServerHeader, msgName string, msgData []byte) *errmsg.ErrMsg {
	if h != nil && (h.ServerId != this_.serverId || h.ServerType != this_.serverType) {
		os, ot := h.ServerId, h.ServerType
		h.ServerId, h.ServerType = this_.serverId, this_.serverType
		defer func() {
			h.ServerId, h.ServerType = os, ot
		}()
	}
	n := protocol.GetEncodeInternalDataToSize(h, msgData)
	d := bytespool.GetSample(n)
	defer bytespool.PutSample(d)
	err := protocol.EncodeInternalDataFrom(d, h, msgData)
	if err != nil {
		return err
	}
	if serverId != 0 {
		n := strings.IndexByte(msgName, '.')
		if n == -1 {
			return errmsg.NewProtocolErrorInfo(fmt.Sprintf("header.Data.TypeUrl is not a proto.MessageName:%s", msgName))
		}

		b := bytespool.GetSample(len(msgName) + 10)
		defer bytespool.PutSample(b)
		b = b[:0]
		b = append(b, msgName[:n]...)
		b = append(b, '.')
		b = strconv.AppendInt(b, serverId, 10)
		b = append(b, msgName[n:]...)
		msgName = *(*string)(unsafe.Pointer(&b))
	}
	// 	this_.log.Info("publish", zap.String("msgName", msgName), zap.Any("msg", msg))
	return errmsg.NewProtocolError(this_.conn.Publish(msgName, d))
}

func (this_ *NatsClient) Shutdown() {
	if atomic.CompareAndSwapInt32(&this_.closed, 0, 1) {
		_ = this_.conn.FlushTimeout(time.Second * 3)
		this_.conn.Close()
	}
}

func (this_ *NatsClient) Subscribe(subj string, h nats.MsgHandler) {
	if _, ok := this_.subs[subj]; ok {
		panic(fmt.Sprintf("subj [%s] had Subscribed", subj))
	}
	this_.log.Info("Subscribe", zap.String("urls", this_.urls), zap.String("subj", subj))
	sub, err := this_.conn.Subscribe(subj, h)
	utils.Must(err)
	this_.subs[subj] = sub
}

const HeaderLen = 2

func (this_ *NatsClient) RequestWithOut(c *ctx.Context, serverId int64, msg proto.Message, out proto.Message, timeout ...time.Duration) *errmsg.ErrMsg {
	o, err := this_.Request(c, serverId, msg, timeout...)
	if err != nil {
		return err
	}
	if o.Resp == nil {
		return nil
	}
	e := proto.Unmarshal(o.Resp.Value, out)
	if e != nil {
		return errmsg.NewProtocolError(err)
	}
	return nil
}

func (this_ *NatsClient) RequestProto(serverId int64, header *models.ServerHeader, msg proto.Message, timeout ...time.Duration) ([]byte, *errmsg.ErrMsg) {
	if header != nil && (header.ServerId != this_.serverId || header.ServerType != this_.serverType) {
		os, ot := header.ServerId, header.ServerType
		header.ServerId, header.ServerType = this_.serverId, this_.serverType
		defer func() {
			header.ServerId, header.ServerType = os, ot
		}()
	}
	n := protocol.GetEncodeInternalToSize(header, msg)
	dh := bytespool.GetSample(n)
	defer bytespool.PutSample(dh)
	err := protocol.EncodeInternalFrom(dh, header, msg)
	if err != nil {
		return nil, err
	}
	msgName := proto.MessageName(msg)
	if serverId != 0 {
		n := strings.IndexByte(msgName, '.')
		if n == -1 {
			return nil, errmsg.NewProtocolErrorInfo(fmt.Sprintf("header.Data.TypeUrl is not a proto.MessageName:%s", msgName))
		}

		b := bytespool.GetSample(len(msgName) + 10)
		defer bytespool.PutSample(b)
		b = b[:0]
		b = append(b, msgName[:n]...)
		b = append(b, '.')
		b = strconv.AppendInt(b, serverId, 10)
		b = append(b, msgName[n:]...)
		msgName = *(*string)(unsafe.Pointer(&b))
	}
	to := time.Second * 10
	if len(timeout) > 0 {
		to = timeout[0]
	}
	outMsg, e := this_.conn.Request(msgName, dh, to)
	if e != nil {
		if e == nats.ErrNoResponders {
			return nil, errmsg.NewErrorNatsNoResponders(msgName + ":" + e.Error())
		}
		return nil, errmsg.NewProtocolErrorInfo(msgName + ":" + e.Error())
	}

	return outMsg.Data, nil
}

func (this_ *NatsClient) RequestData(serverId int64, header *models.ServerHeader, msgName string, data []byte) ([]byte, *errmsg.ErrMsg) {
	if header != nil && (header.ServerId != this_.serverId || header.ServerType != this_.serverType) {
		os, ot := header.ServerId, header.ServerType
		header.ServerId, header.ServerType = this_.serverId, this_.serverType
		defer func() {
			header.ServerId, header.ServerType = os, ot
		}()
	}
	n := protocol.GetEncodeInternalDataToSize(header, data)
	dh := bytespool.GetSample(n)
	defer bytespool.PutSample(dh)
	err := protocol.EncodeInternalDataFrom(dh, header, data)
	if err != nil {
		return nil, err
	}

	if serverId != 0 {
		n := strings.IndexByte(msgName, '.')
		if n == -1 {
			return nil, errmsg.NewProtocolErrorInfo(fmt.Sprintf("header.Data.TypeUrl is not a proto.MessageName:%s", msgName))
		}
		b := bytespool.GetSample(len(msgName) + 10)
		defer bytespool.PutSample(b)
		b = b[:0]
		b = append(b, msgName[:n]...)
		b = append(b, '.')
		b = strconv.AppendInt(b, serverId, 10)
		b = append(b, msgName[n:]...)
		msgName = *(*string)(unsafe.Pointer(&b))
	}
	outMsg, e := this_.conn.Request(msgName, dh, time.Second*10)
	if e != nil {
		if e == nats.ErrNoResponders {
			return nil, errmsg.NewErrorNatsNoResponders(msgName + ":" + e.Error())
		}
		return nil, errmsg.NewProtocolErrorInfo(msgName + ":" + e.Error())
	}
	return outMsg.Data, nil
}

func (this_ *NatsClient) Request(ctx *ctx.Context, serverId values.ServerId, msg proto.Message, timeout ...time.Duration) (*models.Resp, *errmsg.ErrMsg) {
	out, err := this_.RequestProto(serverId, ctx.ServerHeader, msg, timeout...)
	if err != nil {
		return nil, err
	}
	outResp := &models.Resp{}
	err = protocol.DecodeInternal(out, nil, outResp)
	if err != nil {
		return nil, err
	}
	if outResp.ErrCode != 0 {
		return nil, (*errmsg.ErrMsg)(outResp)
	}
	return outResp, nil
}

func (this_ *NatsClient) RequestWithHeader(serverId values.ServerId, header *models.ServerHeader, msg proto.Message) (*models.Resp, *errmsg.ErrMsg) {
	out, err := this_.RequestProto(serverId, header, msg)
	if err != nil {
		return nil, err
	}
	outResp := &models.Resp{}
	err = protocol.DecodeInternal(out, nil, outResp)
	if err != nil {
		return nil, err
	}
	if outResp.ErrCode != 0 {
		return nil, (*errmsg.ErrMsg)(outResp)
	}
	return outResp, nil
}

func (this_ *NatsClient) RequestWithHeaderOut(serverId values.ServerId, header *models.ServerHeader, msg proto.Message, outMsg proto.Message) *errmsg.ErrMsg {
	out, err := this_.RequestWithHeader(serverId, header, msg)
	if err != nil {
		return err
	}
	if out.Resp == nil {
		return nil
	}
	e := proto.Unmarshal(out.Resp.Value, outMsg)
	if e != nil {
		return errmsg.NewProtocolError(err)
	}
	return nil
}

func (this_ *NatsClient) SubscribeHandler(subj string, f func(ctx *ctx.Context)) {
	if _, ok := this_.subs[subj]; ok {
		panic(fmt.Sprintf("subj [%s] had Subscribed", subj))
	}
	group := strings.ReplaceAll(subj, ">", "group")
	this_.log.Info("SubscribeHandler", zap.String("urls", this_.urls), zap.String("subj", subj), zap.String("group", group))
	sub, err := this_.conn.QueueSubscribe(subj, group, this_.defaultSubCB(subj, f))
	utils.Must(err)
	this_.subs[subj] = sub
}

func (this_ *NatsClient) UnSub(subj string) {
	if s, ok := this_.subs[subj]; ok {
		this_.log.Info("Unsubscribe", zap.String("subj", subj))
		s.Unsubscribe()
		delete(this_.subs, subj)
	}
}

func (this_ *NatsClient) SubscribeBroadcast(subj string, f func(ctx *ctx.Context)) {
	if _, ok := this_.subs[subj]; ok {
		panic(fmt.Sprintf("subj [%s] had Subscribed", subj))
	}
	this_.log.Info("SubscribeBroadcast", zap.String("urls", this_.urls), zap.String("subj", subj))
	sub, err := this_.conn.Subscribe(subj, this_.defaultSubCB(subj, f))
	utils.Must(err)
	this_.subs[subj] = sub
}

func DecodeRequest(data []byte, header *models.ServerHeader, msg proto.Message) error {
	if len(data) < HeaderLen {
		return ErrInvalidRequestLength
	}
	hs := int(binary.LittleEndian.Uint16(data))
	err := proto.Unmarshal(data[HeaderLen:HeaderLen+hs], header)
	if err != nil {
		return err
	}
	return proto.Unmarshal(data[hs+HeaderLen:], msg)
}

func parseMsgName(subject string) string {
	n := strings.IndexByte(subject, '.')
	if n == -1 {
		return subject
	}
	b := &strings.Builder{}
	b.WriteString(subject[:n])
	subject = subject[n+1:]
	n = strings.IndexByte(subject, '.')
	b.WriteString(subject[n:])
	return b.String()
}

func (this_ *NatsClient) defaultSubCB(subj string, f func(ctx *ctx.Context)) nats.MsgHandler {
	return func(msg *nats.Msg) {
		data := msg.Data
		var err *errmsg.ErrMsg
		header := &models.ServerHeader{}
		var serverId int64
		var serverType models.ServerType
		var traceId, msgName string
		var roleId values.RoleId
		var req proto.Message
		defer func() {
			if e := recover(); e != nil {
				err = errmsg.NewErrorPanic(e)
			}
			if err != nil {
				if msg.Reply != "" {
					var resp *models.Resp
					resp = (*models.Resp)(err)
					size := protocol.GetEncodeInternalToSize(nil, resp)
					dh := bytespool.GetSample(size)
					defer bytespool.PutSample(dh)
					err = protocol.EncodeInternalFrom(dh, nil, resp)
					if err != nil {
						this_.log.Error("protocol.EncodeInternalFrom error", zap.String(values.HeaderTraceId, traceId), zap.String("roleId", roleId),
							zap.Error(err), zap.Any("resp", resp), zap.String("subject", subj), zap.Int64("serverId", serverId),
							zap.Int64("serverType", int64(serverType)), zap.String("msgName", msgName))
						return
					}
					err1 := msg.Respond(dh)
					if err1 != nil {
						this_.log.Error("msg.Respond error", zap.String(values.HeaderTraceId, traceId), zap.String("roleId", roleId),
							zap.Error(err1), zap.Error(err), zap.String("subject", subj), zap.Int64("serverId", serverId),
							zap.Int64("serverType", int64(serverType)), zap.String("msgName", msgName))
					} else {
						this_.log.Error("handle msg error", zap.String(values.HeaderTraceId, traceId), zap.String("roleId", roleId),
							zap.Any("resp", resp), zap.String("subject", subj), zap.Int64("serverId", serverId),
							zap.Int64("serverType", int64(serverType)), zap.String("msgName", msgName))
					}
				} else {
					this_.log.Error("handle event error", zap.String(values.HeaderTraceId, traceId), zap.String("roleId", roleId),
						zap.Error(err), zap.String("subject", subj), zap.Int64("serverId", serverId),
						zap.Int64("serverType", int64(serverType)), zap.String("msgName", msgName))
				}
			}
		}()
		msgName = msg.Subject

		if this_.serverId > 0 && !strings.HasPrefix(msgName, protocol.TopicBroadcastPre) {
			msgName = parseMsgName(msg.Subject)
		}
		req = msgcreate.NewMessage(msgName)
		err = protocol.DecodeInternal(data, header, req)
		if err != nil {
			return
		}

		serverId = header.ServerId
		serverType = header.ServerType
		if header.StartTime == 0 {
			header.StartTime = timer.Now().UnixNano()
		}
		if header.TraceId == "" {
			header.TraceId = xid.NewWithTime(timer.Now()).String()
		}
		traceId = header.TraceId
		roleId = header.RoleId
		uc := ctx.GetPoolContext()
		uc.ServerHeader = header
		uc.Req = req
		uc.Msg = msg
		uc.TraceLogger.ResetInitFiledS(header.TraceId, header.RoleId)
		f(uc)
	}
}
