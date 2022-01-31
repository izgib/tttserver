// Code generated by protoc-gen-go. DO NOT EDIT.
// source: error_details.proto

package rpc_service

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type InterruptionCause int32

const (
	InterruptionCause_DISCONNECT       InterruptionCause = 0
	InterruptionCause_LEAVE            InterruptionCause = 1
	InterruptionCause_OPP_INVALID_MOVE InterruptionCause = 2
	InterruptionCause_INVALID_MOVE     InterruptionCause = 3
)

var InterruptionCause_name = map[int32]string{
	0: "DISCONNECT",
	1: "LEAVE",
	2: "OPP_INVALID_MOVE",
	3: "INVALID_MOVE",
}

var InterruptionCause_value = map[string]int32{
	"DISCONNECT":       0,
	"LEAVE":            1,
	"OPP_INVALID_MOVE": 2,
	"INVALID_MOVE":     3,
}

func (x InterruptionCause) String() string {
	return proto.EnumName(InterruptionCause_name, int32(x))
}

func (InterruptionCause) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_bbac13548d6353a4, []int{0}
}

type InterruptionInfo struct {
	Cause                InterruptionCause `protobuf:"varint,1,opt,name=cause,proto3,enum=errdetails.InterruptionCause" json:"cause,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *InterruptionInfo) Reset()         { *m = InterruptionInfo{} }
func (m *InterruptionInfo) String() string { return proto.CompactTextString(m) }
func (*InterruptionInfo) ProtoMessage()    {}
func (*InterruptionInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_bbac13548d6353a4, []int{0}
}

func (m *InterruptionInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InterruptionInfo.Unmarshal(m, b)
}
func (m *InterruptionInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InterruptionInfo.Marshal(b, m, deterministic)
}
func (m *InterruptionInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InterruptionInfo.Merge(m, src)
}
func (m *InterruptionInfo) XXX_Size() int {
	return xxx_messageInfo_InterruptionInfo.Size(m)
}
func (m *InterruptionInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_InterruptionInfo.DiscardUnknown(m)
}

var xxx_messageInfo_InterruptionInfo proto.InternalMessageInfo

func (m *InterruptionInfo) GetCause() InterruptionCause {
	if m != nil {
		return m.Cause
	}
	return InterruptionCause_DISCONNECT
}

func init() {
	proto.RegisterEnum("errdetails.InterruptionCause", InterruptionCause_name, InterruptionCause_value)
	proto.RegisterType((*InterruptionInfo)(nil), "errdetails.InterruptionInfo")
}

func init() { proto.RegisterFile("error_details.proto", fileDescriptor_bbac13548d6353a4) }

var fileDescriptor_bbac13548d6353a4 = []byte{
	// 199 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4e, 0x2d, 0x2a, 0xca,
	0x2f, 0x8a, 0x4f, 0x49, 0x2d, 0x49, 0xcc, 0xcc, 0x29, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0xe2, 0x4a, 0x2d, 0x2a, 0x82, 0x8a, 0x28, 0xb9, 0x73, 0x09, 0x78, 0xe6, 0x95, 0xa4, 0x16, 0x15,
	0x95, 0x16, 0x94, 0x64, 0xe6, 0xe7, 0x79, 0xe6, 0xa5, 0xe5, 0x0b, 0x19, 0x73, 0xb1, 0x26, 0x27,
	0x96, 0x16, 0xa7, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0xf0, 0x19, 0xc9, 0xea, 0x21, 0xd4, 0xeb, 0x21,
	0x2b, 0x76, 0x06, 0x29, 0x0a, 0x82, 0xa8, 0xd5, 0x0a, 0xe3, 0x12, 0xc4, 0x90, 0x13, 0xe2, 0xe3,
	0xe2, 0x72, 0xf1, 0x0c, 0x76, 0xf6, 0xf7, 0xf3, 0x73, 0x75, 0x0e, 0x11, 0x60, 0x10, 0xe2, 0xe4,
	0x62, 0xf5, 0x71, 0x75, 0x0c, 0x73, 0x15, 0x60, 0x14, 0x12, 0xe1, 0x12, 0xf0, 0x0f, 0x08, 0x88,
	0xf7, 0xf4, 0x0b, 0x73, 0xf4, 0xf1, 0x74, 0x89, 0xf7, 0xf5, 0x0f, 0x73, 0x15, 0x60, 0x12, 0x12,
	0xe0, 0xe2, 0x41, 0x11, 0x61, 0x76, 0x92, 0xe5, 0x92, 0x4e, 0xce, 0xcf, 0xd5, 0x4b, 0xad, 0x48,
	0xcc, 0x2d, 0xc8, 0x49, 0xd5, 0x4b, 0x4f, 0xcc, 0x4d, 0xd5, 0xcb, 0x4b, 0x2d, 0x29, 0xcf, 0x2f,
	0xca, 0xce, 0xcc, 0x4b, 0x4f, 0x62, 0x03, 0x7b, 0xc9, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x2a,
	0xef, 0x15, 0x25, 0xe9, 0x00, 0x00, 0x00,
}
