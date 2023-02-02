// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/dao/discord.proto

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

type DiscordData struct {
	RoleId        string  `protobuf:"bytes,1,opt,name=role_id,json=roleId,proto3" json:"role_id,omitempty" pk`
	RewardVersion []int64 `protobuf:"varint,2,rep,packed,name=reward_version,json=rewardVersion,proto3" json:"reward_version,omitempty"`
}

func (m *DiscordData) Reset()      { *m = DiscordData{} }
func (*DiscordData) ProtoMessage() {}
func (*DiscordData) Descriptor() ([]byte, []int) {
	return fileDescriptor_89b099086acbaecc, []int{0}
}
func (m *DiscordData) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DiscordData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DiscordData.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DiscordData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DiscordData.Merge(m, src)
}
func (m *DiscordData) XXX_Size() int {
	return m.Size()
}
func (m *DiscordData) XXX_DiscardUnknown() {
	xxx_messageInfo_DiscordData.DiscardUnknown(m)
}

var xxx_messageInfo_DiscordData proto.InternalMessageInfo

func (m *DiscordData) GetRoleId() string {
	if m != nil {
		return m.RoleId
	}
	return ""
}

func (m *DiscordData) GetRewardVersion() []int64 {
	if m != nil {
		return m.RewardVersion
	}
	return nil
}

func (*DiscordData) XXX_MessageName() string {
	return "dao.DiscordData"
}
func init() {
	proto.RegisterType((*DiscordData)(nil), "dao.DiscordData")
}

func init() { proto.RegisterFile("proto/dao/discord.proto", fileDescriptor_89b099086acbaecc) }

var fileDescriptor_89b099086acbaecc = []byte{
	// 234 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2f, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0x4f, 0x49, 0xcc, 0xd7, 0x4f, 0xc9, 0x2c, 0x4e, 0xce, 0x2f, 0x4a, 0xd1, 0x03, 0x8b,
	0x08, 0x31, 0xa7, 0x24, 0xe6, 0x4b, 0x89, 0xa4, 0xe7, 0xa7, 0xe7, 0x43, 0x54, 0x80, 0x58, 0x10,
	0x29, 0x29, 0x19, 0x88, 0x48, 0x6e, 0x7e, 0x4a, 0x6a, 0x4e, 0xb1, 0x7e, 0x52, 0x7e, 0x71, 0x71,
	0x7c, 0x46, 0x62, 0x4e, 0x0e, 0x44, 0x56, 0x29, 0x94, 0x8b, 0xdb, 0x05, 0x62, 0x92, 0x4b, 0x62,
	0x49, 0xa2, 0x90, 0x3c, 0x17, 0x7b, 0x51, 0x7e, 0x4e, 0x6a, 0x7c, 0x66, 0x8a, 0x04, 0xa3, 0x02,
	0xa3, 0x06, 0xa7, 0x13, 0xdb, 0xa7, 0x7b, 0xf2, 0x4c, 0x05, 0xd9, 0x41, 0x6c, 0x20, 0x61, 0xcf,
	0x14, 0x21, 0x55, 0x2e, 0xbe, 0xa2, 0xd4, 0xf2, 0xc4, 0xa2, 0x94, 0xf8, 0xb2, 0xd4, 0xa2, 0xe2,
	0xcc, 0xfc, 0x3c, 0x09, 0x26, 0x05, 0x66, 0x0d, 0xe6, 0x20, 0x5e, 0x88, 0x68, 0x18, 0x44, 0xd0,
	0xc9, 0xeb, 0xc6, 0x43, 0x39, 0x86, 0x15, 0x8f, 0xe4, 0x18, 0x4f, 0x3c, 0x92, 0x63, 0xbc, 0xf0,
	0x48, 0x8e, 0xf1, 0xc1, 0x23, 0x39, 0xc6, 0x17, 0x8f, 0xe4, 0x18, 0x3e, 0x3c, 0x92, 0x63, 0x9c,
	0xf0, 0x58, 0x8e, 0xe1, 0xc4, 0x63, 0x39, 0xc6, 0x0b, 0x8f, 0xe5, 0x18, 0x6e, 0x3c, 0x96, 0x63,
	0x88, 0x92, 0x49, 0xce, 0xcf, 0xcc, 0xd3, 0x2d, 0x4e, 0x2d, 0x2a, 0x4b, 0x2d, 0xd2, 0x4f, 0xce,
	0xcf, 0xcd, 0xcd, 0xcf, 0xd3, 0x87, 0xfb, 0x34, 0x89, 0x0d, 0xcc, 0x34, 0x06, 0x04, 0x00, 0x00,
	0xff, 0xff, 0x52, 0xf6, 0xc7, 0x33, 0xfd, 0x00, 0x00, 0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolDiscordData.Get().(proto.Message)
	})
}

var poolDiscordData = &sync.Pool{New: func() interface{} { return &DiscordData{} }}

func (m *DiscordData) ReleasePool() { m.Reset(); poolDiscordData.Put(m); m = nil }

func (m *DiscordData) PK() string {
	if m == nil {
		return ""
	}
	return m.RoleId
}

func (m *DiscordData) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return append(d, m.RoleId...)
}

func (m *DiscordData) ToKVSave() ([]byte, []byte) {
	msgName := m.XXX_MessageName()
	dk := coin_server_common_bytespool.GetSample(64)
	dk = dk[:0]
	dk = append(dk, msgName...)
	dk = append(dk, ':', 'k', ':')
	dk = m.PKAppendTo(dk)
	return dk, m.ToSave()
}

func (m *DiscordData) ToSave() []byte {
	msgName := m.XXX_MessageName()
	ml := len(msgName)
	d := coin_server_common_bytespool.GetSample(1 + ml + m.Size())
	d[0] = uint8(ml)
	copy(d[1:], msgName)
	_, _ = m.MarshalToSizedBuffer(d[1+ml:])
	return d
}

func (m *DiscordData) KVKey() string {
	return m.XXX_MessageName() + ":k:" + m.PK()
}

func (this *DiscordData) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DiscordData)
	if !ok {
		that2, ok := that.(DiscordData)
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
	if len(this.RewardVersion) != len(that1.RewardVersion) {
		return false
	}
	for i := range this.RewardVersion {
		if this.RewardVersion[i] != that1.RewardVersion[i] {
			return false
		}
	}
	return true
}
func (m *DiscordData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DiscordData) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DiscordData) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.RewardVersion) > 0 {
		dAtA2 := make([]byte, len(m.RewardVersion)*10)
		var j1 int
		for _, num1 := range m.RewardVersion {
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
		i = encodeVarintDiscord(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0x12
	}
	if len(m.RoleId) > 0 {
		i -= len(m.RoleId)
		copy(dAtA[i:], m.RoleId)
		i = encodeVarintDiscord(dAtA, i, uint64(len(m.RoleId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintDiscord(dAtA []byte, offset int, v uint64) int {
	offset -= sovDiscord(v)
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

func (m *DiscordData) JsonBytes(w *coin_server_common_jwriter.Writer) {
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
	w.RawString("reward_version")
	w.RawByte('"')
	w.RawByte(':')
	if m.RewardVersion == nil {
		w.RawString("null")
	} else if len(m.RewardVersion) == 0 {
		w.RawString("[]")
	} else {
		w.RawByte('[')
		for i, v := range m.RewardVersion {
			w.Int64(int64(v))
			if i != len(m.RewardVersion)-1 {
				w.RawByte(',')
			}
		}
		w.RawByte(']')
	}
	needWriteComma = true
	_ = needWriteComma
	w.RawByte('}')

}

func (m *DiscordData) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *DiscordData) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *DiscordData) GoString() string {
	return m.String()
}

func (m *DiscordData) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RoleId)
	if l > 0 {
		n += 1 + l + sovDiscord(uint64(l))
	}
	if len(m.RewardVersion) > 0 {
		l = 0
		for _, e := range m.RewardVersion {
			l += sovDiscord(uint64(e))
		}
		n += 1 + sovDiscord(uint64(l)) + l
	}
	return n
}

func sovDiscord(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozDiscord(x uint64) (n int) {
	return sovDiscord(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *DiscordData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDiscord
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
			return fmt.Errorf("proto: DiscordData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DiscordData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RoleId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDiscord
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
				return ErrInvalidLengthDiscord
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthDiscord
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RoleId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType == 0 {
				var v int64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowDiscord
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
				m.RewardVersion = append(m.RewardVersion, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowDiscord
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
					return ErrInvalidLengthDiscord
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthDiscord
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
				if elementCount != 0 && len(m.RewardVersion) == 0 {
					m.RewardVersion = make([]int64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowDiscord
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
					m.RewardVersion = append(m.RewardVersion, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field RewardVersion", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipDiscord(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthDiscord
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
func skipDiscord(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowDiscord
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
					return 0, ErrIntOverflowDiscord
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
					return 0, ErrIntOverflowDiscord
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
				return 0, ErrInvalidLengthDiscord
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupDiscord
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthDiscord
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthDiscord        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowDiscord          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupDiscord = fmt.Errorf("proto: unexpected end of group")
)