package handler

import (
	"coin-server/common/msgcreate"
	"coin-server/common/utils"

	"github.com/gogo/protobuf/proto"
)

type Parser struct {
	Raw    []byte
	RawMap map[string]string
}

func NewParser(raw []byte, rawMap map[string]string) *Parser {
	return &Parser{
		Raw:    raw,
		RawMap: rawMap,
	}
}

func (p *Parser) KV() interface{} {
	return p.unmarshalData(string(p.Raw))
}

func (p *Parser) Hash() map[string]interface{} {
	if len(p.RawMap) <= 0 {
		return nil
	}
	ret := make(map[string]interface{}, len(p.RawMap))
	for k, v := range p.RawMap {
		ret[k] = p.unmarshalData(v)
	}
	return ret
}

const messageNameLen = 1

func (p *Parser) unmarshalData(v string) interface{} {
	data := []byte(v)
	out := msgcreate.NewMessage(utils.BytesToString(data[messageNameLen : messageNameLen+data[0]]))
	if err := proto.Unmarshal(data[messageNameLen+data[0]:], out); err != nil {
		return v
	}
	return out
}
