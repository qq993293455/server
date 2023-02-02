// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/dao/relics.proto

package dao

import (
	coin_server_common_buffer "coin-server/common/buffer"
	coin_server_common_bytespool "coin-server/common/bytespool"
	coin_server_common_jwriter "coin-server/common/jwriter"
	coin_server_common_msgcreate "coin-server/common/msgcreate"
	coin_server_common_proto_jsonany "coin-server/common/proto/jsonany"
	_ "coin-server/common/proto/models"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
	strconv "strconv"
	sync "sync"
	unsafe "unsafe"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Relics struct {
	RelicsId  int64           `protobuf:"varint,1,opt,name=relics_id,json=relicsId,proto3" json:"relics_id,omitempty" pk`
	Level     int64           `protobuf:"varint,2,opt,name=level,proto3" json:"level,omitempty"`
	Star      int64           `protobuf:"varint,3,opt,name=star,proto3" json:"star,omitempty"`
	Attr      map[int64]int64 `protobuf:"bytes,4,rep,name=attr,proto3" json:"attr,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	IsNew     bool            `protobuf:"varint,5,opt,name=is_new,json=isNew,proto3" json:"is_new,omitempty"`
	FuncAttr  map[int64]int64 `protobuf:"bytes,6,rep,name=func_attr,json=funcAttr,proto3" json:"func_attr,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	FuncTimes int64           `protobuf:"varint,7,opt,name=func_times,json=funcTimes,proto3" json:"func_times,omitempty"`
}

func (m *Relics) Reset()      { *m = Relics{} }
func (*Relics) ProtoMessage() {}
func (*Relics) Descriptor() ([]byte, []int) {
	return fileDescriptor_2421c715e995929d, []int{0}
}
func (m *Relics) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Relics) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Relics.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Relics) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Relics.Merge(m, src)
}
func (m *Relics) XXX_Size() int {
	return m.Size()
}
func (m *Relics) XXX_DiscardUnknown() {
	xxx_messageInfo_Relics.DiscardUnknown(m)
}

var xxx_messageInfo_Relics proto.InternalMessageInfo

func (m *Relics) GetRelicsId() int64 {
	if m != nil {
		return m.RelicsId
	}
	return 0
}

func (m *Relics) GetLevel() int64 {
	if m != nil {
		return m.Level
	}
	return 0
}

func (m *Relics) GetStar() int64 {
	if m != nil {
		return m.Star
	}
	return 0
}

func (m *Relics) GetAttr() map[int64]int64 {
	if m != nil {
		return m.Attr
	}
	return nil
}

func (m *Relics) GetIsNew() bool {
	if m != nil {
		return m.IsNew
	}
	return false
}

func (m *Relics) GetFuncAttr() map[int64]int64 {
	if m != nil {
		return m.FuncAttr
	}
	return nil
}

func (m *Relics) GetFuncTimes() int64 {
	if m != nil {
		return m.FuncTimes
	}
	return 0
}

func (*Relics) XXX_MessageName() string {
	return "dao.Relics"
}

type RelicsSuit struct {
	SuitId int64 `protobuf:"varint,1,opt,name=suit_id,json=suitId,proto3" json:"suit_id,omitempty" pk`
	Level  int64 `protobuf:"varint,2,opt,name=level,proto3" json:"level,omitempty"`
	Star   int64 `protobuf:"varint,3,opt,name=star,proto3" json:"star,omitempty"`
}

func (m *RelicsSuit) Reset()      { *m = RelicsSuit{} }
func (*RelicsSuit) ProtoMessage() {}
func (*RelicsSuit) Descriptor() ([]byte, []int) {
	return fileDescriptor_2421c715e995929d, []int{1}
}
func (m *RelicsSuit) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RelicsSuit) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RelicsSuit.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RelicsSuit) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RelicsSuit.Merge(m, src)
}
func (m *RelicsSuit) XXX_Size() int {
	return m.Size()
}
func (m *RelicsSuit) XXX_DiscardUnknown() {
	xxx_messageInfo_RelicsSuit.DiscardUnknown(m)
}

var xxx_messageInfo_RelicsSuit proto.InternalMessageInfo

func (m *RelicsSuit) GetSuitId() int64 {
	if m != nil {
		return m.SuitId
	}
	return 0
}

func (m *RelicsSuit) GetLevel() int64 {
	if m != nil {
		return m.Level
	}
	return 0
}

func (m *RelicsSuit) GetStar() int64 {
	if m != nil {
		return m.Star
	}
	return 0
}

func (*RelicsSuit) XXX_MessageName() string {
	return "dao.RelicsSuit"
}
func init() {
	proto.RegisterType((*Relics)(nil), "dao.Relics")
	proto.RegisterMapType((map[int64]int64)(nil), "dao.Relics.AttrEntry")
	proto.RegisterMapType((map[int64]int64)(nil), "dao.Relics.FuncAttrEntry")
	proto.RegisterType((*RelicsSuit)(nil), "dao.RelicsSuit")
}

func init() { proto.RegisterFile("proto/dao/relics.proto", fileDescriptor_2421c715e995929d) }

var fileDescriptor_2421c715e995929d = []byte{
	// 390 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0xc1, 0xaa, 0xd3, 0x40,
	0x14, 0x86, 0x33, 0x49, 0x9b, 0xb6, 0x47, 0x04, 0x19, 0x5a, 0x89, 0x41, 0xa7, 0xa5, 0x6e, 0xea,
	0xc2, 0x04, 0x14, 0x54, 0x74, 0x65, 0x41, 0x41, 0x17, 0x2e, 0xa2, 0x20, 0xb8, 0x29, 0x31, 0x99,
	0x96, 0xa1, 0x49, 0xa6, 0xcc, 0x4c, 0x5a, 0xfa, 0x16, 0x3e, 0x86, 0x8f, 0xd2, 0x65, 0x97, 0x5d,
	0x89, 0x26, 0x20, 0x2e, 0xc5, 0x27, 0x90, 0xcc, 0xf4, 0xf6, 0xde, 0xcb, 0x5d, 0x75, 0xf7, 0x9f,
	0x7f, 0xce, 0xff, 0xcd, 0x9c, 0xc3, 0xc0, 0xdd, 0x95, 0xe0, 0x8a, 0x87, 0x69, 0xcc, 0x43, 0x41,
	0x33, 0x96, 0xc8, 0x40, 0x1b, 0xd8, 0x49, 0x63, 0xee, 0xf7, 0x17, 0x7c, 0xc1, 0x4d, 0x43, 0xa3,
	0xcc, 0x91, 0xef, 0x19, 0x27, 0xe7, 0x29, 0xcd, 0x64, 0x48, 0x8b, 0x32, 0x3f, 0x86, 0xc6, 0xbf,
	0x6d, 0x70, 0x23, 0x4d, 0xc1, 0x0f, 0xa1, 0x67, 0x78, 0x33, 0x96, 0x7a, 0x68, 0x84, 0x26, 0xce,
	0xd4, 0xfd, 0xf7, 0x63, 0x68, 0xaf, 0x96, 0x51, 0xd7, 0x1c, 0xbc, 0x4b, 0x71, 0x1f, 0xda, 0x19,
	0x5d, 0xd3, 0xcc, 0xb3, 0x9b, 0x86, 0xc8, 0x14, 0x18, 0x43, 0x4b, 0xaa, 0x58, 0x78, 0x8e, 0x36,
	0xb5, 0xc6, 0x8f, 0xa0, 0x15, 0x2b, 0x25, 0xbc, 0xd6, 0xc8, 0x99, 0xdc, 0x7a, 0x32, 0x08, 0xd2,
	0x98, 0x07, 0xe6, 0xa6, 0xe0, 0xb5, 0x52, 0xe2, 0x4d, 0xa1, 0xc4, 0x36, 0xd2, 0x2d, 0x78, 0x00,
	0x2e, 0x93, 0xb3, 0x82, 0x6e, 0xbc, 0xf6, 0x08, 0x4d, 0xba, 0x51, 0x9b, 0xc9, 0x0f, 0x74, 0x83,
	0x9f, 0x41, 0x6f, 0x5e, 0x16, 0xc9, 0x4c, 0x63, 0x5c, 0x8d, 0xb9, 0x77, 0x15, 0xf3, 0xb6, 0x2c,
	0x92, 0x4b, 0x54, 0x77, 0x7e, 0x2c, 0xf1, 0x03, 0x00, 0x9d, 0x53, 0x2c, 0xa7, 0xd2, 0xeb, 0xe8,
	0x37, 0x69, 0xd2, 0xa7, 0xc6, 0xf0, 0x9f, 0x43, 0xef, 0x94, 0xc2, 0x77, 0xc0, 0x59, 0xd2, 0xad,
	0x19, 0x37, 0x6a, 0x64, 0x33, 0xe1, 0x3a, 0xce, 0x4a, 0x7a, 0x31, 0xa1, 0x2e, 0x5e, 0xda, 0x2f,
	0x90, 0xff, 0x0a, 0x6e, 0x5f, 0xbb, 0xf2, 0x9c, 0xf0, 0xf8, 0x33, 0x80, 0x79, 0xf6, 0xc7, 0x92,
	0x29, 0x3c, 0x84, 0x8e, 0x2c, 0x99, 0xba, 0xb9, 0x69, 0xb7, 0xb1, 0xcf, 0xd9, 0xf3, 0xf4, 0xfd,
	0xe1, 0x17, 0xb1, 0xbe, 0x57, 0x04, 0xed, 0x2a, 0x82, 0xf6, 0x15, 0x41, 0x3f, 0x2b, 0x82, 0xfe,
	0x54, 0xc4, 0xfa, 0x5b, 0x11, 0xf4, 0xad, 0x26, 0xd6, 0xae, 0x26, 0x68, 0x5f, 0x13, 0xeb, 0x50,
	0x13, 0xeb, 0xcb, 0xfd, 0x84, 0xb3, 0xe2, 0xb1, 0xa4, 0x62, 0x4d, 0x45, 0x98, 0xf0, 0x3c, 0xe7,
	0x45, 0x78, 0xfa, 0x4f, 0x5f, 0x5d, 0x2d, 0x9f, 0xfe, 0x0f, 0x00, 0x00, 0xff, 0xff, 0x4e, 0xf1,
	0xf4, 0xc3, 0x63, 0x02, 0x00, 0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolRelics.Get().(proto.Message)
	})
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolRelicsSuit.Get().(proto.Message)
	})
}

var poolRelics = &sync.Pool{New: func() interface{} { return &Relics{} }}

func (m *Relics) ReleasePool() { m.Reset(); poolRelics.Put(m); m = nil }

var poolRelicsSuit = &sync.Pool{New: func() interface{} { return &RelicsSuit{} }}

func (m *RelicsSuit) ReleasePool() { m.Reset(); poolRelicsSuit.Put(m); m = nil }

func (m *Relics) PK() string {
	if m == nil {
		return ""
	}
	return strconv.FormatInt(int64(m.RelicsId), 10)
}

func (m *Relics) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return strconv.AppendInt(d, int64(m.RelicsId), 10)
}

func (m *Relics) ToKVSave() ([]byte, []byte) {
	msgName := m.XXX_MessageName()
	dk := coin_server_common_bytespool.GetSample(64)
	dk = dk[:0]
	dk = append(dk, msgName...)
	dk = append(dk, ':', 'k', ':')
	dk = m.PKAppendTo(dk)
	return dk, m.ToSave()
}

func (m *Relics) ToSave() []byte {
	msgName := m.XXX_MessageName()
	ml := len(msgName)
	d := coin_server_common_bytespool.GetSample(1 + ml + m.Size())
	d[0] = uint8(ml)
	copy(d[1:], msgName)
	_, _ = m.MarshalToSizedBuffer(d[1+ml:])
	return d
}

func (m *Relics) KVKey() string {
	return m.XXX_MessageName() + ":k:" + m.PK()
}

func (m *RelicsSuit) PK() string {
	if m == nil {
		return ""
	}
	return strconv.FormatInt(int64(m.SuitId), 10)
}

func (m *RelicsSuit) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return strconv.AppendInt(d, int64(m.SuitId), 10)
}

func (m *RelicsSuit) ToKVSave() ([]byte, []byte) {
	msgName := m.XXX_MessageName()
	dk := coin_server_common_bytespool.GetSample(64)
	dk = dk[:0]
	dk = append(dk, msgName...)
	dk = append(dk, ':', 'k', ':')
	dk = m.PKAppendTo(dk)
	return dk, m.ToSave()
}

func (m *RelicsSuit) ToSave() []byte {
	msgName := m.XXX_MessageName()
	ml := len(msgName)
	d := coin_server_common_bytespool.GetSample(1 + ml + m.Size())
	d[0] = uint8(ml)
	copy(d[1:], msgName)
	_, _ = m.MarshalToSizedBuffer(d[1+ml:])
	return d
}

func (m *RelicsSuit) KVKey() string {
	return m.XXX_MessageName() + ":k:" + m.PK()
}

func (this *Relics) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Relics)
	if !ok {
		that2, ok := that.(Relics)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.RelicsId != that1.RelicsId {
		return false
	}
	if this.Level != that1.Level {
		return false
	}
	if this.Star != that1.Star {
		return false
	}
	if len(this.Attr) != len(that1.Attr) {
		return false
	}
	for i := range this.Attr {
		if this.Attr[i] != that1.Attr[i] {
			return false
		}
	}
	if this.IsNew != that1.IsNew {
		return false
	}
	if len(this.FuncAttr) != len(that1.FuncAttr) {
		return false
	}
	for i := range this.FuncAttr {
		if this.FuncAttr[i] != that1.FuncAttr[i] {
			return false
		}
	}
	if this.FuncTimes != that1.FuncTimes {
		return false
	}
	return true
}
func (this *RelicsSuit) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RelicsSuit)
	if !ok {
		that2, ok := that.(RelicsSuit)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.SuitId != that1.SuitId {
		return false
	}
	if this.Level != that1.Level {
		return false
	}
	if this.Star != that1.Star {
		return false
	}
	return true
}
func (m *Relics) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Relics) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Relics) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.FuncTimes != 0 {
		i = encodeVarintRelics(dAtA, i, uint64(m.FuncTimes))
		i--
		dAtA[i] = 0x38
	}
	if len(m.FuncAttr) > 0 {
		for k := range m.FuncAttr {
			v := m.FuncAttr[k]
			baseI := i
			i = encodeVarintRelics(dAtA, i, uint64(v))
			i--
			dAtA[i] = 0x10
			i = encodeVarintRelics(dAtA, i, uint64(k))
			i--
			dAtA[i] = 0x8
			i = encodeVarintRelics(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x32
		}
	}
	if m.IsNew {
		i--
		if m.IsNew {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x28
	}
	if len(m.Attr) > 0 {
		for k := range m.Attr {
			v := m.Attr[k]
			baseI := i
			i = encodeVarintRelics(dAtA, i, uint64(v))
			i--
			dAtA[i] = 0x10
			i = encodeVarintRelics(dAtA, i, uint64(k))
			i--
			dAtA[i] = 0x8
			i = encodeVarintRelics(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x22
		}
	}
	if m.Star != 0 {
		i = encodeVarintRelics(dAtA, i, uint64(m.Star))
		i--
		dAtA[i] = 0x18
	}
	if m.Level != 0 {
		i = encodeVarintRelics(dAtA, i, uint64(m.Level))
		i--
		dAtA[i] = 0x10
	}
	if m.RelicsId != 0 {
		i = encodeVarintRelics(dAtA, i, uint64(m.RelicsId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *RelicsSuit) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RelicsSuit) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RelicsSuit) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Star != 0 {
		i = encodeVarintRelics(dAtA, i, uint64(m.Star))
		i--
		dAtA[i] = 0x18
	}
	if m.Level != 0 {
		i = encodeVarintRelics(dAtA, i, uint64(m.Level))
		i--
		dAtA[i] = 0x10
	}
	if m.SuitId != 0 {
		i = encodeVarintRelics(dAtA, i, uint64(m.SuitId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintRelics(dAtA []byte, offset int, v uint64) int {
	offset -= sovRelics(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

var _ = coin_server_common_proto_jsonany.Any{}

func (m *Relics) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	if m.RelicsId != 0 {
		w.RawByte('"')
		w.RawString("relics_id")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.RelicsId))
		needWriteComma = true
	}
	if m.Level != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("level")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.Level))
		needWriteComma = true
	}
	if m.Star != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("star")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.Star))
		needWriteComma = true
	}
	if needWriteComma {
		w.RawByte(',')
	}
	w.RawByte('"')
	w.RawString("attr")
	w.RawByte('"')
	w.RawByte(':')
	if m.Attr == nil {
		w.RawString("null")
	} else if len(m.Attr) == 0 {
		w.RawString("{}")
	} else {
		w.RawByte('{')
		mlAttr := len(m.Attr)
		for k, v := range m.Attr {
			w.RawByte('"')
			w.Int64(int64(k))
			w.RawByte('"')
			w.RawByte(':')
			w.Int64(int64(v))
			mlAttr--
			if mlAttr != 0 {
				w.RawByte(',')
			}
		}
		w.RawByte('}')
	}
	needWriteComma = true
	if m.IsNew {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("is_new")
		w.RawByte('"')
		w.RawByte(':')
		w.Bool(m.IsNew)
		needWriteComma = true
	}
	if needWriteComma {
		w.RawByte(',')
	}
	w.RawByte('"')
	w.RawString("func_attr")
	w.RawByte('"')
	w.RawByte(':')
	if m.FuncAttr == nil {
		w.RawString("null")
	} else if len(m.FuncAttr) == 0 {
		w.RawString("{}")
	} else {
		w.RawByte('{')
		mlFuncAttr := len(m.FuncAttr)
		for k, v := range m.FuncAttr {
			w.RawByte('"')
			w.Int64(int64(k))
			w.RawByte('"')
			w.RawByte(':')
			w.Int64(int64(v))
			mlFuncAttr--
			if mlFuncAttr != 0 {
				w.RawByte(',')
			}
		}
		w.RawByte('}')
	}
	needWriteComma = true
	if m.FuncTimes != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("func_times")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.FuncTimes))
		needWriteComma = true
	}
	_ = needWriteComma
	w.RawByte('}')

}

func (m *RelicsSuit) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	if m.SuitId != 0 {
		w.RawByte('"')
		w.RawString("suit_id")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.SuitId))
		needWriteComma = true
	}
	if m.Level != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("level")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.Level))
		needWriteComma = true
	}
	if m.Star != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("star")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.Star))
		needWriteComma = true
	}
	_ = needWriteComma
	w.RawByte('}')

}

func (m *Relics) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *Relics) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *Relics) GoString() string {
	return m.String()
}

func (m *RelicsSuit) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *RelicsSuit) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *RelicsSuit) GoString() string {
	return m.String()
}

func (m *Relics) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.RelicsId != 0 {
		n += 1 + sovRelics(uint64(m.RelicsId))
	}
	if m.Level != 0 {
		n += 1 + sovRelics(uint64(m.Level))
	}
	if m.Star != 0 {
		n += 1 + sovRelics(uint64(m.Star))
	}
	if len(m.Attr) > 0 {
		for k, v := range m.Attr {
			_ = k
			_ = v
			mapEntrySize := 1 + sovRelics(uint64(k)) + 1 + sovRelics(uint64(v))
			n += mapEntrySize + 1 + sovRelics(uint64(mapEntrySize))
		}
	}
	if m.IsNew {
		n += 2
	}
	if len(m.FuncAttr) > 0 {
		for k, v := range m.FuncAttr {
			_ = k
			_ = v
			mapEntrySize := 1 + sovRelics(uint64(k)) + 1 + sovRelics(uint64(v))
			n += mapEntrySize + 1 + sovRelics(uint64(mapEntrySize))
		}
	}
	if m.FuncTimes != 0 {
		n += 1 + sovRelics(uint64(m.FuncTimes))
	}
	return n
}

func (m *RelicsSuit) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.SuitId != 0 {
		n += 1 + sovRelics(uint64(m.SuitId))
	}
	if m.Level != 0 {
		n += 1 + sovRelics(uint64(m.Level))
	}
	if m.Star != 0 {
		n += 1 + sovRelics(uint64(m.Star))
	}
	return n
}

func sovRelics(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozRelics(x uint64) (n int) {
	return sovRelics(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Relics) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRelics
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Relics: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Relics: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RelicsId", wireType)
			}
			m.RelicsId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RelicsId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Level", wireType)
			}
			m.Level = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Level |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Star", wireType)
			}
			m.Star = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Star |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Attr", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthRelics
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRelics
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Attr == nil {
				m.Attr = make(map[int64]int64)
			}
			var mapkey int64
			var mapvalue int64
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowRelics
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowRelics
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapkey |= int64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
				} else if fieldNum == 2 {
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowRelics
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapvalue |= int64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipRelics(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return ErrInvalidLengthRelics
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Attr[mapkey] = mapvalue
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsNew", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.IsNew = bool(v != 0)
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FuncAttr", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthRelics
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRelics
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.FuncAttr == nil {
				m.FuncAttr = make(map[int64]int64)
			}
			var mapkey int64
			var mapvalue int64
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowRelics
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowRelics
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapkey |= int64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
				} else if fieldNum == 2 {
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowRelics
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapvalue |= int64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipRelics(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return ErrInvalidLengthRelics
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.FuncAttr[mapkey] = mapvalue
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FuncTimes", wireType)
			}
			m.FuncTimes = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.FuncTimes |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipRelics(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthRelics
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *RelicsSuit) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRelics
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RelicsSuit: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RelicsSuit: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SuitId", wireType)
			}
			m.SuitId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SuitId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Level", wireType)
			}
			m.Level = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Level |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Star", wireType)
			}
			m.Star = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Star |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipRelics(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthRelics
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipRelics(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowRelics
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowRelics
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthRelics
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupRelics
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthRelics
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthRelics        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowRelics          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupRelics = fmt.Errorf("proto: unexpected end of group")
)
