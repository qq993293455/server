package cache

import "sync"

type LRU[K comparable, V any] struct {
	cache  map[K]*Node[K, V]
	head   *Node[K, V]
	tail   *Node[K, V]
	lock   sync.Mutex
	cap    int
	length int
}

type Node[K comparable, V any] struct {
	key   K
	value V
	prev  *Node[K, V]
	next  *Node[K, V]
}

func NewLRU[K comparable, V any](cap int) *LRU[K, V] {
	return &LRU[K, V]{
		cache:  make(map[K]*Node[K, V], cap),
		head:   nil,
		tail:   nil,
		lock:   sync.Mutex{},
		cap:    cap,
		length: 0,
	}
}

func (l *LRU[K, V]) Add(key K, value V) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if node, ok := l.cache[key]; ok {
		node.value = value
		l.moveToHead(node)
		return
	}
	if l.length == l.cap {
		node := l.tail
		delete(l.cache, node.key)
		node.key = key
		node.value = value
		l.moveToHead(node)
		l.cache[key] = node
		return
	}
	node := new(Node[K, V])
	node.key = key
	node.value = value
	if l.head == nil {
		l.head = node
		l.tail = node
		node.next = l.tail
		node.prev = l.head
	} else {
		head := l.head
		head.prev = node
		node.next = head
		l.head = node
		node.prev = l.head
	}
	l.cache[key] = node
	l.length++
}

func (l *LRU[K, V]) Get(key K) V {
	var v V
	node, ok := l.cache[key]
	if !ok {
		return v
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	l.moveToHead(node)
	return node.value
}

func (l *LRU[K, V]) GetHead() V {
	var v V
	if l.head == nil {
		return v
	}
	return l.head.value
}

func (l *LRU[K, V]) GetTail() V {
	if l.tail == nil {
		return nil
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	l.moveToHead(l.tail)
	return l.tail.value
}

//尾节点移到头结点
func (l *LRU[K, V]) moveToHead(node *Node[K, V]) {
	if l.head == node {
		return
	}
	if l.tail == node {
		oldH := l.head
		head := l.tail
		//尾指针
		tail := head.prev
		l.tail = tail
		tail.next = l.tail
		//头指针
		oldH.prev = head
		head.next = oldH
		l.head = head
		head.prev = l.head
		return
	}
	oldH := l.head
	//node前后节点指针
	node.prev.next = node.next
	node.next.prev = node.prev
	//头指针
	oldH.prev = node
	node.next = oldH
	l.head = node
	node.prev = l.head
}
