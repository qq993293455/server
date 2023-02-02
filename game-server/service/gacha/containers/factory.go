package containers

import (
	"time"

	"coin-server/common/values"
)

const (
	DefaultCap     = 5000
	DefaultTimeOut = 1 * time.Minute
	DefaultPeriod  = 1 * time.Minute
)

// NewLRU
// capacity 容量
// timeOut  剩余过期时间
// period   清理周期
func NewLRUC(capacity int, timeOut, period time.Duration) LRU {
	head := &linkNode{"head", 0, -1, nil, nil}
	tail := &linkNode{"tail", 0, -1, head, nil}
	head.next = tail
	n := &lruCache{make(map[string]*linkNode), capacity, timeOut, period, head, tail}
	nl := &currencyLRU{n: n}
	return nl
}

func NewLRU(capacity int, timeOut, period time.Duration) LRU {
	head := &linkNode{"head", 0, -1, nil, nil}
	tail := &linkNode{"tail", 0, -1, head, nil}
	head.next = tail
	n := &lruCache{make(map[string]*linkNode), capacity, timeOut, period, head, tail}
	return n
}

//nolint
func NewDefaultLRUC() LRU {
	return NewLRUC(DefaultCap, DefaultTimeOut, DefaultPeriod)
}

//nolint
func NewDefaultLRU() LRU {
	return NewLRU(DefaultCap, DefaultTimeOut, DefaultPeriod)
}

type LRU interface {
	Get(string) values.Integer
	Put(string) values.Integer
	Delete(key string)
	Clear()
	GetPeriod() time.Duration
	TTL()
}
