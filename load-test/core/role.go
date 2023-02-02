package core

import (
	"context"
	"fmt"
	"sync"

	"coin-server/common/idgenerate"
	"coin-server/common/redisclient"
	"coin-server/common/values"
	defaultEnv "coin-server/load-test/env"
)

const (
	UID_PREFIX = "uid_"
	Key        = "load_test_uids"
)

type account struct {
	cursor uint64 // redis set 的游标
	lock   sync.Mutex
}

var ac = &account{}

func (a *account) GetUserIds(count values.Integer) ([]string, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	res := make([]string, 0, count)
	client := redisclient.GetCommonRedis()
	//client.Del(context.Background(), Key)

	// 如果使用已有账号，则从redis里提取
	if defaultEnv.UseSaveUser() {
		keys, cursor, err := client.SScan(context.Background(), Key, a.cursor, "*", count).Result()
		if err != nil {
			return nil, err
		}
		if len(keys) > 0 {
			a.cursor = cursor
			for _, key := range keys {
				res = append(res, key)
			}
		}
	}
	if len(res) == int(count) {
		return res, nil
	}

	// redis里用户不足，则创建新的用户
	diffNum := int(count) - len(res)
	for i := 1; i <= diffNum; i++ {
		id, err := idgenerate.GenerateID(context.Background(), idgenerate.LoadTestUid)
		if err != nil {
			return nil, err
		}
		uid := fmt.Sprintf("%s%d", UID_PREFIX, id)
		_, err1 := client.SAdd(context.Background(), Key, uid).Result()
		if err1 != nil {
			return nil, err1
		}
		res = append(res, uid)
	}
	return res, nil
}

func GetUserIds(count values.Integer) ([]string, error) {
	return ac.GetUserIds(count)
}

func GetUserId() (string, error) {
	list, err := ac.GetUserIds(1)
	if err != nil {
		return "", err
	}
	return list[0], nil
}
