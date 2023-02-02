package orm

import "github.com/gogo/protobuf/proto"

type RedisInterface interface {
	XXX_MessageName() string
	ToKVSave() ([]byte, []byte)
	ToSave() []byte
	PK() string
	KVKey() string
	proto.Message
}
