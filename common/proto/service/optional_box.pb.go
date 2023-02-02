// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/service/optional_box.proto

package service

import (
	coin_server_common_buffer "coin-server/common/buffer"
	coin_server_common_jwriter "coin-server/common/jwriter"
	coin_server_common_msgcreate "coin-server/common/msgcreate"
	coin_server_common_proto_jsonany "coin-server/common/proto/jsonany"
	models "coin-server/common/proto/models"
	fmt "fmt"
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

type OptionalBoxErrorCode int32

const (
	OptionalBoxErrorCode_ErrOptionalBoxExist        OptionalBoxErrorCode = 0
	OptionalBoxErrorCode_ErrOptionalBoxType         OptionalBoxErrorCode = 1
	OptionalBoxErrorCode_ErrOptionalBoxNoSelect     OptionalBoxErrorCode = 2
	OptionalBoxErrorCode_ErrOptionalBoxNotEnough    OptionalBoxErrorCode = 3
	OptionalBoxErrorCode_ErrOptionalBoxSelect       OptionalBoxErrorCode = 4
	OptionalBoxErrorCode_ErrOptionalBoxNoSelectItem OptionalBoxErrorCode = 5
	OptionalBoxErrorCode_ErrOptionalBoxRandom       OptionalBoxErrorCode = 6
	OptionalBoxErrorCode_ErrOptionalBoxCnf          OptionalBoxErrorCode = 7
)

var OptionalBoxErrorCode_name = map[int32]string{
	0: "ErrOptionalBoxExist",
	1: "ErrOptionalBoxType",
	2: "ErrOptionalBoxNoSelect",
	3: "ErrOptionalBoxNotEnough",
	4: "ErrOptionalBoxSelect",
	5: "ErrOptionalBoxNoSelectItem",
	6: "ErrOptionalBoxRandom",
	7: "ErrOptionalBoxCnf",
}

var OptionalBoxErrorCode_value = map[string]int32{
	"ErrOptionalBoxExist":        0,
	"ErrOptionalBoxType":         1,
	"ErrOptionalBoxNoSelect":     2,
	"ErrOptionalBoxNotEnough":    3,
	"ErrOptionalBoxSelect":       4,
	"ErrOptionalBoxNoSelectItem": 5,
	"ErrOptionalBoxRandom":       6,
	"ErrOptionalBoxCnf":          7,
}

func (OptionalBoxErrorCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_607ecfd417707e5b, []int{0}
}

type OptionalBox struct {
}

func (m *OptionalBox) Reset()      { *m = OptionalBox{} }
func (*OptionalBox) ProtoMessage() {}
func (*OptionalBox) Descriptor() ([]byte, []int) {
	return fileDescriptor_607ecfd417707e5b, []int{0}
}
func (m *OptionalBox) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OptionalBox) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OptionalBox.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OptionalBox) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OptionalBox.Merge(m, src)
}
func (m *OptionalBox) XXX_Size() int {
	return m.Size()
}
func (m *OptionalBox) XXX_DiscardUnknown() {
	xxx_messageInfo_OptionalBox.DiscardUnknown(m)
}

var xxx_messageInfo_OptionalBox proto.InternalMessageInfo

func (*OptionalBox) XXX_MessageName() string {
	return "service.OptionalBox"
}

type OptionalBox_OpenRequest struct {
	Typ        models.OptionalBoxType `protobuf:"varint,1,opt,name=typ,proto3,enum=models.OptionalBoxType" json:"typ,omitempty"`
	Item       *models.Item           `protobuf:"bytes,2,opt,name=item,proto3" json:"item,omitempty"`
	SelectItem []*models.Item         `protobuf:"bytes,3,rep,name=select_item,json=selectItem,proto3" json:"select_item,omitempty"`
}

func (m *OptionalBox_OpenRequest) Reset()      { *m = OptionalBox_OpenRequest{} }
func (*OptionalBox_OpenRequest) ProtoMessage() {}
func (*OptionalBox_OpenRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_607ecfd417707e5b, []int{0, 0}
}
func (m *OptionalBox_OpenRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OptionalBox_OpenRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OptionalBox_OpenRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OptionalBox_OpenRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OptionalBox_OpenRequest.Merge(m, src)
}
func (m *OptionalBox_OpenRequest) XXX_Size() int {
	return m.Size()
}
func (m *OptionalBox_OpenRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_OptionalBox_OpenRequest.DiscardUnknown(m)
}

var xxx_messageInfo_OptionalBox_OpenRequest proto.InternalMessageInfo

func (m *OptionalBox_OpenRequest) GetTyp() models.OptionalBoxType {
	if m != nil {
		return m.Typ
	}
	return models.OptionalBoxType_OptionalBox_None
}

func (m *OptionalBox_OpenRequest) GetItem() *models.Item {
	if m != nil {
		return m.Item
	}
	return nil
}

func (m *OptionalBox_OpenRequest) GetSelectItem() []*models.Item {
	if m != nil {
		return m.SelectItem
	}
	return nil
}

func (*OptionalBox_OpenRequest) XXX_MessageName() string {
	return "service.OptionalBox.OpenRequest"
}

type OptionalBox_OpenResponse struct {
	Items []*models.Item `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (m *OptionalBox_OpenResponse) Reset()      { *m = OptionalBox_OpenResponse{} }
func (*OptionalBox_OpenResponse) ProtoMessage() {}
func (*OptionalBox_OpenResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_607ecfd417707e5b, []int{0, 1}
}
func (m *OptionalBox_OpenResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OptionalBox_OpenResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OptionalBox_OpenResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OptionalBox_OpenResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OptionalBox_OpenResponse.Merge(m, src)
}
func (m *OptionalBox_OpenResponse) XXX_Size() int {
	return m.Size()
}
func (m *OptionalBox_OpenResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_OptionalBox_OpenResponse.DiscardUnknown(m)
}

var xxx_messageInfo_OptionalBox_OpenResponse proto.InternalMessageInfo

func (m *OptionalBox_OpenResponse) GetItems() []*models.Item {
	if m != nil {
		return m.Items
	}
	return nil
}

func (*OptionalBox_OpenResponse) XXX_MessageName() string {
	return "service.OptionalBox.OpenResponse"
}
func init() {
	proto.RegisterEnum("service.OptionalBoxErrorCode", OptionalBoxErrorCode_name, OptionalBoxErrorCode_value)
	proto.RegisterType((*OptionalBox)(nil), "service.OptionalBox")
	proto.RegisterType((*OptionalBox_OpenRequest)(nil), "service.OptionalBox.OpenRequest")
	proto.RegisterType((*OptionalBox_OpenResponse)(nil), "service.OptionalBox.OpenResponse")
}

func init() { proto.RegisterFile("proto/service/optional_box.proto", fileDescriptor_607ecfd417707e5b) }

var fileDescriptor_607ecfd417707e5b = []byte{
	// 552 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x93, 0xcd, 0x6b, 0x13, 0x4f,
	0x18, 0xc7, 0x77, 0x9a, 0xbe, 0xc0, 0x24, 0x94, 0xfd, 0xcd, 0x2f, 0x36, 0xcb, 0x0a, 0xe3, 0xd2,
	0x53, 0x5b, 0x6d, 0xb6, 0xc4, 0x83, 0x9e, 0x5b, 0x72, 0xd0, 0x43, 0x85, 0xe8, 0xc9, 0x4b, 0xc9,
	0xcb, 0x58, 0x03, 0xd9, 0x9d, 0x75, 0x77, 0x2b, 0xe9, 0x4d, 0x0f, 0x45, 0x93, 0xba, 0x56, 0xd0,
	0x52, 0x94, 0x2a, 0x45, 0x3c, 0x58, 0xc1, 0xc4, 0x22, 0x91, 0x56, 0xf1, 0xe6, 0xa1, 0x17, 0xa1,
	0xc7, 0x1e, 0x75, 0xd3, 0xdd, 0x64, 0x6f, 0xfe, 0x09, 0xb2, 0x6f, 0xb2, 0x9b, 0xf6, 0x36, 0xf3,
	0x7c, 0x3f, 0xdf, 0xe7, 0xf9, 0x3e, 0x03, 0x03, 0x05, 0x45, 0xa5, 0x3a, 0x15, 0x35, 0xa2, 0xde,
	0xaf, 0x96, 0x89, 0x48, 0x15, 0xbd, 0x4a, 0xe5, 0x62, 0x6d, 0xa9, 0x44, 0xeb, 0x59, 0x4f, 0x42,
	0x63, 0x81, 0xc6, 0x67, 0x7c, 0x54, 0xa2, 0x15, 0x52, 0xd3, 0xc4, 0x52, 0x51, 0x23, 0x3e, 0xc1,
	0x4f, 0x0c, 0x08, 0xcb, 0x41, 0xfd, 0x42, 0xac, 0x7e, 0xba, 0xf5, 0xe4, 0x4f, 0x00, 0x93, 0x37,
	0x82, 0xf2, 0x3c, 0xad, 0xf3, 0x0d, 0xef, 0x4e, 0xe4, 0x02, 0xb9, 0xb7, 0x42, 0x34, 0x1d, 0x4d,
	0xc3, 0x84, 0xbe, 0xaa, 0x70, 0x40, 0x00, 0x53, 0xe3, 0xb9, 0x4c, 0xd6, 0x6f, 0x94, 0x8d, 0x38,
	0x6e, 0xad, 0x2a, 0xa4, 0xe0, 0x32, 0x48, 0x80, 0xc3, 0x55, 0x9d, 0x48, 0xdc, 0x90, 0x00, 0xa6,
	0x92, 0xb9, 0x54, 0xc8, 0x5e, 0xd3, 0x89, 0x54, 0xf0, 0x14, 0x34, 0x0b, 0x93, 0x1a, 0xa9, 0x91,
	0xb2, 0xbe, 0xe4, 0x81, 0x09, 0x21, 0x71, 0x0a, 0x84, 0x3e, 0xe0, 0x9e, 0xf9, 0x1c, 0x4c, 0xf9,
	0x51, 0x34, 0x85, 0xca, 0x1a, 0x41, 0x93, 0x70, 0xc4, 0xf5, 0x69, 0x1c, 0x38, 0xc3, 0xe8, 0x4b,
	0x33, 0x3f, 0x12, 0x30, 0x1d, 0x49, 0x97, 0x57, 0x55, 0xaa, 0x2e, 0xd0, 0x0a, 0x41, 0x17, 0xe1,
	0xff, 0x79, 0x55, 0x8d, 0x4a, 0xf5, 0xaa, 0xa6, 0xb3, 0x0c, 0x8f, 0x9a, 0x1d, 0x6e, 0x1c, 0xa5,
	0xec, 0xad, 0x1d, 0xfb, 0xe3, 0x3b, 0xab, 0xf5, 0xba, 0xd7, 0xda, 0x43, 0x73, 0x10, 0xc5, 0x61,
	0x77, 0x4b, 0x16, 0xf0, 0x5c, 0xb3, 0xc3, 0xa5, 0x11, 0xf2, 0xa9, 0xde, 0xde, 0x37, 0xeb, 0xcd,
	0xba, 0xf3, 0xe2, 0x65, 0xbf, 0xdd, 0x46, 0x57, 0xe0, 0x44, 0xdc, 0xb1, 0x48, 0x6f, 0x7a, 0x9b,
	0xb0, 0x43, 0xfc, 0xf9, 0x66, 0x87, 0xcb, 0xa0, 0x73, 0xf6, 0xf6, 0x07, 0xe7, 0x41, 0xc3, 0x5e,
	0xdf, 0xb5, 0xda, 0x07, 0xb6, 0xf1, 0xb6, 0xd7, 0xd8, 0xb5, 0x9e, 0x3f, 0x44, 0x22, 0xcc, 0x0c,
	0x1a, 0xf5, 0xbc, 0x4c, 0x57, 0x96, 0xef, 0xb2, 0x89, 0x30, 0x9b, 0x3f, 0xef, 0x64, 0xdf, 0xe8,
	0x7f, 0xf9, 0x8c, 0xae, 0xc3, 0x74, 0xdc, 0x10, 0xcc, 0x19, 0xe6, 0xe7, 0x9a, 0x1d, 0xee, 0x12,
	0x9a, 0xf1, 0x69, 0x7b, 0xf3, 0x93, 0xf3, 0x68, 0xe3, 0x64, 0xff, 0x69, 0x74, 0x5a, 0x58, 0x34,
	0xac, 0x27, 0x07, 0xce, 0x9a, 0x81, 0xae, 0x42, 0xfe, 0xec, 0xd4, 0xee, 0x93, 0xb2, 0x23, 0xff,
	0xf6, 0x8d, 0xb6, 0xd8, 0xda, 0xb1, 0x37, 0x1f, 0xa3, 0xdc, 0x60, 0x8a, 0x42, 0x51, 0xae, 0x50,
	0x89, 0x1d, 0x0d, 0x3d, 0xce, 0xab, 0x0d, 0x7b, 0xfb, 0xab, 0xb3, 0x66, 0xf4, 0xbe, 0xb7, 0x82,
	0x37, 0x9a, 0x86, 0xff, 0xc5, 0x3d, 0x0b, 0xf2, 0x1d, 0x76, 0x2c, 0x5c, 0x32, 0x8a, 0xce, 0x2f,
	0x1e, 0xff, 0xc6, 0xcc, 0x7b, 0x13, 0x83, 0x43, 0x13, 0x83, 0x23, 0x13, 0x83, 0x5f, 0x26, 0x06,
	0x7d, 0x13, 0x33, 0x7f, 0x4c, 0x0c, 0x9e, 0x75, 0x31, 0x73, 0xd8, 0xc5, 0xe0, 0xa8, 0x8b, 0x99,
	0xe3, 0x2e, 0x66, 0x6e, 0x0b, 0x65, 0x5a, 0x95, 0x67, 0xdd, 0x4f, 0x42, 0x54, 0xb1, 0x4c, 0x25,
	0x89, 0xca, 0x62, 0xec, 0x53, 0x95, 0x46, 0xbd, 0xeb, 0xe5, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff,
	0x3e, 0x2f, 0x5f, 0xb2, 0x6c, 0x03, 0x00, 0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolOptionalBox.Get().(proto.Message)
	})
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolOptionalBox_OpenRequest.Get().(proto.Message)
	})
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolOptionalBox_OpenResponse.Get().(proto.Message)
	})
}

var poolOptionalBox = &sync.Pool{New: func() interface{} { return &OptionalBox{} }}

func (m *OptionalBox) ReleasePool() { m.Reset(); poolOptionalBox.Put(m); m = nil }

var poolOptionalBox_OpenRequest = &sync.Pool{New: func() interface{} { return &OptionalBox_OpenRequest{} }}

func (m *OptionalBox_OpenRequest) ReleasePool() {
	m.Reset()
	poolOptionalBox_OpenRequest.Put(m)
	m = nil
}

var poolOptionalBox_OpenResponse = &sync.Pool{New: func() interface{} { return &OptionalBox_OpenResponse{} }}

func (m *OptionalBox_OpenResponse) ReleasePool() {
	m.Reset()
	poolOptionalBox_OpenResponse.Put(m)
	m = nil
}
func (x OptionalBoxErrorCode) String() string {
	s, ok := OptionalBoxErrorCode_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
func (this *OptionalBox) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*OptionalBox)
	if !ok {
		that2, ok := that.(OptionalBox)
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
func (this *OptionalBox_OpenRequest) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*OptionalBox_OpenRequest)
	if !ok {
		that2, ok := that.(OptionalBox_OpenRequest)
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
	if this.Typ != that1.Typ {
		return false
	}
	if !this.Item.Equal(that1.Item) {
		return false
	}
	if len(this.SelectItem) != len(that1.SelectItem) {
		return false
	}
	for i := range this.SelectItem {
		if !this.SelectItem[i].Equal(that1.SelectItem[i]) {
			return false
		}
	}
	return true
}
func (this *OptionalBox_OpenResponse) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*OptionalBox_OpenResponse)
	if !ok {
		that2, ok := that.(OptionalBox_OpenResponse)
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
	if len(this.Items) != len(that1.Items) {
		return false
	}
	for i := range this.Items {
		if !this.Items[i].Equal(that1.Items[i]) {
			return false
		}
	}
	return true
}
func (m *OptionalBox) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OptionalBox) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OptionalBox) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *OptionalBox_OpenRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OptionalBox_OpenRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OptionalBox_OpenRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SelectItem) > 0 {
		for iNdEx := len(m.SelectItem) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.SelectItem[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintOptionalBox(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if m.Item != nil {
		{
			size, err := m.Item.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintOptionalBox(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.Typ != 0 {
		i = encodeVarintOptionalBox(dAtA, i, uint64(m.Typ))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *OptionalBox_OpenResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OptionalBox_OpenResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OptionalBox_OpenResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Items) > 0 {
		for iNdEx := len(m.Items) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Items[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintOptionalBox(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintOptionalBox(dAtA []byte, offset int, v uint64) int {
	offset -= sovOptionalBox(v)
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

func (m *OptionalBox) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	w.RawByte('}')

}

func (m *OptionalBox_OpenRequest) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	if m.Typ != 0 {
		w.RawByte('"')
		w.RawString("typ")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.Typ))
		needWriteComma = true
	}
	if needWriteComma {
		w.RawByte(',')
	}
	w.RawByte('"')
	w.RawString("item")
	w.RawByte('"')
	w.RawByte(':')
	m.Item.JsonBytes(w)
	needWriteComma = true
	if needWriteComma {
		w.RawByte(',')
	}
	w.RawByte('"')
	w.RawString("select_item")
	w.RawByte('"')
	w.RawByte(':')
	if m.SelectItem == nil {
		w.RawString("null")
	} else if len(m.SelectItem) == 0 {
		w.RawString("[]")
	} else {
		w.RawByte('[')
		for i, v := range m.SelectItem {
			v.JsonBytes(w)
			if i != len(m.SelectItem)-1 {
				w.RawByte(',')
			}
		}
		w.RawByte(']')
	}
	needWriteComma = true
	_ = needWriteComma
	w.RawByte('}')

}

func (m *OptionalBox_OpenResponse) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	w.RawByte('"')
	w.RawString("items")
	w.RawByte('"')
	w.RawByte(':')
	if m.Items == nil {
		w.RawString("null")
	} else if len(m.Items) == 0 {
		w.RawString("[]")
	} else {
		w.RawByte('[')
		for i, v := range m.Items {
			v.JsonBytes(w)
			if i != len(m.Items)-1 {
				w.RawByte(',')
			}
		}
		w.RawByte(']')
	}
	needWriteComma = true
	_ = needWriteComma
	w.RawByte('}')

}

func (m *OptionalBox) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *OptionalBox) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *OptionalBox) GoString() string {
	return m.String()
}

func (m *OptionalBox_OpenRequest) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *OptionalBox_OpenRequest) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *OptionalBox_OpenRequest) GoString() string {
	return m.String()
}

func (m *OptionalBox_OpenResponse) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *OptionalBox_OpenResponse) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *OptionalBox_OpenResponse) GoString() string {
	return m.String()
}

func (m *OptionalBox) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *OptionalBox_OpenRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Typ != 0 {
		n += 1 + sovOptionalBox(uint64(m.Typ))
	}
	if m.Item != nil {
		l = m.Item.Size()
		n += 1 + l + sovOptionalBox(uint64(l))
	}
	if len(m.SelectItem) > 0 {
		for _, e := range m.SelectItem {
			l = e.Size()
			n += 1 + l + sovOptionalBox(uint64(l))
		}
	}
	return n
}

func (m *OptionalBox_OpenResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Items) > 0 {
		for _, e := range m.Items {
			l = e.Size()
			n += 1 + l + sovOptionalBox(uint64(l))
		}
	}
	return n
}

func sovOptionalBox(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozOptionalBox(x uint64) (n int) {
	return sovOptionalBox(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *OptionalBox) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOptionalBox
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
			return fmt.Errorf("proto: OptionalBox: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OptionalBox: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipOptionalBox(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthOptionalBox
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
func (m *OptionalBox_OpenRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOptionalBox
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
			return fmt.Errorf("proto: OpenRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OpenRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Typ", wireType)
			}
			m.Typ = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOptionalBox
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Typ |= models.OptionalBoxType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Item", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOptionalBox
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
				return ErrInvalidLengthOptionalBox
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthOptionalBox
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Item == nil {
				m.Item = &models.Item{}
			}
			if err := m.Item.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SelectItem", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOptionalBox
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
				return ErrInvalidLengthOptionalBox
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthOptionalBox
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SelectItem = append(m.SelectItem, &models.Item{})
			if err := m.SelectItem[len(m.SelectItem)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOptionalBox(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthOptionalBox
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
func (m *OptionalBox_OpenResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOptionalBox
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
			return fmt.Errorf("proto: OpenResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OpenResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Items", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOptionalBox
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
				return ErrInvalidLengthOptionalBox
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthOptionalBox
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Items = append(m.Items, &models.Item{})
			if err := m.Items[len(m.Items)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOptionalBox(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthOptionalBox
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
func skipOptionalBox(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowOptionalBox
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
					return 0, ErrIntOverflowOptionalBox
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
					return 0, ErrIntOverflowOptionalBox
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
				return 0, ErrInvalidLengthOptionalBox
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupOptionalBox
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthOptionalBox
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthOptionalBox        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowOptionalBox          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupOptionalBox = fmt.Errorf("proto: unexpected end of group")
)
