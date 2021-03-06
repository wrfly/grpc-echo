// Code generated by protoc-gen-go. DO NOT EDIT.
// source: echo.proto

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	echo.proto

It has these top-level messages:
	Msg
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Msg struct {
	Msg   string `protobuf:"bytes,1,opt,name=msg" json:"msg,omitempty"`
	Sleep int32  `protobuf:"varint,2,opt,name=sleep" json:"sleep,omitempty"`
}

func (m *Msg) Reset()                    { *m = Msg{} }
func (m *Msg) String() string            { return proto.CompactTextString(m) }
func (*Msg) ProtoMessage()               {}
func (*Msg) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Msg) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func (m *Msg) GetSleep() int32 {
	if m != nil {
		return m.Sleep
	}
	return 0
}

func init() {
	proto.RegisterType((*Msg)(nil), "msg")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Echo service

type EchoClient interface {
	Hi(ctx context.Context, in *Msg, opts ...grpc.CallOption) (*Msg, error)
	Sleep(ctx context.Context, in *Msg, opts ...grpc.CallOption) (*Msg, error)
}

type echoClient struct {
	cc *grpc.ClientConn
}

func NewEchoClient(cc *grpc.ClientConn) EchoClient {
	return &echoClient{cc}
}

func (c *echoClient) Hi(ctx context.Context, in *Msg, opts ...grpc.CallOption) (*Msg, error) {
	out := new(Msg)
	err := grpc.Invoke(ctx, "/echo/Hi", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *echoClient) Sleep(ctx context.Context, in *Msg, opts ...grpc.CallOption) (*Msg, error) {
	out := new(Msg)
	err := grpc.Invoke(ctx, "/echo/Sleep", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Echo service

type EchoServer interface {
	Hi(context.Context, *Msg) (*Msg, error)
	Sleep(context.Context, *Msg) (*Msg, error)
}

func RegisterEchoServer(s *grpc.Server, srv EchoServer) {
	s.RegisterService(&_Echo_serviceDesc, srv)
}

func _Echo_Hi_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Msg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EchoServer).Hi(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/echo/Hi",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EchoServer).Hi(ctx, req.(*Msg))
	}
	return interceptor(ctx, in, info, handler)
}

func _Echo_Sleep_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Msg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EchoServer).Sleep(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/echo/Sleep",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EchoServer).Sleep(ctx, req.(*Msg))
	}
	return interceptor(ctx, in, info, handler)
}

var _Echo_serviceDesc = grpc.ServiceDesc{
	ServiceName: "echo",
	HandlerType: (*EchoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Hi",
			Handler:    _Echo_Hi_Handler,
		},
		{
			MethodName: "Sleep",
			Handler:    _Echo_Sleep_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "echo.proto",
}

func init() { proto.RegisterFile("echo.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 113 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4a, 0x4d, 0xce, 0xc8,
	0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x57, 0xd2, 0xe5, 0x62, 0xce, 0x2d, 0x4e, 0x17, 0x12, 0x00,
	0x53, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x60, 0x11, 0x11, 0x2e, 0xd6, 0xe2, 0x9c, 0xd4,
	0xd4, 0x02, 0x09, 0x26, 0x05, 0x46, 0x0d, 0xd6, 0x20, 0x08, 0xc7, 0x48, 0x97, 0x8b, 0x05, 0xa4,
	0x59, 0x48, 0x80, 0x8b, 0xc9, 0x23, 0x53, 0x88, 0x45, 0x2f, 0xb7, 0x38, 0x5d, 0x0a, 0x4c, 0x0a,
	0x09, 0x73, 0xb1, 0x06, 0x83, 0x94, 0x20, 0x0b, 0x3a, 0xb1, 0x44, 0x31, 0x15, 0x24, 0x25, 0xb1,
	0x81, 0xad, 0x32, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0xde, 0x18, 0xfd, 0x60, 0x78, 0x00, 0x00,
	0x00,
}
