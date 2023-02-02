package dlock

import (
	"context"
	"sync"
	"time"

	"coin-server/common/errmsg"

	"github.com/FJSDS/redislock"
)

var (
	lockerPool = sync.Pool{New: func() interface{} {
		return NewRecursiveLock()
	}}

	lockerKey = "_dr_locker"
)

type RecursiveLock struct {
	lockers []*redislock.Lock
}

func GetLocker() *RecursiveLock {
	return lockerPool.Get().(*RecursiveLock)
}

func PutLocker(l *RecursiveLock) {
	l.lockers = l.lockers[:0]
	lockerPool.Put(l)
}

func NewRecursiveLock() *RecursiveLock {
	return &RecursiveLock{
		lockers: make([]*redislock.Lock, 0, 1),
	}
}

func (this_ *RecursiveLock) Lock(client *redislock.Client, ttl time.Duration, keys ...string) error {
	lks := len(keys)
	if lks == 0 {
		return nil
	}
	if lks == 1 {
		for _, v := range this_.lockers {
			if v.Key() == keys[0] {
				return nil
			}
		}
		locker, err := DistributedLockWithTTL(client, keys[0], ttl)
		if err != nil {
			return err
		}
		this_.lockers = append(this_.lockers, locker)
		return nil
	}
	keysMap := make(map[string]struct{}, lks)
	for _, v := range keys {
		keysMap[v] = struct{}{}
	}
	for _, v := range this_.lockers {
		vKeys := v.Keys()
		lVs := len(vKeys)
		if lVs > 0 {
			vKeysMap := make(map[string]struct{}, lVs)
			for _, vv := range vKeys {
				vKeysMap[vv] = struct{}{}
			}
			if len(keysMap) == lVs { // 长度相等的情况下检查是否全部相等
				allSame := true
				allNotSame := true
				for k := range keysMap {
					if _, ok := vKeysMap[k]; !ok {
						allSame = false
					} else {
						allNotSame = false
					}
				}

				if !allSame && !allNotSame { //又不是全部相等，也不是全部不相等，则加锁失败
					return errmsg.NewErrorDBInfo("locker keys is not all equal") // 不允许重复锁定部分相同的keys
				}
				if allSame { // 如果全相等，则直接返回
					return nil
				}
				// 如果全部不相等,则不做任何处理
			} else { // 长度不等的情况下检查是否有相同的key
				for k := range keysMap {
					if _, ok := vKeysMap[k]; ok {
						return errmsg.NewErrorDBInfo("locker key is exist:" + k) // 不允许重复锁定部分相同的keys
					}
				}
			}
		}
	}
	locker, err := DistributedLockKeysWithTTL(client, keys, ttl)
	if err != nil {
		return err
	}
	this_.lockers = append(this_.lockers, locker)
	return nil
}

func (this_ *RecursiveLock) Unlock() error {
	l := len(this_.lockers) - 1
	for i := l; i >= 0; i-- {
		err := this_.lockers[i].Release(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
}

func DistributedLockWithTTL(client *redislock.Client, key string, ttl time.Duration) (*redislock.Lock, error) {
	return client.Obtain(context.Background(), key, ttl, &redislock.Options{
		RetryStrategy: redislock.ExponentialBackoff(16, 512),
	})
}

func DistributedLockKeysWithTTL(client *redislock.Client, keys []string, ttl time.Duration) (*redislock.Lock, error) {
	ttlS := int64(ttl / time.Second)
	if ttlS <= 0 {
		ttlS = 1
	}
	return client.ObtainMany(context.Background(), ttlS, &redislock.Options{
		RetryStrategy: redislock.ExponentialBackoff(16, 512),
	}, keys...)
}

func DistributedLock(client *redislock.Client, key string) (*redislock.Lock, error) {
	return client.Obtain(context.Background(), key, time.Second*5, &redislock.Options{
		RetryStrategy: redislock.ExponentialBackoff(16, 512),
	})
}

func DistributedLockKeys(client *redislock.Client, keys ...string) (*redislock.Lock, error) {
	return client.ObtainMany(context.Background(), 5, &redislock.Options{
		RetryStrategy: redislock.ExponentialBackoff(16, 512),
	}, keys...)
}

func DistributedLockWithFunc(client *redislock.Client, key string, f func(gotLock bool) error) error {
	c := context.Background()
	l, err := client.Obtain(c, key, time.Second*10, &redislock.Options{
		RetryStrategy: redislock.ExponentialBackoff(16, 512),
	})
	if err == redislock.ErrNotObtained {
		return f(false)
	} else if err != nil {
		return err
	}
	defer l.Release(c)
	return f(true)
}

func DistributedLockKeysWithFunc(client *redislock.Client, keys []string, f func(gotLock bool) error) error {
	c := context.Background()
	l, err := client.ObtainMany(c, 10, &redislock.Options{
		RetryStrategy: redislock.ExponentialBackoff(16, 512),
	}, keys...)
	if err == redislock.ErrNotObtained {
		return f(false)
	} else if err != nil {
		return err
	}
	context.Background()
	defer l.Release(c)
	return f(true)
}

func DistributedLockWithFuncTTL(client *redislock.Client, key string, ttl time.Duration, f func(gotLock bool) error) error {
	c := context.Background()
	l, err := client.Obtain(c, key, ttl, &redislock.Options{
		RetryStrategy: redislock.ExponentialBackoff(16, 512),
	})
	if err == redislock.ErrNotObtained {
		return f(false)
	} else if err != nil {
		return err
	}
	defer l.Release(c)
	return f(true)
}

func DistributedLockKeysWithFuncTTL(client *redislock.Client, keys []string, ttl time.Duration, f func(gotLock bool) error) error {
	ttlS := int64(ttl / time.Second)
	if ttlS <= 0 {
		ttlS = 1
	}
	c := context.Background()
	l, err := client.ObtainMany(c, ttlS, &redislock.Options{
		RetryStrategy: redislock.ExponentialBackoff(16, 512),
	}, keys...)
	if err == redislock.ErrNotObtained {
		return f(false)
	} else if err != nil {
		return err
	}
	defer l.Release(c)
	return f(true)
}
