package bloomfilter

import (
	"context"
	"strconv"
	"time"

	"coin-server/common/errmsg"

	"github.com/go-redis/redis/v8"
	"github.com/spaolacci/murmur3"
)

const (
	maps      = 5
	setScript = `
			for _, offset in ipairs(ARGV) do
				redis.call("setbit", KEYS[1], offset, 1)
			end
			`
	testScript = `
			for _, offset in ipairs(ARGV) do
				if tonumber(redis.call("getbit", KEYS[1], offset)) == 0 then
					return false
				end
			end
			return true
			`
)

type (
	Filter struct {
		bits   uint
		bitSet bitSetProvider
	}

	bitSetProvider interface {
		check([]uint) (bool, *errmsg.ErrMsg)
		set([]uint) *errmsg.ErrMsg
	}
)

func New(store redis.Cmdable, key string, bits uint) *Filter {
	return &Filter{
		bits:   bits,
		bitSet: newRedisBitSet(store, key, bits),
	}
}

func (f *Filter) Add(data []byte) *errmsg.ErrMsg {
	locations := f.getLocations(data)
	return f.bitSet.set(locations)
}

func (f *Filter) Exists(data []byte) (bool, *errmsg.ErrMsg) {
	locations := f.getLocations(data)
	isSet, err := f.bitSet.check(locations)
	if err != nil {
		return false, err
	}
	if !isSet {
		return false, nil
	}

	return true, nil
}

func (f *Filter) getLocations(data []byte) []uint {
	locations := make([]uint, maps)
	for i := uint(0); i < maps; i++ {
		hashValue := murmur3.Sum64(append(data, byte(i)))
		locations[i] = uint(hashValue % uint64(f.bits))
	}

	return locations
}

type redisBitSet struct {
	store redis.Cmdable
	key   string
	bits  uint
}

func newRedisBitSet(store redis.Cmdable, key string, bits uint) *redisBitSet {
	return &redisBitSet{
		store: store,
		key:   key,
		bits:  bits,
	}
}

func (r *redisBitSet) buildOffsetArgs(offsets []uint) ([]string, *errmsg.ErrMsg) {
	var args []string

	for _, offset := range offsets {
		if offset >= r.bits {
			return nil, errmsg.NewInternalErr("offset is out of range")
		}

		args = append(args, strconv.FormatUint(uint64(offset), 10))
	}

	return args, nil
}

func (r *redisBitSet) check(offsets []uint) (bool, *errmsg.ErrMsg) {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return false, err
	}
	resp, err1 := r.store.Eval(context.Background(), testScript, []string{r.key}, args).Result()
	if err1 == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, errmsg.NewErrorDB(err)
	}

	exists, ok := resp.(int64)
	if !ok {
		return false, nil
	}

	return exists == 1, nil
}

func (r *redisBitSet) set(offsets []uint) *errmsg.ErrMsg {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return err
	}

	_, err1 := r.store.Eval(context.Background(), setScript, []string{r.key}, args).Result()
	if err1 == redis.Nil {
		return nil
	}

	return err
}

func (r *redisBitSet) del() *errmsg.ErrMsg {
	if err := r.store.Del(context.Background(), r.key).Err(); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}

func (r *redisBitSet) expire(seconds time.Duration) *errmsg.ErrMsg {
	if err := r.store.Expire(context.Background(), r.key, seconds*time.Second).Err(); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}
