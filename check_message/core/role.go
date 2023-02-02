package core

import (
	"fmt"
	"sync"

	"coin-server/common/values"
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
	res := make([]string, 0, count)
	res = append(res, fmt.Sprintf("%s%d", UID_PREFIX, 10000))
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
