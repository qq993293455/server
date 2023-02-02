package ctx

import (
	"context"
	"sync"
	"time"

	"coin-server/common/statistical2"
	"coin-server/common/values/env"

	"github.com/rs/xid"

	"go.opentelemetry.io/otel"

	"coin-server/common/dlock"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/orm"
	"coin-server/common/proto/models"
	readpb "coin-server/common/proto/role_state_read"
	"coin-server/common/protocol"
	"coin-server/common/statistical"
	"coin-server/common/values"

	"github.com/FJSDS/redislock"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/nats-io/nats.go"
)

type Context struct {
	context.Context
	*models.ServerHeader
	*logger.TraceLogger

	data   map[string]interface{}
	locker *sync.RWMutex

	Module       string // 是属于哪个模块
	Msg          *nats.Msg
	RespondFunc  func(string, []byte) *errmsg.ErrMsg
	Req          proto.Message
	Resp         proto.Message
	OtherMsg     []proto.Message //需要连带发给客户端的消息
	OtherRequest []string        // 需要客户端接下来请求的消息
	F            func(ctx *Context)

	// 推送给其他玩家的消息
	PushMessages []*PushMessage

	// 本地事件
	EventLocal  []interface{}
	EventRemote []*EventRemote
	db          *orm.Orm
	ls          *statistical.LogServer
	ls2         *statistical2.LogServer

	drLocker *dlock.RecursiveLock
}

type PushMessage struct {
	Messages []proto.Message
	Roles    []string
}

type EventRemote struct {
	RoleId      values.RoleId
	ServerId    values.Integer
	Data        proto.Message
	StartTime   values.Integer
	RuleVersion string
	TraceId     string
	UserId      string
}

func (c *Context) PublishEventLocal(i interface{}) {
	c.EventLocal = append(c.EventLocal, i)
}

func (c *Context) PublishEventRemote(roleId values.RoleId, serverId values.Integer, userId string, data proto.Message) {
	if len(c.EventRemote) == 0 {
		c.EventRemote = make([]*EventRemote, 0, 8)
	}
	c.EventRemote = append(c.EventRemote, &EventRemote{
		RoleId:      roleId,
		ServerId:    serverId,
		Data:        data,
		StartTime:   c.StartTime,
		RuleVersion: c.RuleVersion,
		TraceId:     c.TraceId,
		UserId:      userId,
	})
}

func (c *Context) MarshalPush() []*readpb.RoleStateROnly_PushToClient {
	if len(c.PushMessages) == 0 {
		return nil
	}
	pcs := make([]*readpb.RoleStateROnly_PushToClient, 0, len(c.PushMessages))
	for _, v := range c.PushMessages {
		pc := &readpb.RoleStateROnly_PushToClient{
			Roles:    make([]string, 0, len(v.Roles)),
			Messages: make([]*types.Any, 0, len(v.Messages)),
		}
		pc.Roles = append(pc.Roles, v.Roles...)
		for _, msg := range v.Messages {
			pc.Messages = append(pc.Messages, protocol.MsgToAny(msg))
		}
		pcs = append(pcs, pc)
	}
	return pcs
}

func (c *Context) MarshalEventRemote() []*EventRemote {
	if len(c.EventRemote) == 0 {
		return nil
	}
	list := make([]*EventRemote, 0, len(c.EventRemote))
	for _, v := range c.EventRemote {
		list = append(list, &EventRemote{
			RoleId:      v.RoleId,
			ServerId:    v.ServerId,
			StartTime:   v.StartTime,
			RuleVersion: v.RuleVersion,
			TraceId:     v.TraceId,
			Data:        v.Data,
			UserId:      v.UserId,
		})
	}
	return list
}

func (c *Context) MergePushMessage() {
	if len(c.OtherMsg) > 0 && c.RoleId != "" {
		pm := pmPool.Get().(*PushMessage)
		pm.Roles = append(pm.Roles, c.RoleId)
		pm.Messages = c.OtherMsg
		c.PushMessages = append(c.PushMessages, pm)
	}
}

var pmPool = sync.Pool{New: func() interface{} {
	return &PushMessage{}
}}

func GetPushMessage() *PushMessage {
	return pmPool.Get().(*PushMessage)
}

func PutPushMessage(pm *PushMessage) {
	pmPool.Put(pm)
}

func (c *Context) Reset() {

	c.RespondFunc = nil
	c.ServerHeader = nil
	// c.TraceLogger.ResetInitFiledS(traceId, roleId)
	for k := range c.data {
		delete(c.data, k)
	}
	c.TraceLogger.ReleasePool()
	c.TraceLogger = nil
	c.Msg = nil
	c.Module = ""
	c.Req = nil
	c.Resp = nil
	c.OtherMsg = c.OtherMsg[:0]
	c.OtherRequest = c.OtherRequest[:0]
	c.EventLocal = c.EventLocal[:0]
	c.EventRemote = c.EventRemote[:0]
	c.Context = context.Background()
	c.F = nil

	for _, v := range c.PushMessages {
		v.Roles = v.Roles[:0]
		v.Messages = v.Messages[:0]
		pm := v
		pmPool.Put(pm)
	}
	c.PushMessages = c.PushMessages[:0]
	if c.db != nil {
		orm.PutOrm(c.db)
		c.db = nil
	}
	if c.ls != nil {
		statistical.PutLogServer(c.ls)
		c.ls = nil
	}
	if c.ls2 != nil {
		statistical2.PutLogServer(c.ls2)
		c.ls2 = nil
	}
	_ = c.DRUnlock()
}

func (c *Context) GetOrmForMiddleWare() *orm.Orm {
	return c.db
}

func (c *Context) NewOrm() *orm.Orm {
	if c.db == nil {
		c.db = orm.GetOrm(c)
	}
	return c.db
}

func (c *Context) NewLogServer() *statistical.LogServer {
	if c.ls == nil {
		c.ls = statistical.GetLogServer(c)
	}
	return c.ls
}

func (c *Context) NewLogServer2() *statistical2.LogServer {
	if c.ls2 == nil {
		c.ls2 = statistical2.GetLogServer(c)
	}
	return c.ls2
}

func (c *Context) Release() {
	c.Reset()
	ctxPool.Put(c)
}

var lockTracer = otel.Tracer("Lock")

func (c *Context) DRLock(client *redislock.Client, key ...string) *errmsg.ErrMsg {
	if env.OpenTracing() {
		_, span := lockTracer.Start(c.Context, "DRLock")
		defer span.End()
	}
	if c.drLocker == nil {
		c.drLocker = dlock.GetLocker()
	}
	e := c.drLocker.Lock(client, time.Second*5, key...)
	if e != nil {
		return errmsg.NewErrorDB(e)
	}
	return nil
}

func (c *Context) DRUnlock() *errmsg.ErrMsg {
	if c.drLocker != nil {
		e := c.drLocker.Unlock()
		dlock.PutLocker(c.drLocker)
		c.drLocker = nil
		if e != nil {
			return errmsg.NewErrorDB(e)
		}
	}
	return nil
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *Context) Err() error {
	return c.Context.Err()
}

func (c *Context) Value(key interface{}) interface{} {
	switch k := key.(type) {
	case string:
		return c.GetValue(k)
	default:
		return c.Context.Value(key)
	}
}

func (c *Context) PushMessage(msg proto.Message) {
	c.OtherMsg = append(c.OtherMsg, msg)
}

// PushMessageToRole 给其他用户推送消息
func (c *Context) PushMessageToRole(roleId values.RoleId, msg proto.Message, msgS ...proto.Message) {
	if roleId == c.RoleId {
		c.OtherMsg = append(c.OtherMsg, msg)
		if len(msgS) > 0 {
			c.OtherMsg = append(c.OtherMsg, msgS...)
		}
		return
	}
	pm := pmPool.Get().(*PushMessage)
	pm.Roles = append(pm.Roles, roleId)
	pm.Messages = append(pm.Messages, msg)
	if len(msgS) > 0 {
		pm.Messages = append(pm.Messages, msgS...)
	}
	c.PushMessages = append(c.PushMessages, pm)
}

// PushMessageToRoles 给其他用户推送消息
func (c *Context) PushMessageToRoles(roleIds []values.RoleId, msg proto.Message, msgS ...proto.Message) {
	if len(roleIds) == 0 {
		return
	}
	pm := pmPool.Get().(*PushMessage)
	pm.Roles = append(pm.Roles, roleIds...)
	pm.Messages = append(pm.Messages, msg)
	if len(msgS) > 0 {
		pm.Messages = append(pm.Messages, msgS...)
	}
	c.PushMessages = append(c.PushMessages, pm)
}

func (c *Context) AppendOtherRequest(req string) {
	c.OtherRequest = append(c.OtherRequest, req)
}

var ctxPool = sync.Pool{New: func() interface{} {
	return &Context{
		Context: context.Background(),
		data:    map[string]interface{}{},
		locker:  &sync.RWMutex{},
	}
}}

func GetPoolContext() *Context {
	c := ctxPool.Get().(*Context)
	if c.TraceLogger != nil {
		panic("c.TraceLogger !=nil")
	}
	c.TraceLogger = logger.GetTraceLogger()
	return c
}

func GetContext() *Context {
	traceId := xid.New().String()
	res := &Context{
		Context: context.Background(),
		ServerHeader: &models.ServerHeader{
			TraceId: traceId,
		},
		TraceLogger: logger.DefaultLogger.WithTrace(traceId, ""),
		data:        map[string]interface{}{},
		locker:      &sync.RWMutex{},
	}
	return res
}

func NewContext(header *models.ServerHeader, req proto.Message, msg *nats.Msg) *Context {
	res := &Context{
		Context:      context.Background(),
		ServerHeader: header,
		TraceLogger:  logger.DefaultLogger.WithTrace(header.TraceId, header.RoleId),
		data:         map[string]interface{}{},
		locker:       &sync.RWMutex{},
		Msg:          msg,
		Req:          req,
	}
	return res
}

func (c *Context) Respond(msgName string, data []byte) error {
	if c.RespondFunc != nil {
		e := c.RespondFunc(msgName, data)
		if e != nil {
			return e
		}
		return nil
	}
	return c.Msg.Respond(data)
}

// SetValue set a key and value to parent, lazy load
func (c *Context) SetValue(key string, value interface{}) {
	// 增加锁逻辑，保证数据正常读写
	c.locker.Lock()
	defer c.locker.Unlock()

	if c.data == nil {
		c.data = make(map[string]interface{})
	}

	c.data[key] = value
}

// GetValue get value
func (c *Context) GetValue(key string) interface{} {
	// 增加锁逻辑，保证数据正常读写
	c.locker.RLock()
	defer c.locker.RUnlock()

	if c.data != nil {
		if value, ok := c.data[key]; ok {
			return value
		}
	}
	return c.Context.Value(key)
}

func NewHeader(roleId values.RoleId, thisServerId values.ServerId, thisServerType models.ServerType, c *Context) *models.ServerHeader {
	newC := &models.ServerHeader{}
	newC.RoleId = roleId
	newC.ServerId = thisServerId
	newC.ServerType = thisServerType
	if c != nil {
		newC.StartTime = c.StartTime
		newC.TraceId = c.TraceId
		newC.RuleVersion = c.RuleVersion
	}
	return newC
}

func NewHeaderWithOutCtx(roleId values.RoleId, userId string, thisServerId values.ServerId, thisServerType models.ServerType, startTime values.Integer, tracId, ruleVersion string) *models.ServerHeader {
	newC := &models.ServerHeader{}
	newC.RoleId = roleId
	newC.ServerId = thisServerId
	newC.ServerType = thisServerType
	newC.StartTime = startTime
	newC.TraceId = tracId
	newC.RuleVersion = ruleVersion
	newC.UserId = userId
	return newC
}
