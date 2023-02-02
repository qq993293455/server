package containers

import (
	"sync"
	"time"

	"coin-server/common/values"
)

type currencyLRU struct {
	n *lruCache
	m sync.Mutex
}

func (lc *currencyLRU) Get(key string) values.Integer {
	lc.m.Lock()
	defer lc.m.Unlock()
	return lc.n.Get(key) // Get会移动链表操作
}

func (lc *currencyLRU) Delete(key string) {
	lc.m.Lock()
	defer lc.m.Unlock()
	lc.n.Delete(key)
}

func (lc *currencyLRU) Clear() {
	lc.m.Lock()
	defer lc.m.Unlock()
	lc.n.Clear()
}

func (lc *currencyLRU) Put(key string) values.Integer {
	lc.m.Lock()
	defer lc.m.Unlock()
	return lc.n.Put(key)
}

func (lc *currencyLRU) GetPeriod() time.Duration {
	return lc.n.GetPeriod()
}

func (lc *currencyLRU) TTL() {
	lc.m.Lock()
	defer lc.m.Unlock()
	lc.n.TTL()
}
