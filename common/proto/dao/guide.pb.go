// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/dao/guide.proto

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

// 引导记录
type Guide struct {
	RoleId           string  `protobuf:"bytes,1,opt,name=role_id,json=roleId,proto3" json:"role_id,omitempty" pk`
	GuideId          int64   `protobuf:"varint,2,opt,name=guide_id,json=guideId,proto3" json:"guide_id,omitempty"`
	StepId           int64   `protobuf:"varint,3,opt,name=step_id,json=stepId,proto3" json:"step_id,omitempty"`
	FinishedGuideIds []int64 `protobuf:"varint,4,rep,packed,name=finished_guide_ids,json=finishedGuideIds,proto3" json:"finished_guide_ids,omitempty"`
}

func (m *Guide) Reset()      { *m = Guide{} }
func (*Guide) ProtoMessage() {}
func (*Guide) Descriptor() ([]byte, []int) {
	return fileDescriptor_26c758c7a973279b, []int{0}
}
func (m *Guide) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Guide) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Guide.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Guide) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Guide.Merge(m, src)
}
func (m *Guide) XXX_Size() int {
	return m.Size()
}
func (m *Guide) XXX_DiscardUnknown() {
	xxx_messageInfo_Guide.DiscardUnknown(m)
}

var xxx_messageInfo_Guide proto.InternalMessageInfo

func (m *Guide) GetRoleId() string {
	if m != nil {
		return m.RoleId
	}
	return ""
}

func (m *Guide) GetGuideId() int64 {
	if m != nil {
		return m.GuideId
	}
	return 0
}

func (m *Guide) GetStepId() int64 {
	if m != nil {
		return m.StepId
	}
	return 0
}

func (m *Guide) GetFinishedGuideIds() []int64 {
	if m != nil {
		return m.FinishedGuideIds
	}
	return nil
}

func (*Guide) XXX_MessageName() string {
	return "dao.Guide"
}
func init() {
	proto.RegisterType((*Guide)(nil), "dao.Guide")
}

func init() { proto.RegisterFile("proto/dao/guide.proto", fileDescriptor_26c758c7a973279b) }

var fileDescriptor_26c758c7a973279b = []byte{
	// 243 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2d, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0x4f, 0x49, 0xcc, 0xd7, 0x4f, 0x2f, 0xcd, 0x4c, 0x49, 0xd5, 0x03, 0xf3, 0x85, 0x98,
	0x53, 0x12, 0xf3, 0xa5, 0x44, 0xd2, 0xf3, 0xd3, 0xf3, 0x21, 0xf2, 0x20, 0x16, 0x44, 0x4a, 0xa9,
	0x8b, 0x91, 0x8b, 0xd5, 0x1d, 0xa4, 0x54, 0x48, 0x9e, 0x8b, 0xbd, 0x28, 0x3f, 0x27, 0x35, 0x3e,
	0x33, 0x45, 0x82, 0x51, 0x81, 0x51, 0x83, 0xd3, 0x89, 0xed, 0xd3, 0x3d, 0x79, 0xa6, 0x82, 0xec,
	0x20, 0x36, 0x90, 0xb0, 0x67, 0x8a, 0x90, 0x24, 0x17, 0x07, 0xd8, 0x50, 0x90, 0x0a, 0x26, 0x05,
	0x46, 0x0d, 0xe6, 0x20, 0x76, 0x30, 0xdf, 0x33, 0x45, 0x48, 0x9c, 0x8b, 0xbd, 0xb8, 0x24, 0xb5,
	0x00, 0x24, 0xc3, 0x0c, 0x96, 0x61, 0x03, 0x71, 0x3d, 0x53, 0x84, 0x74, 0xb8, 0x84, 0xd2, 0x32,
	0xf3, 0x32, 0x8b, 0x33, 0x52, 0x53, 0xe2, 0x61, 0x9a, 0x8b, 0x25, 0x58, 0x14, 0x98, 0x35, 0x98,
	0x83, 0x04, 0x60, 0x32, 0xee, 0x10, 0x53, 0x8a, 0x9d, 0xbc, 0x6e, 0x3c, 0x94, 0x63, 0x58, 0xf1,
	0x48, 0x8e, 0xf1, 0xc4, 0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c, 0x92, 0x63, 0x7c,
	0xf1, 0x48, 0x8e, 0xe1, 0xc3, 0x23, 0x39, 0xc6, 0x09, 0x8f, 0xe5, 0x18, 0x4e, 0x3c, 0x96, 0x63,
	0xbc, 0xf0, 0x58, 0x8e, 0xe1, 0xc6, 0x63, 0x39, 0x86, 0x28, 0x99, 0xe4, 0xfc, 0xcc, 0x3c, 0xdd,
	0xe2, 0xd4, 0xa2, 0xb2, 0xd4, 0x22, 0xfd, 0xe4, 0xfc, 0xdc, 0xdc, 0xfc, 0x3c, 0x7d, 0xb8, 0xff,
	0x93, 0xd8, 0xc0, 0x4c, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x7b, 0x07, 0x77, 0x78, 0x13,
	0x01, 0x00, 0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolGuide.Get().(proto.Message)
	})
}

var poolGuide = &sync.Pool{New: func() interface{} { return &Guide{} }}

func (m *Guide) ReleasePool() { m.Reset(); poolGuide.Put(m); m = nil }

func (m *Guide) PK() string {
	if m == nil {
		return ""
	}
	return m.RoleId
}

func (m *Guide) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return append(d, m.RoleId...)
}

func (m *Guide) ToKVSave() ([]byte, []byte) {
	msgName := m.XXX_MessageName()
	dk := coin_server_common_bytespool.GetSample(64)
	dk = dk[:0]
	dk = append(dk, msgName...)
	dk = append(dk, ':', 'k', ':')
	dk = m.PKAppendTo(dk)
	return dk, m.ToSave()
}

func (m *Guide) ToSave() []byte {
	msgName := m.XXX_MessageName()
	ml := len(msgName)
	d := coin_server_common_bytespool.GetSample(1 + ml + m.Size())
	d[0] = uint8(ml)
	copy(d[1:], msgName)
	_, _ = m.MarshalToSizedBuffer(d[1+ml:])
	return d
}

func (m *Guide) KVKey() string {
	return m.XXX_MessageName() + ":k:" + m.PK()
}

func (this *Guide) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Guide)
	if !ok {
		that2, ok := that.(Guide)
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
	if this.GuideId != that1.GuideId {
		return false
	}
	if this.StepId != that1.StepId {
		return false
	}
	if len(this.FinishedGuideIds) != len(that1.FinishedGuideIds) {
		return false
	}
	for i := range this.FinishedGuideIds {
		if this.FinishedGuideIds[i] != that1.FinishedGuideIds[i] {
			return false
		}
	}
	return true
}
func (m *Guide) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Guide) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Guide) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.FinishedGuideIds) > 0 {
		dAtA2 := make([]byte, len(m.FinishedGuideIds)*10)
		var j1 int
		for _, num1 := range m.FinishedGuideIds {
			num := uint64(num1)
			for num >= 1<<7 {
				dAtA2[j1] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j1++
			}
			dAtA2[j1] = uint8(num)
			j1++
		}
		i -= j1
		copy(dAtA[i:], dAtA2[:j1])
		i = encodeVarintGuide(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0x22
	}
	if m.StepId != 0 {
		i = encodeVarintGuide(dAtA, i, uint64(m.StepId))
		i--
		dAtA[i] = 0x18
	}
	if m.GuideId != 0 {
		i = encodeVarintGuide(dAtA, i, uint64(m.GuideId))
		i--
		dAtA[i] = 0x10
	}
	if len(m.RoleId) > 0 {
		i -= len(m.RoleId)
		copy(dAtA[i:], m.RoleId)
		i = encodeVarintGuide(dAtA, i, uint64(len(m.RoleId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGuide(dAtA []byte, offset int, v uint64) int {
	offset -= sovGuide(v)
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

func (m *Guide) JsonBytes(w *coin_server_common_jwriter.Writer) {
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
	if m.GuideId != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("guide_id")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.GuideId))
		needWriteComma = true
	}
	if m.StepId != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("step_id")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.StepId))
		needWriteComma = true
	}
	if needWriteComma {
		w.RawByte(',')
	}
	w.RawByte('"')
	w.RawString("finished_guide_ids")
	w.RawByte('"')
	w.RawByte(':')
	if m.FinishedGuideIds == nil {
		w.RawString("null")
	} else if len(m.FinishedGuideIds) == 0 {
		w.RawString("[]")
	} else {
		w.RawByte('[')
		for i, v := range m.FinishedGuideIds {
			w.Int64(int64(v))
			if i != len(m.FinishedGuideIds)-1 {
				w.RawByte(',')
			}
		}
		w.RawByte(']')
	}
	needWriteComma = true
	_ = needWriteComma
	w.RawByte('}')

}

func (m *Guide) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *Guide) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *Guide) GoString() string {
	return m.String()
}

func (m *Guide) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RoleId)
	if l > 0 {
		n += 1 + l + sovGuide(uint64(l))
	}
	if m.GuideId != 0 {
		n += 1 + sovGuide(uint64(m.GuideId))
	}
	if m.StepId != 0 {
		n += 1 + sovGuide(uint64(m.StepId))
	}
	if len(m.FinishedGuideIds) > 0 {
		l = 0
		for _, e := range m.FinishedGuideIds {
			l += sovGuide(uint64(e))
		}
		n += 1 + sovGuide(uint64(l)) + l
	}
	return n
}

func sovGuide(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGuide(x uint64) (n int) {
	return sovGuide(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Guide) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGuide
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
			return fmt.Errorf("proto: Guide: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Guide: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RoleId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuide
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
				return ErrInvalidLengthGuide
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGuide
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RoleId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GuideId", wireType)
			}
			m.GuideId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuide
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GuideId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StepId", wireType)
			}
			m.StepId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGuide
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StepId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType == 0 {
				var v int64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGuide
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= int64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.FinishedGuideIds = append(m.FinishedGuideIds, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGuide
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthGuide
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthGuide
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA[iNdEx:postIndex] {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.FinishedGuideIds) == 0 {
					m.FinishedGuideIds = make([]int64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowGuide
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= int64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.FinishedGuideIds = append(m.FinishedGuideIds, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field FinishedGuideIds", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGuide(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGuide
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
func skipGuide(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGuide
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
					return 0, ErrIntOverflowGuide
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
					return 0, ErrIntOverflowGuide
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
				return 0, ErrInvalidLengthGuide
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGuide
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGuide
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGuide        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGuide          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGuide = fmt.Errorf("proto: unexpected end of group")
)
