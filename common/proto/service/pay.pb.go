// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/service/pay.proto

package service

import (
	coin_server_common_buffer "coin-server/common/buffer"
	coin_server_common_jwriter "coin-server/common/jwriter"
	coin_server_common_msgcreate "coin-server/common/msgcreate"
	coin_server_common_proto_jsonany "coin-server/common/proto/jsonany"
	_ "coin-server/common/proto/models"
	fmt "fmt"
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

type Pay struct {
}

func (m *Pay) Reset()      { *m = Pay{} }
func (*Pay) ProtoMessage() {}
func (*Pay) Descriptor() ([]byte, []int) {
	return fileDescriptor_56bb10b4b524e10d, []int{0}
}
func (m *Pay) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Pay) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Pay.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Pay) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Pay.Merge(m, src)
}
func (m *Pay) XXX_Size() int {
	return m.Size()
}
func (m *Pay) XXX_DiscardUnknown() {
	xxx_messageInfo_Pay.DiscardUnknown(m)
}

var xxx_messageInfo_Pay proto.InternalMessageInfo

func (*Pay) XXX_MessageName() string {
	return "service.Pay"
}

type Pay_Success struct {
	PcId       int64 `protobuf:"varint,1,opt,name=pc_id,json=pcId,proto3" json:"pc_id,omitempty"`
	PaidTime   int64 `protobuf:"varint,2,opt,name=paid_time,json=paidTime,proto3" json:"paid_time,omitempty"`
	ExpireTime int64 `protobuf:"varint,3,opt,name=expire_time,json=expireTime,proto3" json:"expire_time,omitempty"`
}

func (m *Pay_Success) Reset()      { *m = Pay_Success{} }
func (*Pay_Success) ProtoMessage() {}
func (*Pay_Success) Descriptor() ([]byte, []int) {
	return fileDescriptor_56bb10b4b524e10d, []int{0, 0}
}
func (m *Pay_Success) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Pay_Success) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Pay_Success.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Pay_Success) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Pay_Success.Merge(m, src)
}
func (m *Pay_Success) XXX_Size() int {
	return m.Size()
}
func (m *Pay_Success) XXX_DiscardUnknown() {
	xxx_messageInfo_Pay_Success.DiscardUnknown(m)
}

var xxx_messageInfo_Pay_Success proto.InternalMessageInfo

func (m *Pay_Success) GetPcId() int64 {
	if m != nil {
		return m.PcId
	}
	return 0
}

func (m *Pay_Success) GetPaidTime() int64 {
	if m != nil {
		return m.PaidTime
	}
	return 0
}

func (m *Pay_Success) GetExpireTime() int64 {
	if m != nil {
		return m.ExpireTime
	}
	return 0
}

func (*Pay_Success) XXX_MessageName() string {
	return "service.Pay.Success"
}

//---------------------- push ----------------------//
// ?????????????????????????????????
type Pay_NormalSuccessPush struct {
	PcId         int64           `protobuf:"varint,1,opt,name=pc_id,json=pcId,proto3" json:"pc_id,omitempty"`
	ShopNormalId int64           `protobuf:"varint,2,opt,name=shop_normal_id,json=shopNormalId,proto3" json:"shop_normal_id,omitempty"`
	Items        map[int64]int64 `protobuf:"bytes,3,rep,name=items,proto3" json:"items,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

func (m *Pay_NormalSuccessPush) Reset()      { *m = Pay_NormalSuccessPush{} }
func (*Pay_NormalSuccessPush) ProtoMessage() {}
func (*Pay_NormalSuccessPush) Descriptor() ([]byte, []int) {
	return fileDescriptor_56bb10b4b524e10d, []int{0, 1}
}
func (m *Pay_NormalSuccessPush) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Pay_NormalSuccessPush) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Pay_NormalSuccessPush.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Pay_NormalSuccessPush) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Pay_NormalSuccessPush.Merge(m, src)
}
func (m *Pay_NormalSuccessPush) XXX_Size() int {
	return m.Size()
}
func (m *Pay_NormalSuccessPush) XXX_DiscardUnknown() {
	xxx_messageInfo_Pay_NormalSuccessPush.DiscardUnknown(m)
}

var xxx_messageInfo_Pay_NormalSuccessPush proto.InternalMessageInfo

func (m *Pay_NormalSuccessPush) GetPcId() int64 {
	if m != nil {
		return m.PcId
	}
	return 0
}

func (m *Pay_NormalSuccessPush) GetShopNormalId() int64 {
	if m != nil {
		return m.ShopNormalId
	}
	return 0
}

func (m *Pay_NormalSuccessPush) GetItems() map[int64]int64 {
	if m != nil {
		return m.Items
	}
	return nil
}

func (*Pay_NormalSuccessPush) XXX_MessageName() string {
	return "service.Pay.NormalSuccessPush"
}
func init() {
	proto.RegisterType((*Pay)(nil), "service.Pay")
	proto.RegisterType((*Pay_Success)(nil), "service.Pay.Success")
	proto.RegisterType((*Pay_NormalSuccessPush)(nil), "service.Pay.NormalSuccessPush")
	proto.RegisterMapType((map[int64]int64)(nil), "service.Pay.NormalSuccessPush.ItemsEntry")
}

func init() { proto.RegisterFile("proto/service/pay.proto", fileDescriptor_56bb10b4b524e10d) }

var fileDescriptor_56bb10b4b524e10d = []byte{
	// 336 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x91, 0x41, 0x4b, 0xfb, 0x30,
	0x18, 0xc6, 0x9b, 0xf5, 0xbf, 0xff, 0xf4, 0x9d, 0x88, 0x56, 0xc1, 0x51, 0x21, 0x0e, 0xf1, 0x30,
	0x0f, 0xb6, 0xa0, 0x97, 0xe1, 0x45, 0x10, 0x3c, 0xec, 0x32, 0xc6, 0xf4, 0x24, 0xc2, 0xc8, 0xd2,
	0xc0, 0x82, 0x4b, 0x13, 0x9a, 0x6e, 0xd8, 0x6f, 0xe1, 0xc7, 0xf0, 0xe6, 0xd7, 0x98, 0xb7, 0x1d,
	0x77, 0xd4, 0xf6, 0xe2, 0xd1, 0x8f, 0x20, 0x69, 0x2a, 0x22, 0x7a, 0xcb, 0xfb, 0x7b, 0x9e, 0xe4,
	0x79, 0x1f, 0x02, 0x7b, 0x2a, 0x91, 0xa9, 0x0c, 0x35, 0x4b, 0xe6, 0x9c, 0xb2, 0x50, 0x91, 0x2c,
	0x28, 0x89, 0xd7, 0xa8, 0x90, 0x5f, 0x39, 0x84, 0x8c, 0xd8, 0x54, 0x87, 0x63, 0xa2, 0x99, 0x75,
	0x1c, 0x3e, 0xd7, 0xc0, 0x1d, 0x90, 0xcc, 0xbf, 0x83, 0xc6, 0xf5, 0x8c, 0x52, 0xa6, 0xb5, 0xb7,
	0x03, 0x75, 0x45, 0x47, 0x3c, 0x6a, 0xa1, 0x36, 0xea, 0xb8, 0xc3, 0x7f, 0x8a, 0xf6, 0x22, 0x6f,
	0x1f, 0xd6, 0x15, 0xe1, 0xd1, 0x28, 0xe5, 0x82, 0xb5, 0x6a, 0xa5, 0xb0, 0x66, 0xc0, 0x0d, 0x17,
	0xcc, 0x3b, 0x80, 0x26, 0x7b, 0x50, 0x3c, 0x61, 0x56, 0x76, 0x4b, 0x19, 0x2c, 0x32, 0x06, 0xff,
	0x05, 0xc1, 0x76, 0x5f, 0x26, 0x82, 0x4c, 0xab, 0x90, 0xc1, 0x4c, 0x4f, 0xfe, 0x0e, 0x3a, 0x82,
	0x4d, 0x3d, 0x91, 0x6a, 0x14, 0x97, 0x76, 0xa3, 0xda, 0xb4, 0x0d, 0x43, 0xed, 0x1b, 0xbd, 0xc8,
	0xbb, 0x80, 0x3a, 0x4f, 0x99, 0xd0, 0x2d, 0xb7, 0xed, 0x76, 0x9a, 0xa7, 0xc7, 0x41, 0x55, 0x34,
	0x18, 0x90, 0x2c, 0xf8, 0x95, 0x14, 0xf4, 0x8c, 0xf7, 0x2a, 0x4e, 0x93, 0x6c, 0x68, 0xef, 0xf9,
	0x5d, 0x80, 0x6f, 0xe8, 0x6d, 0x81, 0x7b, 0xcf, 0xb2, 0x6a, 0x0f, 0x73, 0xf4, 0x76, 0xa1, 0x3e,
	0x27, 0xd3, 0xd9, 0x57, 0x57, 0x3b, 0x9c, 0xd7, 0xba, 0xe8, 0xb2, 0xbf, 0x7a, 0xc3, 0xce, 0x53,
	0x8e, 0xd1, 0x22, 0xc7, 0x68, 0x99, 0x63, 0xf4, 0x9a, 0x63, 0xf4, 0x9e, 0x63, 0xe7, 0x23, 0xc7,
	0xe8, 0xb1, 0xc0, 0xce, 0xa2, 0xc0, 0x68, 0x59, 0x60, 0x67, 0x55, 0x60, 0xe7, 0xb6, 0x4d, 0x25,
	0x8f, 0x4f, 0xcc, 0x76, 0x2c, 0x09, 0xa9, 0x14, 0x42, 0xc6, 0xe1, 0x8f, 0xdf, 0x1a, 0xff, 0x2f,
	0xc7, 0xb3, 0xcf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x03, 0xdf, 0xa3, 0xde, 0xc5, 0x01, 0x00, 0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolPay.Get().(proto.Message)
	})
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolPay_Success.Get().(proto.Message)
	})
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolPay_NormalSuccessPush.Get().(proto.Message)
	})
}

var poolPay = &sync.Pool{New: func() interface{} { return &Pay{} }}

func (m *Pay) ReleasePool() { m.Reset(); poolPay.Put(m); m = nil }

var poolPay_Success = &sync.Pool{New: func() interface{} { return &Pay_Success{} }}

func (m *Pay_Success) ReleasePool() { m.Reset(); poolPay_Success.Put(m); m = nil }

var poolPay_NormalSuccessPush = &sync.Pool{New: func() interface{} { return &Pay_NormalSuccessPush{} }}

func (m *Pay_NormalSuccessPush) ReleasePool() { m.Reset(); poolPay_NormalSuccessPush.Put(m); m = nil }
func (this *Pay) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Pay)
	if !ok {
		that2, ok := that.(Pay)
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
	return true
}
func (this *Pay_Success) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Pay_Success)
	if !ok {
		that2, ok := that.(Pay_Success)
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
	if this.PcId != that1.PcId {
		return false
	}
	if this.PaidTime != that1.PaidTime {
		return false
	}
	if this.ExpireTime != that1.ExpireTime {
		return false
	}
	return true
}
func (this *Pay_NormalSuccessPush) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Pay_NormalSuccessPush)
	if !ok {
		that2, ok := that.(Pay_NormalSuccessPush)
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
	if this.PcId != that1.PcId {
		return false
	}
	if this.ShopNormalId != that1.ShopNormalId {
		return false
	}
	if len(this.Items) != len(that1.Items) {
		return false
	}
	for i := range this.Items {
		if this.Items[i] != that1.Items[i] {
			return false
		}
	}
	return true
}
func (m *Pay) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Pay) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Pay) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *Pay_Success) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Pay_Success) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Pay_Success) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ExpireTime != 0 {
		i = encodeVarintPay(dAtA, i, uint64(m.ExpireTime))
		i--
		dAtA[i] = 0x18
	}
	if m.PaidTime != 0 {
		i = encodeVarintPay(dAtA, i, uint64(m.PaidTime))
		i--
		dAtA[i] = 0x10
	}
	if m.PcId != 0 {
		i = encodeVarintPay(dAtA, i, uint64(m.PcId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Pay_NormalSuccessPush) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Pay_NormalSuccessPush) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Pay_NormalSuccessPush) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Items) > 0 {
		for k := range m.Items {
			v := m.Items[k]
			baseI := i
			i = encodeVarintPay(dAtA, i, uint64(v))
			i--
			dAtA[i] = 0x10
			i = encodeVarintPay(dAtA, i, uint64(k))
			i--
			dAtA[i] = 0x8
			i = encodeVarintPay(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x1a
		}
	}
	if m.ShopNormalId != 0 {
		i = encodeVarintPay(dAtA, i, uint64(m.ShopNormalId))
		i--
		dAtA[i] = 0x10
	}
	if m.PcId != 0 {
		i = encodeVarintPay(dAtA, i, uint64(m.PcId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintPay(dAtA []byte, offset int, v uint64) int {
	offset -= sovPay(v)
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

func (m *Pay) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	w.RawByte('}')

}

func (m *Pay_Success) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	if m.PcId != 0 {
		w.RawByte('"')
		w.RawString("pc_id")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.PcId))
		needWriteComma = true
	}
	if m.PaidTime != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("paid_time")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.PaidTime))
		needWriteComma = true
	}
	if m.ExpireTime != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("expire_time")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.ExpireTime))
		needWriteComma = true
	}
	_ = needWriteComma
	w.RawByte('}')

}

func (m *Pay_NormalSuccessPush) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	if m.PcId != 0 {
		w.RawByte('"')
		w.RawString("pc_id")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.PcId))
		needWriteComma = true
	}
	if m.ShopNormalId != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("shop_normal_id")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.ShopNormalId))
		needWriteComma = true
	}
	if needWriteComma {
		w.RawByte(',')
	}
	w.RawByte('"')
	w.RawString("items")
	w.RawByte('"')
	w.RawByte(':')
	if m.Items == nil {
		w.RawString("null")
	} else if len(m.Items) == 0 {
		w.RawString("{}")
	} else {
		w.RawByte('{')
		mlItems := len(m.Items)
		for k, v := range m.Items {
			w.RawByte('"')
			w.Int64(int64(k))
			w.RawByte('"')
			w.RawByte(':')
			w.Int64(int64(v))
			mlItems--
			if mlItems != 0 {
				w.RawByte(',')
			}
		}
		w.RawByte('}')
	}
	needWriteComma = true
	_ = needWriteComma
	w.RawByte('}')

}

func (m *Pay) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *Pay) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *Pay) GoString() string {
	return m.String()
}

func (m *Pay_Success) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *Pay_Success) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *Pay_Success) GoString() string {
	return m.String()
}

func (m *Pay_NormalSuccessPush) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *Pay_NormalSuccessPush) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *Pay_NormalSuccessPush) GoString() string {
	return m.String()
}

func (m *Pay) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *Pay_Success) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PcId != 0 {
		n += 1 + sovPay(uint64(m.PcId))
	}
	if m.PaidTime != 0 {
		n += 1 + sovPay(uint64(m.PaidTime))
	}
	if m.ExpireTime != 0 {
		n += 1 + sovPay(uint64(m.ExpireTime))
	}
	return n
}

func (m *Pay_NormalSuccessPush) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PcId != 0 {
		n += 1 + sovPay(uint64(m.PcId))
	}
	if m.ShopNormalId != 0 {
		n += 1 + sovPay(uint64(m.ShopNormalId))
	}
	if len(m.Items) > 0 {
		for k, v := range m.Items {
			_ = k
			_ = v
			mapEntrySize := 1 + sovPay(uint64(k)) + 1 + sovPay(uint64(v))
			n += mapEntrySize + 1 + sovPay(uint64(mapEntrySize))
		}
	}
	return n
}

func sovPay(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPay(x uint64) (n int) {
	return sovPay(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Pay) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPay
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
			return fmt.Errorf("proto: Pay: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Pay: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipPay(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPay
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
func (m *Pay_Success) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPay
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
			return fmt.Errorf("proto: Success: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Success: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PcId", wireType)
			}
			m.PcId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPay
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PcId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PaidTime", wireType)
			}
			m.PaidTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPay
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PaidTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExpireTime", wireType)
			}
			m.ExpireTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPay
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExpireTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipPay(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPay
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
func (m *Pay_NormalSuccessPush) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPay
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
			return fmt.Errorf("proto: NormalSuccessPush: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NormalSuccessPush: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PcId", wireType)
			}
			m.PcId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPay
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PcId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ShopNormalId", wireType)
			}
			m.ShopNormalId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPay
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ShopNormalId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Items", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPay
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
				return ErrInvalidLengthPay
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPay
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Items == nil {
				m.Items = make(map[int64]int64)
			}
			var mapkey int64
			var mapvalue int64
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowPay
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
							return ErrIntOverflowPay
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
							return ErrIntOverflowPay
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
					skippy, err := skipPay(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return ErrInvalidLengthPay
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Items[mapkey] = mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPay(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPay
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
func skipPay(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPay
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
					return 0, ErrIntOverflowPay
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
					return 0, ErrIntOverflowPay
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
				return 0, ErrInvalidLengthPay
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPay
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPay
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPay        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPay          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPay = fmt.Errorf("proto: unexpected end of group")
)
