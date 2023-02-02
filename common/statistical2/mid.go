package statistical2

import (
	"context"
	"sync"

	"coin-server/common/statistical2/models"
)

type LogServer struct {
	ctx   context.Context
	cache map[string][]models.Model
}

func (l *LogServer) GetCache() map[string][]models.Model {
	return l.cache
}

var cachePool = sync.Pool{New: func() interface{} {
	return &LogServer{
		ctx:   nil,
		cache: map[string][]models.Model{},
	}
}}

func GetLogServer(ctx context.Context) *LogServer {
	o := cachePool.Get().(*LogServer)
	o.ctx = ctx
	return o
}

func PutLogServer(ls *LogServer) {
	ls.ctx = nil
	for k := range ls.cache {
		delete(ls.cache, k)
	}
	cachePool.Put(ls)
}
