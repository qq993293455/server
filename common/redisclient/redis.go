package redisclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"runtime"
	"time"

	"coin-server/common/bytespool"
	"coin-server/common/consulkv"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/gopool"
	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	tendissync "coin-server/common/tendis-sync"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/env"

	"github.com/FJSDS/redislock"
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
	"stathat.com/c/consistent"
)

var (
	Nil = redis.Nil
)

var (
	redisMgr  *pikaManager
	lockerKey = "locker"
	commonKey = "common"
	userKey   = "user"
	guildKey  = "guild"
	gameKey   = "game"
)

func Init(cnf *consulkv.Config) {
	var pikaConfigs []*PikaConfig
	utils.Must(cnf.Unmarshal("pikas", &pikaConfigs))
	pr := &PikaRouter{}
	utils.Must(cnf.Unmarshal("pika-router", pr))
	rc := &RedisCluster{}
	utils.Must(cnf.Unmarshal("redis-cluster", rc))
	redisMgr = newPikaManager(pikaConfigs, pr, rc)
}

func Close() {
	redisMgr.Close()
	redisMgr = nil
}

func GetLocker() *redislock.Client {
	return redisMgr.locker
}

func GetLockerRedis() *redis.Client {
	return redisMgr.lockerRedis
}

func GetCommonRedis() *redis.Client {
	return redisMgr.commonRedis
}

func GetUserRedis() redis.Cmdable {
	if redisMgr.redisCluster != nil {
		return redisMgr.redisCluster
	}
	return redisMgr.gameRedis
}

func GetDefaultRedis() redis.Cmdable {
	if redisMgr.redisCluster != nil {
		return redisMgr.redisCluster
	}
	return redisMgr.gameRedis
}

func GetGuildRedis() redis.Cmdable {
	if redisMgr.redisCluster != nil {
		return redisMgr.redisCluster
	}
	return redisMgr.gameRedis
}

type PikaConfig struct {
	Name     string `json:"name"`
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	TLS      bool   `json:"tls"`
}

type RedisCluster struct {
	AddrS []string `json:"addrs"`
	Pwd   string   `json:"password"`
	TLS   bool     `json:"tls"`
}

type PikaRouter struct {
	Module  map[string]string `json:"module"`
	Servers map[int64]string
}

type RoleRouter struct {
	Start uint64
	End   uint64
	c     *consistent.Consistent
}

type pikaManager struct {
	pikaConfigs       []*PikaConfig
	pikaRouterConfigs *PikaRouter
	pikaMap           map[string]*redis.Client
	// moduleRouterHash  map[string]*consistent.Consistent
	// roleRouterHash []*RoleRouter
	locker       *redislock.Client
	lockerRedis  *redis.Client
	commonRedis  *redis.Client
	gameRedis    *redis.Client
	redisCluster *redis.ClusterClient
}

func (this_ *pikaManager) Close() {
	if this_.pikaMap != nil {
		for _, v := range this_.pikaMap {
			if v != nil {
				v.Close()
			}
		}
	}

	if this_.lockerRedis != nil {
		_ = this_.lockerRedis.Close()
	}
	if this_.commonRedis != nil {
		_ = this_.commonRedis.Close()
	}
	if this_.gameRedis != nil {
		_ = this_.gameRedis.Close()
	}
	if this_.redisCluster != nil {
		_ = this_.redisCluster.Close()
	}
}

func newPikaManager(pc []*PikaConfig, pr *PikaRouter, rc *RedisCluster) *pikaManager {
	pikaS := map[string]*redis.Client{}
	addrMap := map[string]bool{}
	for _, v := range pc {
		if _, ok := addrMap[v.Addr]; ok {
			panic(fmt.Sprintf("pika addr repeat:%s", v.Addr))
		} else {
			addrMap[v.Addr] = true
		}
		if _, ok := pikaS[v.Name]; ok {
			panic(fmt.Sprintf("pika name repeat:%s", v.Name))
		}
		cnf := &redis.Options{
			Network:         "tcp",
			Addr:            v.Addr,
			Password:        v.Password,
			DB:              v.DB,
			DialTimeout:     time.Second * 3,
			PoolSize:        runtime.NumCPU() * 25,
			MinIdleConns:    1,
			IdleTimeout:     time.Second * 60,
			MaxRetries:      2,
			MaxRetryBackoff: time.Millisecond * 200,
		}
		if v.TLS {
			cnf.TLSConfig = &tls.Config{}
		}
		c := redis.NewClient(cnf)

		if env.OpenTracing() {
			c.AddHook(redisotel.NewTracingHook())
		}
		if env.OpenMetrics() {
			c.AddHook(NewMetricsHook())
		}
		if tendissync.IsOpen() {
			c.AddHook(NewTendisSyncHook())
		}
		pikaS[v.Name] = c
	}
	var rcc *redis.ClusterClient = nil
	if len(rc.AddrS) > 0 {
		cnf := &redis.ClusterOptions{
			Addrs:           rc.AddrS,
			Password:        rc.Pwd,
			PoolSize:        runtime.NumCPU() * 20,
			MinIdleConns:    1,
			IdleTimeout:     time.Second * 60,
			MaxRetries:      2,
			MaxRetryBackoff: time.Millisecond * 200,
			//NewClient: func(opt *redis.Options) *redis.Client {
			//	node := redis.NewClient(opt)
			//	node.AddHook(redisotel.NewTracingHook())
			//	return node
			//},
		}
		if rc.TLS {
			cnf.TLSConfig = &tls.Config{}
		}
		rcc = redis.NewClusterClient(cnf)
		if env.OpenTracing() {
			rcc.AddHook(redisotel.NewTracingHook())
		}
		if env.OpenMetrics() {
			rcc.AddHook(NewMetricsHook())
		}
		if tendissync.IsOpen() {
			rcc.AddHook(NewTendisSyncHook())
		}
	}
	m := &pikaManager{
		pikaConfigs:       pc,
		pikaRouterConfigs: pr,
		pikaMap:           pikaS,
		redisCluster:      rcc,
	}

	for k, v := range pr.Module {
		if k == lockerKey {
			rl, ok := pikaS[v]
			if !ok {
				panic("pika not set locker")
			}
			m.lockerRedis = rl
			m.locker = redislock.New(rl)
		}
		if k == commonKey {
			rl, ok := pikaS[v]
			if !ok {
				panic("pika not set common")
			}
			m.commonRedis = rl
		}
		if k == gameKey {
			rl, ok := pikaS[v]
			if !ok {
				panic("pika not set common")
			}
			m.gameRedis = rl
		}
	}

	if m.locker == nil {
		panic("pika not set locker")
	}
	if m.commonRedis == nil {
		panic("pika not set common")
	}
	return m
}

func (this_ *pikaManager) Get() redis.Cmdable {
	if this_.gameRedis == nil && this_.redisCluster == nil {
		panic("redisCluster is nil")
	}
	if redisMgr.redisCluster != nil {
		return redisMgr.redisCluster
	}
	return this_.gameRedis
}

const messageNameLen = 1

func SetPBWithClient(ctx context.Context, c redis.Cmdable, key string, value proto.Message, ttl time.Duration) *errmsg.ErrMsg {
	messageName := proto.MessageName(value)
	data := bytespool.GetSample(messageNameLen + len(messageName) + proto.Size(value))
	defer bytespool.PutSample(data)
	data[0] = byte(len(messageName))
	copy(data[messageNameLen:], messageName)
	v := value.(values.MarshallerTo)
	_, err := v.MarshalToSizedBuffer(data[messageNameLen+len(messageName):])
	if err != nil {
		return errmsg.NewProtocolErrorInfo(err.Error())
	}
	return errmsg.NewErrorDB(c.Set(ctx, key, data, ttl).Err())
}

func SetPB(ctx context.Context, key string, value proto.Message, ttl time.Duration) *errmsg.ErrMsg {
	messageName := proto.MessageName(value)
	data := bytespool.GetSample(messageNameLen + len(messageName) + proto.Size(value))
	defer bytespool.PutSample(data)
	data[0] = byte(len(messageName))
	copy(data[messageNameLen:], messageName)
	v := value.(values.MarshallerTo)
	_, err := v.MarshalToSizedBuffer(data[messageNameLen+len(messageName):])
	if err != nil {
		return errmsg.NewProtocolErrorInfo(err.Error())
	}
	return errmsg.NewErrorDB(redisMgr.redisCluster.Set(ctx, key, data, ttl).Err())
}

func SetManyPBWithClient(ctx context.Context, c redis.Cmdable, ps []Pair) *errmsg.ErrMsg {
	if len(ps) == 0 {
		return nil
	}
	wr := make([]interface{}, 0, len(ps)*2)
	for i := range ps {
		p := &ps[i]
		messageName := proto.MessageName(p.Value)
		data := bytespool.GetSample(messageNameLen + len(messageName) + proto.Size(p.Value))
		defer bytespool.PutSample(data)
		data[0] = byte(len(messageName))
		copy(data[messageNameLen:], messageName)
		v := p.Value.(values.MarshallerTo)
		_, err := v.MarshalToSizedBuffer(data[messageNameLen+len(messageName):])
		if err != nil {
			return errmsg.NewProtocolErrorInfo(err.Error())
		}
		wr = append(wr, p.Key, data)
	}

	return errmsg.NewErrorDB(c.MSet(ctx, wr...).Err())
}

func MSetWithClient(logger *logger.TraceLogger, c redis.Cmdable, data []interface{}) {
	gopool.Submit(func() {
		err := c.MSet(context.Background(), data...).Err()
		if err != nil {
			logger.Error("SaveFriends error", zap.Error(err))
		}
	})
}

func GetPBWithClient(ctx context.Context, c redis.Cmdable, key string, out proto.Message) (bool, *errmsg.ErrMsg) {
	data, err := c.Get(ctx, key).Bytes()
	if err != nil {
		if err == Nil {
			return false, nil
		}
		return false, errmsg.NewErrorDB(err)
	}
	err = proto.Unmarshal(data[messageNameLen+data[0]:], out)
	if err != nil {
		return false, errmsg.NewProtocolErrorInfo(err.Error())
	}
	return true, nil
}

func PBUnmarshal(d []byte, msg proto.Message) *errmsg.ErrMsg {
	err := proto.Unmarshal(d[messageNameLen+d[0]:], msg)
	if err != nil {
		return errmsg.NewProtocolError(err)
	}
	return nil
}

type Pair struct {
	Key   string
	Value proto.Message
	Found bool
}

func GetManyPBWithClient(ctx context.Context, c redis.Cmdable, p []Pair) *errmsg.ErrMsg {
	if len(p) == 0 {
		return nil
	}
	keys := make([]string, len(p))
	for i := range p {
		keys[i] = p[i].Key
	}
	sc, err := c.MGet(ctx, keys...).Result()
	if err != nil {
		return errmsg.NewErrorDB(err)
	}

	for i, v := range sc {
		var oneData []byte
		found := false
		if v != nil {
			e, ok := v.(error)
			if ok {
				panic(e)
			}
			oneData = v.([]byte)
			found = true
		}
		if found {
			e := PBUnmarshal(oneData, p[i].Value)
			if e != nil {
				return e
			}
		}
	}
	return nil
}

func GetPB(ctx context.Context, key string, out proto.Message) (bool, *errmsg.ErrMsg) {
	data, err := redisMgr.redisCluster.Get(ctx, key).Bytes()
	if err != nil {
		if err == Nil {
			return false, nil
		}
		return false, errmsg.NewErrorDB(err)
	}
	err = proto.Unmarshal(data[messageNameLen+data[0]:], out)
	if err != nil {
		return false, errmsg.NewProtocolErrorInfo(err.Error())
	}
	return true, nil
}

func GetRolePB(ctx *ctx.Context, out proto.Message) (bool, *errmsg.ErrMsg) {
	roleId := ctx.RoleId
	key := proto.MessageName(out) + ":" + roleId
	return GetPB(ctx.Context, key, out)
}

func SetRolePB(ctx *ctx.Context, in proto.Message, ttl time.Duration) *errmsg.ErrMsg {
	roleId := ctx.RoleId
	key := proto.MessageName(in) + ":" + roleId
	return SetPB(ctx.Context, key, in, ttl)
}

func GetAutoPBWithClient(ctx context.Context, c redis.Cmdable, key string) (proto.Message, bool, *errmsg.ErrMsg) {
	data, err := c.Get(ctx, key).Bytes()
	if err != nil {
		if err == Nil {
			return nil, false, nil
		}
		return nil, false, errmsg.NewErrorDB(err)
	}
	out := msgcreate.NewMessage(utils.BytesToString(data[messageNameLen : messageNameLen+data[0]]))
	err = proto.Unmarshal(data[messageNameLen+data[0]:], out)
	if err != nil {
		return nil, false, errmsg.NewProtocolErrorInfo(err.Error())
	}
	return out, true, nil
}

func GetAutoPB(ctx context.Context, key string) (proto.Message, bool, *errmsg.ErrMsg) {
	return GetAutoPBWithClient(ctx, redisMgr.redisCluster, key)
}

func GetPikaNodes() []string {
	list := make([]string, 0, len(redisMgr.pikaMap))
	for k := range redisMgr.pikaMap {
		list = append(list, k)
	}
	return list
}

func GetInstanceByName(name string) (redis.Cmdable, bool) {
	c, ok := redisMgr.pikaMap[name]
	return c, ok
}
