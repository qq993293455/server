package service

import (
	"fmt"
	"sync"
)

type serverElem struct {
	key      float64
	serverId int64
	now      int64
}

func (this_ *serverElem) ExtractKey() float64 {
	return this_.key
}

func (this_ *serverElem) String() string {
	return fmt.Sprintf("server_id:%d,key:%f", this_.serverId, this_.key)
}

var serverElemPool = sync.Pool{New: func() interface{} {
	return &serverElem{}
}}

func getServerElem() *serverElem {
	se := serverElemPool.Get().(*serverElem)
	se.key = 0
	se.serverId = 0
	return se
}

func putServerElem(elem *serverElem) {
	serverElemPool.Put(elem)
}
