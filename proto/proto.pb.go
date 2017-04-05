// Code generated by protoc-gen-go.
// source: proto.proto
// DO NOT EDIT!

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	proto.proto

It has these top-level messages:
	Empty
	ID
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto1.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type ID struct {
	ID string `protobuf:"bytes,1,opt,name=ID,json=iD" json:"ID,omitempty"`
}

func (m *ID) Reset()                    { *m = ID{} }
func (m *ID) String() string            { return proto1.CompactTextString(m) }
func (*ID) ProtoMessage()               {}
func (*ID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ID) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func init() {
	proto1.RegisterType((*Empty)(nil), "proto.Empty")
	proto1.RegisterType((*ID)(nil), "proto.ID")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for LBTest service

type LBTestClient interface {
	Who(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ID, error)
}

type lBTestClient struct {
	cc *grpc.ClientConn
}

func NewLBTestClient(cc *grpc.ClientConn) LBTestClient {
	return &lBTestClient{cc}
}

func (c *lBTestClient) Who(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ID, error) {
	out := new(ID)
	err := grpc.Invoke(ctx, "/proto.LBTest/Who", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for LBTest service

type LBTestServer interface {
	Who(context.Context, *Empty) (*ID, error)
}

func RegisterLBTestServer(s *grpc.Server, srv LBTestServer) {
	s.RegisterService(&_LBTest_serviceDesc, srv)
}

func _LBTest_Who_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LBTestServer).Who(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.LBTest/Who",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LBTestServer).Who(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _LBTest_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.LBTest",
	HandlerType: (*LBTestServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Who",
			Handler:    _LBTest_Who_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto.proto",
}

func init() { proto1.RegisterFile("proto.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 102 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2e, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0x03, 0x93, 0x42, 0xac, 0x60, 0x4a, 0x89, 0x9d, 0x8b, 0xd5, 0x35, 0xb7, 0xa0, 0xa4,
	0x52, 0x49, 0x84, 0x8b, 0xc9, 0xd3, 0x45, 0x88, 0x0f, 0x44, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0x70,
	0x06, 0x31, 0x65, 0xba, 0x18, 0x69, 0x71, 0xb1, 0xf9, 0x38, 0x85, 0xa4, 0x16, 0x97, 0x08, 0x29,
	0x70, 0x31, 0x87, 0x67, 0xe4, 0x0b, 0xf1, 0x40, 0xb4, 0xeb, 0x81, 0x35, 0x49, 0x71, 0x42, 0x79,
	0x9e, 0x2e, 0x4a, 0x0c, 0x49, 0x6c, 0x60, 0xb6, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x61, 0xe4,
	0x27, 0xb9, 0x67, 0x00, 0x00, 0x00,
}