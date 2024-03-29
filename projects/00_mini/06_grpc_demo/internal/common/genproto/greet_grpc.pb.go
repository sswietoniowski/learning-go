// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: api/protobuf/v1/greet.proto

package genproto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	GreetService_SayHello_FullMethodName                       = "/api.protobuf.v1.GreetService/SayHello"
	GreetService_SayHelloServerStreaming_FullMethodName        = "/api.protobuf.v1.GreetService/SayHelloServerStreaming"
	GreetService_SayHelloClientStreaming_FullMethodName        = "/api.protobuf.v1.GreetService/SayHelloClientStreaming"
	GreetService_SayHelloBidirectionalStreaming_FullMethodName = "/api.protobuf.v1.GreetService/SayHelloBidirectionalStreaming"
)

// GreetServiceClient is the client API for GreetService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GreetServiceClient interface {
	SayHello(ctx context.Context, in *SayHelloRequest, opts ...grpc.CallOption) (*SayHelloResponse, error)
	SayHelloServerStreaming(ctx context.Context, in *SayHelloServerStreamingRequest, opts ...grpc.CallOption) (GreetService_SayHelloServerStreamingClient, error)
	SayHelloClientStreaming(ctx context.Context, opts ...grpc.CallOption) (GreetService_SayHelloClientStreamingClient, error)
	SayHelloBidirectionalStreaming(ctx context.Context, opts ...grpc.CallOption) (GreetService_SayHelloBidirectionalStreamingClient, error)
}

type greetServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGreetServiceClient(cc grpc.ClientConnInterface) GreetServiceClient {
	return &greetServiceClient{cc}
}

func (c *greetServiceClient) SayHello(ctx context.Context, in *SayHelloRequest, opts ...grpc.CallOption) (*SayHelloResponse, error) {
	out := new(SayHelloResponse)
	err := c.cc.Invoke(ctx, GreetService_SayHello_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *greetServiceClient) SayHelloServerStreaming(ctx context.Context, in *SayHelloServerStreamingRequest, opts ...grpc.CallOption) (GreetService_SayHelloServerStreamingClient, error) {
	stream, err := c.cc.NewStream(ctx, &GreetService_ServiceDesc.Streams[0], GreetService_SayHelloServerStreaming_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &greetServiceSayHelloServerStreamingClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GreetService_SayHelloServerStreamingClient interface {
	Recv() (*SayHelloServerStreamingResponse, error)
	grpc.ClientStream
}

type greetServiceSayHelloServerStreamingClient struct {
	grpc.ClientStream
}

func (x *greetServiceSayHelloServerStreamingClient) Recv() (*SayHelloServerStreamingResponse, error) {
	m := new(SayHelloServerStreamingResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *greetServiceClient) SayHelloClientStreaming(ctx context.Context, opts ...grpc.CallOption) (GreetService_SayHelloClientStreamingClient, error) {
	stream, err := c.cc.NewStream(ctx, &GreetService_ServiceDesc.Streams[1], GreetService_SayHelloClientStreaming_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &greetServiceSayHelloClientStreamingClient{stream}
	return x, nil
}

type GreetService_SayHelloClientStreamingClient interface {
	Send(*SayHelloClientStreamingRequest) error
	CloseAndRecv() (*SayHelloClientStreamingResponse, error)
	grpc.ClientStream
}

type greetServiceSayHelloClientStreamingClient struct {
	grpc.ClientStream
}

func (x *greetServiceSayHelloClientStreamingClient) Send(m *SayHelloClientStreamingRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *greetServiceSayHelloClientStreamingClient) CloseAndRecv() (*SayHelloClientStreamingResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(SayHelloClientStreamingResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *greetServiceClient) SayHelloBidirectionalStreaming(ctx context.Context, opts ...grpc.CallOption) (GreetService_SayHelloBidirectionalStreamingClient, error) {
	stream, err := c.cc.NewStream(ctx, &GreetService_ServiceDesc.Streams[2], GreetService_SayHelloBidirectionalStreaming_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &greetServiceSayHelloBidirectionalStreamingClient{stream}
	return x, nil
}

type GreetService_SayHelloBidirectionalStreamingClient interface {
	Send(*SayHelloBidirectionalStreamingRequest) error
	Recv() (*SayHelloBidirectionalStreamingResponse, error)
	grpc.ClientStream
}

type greetServiceSayHelloBidirectionalStreamingClient struct {
	grpc.ClientStream
}

func (x *greetServiceSayHelloBidirectionalStreamingClient) Send(m *SayHelloBidirectionalStreamingRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *greetServiceSayHelloBidirectionalStreamingClient) Recv() (*SayHelloBidirectionalStreamingResponse, error) {
	m := new(SayHelloBidirectionalStreamingResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GreetServiceServer is the server API for GreetService service.
// All implementations must embed UnimplementedGreetServiceServer
// for forward compatibility
type GreetServiceServer interface {
	SayHello(context.Context, *SayHelloRequest) (*SayHelloResponse, error)
	SayHelloServerStreaming(*SayHelloServerStreamingRequest, GreetService_SayHelloServerStreamingServer) error
	SayHelloClientStreaming(GreetService_SayHelloClientStreamingServer) error
	SayHelloBidirectionalStreaming(GreetService_SayHelloBidirectionalStreamingServer) error
	mustEmbedUnimplementedGreetServiceServer()
}

// UnimplementedGreetServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGreetServiceServer struct {
}

func (UnimplementedGreetServiceServer) SayHello(context.Context, *SayHelloRequest) (*SayHelloResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SayHello not implemented")
}
func (UnimplementedGreetServiceServer) SayHelloServerStreaming(*SayHelloServerStreamingRequest, GreetService_SayHelloServerStreamingServer) error {
	return status.Errorf(codes.Unimplemented, "method SayHelloServerStreaming not implemented")
}
func (UnimplementedGreetServiceServer) SayHelloClientStreaming(GreetService_SayHelloClientStreamingServer) error {
	return status.Errorf(codes.Unimplemented, "method SayHelloClientStreaming not implemented")
}
func (UnimplementedGreetServiceServer) SayHelloBidirectionalStreaming(GreetService_SayHelloBidirectionalStreamingServer) error {
	return status.Errorf(codes.Unimplemented, "method SayHelloBidirectionalStreaming not implemented")
}
func (UnimplementedGreetServiceServer) mustEmbedUnimplementedGreetServiceServer() {}

// UnsafeGreetServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GreetServiceServer will
// result in compilation errors.
type UnsafeGreetServiceServer interface {
	mustEmbedUnimplementedGreetServiceServer()
}

func RegisterGreetServiceServer(s grpc.ServiceRegistrar, srv GreetServiceServer) {
	s.RegisterService(&GreetService_ServiceDesc, srv)
}

func _GreetService_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SayHelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreetServiceServer).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GreetService_SayHello_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreetServiceServer).SayHello(ctx, req.(*SayHelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GreetService_SayHelloServerStreaming_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SayHelloServerStreamingRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GreetServiceServer).SayHelloServerStreaming(m, &greetServiceSayHelloServerStreamingServer{stream})
}

type GreetService_SayHelloServerStreamingServer interface {
	Send(*SayHelloServerStreamingResponse) error
	grpc.ServerStream
}

type greetServiceSayHelloServerStreamingServer struct {
	grpc.ServerStream
}

func (x *greetServiceSayHelloServerStreamingServer) Send(m *SayHelloServerStreamingResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _GreetService_SayHelloClientStreaming_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GreetServiceServer).SayHelloClientStreaming(&greetServiceSayHelloClientStreamingServer{stream})
}

type GreetService_SayHelloClientStreamingServer interface {
	SendAndClose(*SayHelloClientStreamingResponse) error
	Recv() (*SayHelloClientStreamingRequest, error)
	grpc.ServerStream
}

type greetServiceSayHelloClientStreamingServer struct {
	grpc.ServerStream
}

func (x *greetServiceSayHelloClientStreamingServer) SendAndClose(m *SayHelloClientStreamingResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *greetServiceSayHelloClientStreamingServer) Recv() (*SayHelloClientStreamingRequest, error) {
	m := new(SayHelloClientStreamingRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _GreetService_SayHelloBidirectionalStreaming_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GreetServiceServer).SayHelloBidirectionalStreaming(&greetServiceSayHelloBidirectionalStreamingServer{stream})
}

type GreetService_SayHelloBidirectionalStreamingServer interface {
	Send(*SayHelloBidirectionalStreamingResponse) error
	Recv() (*SayHelloBidirectionalStreamingRequest, error)
	grpc.ServerStream
}

type greetServiceSayHelloBidirectionalStreamingServer struct {
	grpc.ServerStream
}

func (x *greetServiceSayHelloBidirectionalStreamingServer) Send(m *SayHelloBidirectionalStreamingResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *greetServiceSayHelloBidirectionalStreamingServer) Recv() (*SayHelloBidirectionalStreamingRequest, error) {
	m := new(SayHelloBidirectionalStreamingRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GreetService_ServiceDesc is the grpc.ServiceDesc for GreetService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GreetService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.protobuf.v1.GreetService",
	HandlerType: (*GreetServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _GreetService_SayHello_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SayHelloServerStreaming",
			Handler:       _GreetService_SayHelloServerStreaming_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SayHelloClientStreaming",
			Handler:       _GreetService_SayHelloClientStreaming_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "SayHelloBidirectionalStreaming",
			Handler:       _GreetService_SayHelloBidirectionalStreaming_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "api/protobuf/v1/greet.proto",
}
