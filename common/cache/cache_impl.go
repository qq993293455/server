package cache

import (
	"github.com/dgraph-io/ristretto"
)

func Init() {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}
	gCache = &ristrettoCache{cache: cache}
}

type ristrettoCache struct {
	cache *ristretto.Cache
}

func (c *ristrettoCache) Get(key string) (val interface{}, exist bool) {
	return c.cache.Get(key)
}

func (c *ristrettoCache) Set(key string, val interface{}) {
	c.cache.Set(key, val, 1)
	c.cache.Wait()
}

func (c *ristrettoCache) Del(key string) {
	c.cache.Del(key)
}
