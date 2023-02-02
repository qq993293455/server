// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/models/edge.proto

package models

import (
	coin_server_common_proto_jsonany "coin-server/common/proto/jsonany"
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	math "math"
	strconv "strconv"
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

type EdgeType int32

const (
	EdgeType_StatelessServer EdgeType = 0
	EdgeType_StaticServer    EdgeType = 1
	EdgeType_DynamicServer   EdgeType = 2
)

var EdgeType_name = map[int32]string{
	0: "StatelessServer",
	1: "StaticServer",
	2: "DynamicServer",
}

var EdgeType_value = map[string]int32{
	"StatelessServer": 0,
	"StaticServer":    1,
	"DynamicServer":   2,
}

func (EdgeType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_f55364a69f959c73, []int{0}
}

func init() {
	proto.RegisterEnum("models.EdgeType", EdgeType_name, EdgeType_value)
}

func init() { proto.RegisterFile("proto/models/edge.proto", fileDescriptor_f55364a69f959c73) }

var fileDescriptor_f55364a69f959c73 = []byte{
	// 179 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2f, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0xcf, 0xcd, 0x4f, 0x49, 0xcd, 0x29, 0xd6, 0x4f, 0x4d, 0x49, 0x4f, 0xd5, 0x03, 0x8b,
	0x08, 0xb1, 0x41, 0x84, 0xb4, 0x5c, 0xb8, 0x38, 0x5c, 0x53, 0xd2, 0x53, 0x43, 0x2a, 0x0b, 0x52,
	0x85, 0x84, 0xb9, 0xf8, 0x83, 0x4b, 0x12, 0x4b, 0x52, 0x73, 0x52, 0x8b, 0x8b, 0x83, 0x53, 0x8b,
	0xca, 0x52, 0x8b, 0x04, 0x18, 0x84, 0x04, 0xb8, 0x78, 0x40, 0x82, 0x99, 0xc9, 0x50, 0x11, 0x46,
	0x21, 0x41, 0x2e, 0x5e, 0x97, 0xca, 0xbc, 0xc4, 0x5c, 0xb8, 0x10, 0x93, 0x93, 0xef, 0x8d, 0x87,
	0x72, 0x0c, 0x2b, 0x1e, 0xc9, 0x31, 0x9e, 0x78, 0x24, 0xc7, 0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83,
	0x47, 0x72, 0x8c, 0x2f, 0x1e, 0xc9, 0x31, 0x7c, 0x78, 0x24, 0xc7, 0x38, 0xe1, 0xb1, 0x1c, 0xc3,
	0x89, 0xc7, 0x72, 0x8c, 0x17, 0x1e, 0xcb, 0x31, 0xdc, 0x78, 0x2c, 0xc7, 0x10, 0x25, 0x9f, 0x9c,
	0x9f, 0x99, 0xa7, 0x5b, 0x0c, 0xd6, 0xac, 0x9f, 0x9c, 0x9f, 0x9b, 0x9b, 0x9f, 0xa7, 0x8f, 0xec,
	0xce, 0x24, 0x36, 0x30, 0xcf, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x15, 0xdb, 0x5a, 0xe6, 0xbe,
	0x00, 0x00, 0x00,
}

func init() {
}
func (x EdgeType) String() string {
	s, ok := EdgeType_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}

var _ = coin_server_common_proto_jsonany.Any{}