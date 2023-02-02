package statistical

import (
	"context"
	"sync"

	"coin-server/common/statistical/models"
)

type LogServer struct {
	ctx   context.Context
	cache []models.Model
}

func (l *LogServer) GetCache() []models.Model {
	return l.cache
}

var cachePool = sync.Pool{New: func() interface{} {
	return &LogServer{
		ctx:   nil,
		cache: []models.Model{},
	}
}}

func GetLogServer(ctx context.Context) *LogServer {
	o := cachePool.Get().(*LogServer)
	o.ctx = ctx
	return o
}

func PutLogServer(ls *LogServer) {
	ls.ctx = nil
	ls.cache = ls.cache[:0]
	cachePool.Put(ls)
}
