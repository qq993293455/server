// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/dao/white_list.proto

package dao

import (
	coin_server_common_buffer "coin-server/common/buffer"
	coin_server_common_bytespool "coin-server/common/bytespool"
	coin_server_common_jwriter "coin-server/common/jwriter"
	coin_server_common_msgcreate "coin-server/common/msgcreate"
	coin_server_common_proto_jsonany "coin-server/common/proto/jsonany"
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

type WhiteList struct {
	Device    string `protobuf:"bytes,1,opt,name=device,proto3" json:"device,omitempty" pk`
	Enable    bool   `protobuf:"varint,2,opt,name=enable,proto3" json:"enable,omitempty"`
	Comment   string `protobuf:"bytes,3,opt,name=comment,proto3" json:"comment,omitempty"`
	UpdatedAt int64  `protobuf:"varint,4,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

func (m *WhiteList) Reset()      { *m = WhiteList{} }
func (*WhiteList) ProtoMessage() {}
func (*WhiteList) Descriptor() ([]byte, []int) {
	return fileDescriptor_1d8113b0214d666a, []int{0}
}
func (m *WhiteList) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *WhiteList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_WhiteList.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *WhiteList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WhiteList.Merge(m, src)
}
func (m *WhiteList) XXX_Size() int {
	return m.Size()
}
func (m *WhiteList) XXX_DiscardUnknown() {
	xxx_messageInfo_WhiteList.DiscardUnknown(m)
}

var xxx_messageInfo_WhiteList proto.InternalMessageInfo

func (m *WhiteList) GetDevice() string {
	if m != nil {
		return m.Device
	}
	return ""
}

func (m *WhiteList) GetEnable() bool {
	if m != nil {
		return m.Enable
	}
	return false
}

func (m *WhiteList) GetComment() string {
	if m != nil {
		return m.Comment
	}
	return ""
}

func (m *WhiteList) GetUpdatedAt() int64 {
	if m != nil {
		return m.UpdatedAt
	}
	return 0
}

func (*WhiteList) XXX_MessageName() string {
	return "dao.WhiteList"
}
func init() {
	proto.RegisterType((*WhiteList)(nil), "dao.WhiteList")
}

func init() { proto.RegisterFile("proto/dao/white_list.proto", fileDescriptor_1d8113b0214d666a) }

var fileDescriptor_1d8113b0214d666a = []byte{
	// 241 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2a, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0x4f, 0x49, 0xcc, 0xd7, 0x2f, 0xcf, 0xc8, 0x2c, 0x49, 0x8d, 0xcf, 0xc9, 0x2c, 0x2e,
	0xd1, 0x03, 0x0b, 0x0a, 0x31, 0xa7, 0x24, 0xe6, 0x4b, 0x89, 0xa4, 0xe7, 0xa7, 0xe7, 0x43, 0x14,
	0x81, 0x58, 0x10, 0x29, 0xa5, 0x1a, 0x2e, 0xce, 0x70, 0x90, 0x72, 0x9f, 0xcc, 0xe2, 0x12, 0x21,
	0x39, 0x2e, 0xb6, 0x94, 0xd4, 0xb2, 0xcc, 0xe4, 0x54, 0x09, 0x46, 0x05, 0x46, 0x0d, 0x4e, 0x27,
	0xb6, 0x4f, 0xf7, 0xe4, 0x99, 0x0a, 0xb2, 0x83, 0xa0, 0xa2, 0x42, 0x62, 0x5c, 0x6c, 0xa9, 0x79,
	0x89, 0x49, 0x39, 0xa9, 0x12, 0x4c, 0x0a, 0x8c, 0x1a, 0x1c, 0x41, 0x50, 0x9e, 0x90, 0x04, 0x17,
	0x7b, 0x72, 0x7e, 0x6e, 0x6e, 0x6a, 0x5e, 0x89, 0x04, 0x33, 0x48, 0x63, 0x10, 0x8c, 0x2b, 0x24,
	0xcb, 0xc5, 0x55, 0x5a, 0x90, 0x92, 0x58, 0x92, 0x9a, 0x12, 0x9f, 0x58, 0x22, 0xc1, 0xa2, 0xc0,
	0xa8, 0xc1, 0x1c, 0xc4, 0x09, 0x15, 0x71, 0x2c, 0x71, 0xf2, 0xba, 0xf1, 0x50, 0x8e, 0x61, 0xc5,
	0x23, 0x39, 0xc6, 0x13, 0x8f, 0xe4, 0x18, 0x2f, 0x3c, 0x92, 0x63, 0x7c, 0xf0, 0x48, 0x8e, 0xf1,
	0xc5, 0x23, 0x39, 0x86, 0x0f, 0x8f, 0xe4, 0x18, 0x27, 0x3c, 0x96, 0x63, 0x38, 0xf1, 0x58, 0x8e,
	0xf1, 0xc2, 0x63, 0x39, 0x86, 0x1b, 0x8f, 0xe5, 0x18, 0xa2, 0x64, 0x92, 0xf3, 0x33, 0xf3, 0x74,
	0x8b, 0x53, 0x8b, 0xca, 0x52, 0x8b, 0xf4, 0x41, 0x36, 0xe4, 0xe7, 0xe9, 0xc3, 0x7d, 0x9d, 0xc4,
	0x06, 0x66, 0x1a, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x47, 0x8a, 0x1a, 0x95, 0x09, 0x01, 0x00,
	0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolWhiteList.Get().(proto.Message)
	})
}

var poolWhiteList = &sync.Pool{New: func() interface{} { return &WhiteList{} }}

func (m *WhiteList) ReleasePool() { m.Reset(); poolWhiteList.Put(m); m = nil }

func (m *WhiteList) PK() string {
	if m == nil {
		return ""
	}
	return m.Device
}

func (m *WhiteList) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return append(d, m.Device...)
}

func (m *WhiteList) ToKVSave() ([]byte, []byte) {
	msgName := m.XXX_MessageName()
	dk := coin_server_common_bytespool.GetSample(64)
	dk = dk[:0]
	dk = append(dk, msgName...)
	dk = append(dk, ':', 'k', ':')
	dk = m.PKAppendTo(dk)
	return dk, m.ToSave()
}

func (m *WhiteList) ToSave() []byte {
	msgName := m.XXX_MessageName()
	ml := len(msgName)
	d := coin_server_common_bytespool.GetSample(1 + ml + m.Size())
	d[0] = uint8(ml)
	copy(d[1:], msgName)
	_, _ = m.MarshalToSizedBuffer(d[1+ml:])
	return d
}

func (m *WhiteList) KVKey() string {
	return m.XXX_MessageName() + ":k:" + m.PK()
}

func (this *WhiteList) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*WhiteList)
	if !ok {
		that2, ok := that.(WhiteList)
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
	if this.Device != that1.Device {
		return false
	}
	if this.Enable != that1.Enable {
		return false
	}
	if this.Comment != that1.Comment {
		return false
	}
	if this.UpdatedAt != that1.UpdatedAt {
		return false
	}
	return true
}
func (m *WhiteList) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *WhiteList) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *WhiteList) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.UpdatedAt != 0 {
		i = encodeVarintWhiteList(dAtA, i, uint64(m.UpdatedAt))
		i--
		dAtA[i] = 0x20
	}
	if len(m.Comment) > 0 {
		i -= len(m.Comment)
		copy(dAtA[i:], m.Comment)
		i = encodeVarintWhiteList(dAtA, i, uint64(len(m.Comment)))
		i--
		dAtA[i] = 0x1a
	}
	if m.Enable {
		i--
		if m.Enable {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x10
	}
	if len(m.Device) > 0 {
		i -= len(m.Device)
		copy(dAtA[i:], m.Device)
		i = encodeVarintWhiteList(dAtA, i, uint64(len(m.Device)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintWhiteList(dAtA []byte, offset int, v uint64) int {
	offset -= sovWhiteList(v)
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

func (m *WhiteList) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	if m.Device != "" {
		w.RawByte('"')
		w.RawString("device")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.Device)
		needWriteComma = true
	}
	if m.Enable {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("enable")
		w.RawByte('"')
		w.RawByte(':')
		w.Bool(m.Enable)
		needWriteComma = true
	}
	if m.Comment != "" {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("comment")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.Comment)
		needWriteComma = true
	}
	if m.UpdatedAt != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("updated_at")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.UpdatedAt))
		needWriteComma = true
	}
	_ = needWriteComma
	w.RawByte('}')

}

func (m *WhiteList) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *WhiteList) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *WhiteList) GoString() string {
	return m.String()
}

func (m *WhiteList) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Device)
	if l > 0 {
		n += 1 + l + sovWhiteList(uint64(l))
	}
	if m.Enable {
		n += 2
	}
	l = len(m.Comment)
	if l > 0 {
		n += 1 + l + sovWhiteList(uint64(l))
	}
	if m.UpdatedAt != 0 {
		n += 1 + sovWhiteList(uint64(m.UpdatedAt))
	}
	return n
}

func sovWhiteList(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozWhiteList(x uint64) (n int) {
	return sovWhiteList(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *WhiteList) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowWhiteList
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
			return fmt.Errorf("proto: WhiteList: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: WhiteList: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Device", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWhiteList
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
				return ErrInvalidLengthWhiteList
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthWhiteList
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Device = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Enable", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWhiteList
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
			m.Enable = bool(v != 0)
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Comment", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWhiteList
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
				return ErrInvalidLengthWhiteList
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthWhiteList
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Comment = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdatedAt", wireType)
			}
			m.UpdatedAt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWhiteList
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdatedAt |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipWhiteList(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthWhiteList
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
func skipWhiteList(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowWhiteList
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
					return 0, ErrIntOverflowWhiteList
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
					return 0, ErrIntOverflowWhiteList
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
				return 0, ErrInvalidLengthWhiteList
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupWhiteList
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthWhiteList
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthWhiteList        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowWhiteList          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupWhiteList = fmt.Errorf("proto: unexpected end of group")
)
