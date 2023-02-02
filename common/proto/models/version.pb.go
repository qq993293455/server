// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/models/version.proto

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

type Version struct {
	Version      string `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	MinVersion   string `protobuf:"bytes,2,opt,name=min_version,json=minVersion,proto3" json:"min_version,omitempty"`
	Gateway      string `protobuf:"bytes,3,opt,name=gateway,proto3" json:"gateway,omitempty"`
	Cdn          string `protobuf:"bytes,4,opt,name=cdn,proto3" json:"cdn,omitempty"`
	Announcement string `protobuf:"bytes,5,opt,name=announcement,proto3" json:"announcement,omitempty"`
	VersionFile  string `protobuf:"bytes,6,opt,name=version_file,json=versionFile,proto3" json:"version_file,omitempty"`
	Activate     bool   `protobuf:"varint,7,opt,name=activate,proto3" json:"activate,omitempty"`
}

func (m *Version) Reset()      { *m = Version{} }
func (*Version) ProtoMessage() {}
func (*Version) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dbde49cacaff7aa, []int{0}
}
func (m *Version) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Version) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Version.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Version) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Version.Merge(m, src)
}
func (m *Version) XXX_Size() int {
	return m.Size()
}
func (m *Version) XXX_DiscardUnknown() {
	xxx_messageInfo_Version.DiscardUnknown(m)
}

var xxx_messageInfo_Version proto.InternalMessageInfo

func (m *Version) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *Version) GetMinVersion() string {
	if m != nil {
		return m.MinVersion
	}
	return ""
}

func (m *Version) GetGateway() string {
	if m != nil {
		return m.Gateway
	}
	return ""
}

func (m *Version) GetCdn() string {
	if m != nil {
		return m.Cdn
	}
	return ""
}

func (m *Version) GetAnnouncement() string {
	if m != nil {
		return m.Announcement
	}
	return ""
}

func (m *Version) GetVersionFile() string {
	if m != nil {
		return m.VersionFile
	}
	return ""
}

func (m *Version) GetActivate() bool {
	if m != nil {
		return m.Activate
	}
	return false
}

func (*Version) XXX_MessageName() string {
	return "models.Version"
}
func init() {
	proto.RegisterType((*Version)(nil), "models.Version")
}

func init() { proto.RegisterFile("proto/models/version.proto", fileDescriptor_4dbde49cacaff7aa) }

var fileDescriptor_4dbde49cacaff7aa = []byte{
	// 259 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x40, 0x73, 0x14, 0x9a, 0xe2, 0x76, 0x40, 0x9e, 0xac, 0x0e, 0xd7, 0xd2, 0xa9, 0x0b, 0x64,
	0xe0, 0x0f, 0x18, 0xd8, 0x58, 0x3a, 0x30, 0xb0, 0x54, 0xc1, 0x3d, 0x90, 0xa5, 0xf8, 0x8c, 0x12,
	0x13, 0xc4, 0x5f, 0xf0, 0x19, 0x7c, 0x4a, 0xc7, 0x4a, 0x2c, 0x1d, 0xc1, 0x59, 0x18, 0xf9, 0x04,
	0x84, 0x93, 0xa2, 0xb2, 0xf9, 0xbd, 0x77, 0x27, 0x4b, 0x27, 0xc6, 0x8f, 0xa5, 0xf3, 0x2e, 0xb3,
	0x6e, 0x45, 0x45, 0x95, 0xd5, 0x54, 0x56, 0xc6, 0xf1, 0x79, 0x94, 0xb2, 0xdf, 0xda, 0xd9, 0x3b,
	0x88, 0xf4, 0xa6, 0x2d, 0x52, 0x89, 0xb4, 0x1b, 0x52, 0x30, 0x85, 0xf9, 0xf1, 0x62, 0x87, 0x72,
	0x22, 0x86, 0xd6, 0xf0, 0x72, 0x57, 0x0f, 0x62, 0x15, 0xd6, 0xf0, 0xde, 0xea, 0x43, 0xee, 0xe9,
	0x39, 0x7f, 0x51, 0xbd, 0x76, 0xb5, 0x43, 0x79, 0x22, 0x7a, 0x7a, 0xc5, 0xea, 0x30, 0xda, 0xdf,
	0xa7, 0x9c, 0x89, 0x51, 0xce, 0xec, 0x9e, 0x58, 0x93, 0x25, 0xf6, 0xea, 0x28, 0xa6, 0x7f, 0x4e,
	0x9e, 0x8a, 0x51, 0xf7, 0xd9, 0xf2, 0xde, 0x14, 0xa4, 0xfa, 0x71, 0x66, 0xd8, 0xb9, 0x2b, 0x53,
	0x90, 0x1c, 0x8b, 0x41, 0xae, 0xbd, 0xa9, 0x73, 0x4f, 0x2a, 0x9d, 0xc2, 0x7c, 0xb0, 0xf8, 0xe3,
	0xcb, 0xeb, 0xed, 0x27, 0x26, 0x6f, 0x01, 0x61, 0x1d, 0x10, 0x36, 0x01, 0xe1, 0x23, 0x20, 0x7c,
	0x05, 0x4c, 0xbe, 0x03, 0xc2, 0x6b, 0x83, 0xc9, 0xba, 0x41, 0xd8, 0x34, 0x98, 0x6c, 0x1b, 0x4c,
	0x6e, 0x27, 0xda, 0x19, 0x3e, 0xab, 0xa8, 0xac, 0xa9, 0xcc, 0xb4, 0xb3, 0xd6, 0x71, 0xb6, 0x7f,
	0xba, 0xbb, 0x7e, 0xa4, 0x8b, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x72, 0xbc, 0xf9, 0x9f, 0x51,
	0x01, 0x00, 0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolVersion.Get().(proto.Message)
	})
}

var poolVersion = &sync.Pool{New: func() interface{} { return &Version{} }}

func (m *Version) ReleasePool() { m.Reset(); poolVersion.Put(m); m = nil }
func (this *Version) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Version)
	if !ok {
		that2, ok := that.(Version)
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
	if this.Version != that1.Version {
		return false
	}
	if this.MinVersion != that1.MinVersion {
		return false
	}
	if this.Gateway != that1.Gateway {
		return false
	}
	if this.Cdn != that1.Cdn {
		return false
	}
	if this.Announcement != that1.Announcement {
		return false
	}
	if this.VersionFile != that1.VersionFile {
		return false
	}
	if this.Activate != that1.Activate {
		return false
	}
	return true
}
func (m *Version) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Version) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Version) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Activate {
		i--
		if m.Activate {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x38
	}
	if len(m.VersionFile) > 0 {
		i -= len(m.VersionFile)
		copy(dAtA[i:], m.VersionFile)
		i = encodeVarintVersion(dAtA, i, uint64(len(m.VersionFile)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Announcement) > 0 {
		i -= len(m.Announcement)
		copy(dAtA[i:], m.Announcement)
		i = encodeVarintVersion(dAtA, i, uint64(len(m.Announcement)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Cdn) > 0 {
		i -= len(m.Cdn)
		copy(dAtA[i:], m.Cdn)
		i = encodeVarintVersion(dAtA, i, uint64(len(m.Cdn)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Gateway) > 0 {
		i -= len(m.Gateway)
		copy(dAtA[i:], m.Gateway)
		i = encodeVarintVersion(dAtA, i, uint64(len(m.Gateway)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.MinVersion) > 0 {
		i -= len(m.MinVersion)
		copy(dAtA[i:], m.MinVersion)
		i = encodeVarintVersion(dAtA, i, uint64(len(m.MinVersion)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Version) > 0 {
		i -= len(m.Version)
		copy(dAtA[i:], m.Version)
		i = encodeVarintVersion(dAtA, i, uint64(len(m.Version)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintVersion(dAtA []byte, offset int, v uint64) int {
	offset -= sovVersion(v)
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

func (m *Version) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	if m.Version != "" {
		w.RawByte('"')
		w.RawString("version")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.Version)
		needWriteComma = true
	}
	if m.MinVersion != "" {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("min_version")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.MinVersion)
		needWriteComma = true
	}
	if m.Gateway != "" {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("gateway")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.Gateway)
		needWriteComma = true
	}
	if m.Cdn != "" {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("cdn")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.Cdn)
		needWriteComma = true
	}
	if m.Announcement != "" {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("announcement")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.Announcement)
		needWriteComma = true
	}
	if m.VersionFile != "" {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("version_file")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.VersionFile)
		needWriteComma = true
	}
	if m.Activate {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("activate")
		w.RawByte('"')
		w.RawByte(':')
		w.Bool(m.Activate)
		needWriteComma = true
	}
	_ = needWriteComma
	w.RawByte('}')

}

func (m *Version) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *Version) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *Version) GoString() string {
	return m.String()
}

func (m *Version) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Version)
	if l > 0 {
		n += 1 + l + sovVersion(uint64(l))
	}
	l = len(m.MinVersion)
	if l > 0 {
		n += 1 + l + sovVersion(uint64(l))
	}
	l = len(m.Gateway)
	if l > 0 {
		n += 1 + l + sovVersion(uint64(l))
	}
	l = len(m.Cdn)
	if l > 0 {
		n += 1 + l + sovVersion(uint64(l))
	}
	l = len(m.Announcement)
	if l > 0 {
		n += 1 + l + sovVersion(uint64(l))
	}
	l = len(m.VersionFile)
	if l > 0 {
		n += 1 + l + sovVersion(uint64(l))
	}
	if m.Activate {
		n += 2
	}
	return n
}

func sovVersion(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozVersion(x uint64) (n int) {
	return sovVersion(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Version) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowVersion
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
			return fmt.Errorf("proto: Version: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Version: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVersion
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
				return ErrInvalidLengthVersion
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVersion
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Version = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinVersion", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVersion
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
				return ErrInvalidLengthVersion
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVersion
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MinVersion = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Gateway", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVersion
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
				return ErrInvalidLengthVersion
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVersion
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Gateway = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cdn", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVersion
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
				return ErrInvalidLengthVersion
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVersion
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Cdn = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Announcement", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVersion
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
				return ErrInvalidLengthVersion
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVersion
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Announcement = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field VersionFile", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVersion
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
				return ErrInvalidLengthVersion
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVersion
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.VersionFile = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Activate", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVersion
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
			m.Activate = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipVersion(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthVersion
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
func skipVersion(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowVersion
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
					return 0, ErrIntOverflowVersion
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
					return 0, ErrIntOverflowVersion
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
				return 0, ErrInvalidLengthVersion
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupVersion
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthVersion
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthVersion        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowVersion          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupVersion = fmt.Errorf("proto: unexpected end of group")
)
