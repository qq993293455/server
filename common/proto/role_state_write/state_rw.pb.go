// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/role-state-write-server/state_rw.proto

package role_state_write

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

// 只有 RoleState 服务的 master 会sub下列msg
type RoleStateRW struct {
}

func (m *RoleStateRW) Reset()      { *m = RoleStateRW{} }
func (*RoleStateRW) ProtoMessage() {}
func (*RoleStateRW) Descriptor() ([]byte, []int) {
	return fileDescriptor_2ce8afeb98d53622, []int{0}
}
func (m *RoleStateRW) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RoleStateRW) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RoleStateRW.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RoleStateRW) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RoleStateRW.Merge(m, src)
}
func (m *RoleStateRW) XXX_Size() int {
	return m.Size()
}
func (m *RoleStateRW) XXX_DiscardUnknown() {
	xxx_messageInfo_RoleStateRW.DiscardUnknown(m)
}

var xxx_messageInfo_RoleStateRW proto.InternalMessageInfo

func (*RoleStateRW) XXX_MessageName() string {
	return "role_state_write.RoleStateRW"
}

type RoleStateRW_LoginNotifyEvent struct {
}

func (m *RoleStateRW_LoginNotifyEvent) Reset()      { *m = RoleStateRW_LoginNotifyEvent{} }
func (*RoleStateRW_LoginNotifyEvent) ProtoMessage() {}
func (*RoleStateRW_LoginNotifyEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_2ce8afeb98d53622, []int{0, 0}
}
func (m *RoleStateRW_LoginNotifyEvent) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RoleStateRW_LoginNotifyEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RoleStateRW_LoginNotifyEvent.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RoleStateRW_LoginNotifyEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RoleStateRW_LoginNotifyEvent.Merge(m, src)
}
func (m *RoleStateRW_LoginNotifyEvent) XXX_Size() int {
	return m.Size()
}
func (m *RoleStateRW_LoginNotifyEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_RoleStateRW_LoginNotifyEvent.DiscardUnknown(m)
}

var xxx_messageInfo_RoleStateRW_LoginNotifyEvent proto.InternalMessageInfo

func (*RoleStateRW_LoginNotifyEvent) XXX_MessageName() string {
	return "role_state_write.RoleStateRW.LoginNotifyEvent"
}

type RoleStateRW_LogoutNotifyEvent struct {
}

func (m *RoleStateRW_LogoutNotifyEvent) Reset()      { *m = RoleStateRW_LogoutNotifyEvent{} }
func (*RoleStateRW_LogoutNotifyEvent) ProtoMessage() {}
func (*RoleStateRW_LogoutNotifyEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_2ce8afeb98d53622, []int{0, 1}
}
func (m *RoleStateRW_LogoutNotifyEvent) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RoleStateRW_LogoutNotifyEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RoleStateRW_LogoutNotifyEvent.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RoleStateRW_LogoutNotifyEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RoleStateRW_LogoutNotifyEvent.Merge(m, src)
}
func (m *RoleStateRW_LogoutNotifyEvent) XXX_Size() int {
	return m.Size()
}
func (m *RoleStateRW_LogoutNotifyEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_RoleStateRW_LogoutNotifyEvent.DiscardUnknown(m)
}

var xxx_messageInfo_RoleStateRW_LogoutNotifyEvent proto.InternalMessageInfo

func (*RoleStateRW_LogoutNotifyEvent) XXX_MessageName() string {
	return "role_state_write.RoleStateRW.LogoutNotifyEvent"
}

type RoleStateRW_JoinClusterRequest struct {
	ClusterTcpAddr  string `protobuf:"bytes,1,opt,name=cluster_tcp_addr,json=clusterTcpAddr,proto3" json:"cluster_tcp_addr,omitempty"`
	ClusterServerId string `protobuf:"bytes,2,opt,name=cluster_server_id,json=clusterServerId,proto3" json:"cluster_server_id,omitempty"`
}

func (m *RoleStateRW_JoinClusterRequest) Reset()      { *m = RoleStateRW_JoinClusterRequest{} }
func (*RoleStateRW_JoinClusterRequest) ProtoMessage() {}
func (*RoleStateRW_JoinClusterRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_2ce8afeb98d53622, []int{0, 2}
}
func (m *RoleStateRW_JoinClusterRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RoleStateRW_JoinClusterRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RoleStateRW_JoinClusterRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RoleStateRW_JoinClusterRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RoleStateRW_JoinClusterRequest.Merge(m, src)
}
func (m *RoleStateRW_JoinClusterRequest) XXX_Size() int {
	return m.Size()
}
func (m *RoleStateRW_JoinClusterRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RoleStateRW_JoinClusterRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RoleStateRW_JoinClusterRequest proto.InternalMessageInfo

func (m *RoleStateRW_JoinClusterRequest) GetClusterTcpAddr() string {
	if m != nil {
		return m.ClusterTcpAddr
	}
	return ""
}

func (m *RoleStateRW_JoinClusterRequest) GetClusterServerId() string {
	if m != nil {
		return m.ClusterServerId
	}
	return ""
}

func (*RoleStateRW_JoinClusterRequest) XXX_MessageName() string {
	return "role_state_write.RoleStateRW.JoinClusterRequest"
}

type RoleStateRW_JoinClusterResponse struct {
}

func (m *RoleStateRW_JoinClusterResponse) Reset()      { *m = RoleStateRW_JoinClusterResponse{} }
func (*RoleStateRW_JoinClusterResponse) ProtoMessage() {}
func (*RoleStateRW_JoinClusterResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_2ce8afeb98d53622, []int{0, 3}
}
func (m *RoleStateRW_JoinClusterResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RoleStateRW_JoinClusterResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RoleStateRW_JoinClusterResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RoleStateRW_JoinClusterResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RoleStateRW_JoinClusterResponse.Merge(m, src)
}
func (m *RoleStateRW_JoinClusterResponse) XXX_Size() int {
	return m.Size()
}
func (m *RoleStateRW_JoinClusterResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RoleStateRW_JoinClusterResponse.DiscardUnknown(m)
}

var xxx_messageInfo_RoleStateRW_JoinClusterResponse proto.InternalMessageInfo

func (*RoleStateRW_JoinClusterResponse) XXX_MessageName() string {
	return "role_state_write.RoleStateRW.JoinClusterResponse"
}
func init() {
	proto.RegisterType((*RoleStateRW)(nil), "role_state_write.RoleStateRW")
	proto.RegisterType((*RoleStateRW_LoginNotifyEvent)(nil), "role_state_write.RoleStateRW.LoginNotifyEvent")
	proto.RegisterType((*RoleStateRW_LogoutNotifyEvent)(nil), "role_state_write.RoleStateRW.LogoutNotifyEvent")
	proto.RegisterType((*RoleStateRW_JoinClusterRequest)(nil), "role_state_write.RoleStateRW.JoinClusterRequest")
	proto.RegisterType((*RoleStateRW_JoinClusterResponse)(nil), "role_state_write.RoleStateRW.JoinClusterResponse")
}

func init() {
	proto.RegisterFile("proto/role-state-write-server/state_rw.proto", fileDescriptor_2ce8afeb98d53622)
}

var fileDescriptor_2ce8afeb98d53622 = []byte{
	// 280 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0xd0, 0x31, 0x4a, 0x03, 0x41,
	0x14, 0xc6, 0xf1, 0x1d, 0x0b, 0xc1, 0x11, 0x34, 0xd9, 0x20, 0x84, 0x14, 0x0f, 0xb1, 0x8a, 0xe2,
	0xba, 0x85, 0x27, 0x50, 0xb1, 0x50, 0x82, 0xc5, 0x46, 0x08, 0xd8, 0x2c, 0x71, 0xe7, 0x19, 0x46,
	0x36, 0xf3, 0xd6, 0x99, 0xd9, 0x04, 0x6f, 0xe1, 0x31, 0xbc, 0x82, 0x37, 0x48, 0x99, 0x32, 0xa5,
	0xce, 0x36, 0x96, 0x1e, 0x41, 0x32, 0x9b, 0x60, 0x6c, 0x7f, 0xf3, 0x87, 0x79, 0x7c, 0xfc, 0xb4,
	0xd0, 0x64, 0x29, 0xd6, 0x94, 0x63, 0x64, 0xec, 0xd0, 0x62, 0x34, 0xd5, 0xd2, 0x62, 0x64, 0x50,
	0x4f, 0x50, 0xc7, 0x9e, 0x52, 0x3d, 0x3d, 0xf3, 0x59, 0xd8, 0x58, 0x76, 0x69, 0x8d, 0xbe, 0x3b,
	0xfa, 0x60, 0x7c, 0x37, 0xa1, 0x1c, 0xfb, 0x4b, 0x4b, 0x06, 0x9d, 0x90, 0x37, 0x7a, 0x34, 0x92,
	0xea, 0x8e, 0xac, 0x7c, 0x7a, 0xbd, 0x9e, 0xa0, 0xb2, 0x9d, 0x16, 0x6f, 0xf6, 0x68, 0x44, 0xa5,
	0xdd, 0xc4, 0x67, 0x1e, 0xde, 0x92, 0x54, 0x57, 0x79, 0x69, 0x2c, 0xea, 0x04, 0x5f, 0x4a, 0x34,
	0x36, 0xec, 0xf2, 0x46, 0x56, 0x4b, 0x6a, 0xb3, 0x22, 0x1d, 0x0a, 0xa1, 0xdb, 0xec, 0x90, 0x75,
	0x77, 0x92, 0xbd, 0x95, 0xdf, 0x67, 0xc5, 0x85, 0x10, 0x3a, 0x3c, 0xe1, 0xcd, 0x75, 0x59, 0xdf,
	0x9a, 0x4a, 0xd1, 0xde, 0xf2, 0xe9, 0xfe, 0xea, 0xa1, 0xef, 0xfd, 0x46, 0x74, 0x0e, 0x78, 0xeb,
	0xdf, 0x5f, 0xa6, 0x20, 0x65, 0xf0, 0x72, 0xb0, 0xf8, 0x82, 0xe0, 0xdd, 0x01, 0x9b, 0x39, 0x60,
	0x73, 0x07, 0xec, 0xd3, 0x01, 0xfb, 0x76, 0x10, 0xfc, 0x38, 0x60, 0x6f, 0x15, 0x04, 0xb3, 0x0a,
	0xd8, 0xbc, 0x82, 0x60, 0x51, 0x41, 0xf0, 0x70, 0x9c, 0x91, 0x54, 0xeb, 0x5d, 0x32, 0x1a, 0x8f,
	0x49, 0xc5, 0x7f, 0xe3, 0x6d, 0x8e, 0xf2, 0xb8, 0xed, 0xfd, 0xfc, 0x37, 0x00, 0x00, 0xff, 0xff,
	0x28, 0x86, 0x41, 0xae, 0x5d, 0x01, 0x00, 0x00,
}

func init() {
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolRoleStateRW.Get().(proto.Message)
	})
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolRoleStateRW_LoginNotifyEvent.Get().(proto.Message)
	})
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolRoleStateRW_LogoutNotifyEvent.Get().(proto.Message)
	})
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolRoleStateRW_JoinClusterRequest.Get().(proto.Message)
	})
	coin_server_common_msgcreate.RegisterNewMessage(func() proto.Message {
		return poolRoleStateRW_JoinClusterResponse.Get().(proto.Message)
	})
}

var poolRoleStateRW = &sync.Pool{New: func() interface{} { return &RoleStateRW{} }}

func (m *RoleStateRW) ReleasePool() { m.Reset(); poolRoleStateRW.Put(m); m = nil }

var poolRoleStateRW_LoginNotifyEvent = &sync.Pool{New: func() interface{} { return &RoleStateRW_LoginNotifyEvent{} }}

func (m *RoleStateRW_LoginNotifyEvent) ReleasePool() {
	m.Reset()
	poolRoleStateRW_LoginNotifyEvent.Put(m)
	m = nil
}

var poolRoleStateRW_LogoutNotifyEvent = &sync.Pool{New: func() interface{} { return &RoleStateRW_LogoutNotifyEvent{} }}

func (m *RoleStateRW_LogoutNotifyEvent) ReleasePool() {
	m.Reset()
	poolRoleStateRW_LogoutNotifyEvent.Put(m)
	m = nil
}

var poolRoleStateRW_JoinClusterRequest = &sync.Pool{New: func() interface{} { return &RoleStateRW_JoinClusterRequest{} }}

func (m *RoleStateRW_JoinClusterRequest) ReleasePool() {
	m.Reset()
	poolRoleStateRW_JoinClusterRequest.Put(m)
	m = nil
}

var poolRoleStateRW_JoinClusterResponse = &sync.Pool{New: func() interface{} { return &RoleStateRW_JoinClusterResponse{} }}

func (m *RoleStateRW_JoinClusterResponse) ReleasePool() {
	m.Reset()
	poolRoleStateRW_JoinClusterResponse.Put(m)
	m = nil
}
func (this *RoleStateRW) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RoleStateRW)
	if !ok {
		that2, ok := that.(RoleStateRW)
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
func (this *RoleStateRW_LoginNotifyEvent) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RoleStateRW_LoginNotifyEvent)
	if !ok {
		that2, ok := that.(RoleStateRW_LoginNotifyEvent)
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
func (this *RoleStateRW_LogoutNotifyEvent) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RoleStateRW_LogoutNotifyEvent)
	if !ok {
		that2, ok := that.(RoleStateRW_LogoutNotifyEvent)
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
func (this *RoleStateRW_JoinClusterRequest) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RoleStateRW_JoinClusterRequest)
	if !ok {
		that2, ok := that.(RoleStateRW_JoinClusterRequest)
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
	if this.ClusterTcpAddr != that1.ClusterTcpAddr {
		return false
	}
	if this.ClusterServerId != that1.ClusterServerId {
		return false
	}
	return true
}
func (this *RoleStateRW_JoinClusterResponse) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RoleStateRW_JoinClusterResponse)
	if !ok {
		that2, ok := that.(RoleStateRW_JoinClusterResponse)
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
func (m *RoleStateRW) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RoleStateRW) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RoleStateRW) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *RoleStateRW_LoginNotifyEvent) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RoleStateRW_LoginNotifyEvent) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RoleStateRW_LoginNotifyEvent) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *RoleStateRW_LogoutNotifyEvent) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RoleStateRW_LogoutNotifyEvent) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RoleStateRW_LogoutNotifyEvent) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *RoleStateRW_JoinClusterRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RoleStateRW_JoinClusterRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RoleStateRW_JoinClusterRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ClusterServerId) > 0 {
		i -= len(m.ClusterServerId)
		copy(dAtA[i:], m.ClusterServerId)
		i = encodeVarintStateRw(dAtA, i, uint64(len(m.ClusterServerId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ClusterTcpAddr) > 0 {
		i -= len(m.ClusterTcpAddr)
		copy(dAtA[i:], m.ClusterTcpAddr)
		i = encodeVarintStateRw(dAtA, i, uint64(len(m.ClusterTcpAddr)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *RoleStateRW_JoinClusterResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RoleStateRW_JoinClusterResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RoleStateRW_JoinClusterResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintStateRw(dAtA []byte, offset int, v uint64) int {
	offset -= sovStateRw(v)
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

func (m *RoleStateRW) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	w.RawByte('}')

}

func (m *RoleStateRW_LoginNotifyEvent) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	w.RawByte('}')

}

func (m *RoleStateRW_LogoutNotifyEvent) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	w.RawByte('}')

}

func (m *RoleStateRW_JoinClusterRequest) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	needWriteComma := false
	if m.ClusterTcpAddr != "" {
		w.RawByte('"')
		w.RawString("cluster_tcp_addr")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.ClusterTcpAddr)
		needWriteComma = true
	}
	if m.ClusterServerId != "" {
		if needWriteComma {
			w.RawByte(',')
		}
		w.RawByte('"')
		w.RawString("cluster_server_id")
		w.RawByte('"')
		w.RawByte(':')
		w.String(m.ClusterServerId)
		needWriteComma = true
	}
	_ = needWriteComma
	w.RawByte('}')

}

func (m *RoleStateRW_JoinClusterResponse) JsonBytes(w *coin_server_common_jwriter.Writer) {
	if m == nil {
		w.RawString("null")
		return
	}

	w.RawByte('{')
	w.RawByte('}')

}

func (m *RoleStateRW) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *RoleStateRW) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *RoleStateRW) GoString() string {
	return m.String()
}

func (m *RoleStateRW_LoginNotifyEvent) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *RoleStateRW_LoginNotifyEvent) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *RoleStateRW_LoginNotifyEvent) GoString() string {
	return m.String()
}

func (m *RoleStateRW_LogoutNotifyEvent) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *RoleStateRW_LogoutNotifyEvent) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *RoleStateRW_LogoutNotifyEvent) GoString() string {
	return m.String()
}

func (m *RoleStateRW_JoinClusterRequest) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *RoleStateRW_JoinClusterRequest) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *RoleStateRW_JoinClusterRequest) GoString() string {
	return m.String()
}

func (m *RoleStateRW_JoinClusterResponse) MarshalJSON() ([]byte, error) {
	w := coin_server_common_jwriter.Writer{Buffer: coin_server_common_buffer.Buffer{Buf: make([]byte, 0, 2048)}}
	m.JsonBytes(&w)
	return w.BuildBytes()
}
func (m *RoleStateRW_JoinClusterResponse) String() string {
	d, _ := m.MarshalJSON()
	return *(*string)(unsafe.Pointer(&d))
}
func (m *RoleStateRW_JoinClusterResponse) GoString() string {
	return m.String()
}

func (m *RoleStateRW) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *RoleStateRW_LoginNotifyEvent) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *RoleStateRW_LogoutNotifyEvent) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *RoleStateRW_JoinClusterRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ClusterTcpAddr)
	if l > 0 {
		n += 1 + l + sovStateRw(uint64(l))
	}
	l = len(m.ClusterServerId)
	if l > 0 {
		n += 1 + l + sovStateRw(uint64(l))
	}
	return n
}

func (m *RoleStateRW_JoinClusterResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovStateRw(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozStateRw(x uint64) (n int) {
	return sovStateRw(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *RoleStateRW) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStateRw
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
			return fmt.Errorf("proto: RoleStateRW: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RoleStateRW: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipStateRw(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthStateRw
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
func (m *RoleStateRW_LoginNotifyEvent) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStateRw
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
			return fmt.Errorf("proto: LoginNotifyEvent: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LoginNotifyEvent: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipStateRw(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthStateRw
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
func (m *RoleStateRW_LogoutNotifyEvent) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStateRw
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
			return fmt.Errorf("proto: LogoutNotifyEvent: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LogoutNotifyEvent: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipStateRw(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthStateRw
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
func (m *RoleStateRW_JoinClusterRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStateRw
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
			return fmt.Errorf("proto: JoinClusterRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: JoinClusterRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClusterTcpAddr", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStateRw
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
				return ErrInvalidLengthStateRw
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthStateRw
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ClusterTcpAddr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClusterServerId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStateRw
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
				return ErrInvalidLengthStateRw
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthStateRw
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ClusterServerId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipStateRw(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthStateRw
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
func (m *RoleStateRW_JoinClusterResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStateRw
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
			return fmt.Errorf("proto: JoinClusterResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: JoinClusterResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipStateRw(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthStateRw
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
func skipStateRw(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowStateRw
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
					return 0, ErrIntOverflowStateRw
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
					return 0, ErrIntOverflowStateRw
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
				return 0, ErrInvalidLengthStateRw
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupStateRw
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthStateRw
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthStateRw        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowStateRw          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupStateRw = fmt.Errorf("proto: unexpected end of group")
)
