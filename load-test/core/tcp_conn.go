package core

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"coin-server/common/bytespool"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	"coin-server/common/network/stdtcp"
	"coin-server/common/proto/models"
	"coin-server/common/protocol"
	"coin-server/common/values"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

type TcpSession struct {
	sess    *stdtcp.Session
	ctx     *RoleContext
	logger  *logger.Logger
	closed  bool
	once    sync.Once
	pushMap map[string][]PushFunc //map[string]CallBackFunc // 回调
	connOk  chan struct{}
}

func NewTcpConn(ctx *RoleContext, addr string) *TcpSession {
	tc := &TcpSession{
		ctx:     ctx,
		logger:  logger.DefaultLogger,
		pushMap: make(map[string][]PushFunc),
		connOk:  make(chan struct{}),
	}
	stdtcp.Connect(addr, time.Second*3, true, tc, ctx.Logger, false)
	t := time.NewTimer(5 * time.Second)
	defer t.Stop()
	select {
	case <-t.C:
		panic(fmt.Sprintf("connect time out!uid:%s addr:%s", ctx.UserId, addr))
	case <-tc.connOk:
	}
	return tc
}

func (c *TcpSession) Close() {
	c.once.Do(func() {
		c.closed = true
		c.sess.Close(nil)
	})
}

func (c *TcpSession) doPush(msgName string, pushMsg proto.Message) *errmsg.ErrMsg {
	list, ok := c.pushMap[msgName]
	if !ok {
		c.logger.Warn("cannot found pushMap func!", zap.String("msgName", msgName),
			zap.String("uid", string(c.ctx.UserId)), zap.String("roleId", c.ctx.RoleId))
		return nil
	}
	for _, f := range list {
		start := time.Now()
		err := f(c.ctx, pushMsg)
		cost := time.Since(start).Milliseconds()
		if err != nil {
			c.ctx.RecordFailure(msgName, cost, err.Error())
		} else {
			c.ctx.RecordSuccess(msgName, cost, values.Integer(proto.Size(pushMsg)))
		}
	}
	return nil
}

func (c *TcpSession) Request(req proto.Message) (response *models.Resp, res proto.Message, err *errmsg.ErrMsg) {
	start := time.Now()
	name := proto.MessageName(req)
	defer func() {
		cost := time.Since(start).Milliseconds()
		if err != nil {
			c.ctx.RecordFailure(name, cost, err.Error())
		} else if response != nil && response.ErrCode == 1 {
			str := fmt.Sprintf("ErrCode:%d ErrMsg:%s ErrInternal:%s otherMsg:%s", response.ErrCode,
				response.ErrMsg, response.ErrInternalInfo, strings.Join(response.OtherRequest, " "))
			c.ctx.RecordFailure(name, cost, str)
		} else {
			c.ctx.RecordSuccess(name, cost, values.Integer(proto.Size(res)))
		}
	}()
	response = &models.Resp{}
	err = c.sess.RPCRequest(nil, req, &response)
	if err != nil {
		return nil, nil, err
	}
	var mErr *errmsg.ErrMsg
	if response.Resp != nil {
		res, mErr = protocol.AnyToMsg(response.Resp)
		if mErr != nil {
			c.logger.Error("unmarshal resp fail", zap.String("uid", c.ctx.UserId), zap.String("roleId", c.ctx.RoleId),
				zap.Any("request", req), zap.Error(mErr))
			return nil, nil, mErr
		}
	}
	for _, v := range response.OtherMsg {
		pushMsg, tmpErr := protocol.AnyToMsg(v)
		if tmpErr != nil {
			c.logger.Error("unmarshal resp fail", zap.String("uid", c.ctx.UserId), zap.String("roleId", c.ctx.RoleId),
				zap.Any("request", req), zap.Error(tmpErr))
			return nil, nil, tmpErr
		}
		if pErr := c.doPush(v.TypeUrl, pushMsg); pErr != nil {
			return nil, nil, pErr
		}
	}
	if response.ErrCode != 0 {
		c.logger.Warn("request with error code", zap.String("name", name), zap.String("uid", c.ctx.UserId), zap.String("roleId", c.ctx.RoleId),
			zap.Any("request", req), zap.Any("response", res))
		return nil, nil, (*errmsg.ErrMsg)(response)
	}
	//c.logger.Debug("request ok", zap.String("name", name), zap.String("uid", c.ctx.UserId), zap.String("roleId", c.ctx.RoleId),
	//	zap.Any("request", req), zap.Any("response", res))
	return
}

func (c *TcpSession) AsyncSend(req proto.Message) (err *errmsg.ErrMsg) {
	start := time.Now()
	name := proto.MessageName(req)
	defer func() {
		cost := time.Since(start).Milliseconds()
		if err != nil {
			c.ctx.RecordFailure(name, cost, err.Error())
		} else {
			c.ctx.RecordSuccess(name, cost, values.Integer(proto.Size(req)))
		}
	}()
	err = c.write(req, 0)
	if err != nil {
		return err
	}
	return nil
}

func (c *TcpSession) write(req proto.Message, seqId uint32) *errmsg.ErrMsg {
	s := protocol.GetEncodeLen(nil, req)
	data := bytespool.GetSample(s)
	defer bytespool.PutSample(data)
	mErr := protocol.EncodeTCPFrom(data, uint8(models.MsgType_request), seqId, nil, req)
	if mErr != nil {
		return mErr
	}
	length := len(data)
	if length <= 0 {
		return nil
	}
	err := c.sess.Write(data)
	if err != nil {
		return err
	}
	c.logger.Debug("sendMsg", zap.String("uid", c.ctx.UserId), zap.String("roleId", c.ctx.RoleId),
		zap.Any("seqId", seqId), zap.Any("name", proto.MessageName(req)), zap.Any("req", req),
		zap.Int("size", length))
	return nil
}

func (c *TcpSession) RegisterPush(name string, f PushFunc) {
	list := c.pushMap[name]
	if len(list) == 0 {
		c.pushMap[name] = []PushFunc{f}
	} else {
		list = append(list, f)
		c.pushMap[name] = list
	}
}

func (c *TcpSession) OnConnected(session *stdtcp.Session) {
	c.sess = session
	session.SetMeta(c)
	c.connOk <- struct{}{}
	c.logger.Debug("connect ok", zap.Any("uid", c.ctx.UserId), zap.Any("roleId", c.ctx.RoleId),
		zap.Any("remote", session.RemoteAddr()))
}

func (c *TcpSession) OnDisconnected(session *stdtcp.Session, err error) {
	c.logger.Debug("connect closed", zap.Any("uid", c.ctx.UserId), zap.Any("roleId", c.ctx.RoleId))
}

func (c *TcpSession) OnRequest(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) {

}

func (c *TcpSession) OnMessage(session *stdtcp.Session, msgName string, data []byte) {
	c.logger.Debug("OnMessage", zap.Any("uid", c.ctx.UserId), zap.Any("roleId", c.ctx.RoleId), zap.String("msgName", msgName))
	if c.ctx.RoleId == "" {
		return
	}
	var err *errmsg.ErrMsg
	header := &models.ServerHeader{}
	req := msgcreate.NewMessage(msgName)
	err = protocol.DecodeInternal(data, header, req)
	if err != nil {
		c.logger.Error("decode fail", zap.String("msgName", msgName), zap.Error(err))
		return
	}
	err = c.doPush(msgName, req)
	if err != nil {
		c.logger.Error("push msg error!", zap.String("msgName", msgName), zap.Error(err))
	}
}

func panicIfError(err error) {
	e, ok := err.(*errmsg.ErrMsg)
	if ok {
		if e != nil {
			panic(e.String())
		}
		return
	}
	if err != nil {
		panic(err)
	}
}
