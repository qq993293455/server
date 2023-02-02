package consulkv

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"coin-server/common/logger"

	"go.uber.org/zap"

	"github.com/hashicorp/consul/api"
)

type Config struct {
	addrS       []string
	namespace   string
	mutex       sync.RWMutex
	kv          map[string]*value
	ccMap       map[string]*api.Client
	log         *logger.Logger
	showRawData bool
}

type value struct {
	sync.RWMutex
	data     []byte
	index    uint64
	onChange func()
}

func (this_ *value) write(f func(*value)) {
	this_.Lock()
	defer this_.Unlock()
	f(this_)
}

func (this_ *value) read(f func(*value)) {
	this_.RLock()
	defer this_.RUnlock()
	f(this_)
}

func NewConfig(namespace string, hosts string, log *logger.Logger, showRawData ...bool) (*Config, error) {
	namespace = strings.TrimPrefix(namespace, "/")
	if !strings.HasSuffix(namespace, "/") {
		namespace += "/"
	}
	addrS := strings.Split(hosts, ",")
	c := &Config{
		addrS:     addrS,
		namespace: namespace,
		kv:        map[string]*value{},
		ccMap:     map[string]*api.Client{},
		log:       log.With(zap.Fields(zap.Strings("consul", addrS))),
	}
	if len(showRawData) > 0 {
		c.showRawData = showRawData[0]
	}
	for _, v := range addrS {
		dc := api.DefaultConfig()
		dc.Address = v
		cc, err := api.NewClient(dc)
		if err != nil {
			return nil, err
		}
		c.ccMap[v] = cc
	}
	err := c.getAll()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (this_ *Config) OnChange(key string, f func()) error {
	this_.mutex.RLock()
	v, ok := this_.kv[key]
	this_.mutex.RUnlock()
	if !ok {
		return fmt.Errorf("not found key:%s", key)
	}
	v.write(func(v *value) {
		v.onChange = f
	})
	return nil
}

func (this_ *Config) getAll() error {
	var err error
	for _, v := range this_.ccMap {
		var kvs api.KVPairs
		kvs, _, err = v.KV().List(this_.namespace, nil)
		if err == nil {
			this_.setRaw(kvs)
			return nil
		}
	}
	if err == nil {
		this_.watch()
	}
	return err
}

func (this_ *Config) setRaw(kvs api.KVPairs) {
	this_.mutex.Lock()
	defer this_.mutex.Unlock()
	for _, v := range kvs {
		if len(v.Value) == 0 || v.Key == this_.namespace {
			continue
		}
		if this_.showRawData {
			this_.log.Info("consul kv", zap.String("key", v.Key), zap.ByteString("value", v.Value))
		}
		this_.kv[v.Key] = &value{
			data:  v.Value,
			index: v.ModifyIndex,
		}
	}
}

func (this_ *Config) watch() {
	this_.mutex.RLock()
	temp := make(map[string]*value, len(this_.kv))
	for k, v := range this_.kv {
		temp[k] = v
	}
	this_.mutex.RUnlock()

	for k, v := range temp {
		this_.watchOne(k, v)
	}
}

func (this_ *Config) watchOne(key string, val *value) {
	go func() {
		for {
			for addr, c := range this_.ccMap {
				for {
					opts := api.QueryOptions{
						WaitIndex: val.index,
					}
					keypair, meta, err := c.KV().Get(key, &opts)
					if err != nil {
						this_.log.Error("get kv error", zap.Error(err), zap.String("addr", addr))
						break
					}
					val.write(func(v *value) {
						v.index = meta.LastIndex
					})
					val.index = meta.LastIndex
					if keypair != nil {
						val.write(func(v *value) {
							v.data = keypair.Value
						})
						this_.setValueAndCall(key, val)
					}
				}
				time.Sleep(time.Second)
			}
		}
	}()
}

func (this_ *Config) setValueAndCall(key string, val *value) {
	this_.setValue(key, val)
	var f func()
	var data []byte
	val.read(func(v *value) {
		f = v.onChange
		data = v.data
	})
	this_.log.Info("consul watch", zap.String("key", key), zap.ByteString("value", data))
	if f != nil {
		f()
	}
}

func (this_ *Config) setValue(key string, val *value) {
	this_.mutex.Lock()
	defer this_.mutex.Unlock()
	this_.kv[key] = val
}

func (this_ *Config) GetRawValue(key string) ([]byte, bool) {
	this_.mutex.RLock()
	v, ok := this_.kv[key]
	this_.mutex.RUnlock()
	if !ok {
		return nil, false
	}
	var data []byte
	v.read(func(v *value) {
		data = v.data
	})
	return data, true
}

func (this_ *Config) Unmarshal(key string, v interface{}) error {
	data, ok := this_.GetRawValue(this_.namespace + key)
	if !ok {
		return fmt.Errorf("not found key:%s", key)
	}
	return json.Unmarshal(data, v)
}

func (this_ *Config) GetString(key string) (string, bool) {
	b, ok := this_.GetRawValue(key)
	if !ok {
		return "", false
	}
	return string(b), true
}
