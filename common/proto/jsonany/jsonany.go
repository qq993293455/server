package jsonany

import (
	"coin-server/common/buffer"
	"coin-server/common/jwriter"

	"github.com/gogo/protobuf/types"
)

type JsonAny interface {
	JsonBytes(w *jwriter.Writer)
}

type Any types.Any

func (m *Any) JsonBytes(w *jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}
	comma := false
	w.RawByte('{')
	if m.TypeUrl != "" {
		w.RawString(`"type_url":`)
		w.String(m.TypeUrl)
		comma = true
	}
	if comma {
		w.RawByte(',')
	}
	if len(m.Value) > 0 {
		w.RawString(`"value":`)
		w.Base64Bytes(m.Value)
	}
	w.RawByte('}')
}

func (m *Any) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{Buffer: buffer.Buffer{Buf: make([]byte, 2048)}}
	m.JsonBytes(&w)
	return w.Buffer.Buf, nil
}
