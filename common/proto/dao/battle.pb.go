// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/dao/battle.proto

package dao

import (
	coin_server_common_buffer "coin-server/common/buffer"
	coin_server_common_bytespool "coin-server/common/bytespool"
	coin_server_common_jwriter "coin-server/common/jwriter"
	coin_server_common_msgcreate "coin-server/common/msgcreate"
	coin_server_common_proto_jsonany "coin-server/common/proto/jsonany"
	models "coin-server/common/proto/models"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
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

type RoleTempBag struct {
	RoleId           string          `protobuf:"bytes,1,opt,name=role_id,json=roleId,proto3" json:"role_id,omitempty" pk`
	TempBag          *models.TempBag `protobuf:"bytes,2,opt,name=temp_bag,json=tempBag,proto3" json:"temp_bag,omitempty"`
	MapId            int64           `protobuf:"varint,3,opt,name=map_id,json=mapId,proto3" json:"map_id,omitempty"`
	ProfitUpper      int64           `protobuf:"varint,4,opt,name=profit_upper,json=profitUpper,proto3" json:"profit_upper,omitempty"`
	LastCalcTime     int64           `protobuf:"varint,5,opt,name=last_calc_time,json=lastCalcTime,proto3" json:"last_calc_time,omitempty"`
	ExpProfitAdd     int64           `protobuf:"varint,6,opt,name=exp_profit_add,json=expProfitAdd,proto3" json:"exp_profit_add,omitempty"`
	ExpProfitPercent int64           `protobuf:"varint,7,opt,name=exp_profit_percent,json=expProfitPercent,proto3" json:"exp_profit_percent,omitempty"`
	KickOutTime      int64           `protobuf:"varint,8,opt,name=kick_out_time,json=kickOutTime,proto3" json:"kick_out_time,omitempty"`
	ExpProfitBase    int64           `protobuf:"varint,9,opt,name=exp_profit_base,json=expProfitBase,proto3" json:"exp_profit_base,omitempty"`
	CapLimit         int64           `protobuf:"varint,10,opt,name=cap_limit,json=capLimit,proto3" json:"cap_limit,omitempty"`
	BagSize          int64           `protobuf:"varint,11,opt,name=bag_size,json=bagSize,proto3" json:"bag_size,omitempty"`
}

func (m *RoleTempBag) Reset()      { *m = RoleTempBag{} }
func (*RoleTempBag) ProtoMessage() {}
func (*RoleTempBag) Descriptor() ([]byte, []int) {
	return fileDescriptor_a99d8f678abf1e03, []int{0}
}
func (m *RoleTempBag) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RoleTempBag) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RoleTempBag.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RoleTempBag) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RoleTempBag.Merge(m, src)
}
func (m *RoleTempBag) XXX_Size() int {
	return m.Size()
}
func (m *RoleTempBag) XXX_DiscardUnknown() {
	xxx_messageInfo_RoleTempBag.DiscardUnknown(m)
}

var xxx_messageInfo_RoleTempBag proto.InternalMessageInfo

func (m *RoleTempBag) GetRoleId() string {
	if m != nil {
		return m.RoleId
	}
	return ""
}

func (m *RoleTempBag) GetTempBag() *models.TempBag {
	if m != nil {
		return m.TempBag
	}
	return nil
}

func (m *RoleTempBag) GetMapId() int64 {
	if m != nil {
		return m.MapId
	}
	return 0
}

func (m *RoleTempBag) GetProfitUpper() int64 {
	if m != nil {
		return m.ProfitUpper
	}
	return 0
}

func (m *RoleTempBag) GetLastCalcTime() int64 {
	if m != nil {
		return m.LastCalcTime
	}
	return 0
}

func (m *RoleTempBag) GetExpProfitAdd() int64 {
	if m != nil {
		return m.ExpProfitAdd
	}
	return 0
}

func (m *RoleTempBag) GetExpProfitPercent() int64 {
	if m != nil {
		return m.ExpProfitPercent
	}
	return 0
}

func (m *RoleTempBag) GetKickOutTime() int64 {
	if m != nil {
		return m.KickOutTime
	}
	return 0
}

func (m *RoleTempBag) GetExpProfitBase() int64 {
	if m != nil {
		return m.ExpProfitBase
	}
	return 0
}

func (m *RoleTempBag) GetCapLimit() int64 {
	if m != nil {
		return m.CapLimit
	}
	return 0
}

func (m *RoleTempBag) GetBagSize() int64 {
	if m != nil {
		return m.BagSize
	}
	return 0
}

func (*RoleTempBag) XXX_MessageName() string {
	return "dao.RoleTempBag"
}
func init() {
	proto.RegisterType((*RoleTempBag)(nil), "dao.RoleTempBag")
}

func init() { proto.RegisterFile("proto/dao/battle.proto", fileDescriptor_a99d8f678abf1e03) }

var fileDescriptor_a99d8f678abf1e03 = []byte{
	// 418 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x92, 0xc1, 0x6e, 0xd3, 0x30,
	0x18, 0x80, 0xe3, 0x95, 0xa5, 0xad, 0xbb, 0x31, 0x64, 0x01, 0xca, 0x06, 0xf2, 0xca, 0x84, 0x50,
	0x85, 0xa0, 0x95, 0xe0, 0x09, 0x28, 0xa7, 0x21, 0x24, 0xa6, 0x32, 0x2e, 0x5c, 0xac, 0x3f, 0xf6,
	0x4f, 0x64, 0x35, 0xae, 0xad, 0xc4, 0x45, 0xd3, 0x9e, 0x02, 0xf1, 0x14, 0x3c, 0xca, 0x8e, 0x3b,
	0xee, 0x84, 0x20, 0xbd, 0x70, 0x44, 0x3c, 0x01, 0x8a, 0x5d, 0x55, 0xdd, 0xed, 0xcf, 0xf7, 0x7f,
	0xf9, 0xfe, 0x1c, 0x42, 0x1f, 0xba, 0xca, 0x7a, 0x3b, 0x51, 0x60, 0x27, 0x39, 0x78, 0x5f, 0xe2,
	0x38, 0x00, 0xd6, 0x51, 0x60, 0x8f, 0xee, 0x17, 0xb6, 0xb0, 0x51, 0x68, 0xa7, 0xb8, 0x3a, 0x3a,
	0x8c, 0xc4, 0x58, 0x85, 0x65, 0x7d, 0xeb, 0xad, 0x93, 0xef, 0x1d, 0x3a, 0x98, 0xd9, 0x12, 0xcf,
	0xd1, 0xb8, 0x29, 0x14, 0xec, 0x98, 0x76, 0x2b, 0x5b, 0xa2, 0xd0, 0x2a, 0x23, 0x43, 0x32, 0xea,
	0x4f, 0xd3, 0x7f, 0x3f, 0x8f, 0x77, 0xdc, 0x7c, 0x96, 0xb6, 0xf8, 0x54, 0xb1, 0xe7, 0xb4, 0xe7,
	0xd1, 0x38, 0x91, 0x43, 0x91, 0xed, 0x0c, 0xc9, 0x68, 0xf0, 0xea, 0x60, 0x1c, 0xc3, 0xe3, 0x75,
	0x63, 0xd6, 0xf5, 0xeb, 0xd8, 0x03, 0x9a, 0x1a, 0x70, 0x6d, 0xab, 0x33, 0x24, 0xa3, 0xce, 0x6c,
	0xd7, 0x80, 0x3b, 0x55, 0xec, 0x09, 0xdd, 0x73, 0x95, 0xfd, 0xa2, 0xbd, 0x58, 0x3a, 0x87, 0x55,
	0x76, 0x27, 0x2c, 0x07, 0x91, 0x7d, 0x6a, 0x11, 0x7b, 0x4a, 0xef, 0x96, 0x50, 0x7b, 0x21, 0xa1,
	0x94, 0xc2, 0x6b, 0x83, 0xd9, 0x6e, 0x90, 0xf6, 0x5a, 0xfa, 0x16, 0x4a, 0x79, 0xae, 0x0d, 0xb6,
	0x16, 0x5e, 0x38, 0xb1, 0x8e, 0x81, 0x52, 0x59, 0x1a, 0x2d, 0xbc, 0x70, 0x67, 0x01, 0xbe, 0x51,
	0x8a, 0xbd, 0xa0, 0x6c, 0xcb, 0x72, 0x58, 0x49, 0x5c, 0xf8, 0xac, 0x1b, 0xcc, 0x7b, 0x1b, 0xf3,
	0x2c, 0x72, 0x76, 0x42, 0xf7, 0xe7, 0x5a, 0xce, 0x85, 0x5d, 0xfa, 0x78, 0xb8, 0x17, 0xbf, 0xae,
	0x85, 0x1f, 0x96, 0x3e, 0xdc, 0x7d, 0x46, 0x0f, 0xb6, 0x8a, 0x39, 0xd4, 0x98, 0xf5, 0x83, 0xb5,
	0xbf, 0xc9, 0x4d, 0xa1, 0x46, 0xf6, 0x88, 0xf6, 0x25, 0x38, 0x51, 0x6a, 0xa3, 0x7d, 0x46, 0x83,
	0xd1, 0x93, 0xe0, 0xde, 0xb7, 0xcf, 0xec, 0x90, 0xf6, 0x72, 0x28, 0x44, 0xad, 0x2f, 0x31, 0x1b,
	0x84, 0x5d, 0x37, 0x87, 0xe2, 0xa3, 0xbe, 0xc4, 0xe9, 0xbb, 0x9b, 0xdf, 0x3c, 0xf9, 0xd1, 0x70,
	0x72, 0xd5, 0x70, 0x72, 0xdd, 0x70, 0xf2, 0xab, 0xe1, 0xe4, 0x4f, 0xc3, 0x93, 0xbf, 0x0d, 0x27,
	0xdf, 0x56, 0x3c, 0xb9, 0x5a, 0x71, 0x72, 0xbd, 0xe2, 0xc9, 0xcd, 0x8a, 0x27, 0x9f, 0x1f, 0x4b,
	0xab, 0x17, 0x2f, 0x6b, 0xac, 0xbe, 0x62, 0x35, 0x91, 0xd6, 0x18, 0xbb, 0x98, 0x6c, 0xfe, 0x91,
	0x3c, 0x0d, 0xe3, 0xeb, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff, 0xd7, 0xf3, 0x85, 0xc7, 0x37, 0x02,
	0x00, 0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolRoleTempBag.Get().(proto.Message)
	})
}

var poolRoleTempBag = &sync.Pool{New: func() interface{} { return &RoleTempBag{} }}

func (m *RoleTempBag) ReleasePool() { m.Reset(); poolRoleTempBag.Put(m); m = nil }

func (m *RoleTempBag) PK() string {
	if m == nil {
		return ""
	}
	return m.RoleId
}

func (m *RoleTempBag) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return append(d, m.RoleId...)
}

func (m *RoleTempBag) ToKVSave() ([]byte, []byte) {
	msgName := m.XXX_MessageName()
	dk := coin_server_common_bytespool.GetSample(64)
	dk = dk[:0]
	dk = append(dk, msgName...)
	dk = append(dk, ':', 'k', ':')
	dk = m.PKAppendTo(dk)
	return dk, m.ToSave()
}

func (m *RoleTempBag) ToSave() []byte {
	msgName := m.XXX_MessageName()
	ml := len(msgName)
	d := coin_server_common_bytespool.GetSample(1 + ml + m.Size())
	d[0] = uint8(ml)
	copy(d[1:], msgName)
	_, _ = m.MarshalToSizedBuffer(d[1+ml:])
	return d
}

func (m *RoleTempBag) KVKey() string {
	return m.XXX_MessageName() + ":k:" + m.PK()
}

func (this *RoleTempBag) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RoleTempBag)
	if !ok {
		that2, ok := that.(RoleTempBag)
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
	if this.RoleId != that1.RoleId {
		return false
	}
	if !this.TempBag.Equal(that1.TempBag) {
		return false
	}
	if this.MapId != that1.MapId {
		return false
	}
	if this.ProfitUpper != that1.ProfitUpper {
		return false
	}
	if this.LastCalcTime != that1.LastCalcTime {
		return false
	}
	if this.ExpProfitAdd != that1.ExpProfitAdd {
		return false
	}
	if this.ExpProfitPercent != that1.ExpProfitPercent {
		return false
	}
	if this.KickOutTime != that1.KickOutTime {
		return false
	}
	if this.ExpProfitBase != that1.ExpProfitBase {
		return false
	}
	if this.CapLimit != that1.CapLimit {
		return false
	}
	if this.BagSize != that1.BagSize {
		return false
	}
	return true
}
func (m *RoleTempBag) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RoleTempBag) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RoleTempBag) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.BagSize != 0 {
		i = encodeVarintBattle(dAtA, i, uint64(m.BagSize))
		i--
		dAtA[i] = 0x58
	}
	if m.CapLimit != 0 {
		i = encodeVarintBattle(dAtA, i, uint64(m.CapLimit))
		i--
		dAtA[i] = 0x50
	}
	if m.ExpProfitBase != 0 {
		i = encodeVarintBattle(dAtA, i, uint64(m.ExpProfitBase))
		i--
		dAtA[i] = 0x48
	}
	if m.KickOutTime != 0 {
		i = encodeVarintBattle(dAtA, i, uint64(m.KickOutTime))
		i--
		dAtA[i] = 0x40
	}
	if m.ExpProfitPercent != 0 {
		i = encodeVarintBattle(dAtA, i, uint64(m.ExpProfitPercent))
		i--
		dAtA[i] = 0x38
	}
	if m.ExpProfitAdd != 0 {
		i = encodeVarintBattle(dAtA, i, uint64(m.ExpProfitAdd))
		i--
		dAtA[i] = 0x30
	}
	if m.LastCalcTime != 0 {
		i = encodeVarintBattle(dAtA, i, uint64(m.LastCalcTime))
		i--
		dAtA[i] = 0x28
	}
	if m.ProfitUpper != 0 {
		i = encodeVarintBattle(dAtA, i, uint64(m.ProfitUpper))
		i--
		dAtA[i] = 0x20
	}
	if m.MapId != 0 {
		i = encodeVarintBattle(dAtA, i, uint64(m.MapId))
		i--
		dAtA[i] = 0x18
	}
	if m.TempBag != nil {
		{
			size, err := m.TempBag.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintBattle(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.RoleId) > 0 {
		i -= len(m.RoleId)
		copy(dAtA[i:], m.RoleId)
		i = encodeVarintBattle(dAtA, i, uint64(len(m.RoleId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintBattle(dAtA []byte, offset int, v uint64) int {
	offset -= sovBattle(v)
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

func (m *RoleTempBag) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	if m.RoleId != "" {
		w.RawByte('"')
		w.RawString("role_id")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.RoleId)
		needWriteComma = true
	}
	if needWriteComma {
		w.RawByte(',')
	}
	w.RawByte('"')
	w.RawString("temp_bag")
	w.RawByte('"')
	w.RawByte(':')
	m.TempBag.JsonBytes(w)
	needWriteComma = true
	if m.MapId != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("map_id")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.MapId))
		needWriteComma = true
	}
	if m.ProfitUpper != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("profit_upper")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.ProfitUpper))
		needWriteComma = true
	}
	if m.LastCalcTime != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("last_calc_time")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.LastCalcTime))
		needWriteComma = true
	}
	if m.ExpProfitAdd != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("exp_profit_add")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.ExpProfitAdd))
		needWriteComma = true
	}
	if m.ExpProfitPercent != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("exp_profit_percent")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.ExpProfitPercent))
		needWriteComma = true
	}
	if m.KickOutTime != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("kick_out_time")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.KickOutTime))
		needWriteComma = true
	}
	if m.ExpProfitBase != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("exp_profit_base")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.ExpProfitBase))
		needWriteComma = true
	}
	if m.CapLimit != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("cap_limit")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.CapLimit))
		needWriteComma = true
	}
	if m.BagSize != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("bag_size")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.BagSize))
		needWriteComma = true
	}
	_ = needWriteComma
	w.RawByte('}')

}

func (m *RoleTempBag) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *RoleTempBag) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *RoleTempBag) GoString() string {
	return m.String()
}

func (m *RoleTempBag) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RoleId)
	if l > 0 {
		n += 1 + l + sovBattle(uint64(l))
	}
	if m.TempBag != nil {
		l = m.TempBag.Size()
		n += 1 + l + sovBattle(uint64(l))
	}
	if m.MapId != 0 {
		n += 1 + sovBattle(uint64(m.MapId))
	}
	if m.ProfitUpper != 0 {
		n += 1 + sovBattle(uint64(m.ProfitUpper))
	}
	if m.LastCalcTime != 0 {
		n += 1 + sovBattle(uint64(m.LastCalcTime))
	}
	if m.ExpProfitAdd != 0 {
		n += 1 + sovBattle(uint64(m.ExpProfitAdd))
	}
	if m.ExpProfitPercent != 0 {
		n += 1 + sovBattle(uint64(m.ExpProfitPercent))
	}
	if m.KickOutTime != 0 {
		n += 1 + sovBattle(uint64(m.KickOutTime))
	}
	if m.ExpProfitBase != 0 {
		n += 1 + sovBattle(uint64(m.ExpProfitBase))
	}
	if m.CapLimit != 0 {
		n += 1 + sovBattle(uint64(m.CapLimit))
	}
	if m.BagSize != 0 {
		n += 1 + sovBattle(uint64(m.BagSize))
	}
	return n
}

func sovBattle(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozBattle(x uint64) (n int) {
	return sovBattle(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *RoleTempBag) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBattle
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
			return fmt.Errorf("proto: RoleTempBag: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RoleTempBag: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RoleId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthBattle
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthBattle
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RoleId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TempBag", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
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
				return ErrInvalidLengthBattle
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBattle
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.TempBag == nil {
				m.TempBag = &models.TempBag{}
			}
			if err := m.TempBag.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MapId", wireType)
			}
			m.MapId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MapId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProfitUpper", wireType)
			}
			m.ProfitUpper = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProfitUpper |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastCalcTime", wireType)
			}
			m.LastCalcTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LastCalcTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExpProfitAdd", wireType)
			}
			m.ExpProfitAdd = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExpProfitAdd |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExpProfitPercent", wireType)
			}
			m.ExpProfitPercent = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExpProfitPercent |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field KickOutTime", wireType)
			}
			m.KickOutTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.KickOutTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExpProfitBase", wireType)
			}
			m.ExpProfitBase = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExpProfitBase |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 10:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CapLimit", wireType)
			}
			m.CapLimit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CapLimit |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BagSize", wireType)
			}
			m.BagSize = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBattle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BagSize |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipBattle(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthBattle
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
func skipBattle(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowBattle
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
					return 0, ErrIntOverflowBattle
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
					return 0, ErrIntOverflowBattle
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
				return 0, ErrInvalidLengthBattle
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupBattle
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthBattle
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthBattle        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowBattle          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupBattle = fmt.Errorf("proto: unexpected end of group")
)
