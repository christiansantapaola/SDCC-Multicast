// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package api

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

// ReceiverLCClient is the client API for ReceiverLC service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReceiverLCClient interface {
	Send(ctx context.Context, in *MessageLC, opts ...grpc.CallOption) (*SendReply, error)
}

type receiverLCClient struct {
	cc grpc.ClientConnInterface
}

func NewReceiverLCClient(cc grpc.ClientConnInterface) ReceiverLCClient {
	return &receiverLCClient{cc}
}

func (c *receiverLCClient) Send(ctx context.Context, in *MessageLC, opts ...grpc.CallOption) (*SendReply, error) {
	out := new(SendReply)
	err := c.cc.Invoke(ctx, "/ReceiverLC/Send", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReceiverLCServer is the server API for ReceiverLC service.
// All implementations must embed UnimplementedReceiverLCServer
// for forward compatibility
type ReceiverLCServer interface {
	Send(context.Context, *MessageLC) (*SendReply, error)
	mustEmbedUnimplementedReceiverLCServer()
}

// UnimplementedReceiverLCServer must be embedded to have forward compatible implementations.
type UnimplementedReceiverLCServer struct {
}

func (UnimplementedReceiverLCServer) Send(context.Context, *MessageLC) (*SendReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (UnimplementedReceiverLCServer) mustEmbedUnimplementedReceiverLCServer() {}

// UnsafeReceiverLCServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReceiverLCServer will
// result in compilation errors.
type UnsafeReceiverLCServer interface {
	mustEmbedUnimplementedReceiverLCServer()
}

func RegisterReceiverLCServer(s grpc.ServiceRegistrar, srv ReceiverLCServer) {
	s.RegisterService(&ReceiverLC_ServiceDesc, srv)
}

func _ReceiverLC_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageLC)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReceiverLCServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ReceiverLC/Send",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReceiverLCServer).Send(ctx, req.(*MessageLC))
	}
	return interceptor(ctx, in, info, handler)
}

// ReceiverLC_ServiceDesc is the grpc.ServiceDesc for ReceiverLC service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ReceiverLC_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ReceiverLC",
	HandlerType: (*ReceiverLCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Send",
			Handler:    _ReceiverLC_Send_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "LamportClockService.proto",
}