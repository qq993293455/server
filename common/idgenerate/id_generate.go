package idgenerate

import (
	"context"
	"sync"

	"coin-server/common/errmsg"
	"coin-server/common/redisclient"
	"coin-server/common/values"

	"github.com/go-redis/redis/v8"
)

type IdInfo struct {
	Start values.Integer
	End   values.Integer
	Lock  *sync.Mutex
}

type idManager struct {
	KeyInfo map[IDGenKey]*IdInfo //只会并发读，不需要加锁
}

var gManager *idManager

func Init(client *redis.Client) {
	gManager = &idManager{
		KeyInfo: map[IDGenKey]*IdInfo{},
	}
	//client := redisclient.GetCommonRedis()
	if client == nil {
		panic("id gen redis nil")
	}
	for k, v := range initValue {
		gManager.registerKey(client, k, v)
	}
}

func GetKeyInitValue(key IDGenKey) (values.Integer, bool) {
	val, ok := initValue[key]
	if ok {
		return val, true
	}
	return -1, false
}

func (g *idManager) registerKey(client *redis.Client, key IDGenKey, initValue values.Integer) {
	gManager.KeyInfo[key] = &IdInfo{Lock: &sync.Mutex{}}
	_, err := client.SetNX(context.Background(), string(key), initValue, 0).Result()
	if err != nil {
		panic(err)
	}
}

func (g *idManager) getKeyInfo(key IDGenKey) (*IdInfo, *errmsg.ErrMsg) {
	keyInfo, ok := g.KeyInfo[key]
	if ok {
		return keyInfo, nil
	}
	return nil, errmsg.NewInternalErr("invalid id gen key:" + string(key))
}

const step = 1000

func GenerateID(ctx context.Context, key IDGenKey) (values.Integer, *errmsg.ErrMsg) {
	info, err := gManager.getKeyInfo(key)
	if err != nil {
		return 0, err
	}
	info.Lock.Lock()
	defer info.Lock.Unlock()

	client := redisclient.GetCommonRedis()
	if client == nil {
		return 0, errmsg.NewInternalErr("id gen redis nil")
	}
	if info.Start >= info.End {
		val, err1 := client.IncrBy(ctx, string(key), step).Result()
		if err1 != nil {
			return 0, errmsg.NewErrorDB(err1)
		}
		info.Start, info.End = val-step, val
	}
	id := info.Start
	info.Start += 1
	return id, nil
}
