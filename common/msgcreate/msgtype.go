package msgcreate

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

var m = map[string]func() proto.Message{}

func NewMessage(typeUrl string) proto.Message {
	v, ok := m[typeUrl]
	if !ok {
		panic(fmt.Sprintf("not found message name:%s", typeUrl))
	}
	return v()
}

func MessageName(msg proto.Message) string {
	return proto.MessageName(msg)
}

func RegisterNewMessage(f func() proto.Message) {
	typeUrl := proto.MessageName(f())
	if _, ok := m[typeUrl]; ok {
		panic("RegisterNewMessage: message '" + typeUrl + "' had registered")
	}

	m[typeUrl] = f
}

func GetAll() <-chan string {
	c := make(chan string, 20)
	go func() {
		for k := range m {
			c <- k
		}
		close(c)
	}()

	return c
}
