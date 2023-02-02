package bloomfilter

import (
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

const (
	key  = "bf-test"
	size = 1e6
)

var store = redis.NewClient(&redis.Options{
	Network:  "tcp",
	Addr:     "10.23.20.53:6379",
	Password: "iggcdl5,.",
	DB:       1,
})

func TestRedisBitSet(t *testing.T) {
	bf := newRedisBitSet(store, key, size)
	isSetBefore, err := bf.check([]uint{0})
	if err != nil {
		t.Fatal(err)
	}
	if isSetBefore {
		t.Fatal("Bit should not be set")
	}
	err = bf.set([]uint{512})
	if err != nil {
		t.Fatal(err)
	}
	isSetAfter, err := bf.check([]uint{512})
	if err != nil {
		t.Fatal(err)
	}
	if !isSetAfter {
		t.Fatal("Bit should be set")
	}
	err = bf.expire(10)
	if err != nil {
		t.Fatal(err)
	}
	err = bf.del()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFilterAdd(t *testing.T) {
	bf := New(store, key, size)
	assert.Nil(t, bf.Add([]byte("公会名字1")))
	assert.Nil(t, bf.Add([]byte("公会名字2")))
	ok, err := bf.Exists([]byte("公会名字2"))
	assert.Nil(t, err)
	assert.True(t, ok)
	ok, err = bf.Exists([]byte("公会名字3"))
	assert.Nil(t, err)
	assert.False(t, ok)
}
