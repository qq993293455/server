package orm

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"unsafe"

	"coin-server/common/bytespool"
	"coin-server/common/errmsg"
	"coin-server/common/msgcreate"
	utils2 "coin-server/common/utils"

	"github.com/go-redis/redis/v8"
	"github.com/gogo/protobuf/proto"
)

type writeType int

const (
	set   writeType = 1
	del   writeType = 2
	mset  writeType = 3
	hset  writeType = 4
	hmset writeType = 5
	hdel  writeType = 6
	incr  writeType = 7
)

type OneCmd struct {
	cmd  writeType
	key  string
	args []interface{}
}

var cmdPool = sync.Pool{New: func() interface{} {
	return &OneCmd{
		cmd:  0,
		key:  "",
		args: make([]interface{}, 0, 4),
	}
}}

func GetOneCmd() *OneCmd {
	return cmdPool.Get().(*OneCmd)
}

func PutOneCmd(v *OneCmd) {
	v.key = ""
	v.cmd = 0
	for _, arg := range v.args {
		switch b := arg.(type) {
		case []byte:
			bytespool.PutSample(b)
		}
	}
	v.args = v.args[:0]
}

type ClientCmd struct {
	cmdS []*OneCmd
}

type Orm struct {
	ctx    context.Context
	cmdMap map[redis.Cmdable]*ClientCmd
	cache  map[redis.Cmdable]map[string]interface{}
}

var ormPool = sync.Pool{New: func() interface{} {
	return &Orm{cmdMap: map[redis.Cmdable]*ClientCmd{}, cache: map[redis.Cmdable]map[string]interface{}{}}
}}

func GetOrm(ctx context.Context) *Orm {
	o := ormPool.Get().(*Orm)
	o.ctx = ctx
	return o
}

func PutOrm(orm *Orm) {
	orm.Reset()
	ormPool.Put(orm)
}

func (this_ *Orm) check() {
	if this_.cmdMap == nil {
		this_.cmdMap = map[redis.Cmdable]*ClientCmd{}
	}
}

func (this_ *Orm) Reset() {
	if this_ == nil {
		return
	}
	this_.ctx = nil
	if this_.cmdMap != nil && len(this_.cmdMap) != 0 {
		for k, v := range this_.cmdMap {
			for _, oc := range v.cmdS {
				PutOneCmd(oc)
			}
			v.cmdS = v.cmdS[:0]
			delete(this_.cmdMap, k)
		}
	}
	if this_.cache != nil && len(this_.cache) != 0 {
		for _, data := range this_.cache {
			for key := range data {
				delete(data, key)
			}
		}
	}
}

func (this_ *Orm) Do() *errmsg.ErrMsg {
	if this_ == nil || this_.cmdMap == nil || len(this_.cmdMap) == 0 {
		return nil
	}
	for c, cmdS := range this_.cmdMap {
		_, err := c.Pipelined(this_.ctx, func(pp redis.Pipeliner) error {
			for _, v := range cmdS.cmdS {
				switch v.cmd {
				case set:
					pp.Set(this_.ctx, v.key, v.args[0], 0)
				case mset:
					for i := 0; i < len(v.args); i += 2 {
						pp.Set(this_.ctx, v.args[i].(string), v.args[i+1], 0)
					}
				case del:
					if len(v.args) == 0 {
						pp.Del(this_.ctx, v.key)
					} else {
						keys := make([]string, 0, len(v.args)+1)
						keys = append(keys, v.key)
						for _, v := range v.args {
							keys = append(keys, v.(string))
						}
						pp.Del(this_.ctx, keys...)
					}
				case hset:
					pp.HSet(this_.ctx, v.key, v.args...)
				case hmset:
					pp.HMSet(this_.ctx, v.key, v.args...)
				case hdel:
					args := make([]string, 0, len(v.args))
					for _, v := range v.args {
						args = append(args, v.(string))
					}
					pp.HDel(this_.ctx, v.key, args...)
				case incr:
					pp.Incr(this_.ctx, v.key)
				}
			}
			return nil
		})
		if err != nil {
			return errmsg.NewErrorDB(err)
		}
	}

	return nil
}

func (this_ *Orm) Incr(c redis.Cmdable, key string) {
	v := this_.checkAndGet(c)
	oc := GetOneCmd()
	oc.cmd = incr
	oc.key = key
	v.cmdS = append(v.cmdS, oc)
}

func (this_ *Orm) SetPB(c redis.Cmdable, in RedisInterface) {
	v := this_.checkAndGet(c)

	oc := GetOneCmd()
	oc.cmd = set
	oc.key = in.KVKey()
	oc.args = append(oc.args, in.ToSave())
	v.cmdS = append(v.cmdS, oc)
	this_.addSet(c, in)
}

func (this_ *Orm) addSet(c redis.Cmdable, in RedisInterface) {
	r, ok := this_.cache[c]
	if !ok {
		r = map[string]interface{}{}
		this_.cache[c] = r
	}
	r[in.KVKey()] = in
}

func (this_ *Orm) DelPB(c redis.Cmdable, in RedisInterface) {
	v := this_.checkAndGet(c)

	oc := GetOneCmd()
	oc.cmd = del
	oc.key = in.KVKey()
	v.cmdS = append(v.cmdS, oc)
	this_.addDel2(c, in)
}

func (this_ *Orm) addDel2(c redis.Cmdable, in RedisInterface) {
	r, ok := this_.cache[c]
	if !ok {
		return
	}
	delete(r, in.KVKey())
}

func (this_ *Orm) DelManyPB(c redis.Cmdable, in ...RedisInterface) {
	v := this_.checkAndGet(c)

	oc := GetOneCmd()
	oc.cmd = set
	oc.key = in[0].KVKey()
	if len(in) > 1 {
		for _, v := range in[1:] {
			oc.args = append(oc.args, v.KVKey())
		}
	}
	v.cmdS = append(v.cmdS, oc)
	this_.addDel1(c, in...)
}

func (this_ *Orm) addDel1(c redis.Cmdable, in ...RedisInterface) {
	r, ok := this_.cache[c]
	if !ok {
		return
	}
	for _, v := range in {
		delete(r, v.KVKey())
	}
}

func (this_ *Orm) Del(c redis.Cmdable, key ...string) {
	v := this_.checkAndGet(c)

	oc := GetOneCmd()
	oc.cmd = del
	oc.key = key[0]
	if len(key) > 1 {
		for _, v := range key[1:] {
			oc.args = append(oc.args, v)
		}
	}
	v.cmdS = append(v.cmdS, oc)
	this_.addDel(c, key...)
}

func (this_ *Orm) addDel(c redis.Cmdable, keys ...string) {
	r, ok := this_.cache[c]
	if !ok {
		return
	}
	for _, v := range keys {
		delete(r, v)
	}
}

func (this_ *Orm) checkAndGet(c redis.Cmdable) *ClientCmd {
	this_.check()
	v, ok := this_.cmdMap[c]
	if !ok {
		v = &ClientCmd{}
		this_.cmdMap[c] = v
	}
	return v
}

func (this_ *Orm) MSetPB(c redis.Cmdable, ins []RedisInterface) {
	li := len(ins)
	if li == 0 {
		return
	}
	if li == 1 {
		this_.SetPB(c, ins[0])
		return
	}
	v := this_.checkAndGet(c)
	oc := GetOneCmd()
	oc.cmd = mset
	for _, in := range ins {
		key, value := in.KVKey(), in.ToSave()
		oc.args = append(oc.args, key, value)
	}
	v.cmdS = append(v.cmdS, oc)
	this_.addMSet(c, ins)
}

func (this_ *Orm) addMSet(c redis.Cmdable, ins []RedisInterface) {
	r, ok := this_.cache[c]
	if !ok {
		r = map[string]interface{}{}
		this_.cache[c] = r
	}
	for _, v := range ins {
		r[v.KVKey()] = v
	}
}

func (this_ *Orm) HSetPB(c redis.Cmdable, key string, in RedisInterface) {
	v := this_.checkAndGet(c)
	oc := GetOneCmd()
	oc.cmd = hset
	oc.key = key
	oc.args = append(oc.args, in.PK(), in.ToSave())
	v.cmdS = append(v.cmdS, oc)
	this_.addHSet(c, key, in)
}
func (this_ *Orm) addHSet(c redis.Cmdable, key string, in RedisInterface) {
	r, ok := this_.cache[c]
	if !ok {
		r = map[string]interface{}{}
		this_.cache[c] = r
	}
	kvi, ok := r[key]
	if !ok {
		kvi = &HALL{all: false, data: map[string]RedisInterface{}}
		r[key] = kvi
	}
	kv := kvi.(*HALL)
	kv.data[in.PK()] = in
}

// func (this_ *Orm) HSet(c redis.Cmdable, key string, args ...interface{}) {
//	this_.check()
//	v, ok := this_.cmdMap[c]
//	if !ok {
//		v = &ClientCmd{}
//		this_.cmdMap[c] = v
//	}
//	oc := GetOneCmd()
//	oc.cmd = hset
//	oc.key = key
//	oc.args = append(oc.args, args...)
//	v.cmdS = append(v.cmdS, oc)
// }
//
// func (this_ *Orm) HMSet(c redis.Cmdable, key string, args ...interface{}) {
//	this_.check()
//	v, ok := this_.cmdMap[c]
//	if !ok {
//		v = &ClientCmd{}
//		this_.cmdMap[c] = v
//	}
//	oc := GetOneCmd()
//	oc.cmd = hmset
//	oc.key = key
//	oc.args = append(oc.args, args...)
//	v.cmdS = append(v.cmdS, oc)
// }

// 支持 map[string]RedisInterface 、 []RedisInterface
// func (this_ *Orm) HMSetPB(c redis.Cmdable, key string, ins interface{}) {
//	v := this_.checkAndGet(c)
//	oc := GetOneCmd()
//	oc.cmd = hmset
//	oc.key = key
//	if reflect.TypeOf(ins).Kind() == reflect.Map {
//		im, ok := ins.(map[string]RedisInterface)
//		if !ok {
//			panic("HMSetPB: Not support type")
//		}
//		for _, in := range im {
//			k, value := in.PK(), in.ToSave()
//			oc.args = append(oc.args, k, value)
//		}
//		v.cmdS = append(v.cmdS, oc)
//		this_.addHMSetMap(c, key, im)
//	} else if reflect.TypeOf(ins).Kind() == reflect.Slice {
//		is, ok := ins.([]RedisInterface)
//		if !ok {
//			panic("HMSetPB: Not support type")
//		}
//		for _, in := range is {
//			k, value := in.PK(), in.ToSave()
//			oc.args = append(oc.args, k, value)
//		}
//		v.cmdS = append(v.cmdS, oc)
//		this_.addHMSet(c, key, is)
//	}
// }

func (this_ *Orm) HMSetPB(c redis.Cmdable, key string, ins []RedisInterface) {
	v := this_.checkAndGet(c)
	oc := GetOneCmd()
	oc.cmd = hmset
	oc.key = key
	for _, in := range ins {
		k, value := in.PK(), in.ToSave()
		oc.args = append(oc.args, k, value)
	}
	v.cmdS = append(v.cmdS, oc)
	this_.addHMSet(c, key, ins)
}

func (this_ *Orm) addHMSet(c redis.Cmdable, key string, ins []RedisInterface) {
	r, ok := this_.cache[c]
	if !ok {
		r = map[string]interface{}{}
		this_.cache[c] = r
	}
	kvi, ok := r[key]
	if !ok {
		kvi = &HALL{all: false, data: map[string]RedisInterface{}}
		r[key] = kvi
	}
	kv := kvi.(*HALL)

	for _, v := range ins {
		kv.data[v.PK()] = v
	}
}

// func (this_ *Orm) addHMSetMap(c redis.Cmdable, key string, ins map[string]RedisInterface) {
//	r, ok := this_.cache[c]
//	if !ok {
//		r = map[string]interface{}{}
//		this_.cache[c] = r
//	}
//	kvi, ok := r[key]
//	if !ok {
//		kvi = &HALL{all: false, data: map[string]RedisInterface{}}
//		r[key] = kvi
//	}
//	kv := kvi.(*HALL)
//
//	for _, v := range ins {
//		kv.data[v.PK()] = v
//	}
// }

func (this_ *Orm) HDelPB(c redis.Cmdable, key string, ins []RedisInterface) {
	v := this_.checkAndGet(c)
	oc := GetOneCmd()
	oc.cmd = hdel
	oc.key = key
	for _, in := range ins {
		oc.args = append(oc.args, in.PK())
	}
	v.cmdS = append(v.cmdS, oc)
	this_.delHDel(c, key, ins)
}

func (this_ *Orm) delHDel(c redis.Cmdable, key string, ins []RedisInterface) {
	r, ok := this_.cache[c]
	if !ok {
		return
	}
	kvi, ok := r[key]
	if !ok {
		return
	}
	kv := kvi.(*HALL)

	for _, v := range ins {
		delete(kv.data, v.PK())
	}
}

func (this_ *Orm) HDel(c redis.Cmdable, key string, fields ...string) {
	v := this_.checkAndGet(c)
	oc := GetOneCmd()
	oc.cmd = hdel
	oc.key = key
	for _, in := range fields {
		oc.args = append(oc.args, in)
	}
	v.cmdS = append(v.cmdS, oc)
	this_.delHDel1(c, key, fields...)
}

func (this_ *Orm) delHDel1(c redis.Cmdable, key string, fields ...string) {
	r, ok := this_.cache[c]
	if !ok {
		return
	}
	kvi, ok := r[key]
	if !ok {
		return
	}
	kv := kvi.(*HALL)

	for _, v := range fields {
		delete(kv.data, v)
	}
}

func (this_ *Orm) GetPB(c redis.Cmdable, out RedisInterface) (bool, *errmsg.ErrMsg) {
	key := out.KVKey()
	r, ok := this_.cache[c]
	if ok {
		v, ok := r[key]
		if ok {
			reflect.ValueOf(out).Elem().Set(reflect.ValueOf(v).Elem())
			return true, nil
		}
	}
	ok, e := GetPBWith(this_.ctx, c, out)
	if e != nil {
		return false, e
	}
	if ok {
		if r == nil {
			r = map[string]interface{}{key: out}
			this_.cache[c] = r
		} else {
			r[key] = out
		}
	}

	return ok, nil
}

func (this_ *Orm) MGetPB(c redis.Cmdable, out ...RedisInterface) (notFoundIndex []int, err *errmsg.ErrMsg) {
	notKeys := make([]RedisInterface, 0, len(out))
	r, ok := this_.cache[c]
	if ok {
		for _, v := range out {
			x, ok := r[v.KVKey()]
			if ok {
				reflect.ValueOf(v).Elem().Set(reflect.ValueOf(x).Elem())
			} else {
				notKeys = append(notKeys, v)
			}
		}
	} else {
		notKeys = out
		r = map[string]interface{}{}
		this_.cache[c] = r
	}
	if len(notKeys) == 0 {
		return nil, nil
	}
	nfi, e := MGetPBWith(this_.ctx, c, notKeys)
	if e != nil {
		return nil, e
	}
	m := make(map[int]struct{}, len(nfi))
	for _, v := range nfi {
		m[v] = struct{}{}
	}
	for i, v := range notKeys {
		if _, ok := m[i]; !ok {
			r[v.KVKey()] = v
		}
	}
	return nfi, nil
}

func (this_ *Orm) MGetPBInSlot(c redis.Cmdable, out ...RedisInterface) (notFoundIndex []int, err *errmsg.ErrMsg) {
	notKeys := make([]RedisInterface, 0, len(out))
	r, ok := this_.cache[c]
	if ok {
		for _, v := range out {
			x, ok := r[v.KVKey()]
			if ok {
				reflect.ValueOf(v).Elem().Set(reflect.ValueOf(x).Elem())
			} else {
				notKeys = append(notKeys, v)
			}
		}
	} else {
		notKeys = out
		r = map[string]interface{}{}
		this_.cache[c] = r
	}
	if len(notKeys) == 0 {
		return nil, nil
	}
	nfi, e := MGetPBInSlotWith(this_.ctx, c, notKeys)
	if e != nil {
		return nil, e
	}
	m := make(map[int]struct{}, len(nfi))
	for _, v := range nfi {
		m[v] = struct{}{}
	}
	for i, v := range notKeys {
		if _, ok := m[i]; !ok {
			r[v.KVKey()] = v
		}
	}
	return nfi, nil
}

func (this_ *Orm) HGetPB(c redis.Cmdable, key string, out RedisInterface) (bool, *errmsg.ErrMsg) {
	r, ok := this_.cache[c]
	var kv *HALL
	if ok {
		kvi, ok := r[key]
		if ok {
			kv = kvi.(*HALL)
			v, ok := kv.data[out.PK()]
			if ok {
				reflect.ValueOf(out).Elem().Set(reflect.ValueOf(v).Elem())
				return true, nil
			}
		}
	}
	ok, err := HGetPBWith(this_.ctx, c, key, out)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	if kv == nil {
		kv = &HALL{data: map[string]RedisInterface{}}
	}
	if r == nil {
		r = map[string]interface{}{key: kv}
	}
	kv.data[out.PK()] = out
	return true, nil
}

func (this_ *Orm) HMGetPB(c redis.Cmdable, key string, outs []RedisInterface) ([]int, *errmsg.ErrMsg) {
	notKeys := make([]RedisInterface, 0, len(outs))
	r, ok := this_.cache[c]
	var kv *HALL
	if ok {
		kvi, ok := r[key]
		if ok {
			kv = kvi.(*HALL)
			for _, out := range outs {
				v, ok := kv.data[out.PK()]
				if ok {
					reflect.ValueOf(out).Elem().Set(reflect.ValueOf(v).Elem())
				} else {
					notKeys = append(notKeys, out)
				}
			}
		} else {
			notKeys = outs
		}
	} else {
		notKeys = outs
	}
	if len(notKeys) == 0 {
		return nil, nil
	}
	nfi, err := HMGetPBWith(this_.ctx, c, key, notKeys)
	if err != nil {
		return nil, err
	}
	if kv == nil {
		kv = &HALL{data: map[string]RedisInterface{}}
	}
	if r == nil {
		r = map[string]interface{}{key: kv}
	}
	m := make(map[int]struct{}, len(nfi))
	for _, v := range nfi {
		m[v] = struct{}{}
	}
	for i, v := range notKeys {
		if _, ok := m[i]; !ok {
			kv.data[v.PK()] = v
		}
	}
	return nfi, nil
}

type HALL struct {
	all  bool
	data map[string]RedisInterface
}

// HGetAll out 支持类型有：[*]map[string][*][proto.any] , *[][*][proto.any]
// 以 dao.Bag 举例，out支持以下类型：
// *map[string]*dao.Bag,map[string]*dao.Bag,map[string]dao.Bag,*map[string]dao.Bag,*[]*dao.Bag,*[]dao.Bag
// 也就是调用时支持：
// m:=map[string]*dao.Bag{}
// m1:=map[string]dao.Bag{}
// var s1 []*dao.Bag
// var s2 []dao.Bag
//
// HGetAll(c,key,&m)
// HGetAll(c,key,m)
// HGetAll(c,key,m1)
// HGetAll(c,key,&m1)
// HGetAll(c,key,&s1)
// HGetAll(c,key,&s2)
func (this_ *Orm) HGetAll(c redis.Cmdable, key string, out interface{}) *errmsg.ErrMsg {
	r, ok := this_.cache[c]
	if !ok {
		r = map[string]interface{}{}
		this_.cache[c] = r
	}
	kvi, ok := r[key]
	if !ok {
		kvi = &HALL{all: false, data: map[string]RedisInterface{}}
		r[key] = kvi
	}
	kv := kvi.(*HALL)
	tv := reflect.ValueOf(out)
	if kv.all {
		if tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Slice {
			elem := tv.Elem()
			ptr := elem.Type().Elem().Kind() == reflect.Ptr
			for _, v := range kv.data {
				rv := reflect.ValueOf(v)
				if !ptr {
					rv = rv.Elem()
				}
				elem = reflect.Append(elem, rv)
			}
			tv.Elem().Set(elem)
			return nil
		}

		if (tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Map) || tv.Kind() == reflect.Map {
			elem := tv
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() != reflect.Map || elem.Type().Key().Kind() != reflect.String {
				panic("not support type:" + tv.String())
			}
			ptr := elem.Type().Elem().Kind() == reflect.Ptr
			for _, v := range kv.data {
				rv := reflect.ValueOf(v)
				if !ptr {
					rv = rv.Elem()
				}
				elem.SetMapIndex(reflect.ValueOf(v.PK()), rv)
			}
			return nil
		}

	} else {
		err := HGetAllWith(this_.ctx, c, key, out)
		if err != nil {
			return err
		}
		use := map[string]RedisInterface{}
		if tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Slice {
			elem := tv.Elem()
			l := elem.Len()
			if l > 0 {
				for i := 0; i < l; i++ {
					v := elem.Index(i)
					x, ok := v.Interface().(RedisInterface)
					if ok {
						key := x.PK()
						n, ok := kv.data[key]
						if ok {
							if v.Type().Kind() == reflect.Ptr {
								v = v.Elem()
							}
							v.Set(reflect.ValueOf(n).Elem())
							use[key] = n
							delete(kv.data, key)
						} else {
							use[key] = x
						}
					}
				}
			}
			if len(kv.data) > 0 {
				ptr := elem.Type().Elem().Kind() == reflect.Ptr
				for _, v := range kv.data {
					rv := reflect.ValueOf(v)
					if !ptr {
						rv = rv.Elem()
					}
					elem = reflect.Append(elem, rv)
					use[v.PK()] = v
				}

				tv.Elem().Set(elem)
			}
			kv.all = true
			kv.data = use
			return nil
		}
		if (tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Map) || tv.Kind() == reflect.Map {
			elem := tv
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() != reflect.Map || elem.Type().Key().Kind() != reflect.String {
				panic("not support type:" + tv.String())
			}
			keys := elem.MapKeys()
			if len(keys) > 0 {
				for _, k := range keys {
					v := elem.MapIndex(k)
					x, ok := v.Interface().(RedisInterface)
					if ok {
						key := x.PK()
						n, ok := kv.data[key]
						if ok {
							if v.Type().Kind() == reflect.Ptr {
								v = v.Elem()
							}
							v.Set(reflect.ValueOf(n).Elem())
							use[key] = n
							delete(kv.data, key)
						} else {
							use[key] = x
						}
					}
				}
			}
			if len(kv.data) > 0 {
				ptr := elem.Type().Elem().Kind() == reflect.Ptr
				for _, v := range kv.data {
					rv := reflect.ValueOf(v)
					if !ptr {
						rv = rv.Elem()
					}
					elem.SetMapIndex(reflect.ValueOf(v.PK()), rv)
					use[v.PK()] = v
				}
			}
			kv.all = true
			kv.data = use
			return nil
		}
	}

	panic("not support type:" + tv.String())
}

func GetPBWith(ctx context.Context, c redis.Cmdable, out RedisInterface) (bool, *errmsg.ErrMsg) {
	sc, err := c.Get(ctx, out.KVKey()).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, errmsg.NewErrorDB(err)
	}

	e := protoUnmarshal(sc, out)
	if e != nil {
		return false, e
	}
	return true, nil
}

func GetPB(ctx context.Context, c redis.Cmdable, key string) (bool, RedisInterface, *errmsg.ErrMsg) {
	sc, err := c.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil, nil
		}
		return false, nil, errmsg.NewErrorDB(err)
	}
	rName := sc[1 : 1+sc[0]]
	msgName := *(*string)(unsafe.Pointer(&rName))
	msg := msgcreate.NewMessage(msgName)
	e := protoUnmarshal(sc, msg)
	if e != nil {
		return false, nil, e
	}
	return true, msg.(RedisInterface), nil
}

func protoUnmarshal(d []byte, out proto.Message) *errmsg.ErrMsg {
	err := proto.Unmarshal(d[1+d[0]:], out)
	if err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}

func ProtoUnmarshal(d []byte, out proto.Message) *errmsg.ErrMsg {
	return protoUnmarshal(d, out)
}

func MGetPBWith(ctx context.Context, c redis.Cmdable, outs []RedisInterface) (notFoundIndex []int, err *errmsg.ErrMsg) {
	ol := len(outs)
	if ol == 0 {
		return nil, nil
	}
	if ol == 1 {
		ok, err := GetPBWith(ctx, c, outs[0])
		if err != nil {
			return nil, err
		}
		if !ok {
			notFoundIndex = append(notFoundIndex, 0)
			return notFoundIndex, nil
		}
		return nil, nil
	}
	keys := make([]string, len(outs))
	for i, v := range outs {
		keys[i] = v.KVKey()
	}
	dataS := make([]*redis.StringCmd, 0, len(keys))
	_, e := c.Pipelined(ctx, func(pp redis.Pipeliner) error {
		for _, v := range keys {
			dataS = append(dataS, c.Get(ctx, v))
		}
		return nil
	})
	if e != nil {
		return nil, errmsg.NewErrorDB(e)
	}

	for i, v := range dataS {
		if v.Err() != nil {
			if v.Err() != redis.Nil {
				return nil, errmsg.NewErrorDB(v.Err())
			}
			notFoundIndex = append(notFoundIndex, i)
		} else {
			b := utils2.StringToBytes(v.Val())
			e := protoUnmarshal(b, outs[i])
			if e != nil {
				return nil, e
			}
		}
	}
	return notFoundIndex, nil
}

func MGetPBInSlotWith(ctx context.Context, c redis.Cmdable, outs []RedisInterface) (notFoundIndex []int, err *errmsg.ErrMsg) {
	ol := len(outs)
	if ol == 0 {
		return nil, nil
	}
	if ol == 1 {
		ok, err := GetPBWith(ctx, c, outs[0])
		if err != nil {
			return nil, err
		}
		if !ok {
			notFoundIndex = append(notFoundIndex, 0)
			return notFoundIndex, nil
		}
		return nil, nil
	}
	keys := make([]string, len(outs))
	for i, v := range outs {
		keys[i] = v.KVKey()
	}
	dataS, e := c.MGet(ctx, keys...).Result()
	if e != nil {
		return nil, errmsg.NewErrorDB(e)
	}
	for i, v := range dataS {
		if v == nil {
			notFoundIndex = append(notFoundIndex, i)
		} else {
			b := utils2.StringToBytes(v.(string))
			e := protoUnmarshal(b, outs[i])
			if e != nil {
				return nil, e
			}
		}
	}
	return notFoundIndex, nil
}

func MGetPB(ctx context.Context, c redis.Cmdable, keys ...string) ([]RedisInterface, *errmsg.ErrMsg) {
	ol := len(keys)
	if ol == 0 {
		return nil, nil
	}
	if ol == 1 {
		ok, p, err := GetPB(ctx, c, keys[0])
		if err != nil {
			return nil, err
		}
		if !ok {
			return []RedisInterface{nil}, nil
		}
		return []RedisInterface{p}, nil
	}
	dataS := make([]*redis.StringCmd, 0, len(keys))
	_, e := c.Pipelined(ctx, func(pp redis.Pipeliner) error {
		for _, v := range keys {
			dataS = append(dataS, c.Get(ctx, v))
		}
		return nil
	})
	if e != nil {
		return nil, errmsg.NewErrorDB(e)
	}
	outS := make([]RedisInterface, len(dataS))
	for i, v := range dataS {
		if v.Err() != nil {
			if v.Err() != redis.Nil {
				return nil, errmsg.NewErrorDB(v.Err())
			}
			outS[i] = nil
		} else {
			data := v.String()
			msgName := data[1 : 1+data[0]]
			outS[i] = msgcreate.NewMessage(msgName).(RedisInterface)
			b := utils2.StringToBytes(data)
			e := protoUnmarshal(b, outS[i])
			if e != nil {
				return nil, e
			}
		}
	}
	return outS, nil
}

func HGetPBWith(ctx context.Context, c redis.Cmdable, key string, field RedisInterface) (bool, *errmsg.ErrMsg) {
	sc, err := c.HGet(ctx, key, field.PK()).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, errmsg.NewErrorDB(err)
	}

	e := protoUnmarshal(sc, field)
	if e != nil {
		return false, e
	}
	return true, nil
}

func HGetPB(ctx context.Context, c redis.Cmdable, key string, field string) (bool, RedisInterface, *errmsg.ErrMsg) {
	sc, err := c.HGet(ctx, key, field).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil, nil
		}
		return false, nil, errmsg.NewErrorDB(err)
	}
	rName := sc[1 : 1+sc[0]]
	msgName := *(*string)(unsafe.Pointer(&rName))
	msg := msgcreate.NewMessage(msgName)
	e := protoUnmarshal(sc, msg)
	if e != nil {
		return false, nil, e
	}
	return true, msg.(RedisInterface), nil
}

func HMGetPBWith(ctx context.Context, c redis.Cmdable, key string, outs []RedisInterface) (notFoundIndex []int, err *errmsg.ErrMsg) {
	ol := len(outs)
	if ol == 0 {
		return nil, nil
	}
	if ol == 1 {
		ok, err := HGetPBWith(ctx, c, key, outs[0])
		if err != nil {
			return nil, err
		}
		if !ok {
			notFoundIndex = append(notFoundIndex, 0)
			return notFoundIndex, nil
		}
		return nil, nil
	}
	fields := make([]string, len(outs))
	for i, v := range outs {
		fields[i] = v.PK()
	}
	dataS, e := c.HMGet(ctx, key, fields...).Result()
	if e != nil {
		return nil, errmsg.NewErrorDB(e)
	}
	for i, d := range dataS {
		if d == nil {
			notFoundIndex = append(notFoundIndex, i)
		} else {
			switch data := d.(type) {
			case []byte:
				e := protoUnmarshal(data, outs[i])
				if e != nil {
					return nil, e
				}
			case string:
				b := utils2.StringToBytes(data)
				e := protoUnmarshal(b, outs[i])
				if e != nil {
					return nil, e
				}
			default:
				return nil, errmsg.NewErrorDBInfo(fmt.Sprintf("not support type: %s", reflect.TypeOf(data).String()))
			}
		}
	}
	return notFoundIndex, nil
}

func HMGetPB(ctx context.Context, c redis.Cmdable, key string, fields ...string) ([]RedisInterface, *errmsg.ErrMsg) {
	ol := len(fields)
	if ol == 0 {
		return nil, nil
	}
	if ol == 1 {
		ok, p, err := GetPB(ctx, c, fields[0])
		if err != nil {
			return nil, err
		}
		if !ok {
			return []RedisInterface{nil}, nil
		}
		return []RedisInterface{p}, nil
	}
	dataS, e := c.HMGet(ctx, key, fields...).Result()
	if e != nil {
		return nil, errmsg.NewErrorDB(e)
	}
	outS := make([]RedisInterface, len(dataS))
	for i, v := range dataS {
		if v == nil {
			outS[i] = nil
		} else {
			switch data := v.(type) {
			case []byte:
				rName := data[1 : 1+data[0]]
				msgName := *(*string)(unsafe.Pointer(&rName))
				outS[i] = msgcreate.NewMessage(msgName).(RedisInterface)
				e := protoUnmarshal(data, outS[i])
				if e != nil {
					return nil, e
				}
			case string:
				msgName := data[1 : 1+data[0]]
				outS[i] = msgcreate.NewMessage(msgName).(RedisInterface)
				b := utils2.StringToBytes(data)
				e := protoUnmarshal(b, outS[i])
				if e != nil {
					return nil, e
				}
			default:
				return nil, errmsg.NewErrorDBInfo(fmt.Sprintf("not support type: %s", reflect.TypeOf(data).String()))
			}
		}
	}
	return outS, nil
}

// HGetAllWith v is a ptr of proto.Message slice
// var out []*dao.User
// HGetAllWith(c,key,&out)
func HGetAllWith(ctx context.Context, c redis.Cmdable, key string, out interface{}) *errmsg.ErrMsg {
	tv := reflect.ValueOf(out)
	if tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Slice {
		m, err := c.HGetAll(ctx, key).Result()
		if err != nil {
			return errmsg.NewErrorDB(err)
		}
		elem := tv.Type().Elem().Elem()
		outR := reflect.MakeSlice(reflect.SliceOf(elem), len(m), len(m))
		index := 0
		te := elem
		if elem.Kind() == reflect.Ptr {
			te = elem.Elem()
		}
		for _, v := range m {
			t := reflect.New(te)
			i := t.Interface().(RedisInterface)
			err := protoUnmarshal(utils2.StringToBytes(v), i)
			if err != nil {
				return err
			}
			if elem.Kind() == reflect.Ptr {
				outR.Index(index).Set(t)
			} else {
				outR.Index(index).Set(t.Elem())
			}

			index++
		}
		tv.Elem().Set(outR)
		return nil
	}

	if (tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Map) || tv.Kind() == reflect.Map {
		elem := tv
		if tv.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Map && elem.Type().Key().Kind() != reflect.String {
			panic("not support type:" + tv.String())
		}
		m, err := c.HGetAll(ctx, key).Result()
		if err != nil {
			return errmsg.NewErrorDB(err)
		}

		te := elem.Type().Elem()
		teKind := te.Kind()
		if teKind == reflect.Ptr {
			te = te.Elem()
		}
		for k, v := range m {
			value := reflect.New(te)
			i := value.Interface().(RedisInterface)
			err := protoUnmarshal(utils2.StringToBytes(v), i)
			if err != nil {
				return err
			}
			if teKind == reflect.Ptr {
				elem.SetMapIndex(reflect.ValueOf(k), value)
			} else {
				elem.SetMapIndex(reflect.ValueOf(k), value.Elem())
			}
		}
		return nil
	}

	panic("not support type :" + tv.String())
}

func HLen(ctx context.Context, c redis.Cmdable, key string) (int64, *errmsg.ErrMsg) {
	count, err := c.HLen(ctx, key).Result()
	if err != nil {
		return 0, errmsg.NewErrorDB(err)
	}
	return count, nil
}

func HScanWith(ctx context.Context, c redis.Cmdable, key string, cursor uint64, match string, count int64, out interface{}) (uint64, *errmsg.ErrMsg) {
	tv := reflect.ValueOf(out)
	if tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Slice {
		keys, cursor, err := c.HScan(ctx, key, cursor, match, count).Result()
		if err != nil {
			return 0, errmsg.NewErrorDB(err)
		}
		elem := tv.Type().Elem().Elem()
		outR := reflect.MakeSlice(reflect.SliceOf(elem), len(keys)/2, len(keys)/2)
		index := 0
		te := elem
		if elem.Kind() == reflect.Ptr {
			te = elem.Elem()
		}
		for j := 0; j < len(keys); j += 2 {
			t := reflect.New(te)
			i := t.Interface().(RedisInterface)
			err := protoUnmarshal(utils2.StringToBytes(keys[j+1]), i)
			if err != nil {
				return 0, err
			}
			if elem.Kind() == reflect.Ptr {
				outR.Index(index).Set(t)
			} else {
				outR.Index(index).Set(t.Elem())
			}
			index++
		}
		tv.Elem().Set(outR)
		return cursor, nil
	}

	if (tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Map) || tv.Kind() == reflect.Map {
		elem := tv
		if tv.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Map && elem.Type().Key().Kind() != reflect.String {
			panic("not support type:" + tv.String())
		}

		keys, cursor, err := c.HScan(ctx, key, cursor, match, count).Result()
		if err != nil {
			return 0, errmsg.NewErrorDB(err)
		}

		te := elem.Type().Elem()
		teKind := te.Kind()
		if teKind == reflect.Ptr {
			te = te.Elem()
		}
		for j := 0; j < len(keys); j += 2 {
			value := reflect.New(te)
			i := value.Interface().(RedisInterface)
			err := protoUnmarshal(utils2.StringToBytes(keys[j+1]), i)
			if err != nil {
				return 0, err
			}
			if teKind == reflect.Ptr {
				elem.SetMapIndex(reflect.ValueOf(keys[j]), value)
			} else {
				elem.SetMapIndex(reflect.ValueOf(keys[j]), value.Elem())
			}
		}
		return cursor, nil
	}
	panic("not support type :" + tv.String())
}

func (this_ *Orm) HScan(c redis.Cmdable, key string, cursor uint64, match string, count int64, out interface{}) (uint64, *errmsg.ErrMsg) {
	cursor, err := HScanWith(this_.ctx, c, key, cursor, match, count, out)
	if err != nil {
		return 0, err
	}
	return cursor, nil
}

func (this_ *Orm) HScanGetAll(c redis.Cmdable, key string, out interface{}) *errmsg.ErrMsg {
	r, ok := this_.cache[c]
	if !ok {
		r = map[string]interface{}{}
		this_.cache[c] = r
	}
	kvi, ok := r[key]
	if !ok {
		kvi = &HALL{all: false, data: map[string]RedisInterface{}}
		r[key] = kvi
	}
	kv := kvi.(*HALL)
	tv := reflect.ValueOf(out)
	if !kv.all {
		if (tv.Kind() != reflect.Ptr && tv.Kind() != reflect.Map) ||
			(tv.Kind() == reflect.Ptr && tv.Elem().Kind() != reflect.Slice && tv.Elem().Kind() != reflect.Map) {
			panic("not support type :" + tv.String())
		}
		if tv.Kind() == reflect.Map || tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Map {
			mType := tv.Type()
			if mType.Kind() == reflect.Ptr {
				mType = mType.Elem()
			}
			if mType.Key().Kind() != reflect.String {
				panic("not support type :" + tv.String())
			}
		}

		elementType := GetElementType(out)
		use := map[string]RedisInterface{}
		err := HScanWithMap(this_.ctx, c, key, elementType, use)
		if err != nil {
			return err
		}

		kv.all = true
		kv.data = use
	}

	if tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Slice {
		elem := tv.Elem()
		ptr := elem.Type().Elem().Kind() == reflect.Ptr
		for _, v := range kv.data {
			rv := reflect.ValueOf(v)
			if !ptr {
				rv = rv.Elem()
			}
			elem = reflect.Append(elem, rv)
		}
		tv.Elem().Set(elem)
		return nil
	}

	if (tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Map) || tv.Kind() == reflect.Map {
		elem := tv
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Map || elem.Type().Key().Kind() != reflect.String {
			panic("not support type:" + tv.String())
		}
		ptr := elem.Type().Elem().Kind() == reflect.Ptr
		for _, v := range kv.data {
			rv := reflect.ValueOf(v)
			if !ptr {
				rv = rv.Elem()
			}
			elem.SetMapIndex(reflect.ValueOf(v.PK()), rv)
		}
		return nil
	}

	panic("not support type:" + tv.String())
}

func GetElementType(out interface{}) reflect.Type {
	tv := reflect.ValueOf(out)
	if tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Slice {
		elem := tv.Type().Elem().Elem()
		if elem.Kind() == reflect.Ptr {
			return elem.Elem()
		}
		return elem
	}

	if (tv.Kind() == reflect.Ptr && tv.Elem().Kind() == reflect.Map) || tv.Kind() == reflect.Map {
		elem := tv
		if tv.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		if elem.Kind() != reflect.Map && elem.Type().Key().Kind() != reflect.String {
			panic("not support type:" + tv.String())
		}

		if elem.Type().Elem().Kind() == reflect.Ptr {
			fmt.Println(elem.Type().Elem().Elem())
			return elem.Type().Elem().Elem()
		}
		return elem.Type().Elem()
	}
	panic("not support type :" + tv.String())
}

func HScanWithMap(ctx context.Context, c redis.Cmdable, key string, elementType reflect.Type, out map[string]RedisInterface) *errmsg.ErrMsg {
	cursor := uint64(0)
	count := int64(20)
	for {
		keys, localCursor, err := c.HScan(ctx, key, cursor, "*", count).Result()
		cursor = localCursor
		if err != nil {
			return errmsg.NewErrorDB(err)
		}
		keysNum := len(keys)
		if keysNum%2 != 0 {
			return errmsg.NewErrorDB(errors.New("HScan Err len :" + strconv.FormatInt(int64(keysNum), 10)))
		}

		for j := 0; j < len(keys); j += 2 {
			value := reflect.New(elementType)
			i := value.Interface().(RedisInterface)
			err := protoUnmarshal(utils2.StringToBytes(keys[j+1]), i)
			if err != nil {
				return err
			}
			out[keys[j]] = i
		}

		if cursor == 0 {
			break
		}
	}
	return nil
}
