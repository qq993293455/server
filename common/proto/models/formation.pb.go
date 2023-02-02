// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/models/formation.proto

package models

import (
	coin_server_common_buffer "coin-server/common/buffer"
	coin_server_common_jwriter "coin-server/common/jwriter"
	coin_server_common_msgcreate "coin-server/common/msgcreate"
	coin_server_common_proto_jsonany "coin-server/common/proto/jsonany"
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

// 一个组合
type Assemble struct {
	Hero_0       int64  `protobuf:"varint,1,opt,name=hero_0,json=hero0,proto3" json:"hero_0,omitempty"`
	HeroOrigin_0 int64  `protobuf:"varint,2,opt,name=hero_origin_0,json=heroOrigin0,proto3" json:"hero_origin_0,omitempty"`
	Hero_1       int64  `protobuf:"varint,3,opt,name=hero_1,json=hero1,proto3" json:"hero_1,omitempty"`
	HeroOrigin_1 int64  `protobuf:"varint,4,opt,name=hero_origin_1,json=heroOrigin1,proto3" json:"hero_origin_1,omitempty"`
	Name         string `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty"`
	// 竞技场
	Hero_0Power int64 `protobuf:"varint,6,opt,name=hero_0_power,json=hero0Power,proto3" json:"hero_0_power,omitempty"`
	Hero_1Power int64 `protobuf:"varint,7,opt,name=hero_1_power,json=hero1Power,proto3" json:"hero_1_power,omitempty"`
}

func (m *Assemble) Reset()      { *m = Assemble{} }
func (*Assemble) ProtoMessage() {}
func (*Assemble) Descriptor() ([]byte, []int) {
	return fileDescriptor_50cc9624c3d20cc2, []int{0}
}
func (m *Assemble) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Assemble) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Assemble.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Assemble) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Assemble.Merge(m, src)
}
func (m *Assemble) XXX_Size() int {
	return m.Size()
}
func (m *Assemble) XXX_DiscardUnknown() {
	xxx_messageInfo_Assemble.DiscardUnknown(m)
}

var xxx_messageInfo_Assemble proto.InternalMessageInfo

func (m *Assemble) GetHero_0() int64 {
	if m != nil {
		return m.Hero_0
	}
	return 0
}

func (m *Assemble) GetHeroOrigin_0() int64 {
	if m != nil {
		return m.HeroOrigin_0
	}
	return 0
}

func (m *Assemble) GetHero_1() int64 {
	if m != nil {
		return m.Hero_1
	}
	return 0
}

func (m *Assemble) GetHeroOrigin_1() int64 {
	if m != nil {
		return m.HeroOrigin_1
	}
	return 0
}

func (m *Assemble) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Assemble) GetHero_0Power() int64 {
	if m != nil {
		return m.Hero_0Power
	}
	return 0
}

func (m *Assemble) GetHero_1Power() int64 {
	if m != nil {
		return m.Hero_1Power
	}
	return 0
}

func (*Assemble) XXX_MessageName() string {
	return "models.Assemble"
}
func init() {
	proto.RegisterType((*Assemble)(nil), "models.Assemble")
}

func init() { proto.RegisterFile("proto/models/formation.proto", fileDescriptor_50cc9624c3d20cc2) }

var fileDescriptor_50cc9624c3d20cc2 = []byte{
	// 254 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x40, 0x7d, 0xb4, 0x0d, 0x60, 0x60, 0xb1, 0x84, 0xe4, 0x01, 0x1d, 0x51, 0xa7, 0x2e, 0x90,
	0x58, 0x7c, 0x01, 0xec, 0x08, 0xd4, 0x91, 0x25, 0x6a, 0x8b, 0x81, 0x48, 0xb5, 0xaf, 0x72, 0x2a,
	0xf8, 0x0d, 0x3e, 0x83, 0x4f, 0xe9, 0xd8, 0x31, 0x23, 0x38, 0x0b, 0x23, 0x9f, 0x80, 0x72, 0x45,
	0x2d, 0xca, 0xe6, 0x7b, 0x7e, 0x4f, 0x3a, 0x9d, 0x3c, 0x5b, 0x04, 0x5a, 0x52, 0xe6, 0xe8, 0xd1,
	0xce, 0xab, 0xec, 0x89, 0x82, 0x9b, 0x2c, 0x4b, 0xf2, 0x97, 0x8c, 0x55, 0xb2, 0xe1, 0xc3, 0x1a,
	0xe4, 0xc1, 0x75, 0x55, 0x59, 0x37, 0x9d, 0x5b, 0x75, 0x2a, 0x93, 0x17, 0x1b, 0xa8, 0xc8, 0x35,
	0xa4, 0x30, 0xea, 0x8d, 0x07, 0xed, 0x94, 0xab, 0xa1, 0x3c, 0x61, 0x4c, 0xa1, 0x7c, 0x2e, 0x7d,
	0x91, 0xeb, 0x3d, 0xfe, 0x3d, 0x6a, 0xe1, 0x1d, 0xb3, 0x7c, 0x9b, 0x1a, 0xdd, 0xdb, 0xa5, 0xa6,
	0x9b, 0x1a, 0xdd, 0xef, 0xa6, 0x46, 0x29, 0xd9, 0xf7, 0x13, 0x67, 0xf5, 0x20, 0x85, 0xd1, 0xe1,
	0x98, 0xdf, 0x2a, 0x95, 0xc7, 0x9b, 0x4d, 0x8a, 0x05, 0xbd, 0xd9, 0xa0, 0x13, 0xce, 0x24, 0xef,
	0x73, 0xdf, 0x92, 0xad, 0x61, 0xfe, 0x8c, 0xfd, 0x9d, 0x61, 0xd8, 0xb8, 0xb9, 0xad, 0xbf, 0x50,
	0x7c, 0x44, 0x84, 0x55, 0x44, 0x58, 0x47, 0x84, 0xcf, 0x88, 0xf0, 0x1d, 0x51, 0xfc, 0x44, 0x84,
	0xf7, 0x06, 0xc5, 0xaa, 0x41, 0x58, 0x37, 0x28, 0xea, 0x06, 0xc5, 0xc3, 0xf9, 0x8c, 0x4a, 0x7f,
	0x51, 0xd9, 0xf0, 0x6a, 0x43, 0x36, 0x23, 0xe7, 0xc8, 0x67, 0xff, 0x2f, 0x38, 0x4d, 0x78, 0xba,
	0xfa, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x4b, 0xf7, 0x57, 0x54, 0x58, 0x01, 0x00, 0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolAssemble.Get().(proto.Message)
	})
}

var poolAssemble = &sync.Pool{New: func() interface{} { return &Assemble{} }}

func (m *Assemble) ReleasePool() { m.Reset(); poolAssemble.Put(m); m = nil }
func (this *Assemble) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Assemble)
	if !ok {
		that2, ok := that.(Assemble)
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
	if this.Hero_0 != that1.Hero_0 {
		return false
	}
	if this.HeroOrigin_0 != that1.HeroOrigin_0 {
		return false
	}
	if this.Hero_1 != that1.Hero_1 {
		return false
	}
	if this.HeroOrigin_1 != that1.HeroOrigin_1 {
		return false
	}
	if this.Name != that1.Name {
		return false
	}
	if this.Hero_0Power != that1.Hero_0Power {
		return false
	}
	if this.Hero_1Power != that1.Hero_1Power {
		return false
	}
	return true
}
func (m *Assemble) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Assemble) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Assemble) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Hero_1Power != 0 {
		i = encodeVarintFormation(dAtA, i, uint64(m.Hero_1Power))
		i--
		dAtA[i] = 0x38
	}
	if m.Hero_0Power != 0 {
		i = encodeVarintFormation(dAtA, i, uint64(m.Hero_0Power))
		i--
		dAtA[i] = 0x30
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintFormation(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0x2a
	}
	if m.HeroOrigin_1 != 0 {
		i = encodeVarintFormation(dAtA, i, uint64(m.HeroOrigin_1))
		i--
		dAtA[i] = 0x20
	}
	if m.Hero_1 != 0 {
		i = encodeVarintFormation(dAtA, i, uint64(m.Hero_1))
		i--
		dAtA[i] = 0x18
	}
	if m.HeroOrigin_0 != 0 {
		i = encodeVarintFormation(dAtA, i, uint64(m.HeroOrigin_0))
		i--
		dAtA[i] = 0x10
	}
	if m.Hero_0 != 0 {
		i = encodeVarintFormation(dAtA, i, uint64(m.Hero_0))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintFormation(dAtA []byte, offset int, v uint64) int {
	offset -= sovFormation(v)
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

func (m *Assemble) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	if m.Hero_0 != 0 {
		w.RawByte('"')
		w.RawString("hero_0")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.Hero_0))
		needWriteComma = true
	}
	if m.HeroOrigin_0 != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("hero_origin_0")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.HeroOrigin_0))
		needWriteComma = true
	}
	if m.Hero_1 != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("hero_1")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.Hero_1))
		needWriteComma = true
	}
	if m.HeroOrigin_1 != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("hero_origin_1")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.HeroOrigin_1))
		needWriteComma = true
	}
	if m.Name != "" {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("name")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.Name)
		needWriteComma = true
	}
	if m.Hero_0Power != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("hero_0_power")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.Hero_0Power))
		needWriteComma = true
	}
	if m.Hero_1Power != 0 {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("hero_1_power")
		w.RawByte('"')
		w.RawByte(':')
		w.Int64(int64(m.Hero_1Power))
		needWriteComma = true
	}
	_ = needWriteComma
	w.RawByte('}')

}

func (m *Assemble) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *Assemble) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *Assemble) GoString() string {
	return m.String()
}

func (m *Assemble) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Hero_0 != 0 {
		n += 1 + sovFormation(uint64(m.Hero_0))
	}
	if m.HeroOrigin_0 != 0 {
		n += 1 + sovFormation(uint64(m.HeroOrigin_0))
	}
	if m.Hero_1 != 0 {
		n += 1 + sovFormation(uint64(m.Hero_1))
	}
	if m.HeroOrigin_1 != 0 {
		n += 1 + sovFormation(uint64(m.HeroOrigin_1))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovFormation(uint64(l))
	}
	if m.Hero_0Power != 0 {
		n += 1 + sovFormation(uint64(m.Hero_0Power))
	}
	if m.Hero_1Power != 0 {
		n += 1 + sovFormation(uint64(m.Hero_1Power))
	}
	return n
}

func sovFormation(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozFormation(x uint64) (n int) {
	return sovFormation(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Assemble) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowFormation
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
			return fmt.Errorf("proto: Assemble: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Assemble: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hero_0", wireType)
			}
			m.Hero_0 = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFormation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Hero_0 |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field HeroOrigin_0", wireType)
			}
			m.HeroOrigin_0 = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFormation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.HeroOrigin_0 |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hero_1", wireType)
			}
			m.Hero_1 = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFormation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Hero_1 |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field HeroOrigin_1", wireType)
			}
			m.HeroOrigin_1 = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFormation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.HeroOrigin_1 |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFormation
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
				return ErrInvalidLengthFormation
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthFormation
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hero_0Power", wireType)
			}
			m.Hero_0Power = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFormation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Hero_0Power |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hero_1Power", wireType)
			}
			m.Hero_1Power = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFormation
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Hero_1Power |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipFormation(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthFormation
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
func skipFormation(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowFormation
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
					return 0, ErrIntOverflowFormation
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
					return 0, ErrIntOverflowFormation
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
				return 0, ErrInvalidLengthFormation
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupFormation
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthFormation
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthFormation        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowFormation          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupFormation = fmt.Errorf("proto: unexpected end of group")
)