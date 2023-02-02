// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/dao/divination.proto

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

type Divination struct {
	RoleId         string `protobuf:"bytes,1,opt,name=role_id,json=roleId,proto3" json:"role_id,omitempty" pk`
	TotalCount     int64  `protobuf:"varint,2,opt,name=total_count,json=totalCount,proto3" json:"total_count,omitempty"`
	AvailableCount int64  `protobuf:"varint,3,opt,name=available_count,json=availableCount,proto3" json:"available_count,omitempty"`
	ResetAt        int64  `protobuf:"varint,4,opt,name=reset_at,json=resetAt,proto3" json:"reset_at,omitempty"`
}

func (m *Divination) Reset()      { *m = Divination{} }
func (*Divination) ProtoMessage() {}
func (*Divination) Descriptor() ([]byte, []int) {
	return fileDescriptor_eaf9ab90e7fd38f3, []int{0}
}
func (m *Divination) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Divination) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Divination.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Divination) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Divination.Merge(m, src)
}
func (m *Divination) XXX_Size() int {
	return m.Size()
}
func (m *Divination) XXX_DiscardUnknown() {
	xxx_messageInfo_Divination.DiscardUnknown(m)
}

var xxx_messageInfo_Divination proto.InternalMessageInfo

func (m *Divination) GetRoleId() string {
	if m != nil {
		return m.RoleId
	}
	return ""
}

func (m *Divination) GetTotalCount() int64 {
	if m != nil {
		return m.TotalCount
	}
	return 0
}

func (m *Divination) GetAvailableCount() int64 {
	if m != nil {
		return m.AvailableCount
	}
	return 0
}

func (m *Divination) GetResetAt() int64 {
	if m != nil {
		return m.ResetAt
	}
	return 0
}

func (*Divination) XXX_MessageName() string {
	return "dao.Divination"
}
func init() {
	proto.RegisterType((*Divination)(nil), "dao.Divination")
}

func init() { proto.RegisterFile("proto/dao/divination.proto", fileDescriptor_eaf9ab90e7fd38f3) }

var fileDescriptor_eaf9ab90e7fd38f3 = []byte{
	// 251 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0xd0, 0x31, 0x4e, 0xc3, 0x30,
	0x14, 0xc6, 0x71, 0xbf, 0x06, 0xa5, 0x60, 0x24, 0x90, 0x22, 0x86, 0x50, 0xa1, 0x97, 0x8a, 0x85,
	0x2e, 0x90, 0x81, 0x13, 0x50, 0x58, 0x60, 0xec, 0xc8, 0x12, 0xb9, 0xb1, 0x55, 0x59, 0xa4, 0x7e,
	0x95, 0x6b, 0x72, 0x0e, 0xc4, 0x29, 0x38, 0x4a, 0xc7, 0x8e, 0x9d, 0x10, 0x38, 0x0b, 0x23, 0xe2,
	0x04, 0x28, 0x4e, 0x95, 0xed, 0xf3, 0xdf, 0xbf, 0xe9, 0xf1, 0xd1, 0xca, 0x92, 0xa3, 0x5c, 0x0a,
	0xca, 0xa5, 0xae, 0xb5, 0x11, 0x4e, 0x93, 0xb9, 0x09, 0x31, 0x89, 0xa4, 0xa0, 0xd1, 0xd9, 0x82,
	0x16, 0xd4, 0xa1, 0x76, 0x75, 0x5f, 0x97, 0xef, 0xc0, 0xf9, 0x43, 0xef, 0x93, 0x8c, 0x0f, 0x2d,
	0x55, 0xaa, 0xd0, 0x32, 0x85, 0x31, 0x4c, 0x8e, 0xa6, 0xf1, 0xdf, 0x67, 0x36, 0x58, 0xbd, 0xcc,
	0xe2, 0x36, 0x3f, 0xca, 0x24, 0xe3, 0xc7, 0x8e, 0x9c, 0xa8, 0x8a, 0x92, 0x5e, 0x8d, 0x4b, 0x07,
	0x63, 0x98, 0x44, 0x33, 0x1e, 0xd2, 0x7d, 0x5b, 0x92, 0x2b, 0x7e, 0x2a, 0x6a, 0xa1, 0x2b, 0x31,
	0xaf, 0xd4, 0x1e, 0x45, 0x01, 0x9d, 0xf4, 0xb9, 0x83, 0xe7, 0xfc, 0xd0, 0xaa, 0xb5, 0x72, 0x85,
	0x70, 0xe9, 0x41, 0x10, 0xc3, 0xf0, 0xbe, 0x73, 0xd3, 0xa7, 0xdd, 0x37, 0xb2, 0x0f, 0x8f, 0xb0,
	0xf1, 0x08, 0x5b, 0x8f, 0xf0, 0xe5, 0x11, 0x7e, 0x3c, 0xb2, 0x5f, 0x8f, 0xf0, 0xd6, 0x20, 0xdb,
	0x34, 0x08, 0xdb, 0x06, 0xd9, 0xae, 0x41, 0xf6, 0x7c, 0x51, 0x92, 0x36, 0xd7, 0x6b, 0x65, 0x6b,
	0x65, 0xf3, 0x92, 0x96, 0x4b, 0x32, 0x79, 0x7f, 0x8c, 0x79, 0x1c, 0xe6, 0xed, 0x7f, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x62, 0x6b, 0xfc, 0x85, 0x20, 0x01, 0x00, 0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolDivination.Get().(proto.Message)
	})
}

var poolDivination = &sync.Pool{New: func() interface{} { return &Divination{} }}

func (m *Divination) ReleasePool() { m.Reset(); poolDivination.Put(m); m = nil }

func (m *Divination) PK() string {
	if m == nil {
		return ""
	}
	return m.RoleId
}

func (m *Divination) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return append(d, m.RoleId...)
}

func (m *Divination) ToKVSave() ([]byte, []byte) {
	msgName := m.XXX_MessageName()
	dk := coin_server_common_bytespool.GetSample(64)
	dk = dk[:0]
	dk = append(dk, msgName...)
	dk = append(dk, ':', 'k', ':')
	dk = m.PKAppendTo(dk)
	return dk, m.ToSave()
}

func (m *Divination) ToSave() []byte {
	msgName := m.XXX_MessageName()
	ml := len(msgName)
	d := coin_server_common_bytespool.GetSample(1 + ml + m.Size())
	d[0] = uint8(ml)
	copy(d[1:], msgName)
	_, _ = m.MarshalToSizedBuffer(d[1+ml:])
	return d
}

func (m *Divination) KVKey() string {
	return m.XXX_MessageName() + ":k:" + m.PK()
}

func (this *Divination) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Divination)
	if !ok {
		that2, ok := that.(Divination)
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
	if this.TotalCount != that1.TotalCount {
		return false
	}
	if this.AvailableCount != that1.AvailableCount {
		return false
	}
	if this.ResetAt != that1.ResetAt {
		return false
	}
	return true
}
func (m *Divination) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Divination) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Divination) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ResetAt != 0 {
		i = encodeVarintDivination(dAtA, i, uint64(m.ResetAt))
		i--
		dAtA[i] = 0x20
	}
	if m.AvailableCount != 0 {
		i = encodeVarintDivination(dAtA, i, uint64(m.AvailableCount))
		i--
		dAtA[i] = 0x18
	}
	if m.TotalCount != 0 {
		i = encodeVarintDivination(dAtA, i, uint64(m.TotalCount))
		i--
		dAtA[i] = 0x10
	}
	if len(m.RoleId) > 0 {
		i -= len(m.RoleId)
		copy(dAtA[i:], m.RoleId)
		i = encodeVarintDivination(dAtA, i, uint64(len(m.RoleId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintDivination(dAtA []byte, offset int, v uint64) int {
	offset -= sovDivination(v)
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

func (m *Divination) JsonBytes(w *coin_server_common_jwriter.Writer) {
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
	if m.TotalCount != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("total_count")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.TotalCount))
		needWriteComma = true
	}
	if m.AvailableCount != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("available_count")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.AvailableCount))
		needWriteComma = true
	}
	if m.ResetAt != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("reset_at")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.ResetAt))
		needWriteComma = true
	}
	_ = needWriteComma
	w.RawByte('}')

}

func (m *Divination) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *Divination) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *Divination) GoString() string {
	return m.String()
}

func (m *Divination) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RoleId)
	if l > 0 {
		n += 1 + l + sovDivination(uint64(l))
	}
	if m.TotalCount != 0 {
		n += 1 + sovDivination(uint64(m.TotalCount))
	}
	if m.AvailableCount != 0 {
		n += 1 + sovDivination(uint64(m.AvailableCount))
	}
	if m.ResetAt != 0 {
		n += 1 + sovDivination(uint64(m.ResetAt))
	}
	return n
}

func sovDivination(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozDivination(x uint64) (n int) {
	return sovDivination(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Divination) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDivination
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
			return fmt.Errorf("proto: Divination: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Divination: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RoleId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDivination
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
				return ErrInvalidLengthDivination
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthDivination
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RoleId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalCount", wireType)
			}
			m.TotalCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDivination
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TotalCount |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AvailableCount", wireType)
			}
			m.AvailableCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDivination
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AvailableCount |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ResetAt", wireType)
			}
			m.ResetAt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDivination
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ResetAt |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipDivination(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthDivination
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
func skipDivination(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowDivination
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
					return 0, ErrIntOverflowDivination
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
					return 0, ErrIntOverflowDivination
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
				return 0, ErrInvalidLengthDivination
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupDivination
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthDivination
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthDivination        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowDivination          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupDivination = fmt.Errorf("proto: unexpected end of group")
)