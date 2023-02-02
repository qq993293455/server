package containers

import (
	"time"

	"coin-server/common/values"
)

type linkNode struct {
	key       string
	val       values.Integer
	unix      values.Integer
	pre, next *linkNode
}

type lruCache struct {
	m               map[string]*linkNode
	cap             int
	timeOut, period time.Duration
	head, tail      *linkNode
}

func (lc *lruCache) delete(node *linkNode) *lruCache {
	pre, next := node.pre, node.next
	pre.next, next.pre = next, pre
	return lc
}

func (lc *lruCache) add(node *linkNode) *lruCache {
	nxt := lc.head.next
	lc.head.next, node.next = node, nxt
	node.pre, nxt.pre = lc.head, node
	return lc
}

func (lc *lruCache) moveToHead(node *linkNode) *lruCache {
	lc.delete(node).add(node)
	return lc
}

func (lc *lruCache) Get(key string) values.Integer {
	if _, exist := lc.m[key]; !exist {
		return 0
	}
	lc.m[key].unix = time.Now().Add(lc.timeOut).Unix()
	lc.moveToHead(lc.m[key])
	return lc.m[key].val
}

func (lc *lruCache) Delete(key string) {
	if _, exist := lc.m[key]; !exist {
		return
	}
	lc.delete(lc.m[key])
	delete(lc.m, key)
}

func (lc *lruCache) Clear() {
	lc.head.next, lc.tail.pre = lc.tail, lc.head
	for key := range lc.m {
		delete(lc.m, key)
	}
}

func (lc *lruCache) Put(key string) values.Integer {
	if curr, exist := lc.m[key]; !exist {
		if len(lc.m) == lc.cap {
			delete(lc.m, lc.tail.pre.key)
			lc.delete(lc.tail.pre)
		}
		node := &linkNode{key, 1, time.Now().Add(lc.timeOut).Unix(), nil, nil}
		lc.m[key] = node
		lc.add(node)
	} else {
		if time.Now().Unix() > curr.unix {
			lc.moveToHead(curr)
			curr.unix = time.Now().Add(lc.timeOut).Unix()
			curr.val = 0
		}
		lc.m[key].val++
	}
	return lc.m[key].val
}

func (lc *lruCache) GetPeriod() time.Duration {
	return lc.period
}

func (lc *lruCache) TTL() {
	curr := lc.tail.pre
	for curr != lc.head {
		pre := curr.pre
		if curr.unix != -1 && time.Now().Unix() > curr.unix {
			delete(lc.m, curr.key)
			lc.delete(curr)
			curr = pre
			continue
		}
		break
	}
}
