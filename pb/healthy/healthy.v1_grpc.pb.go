// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.27.1
// source: pb/healthy/healthy.v1.proto

package healthy

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
	Healthy_Healthy_FullMethodName = "/ops.codo.notice.healthy.v1.Healthy/Healthy"
)

// HealthyClient is the client API for Healthy service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HealthyClient interface {
	// 健康检查
	Healthy(ctx context.Context, in *HealthyRequest, opts ...grpc.CallOption) (*HealthyReply, error)
}

type healthyClient struct {
	cc grpc.ClientConnInterface
}

func NewHealthyClient(cc grpc.ClientConnInterface) HealthyClient {
	return &healthyClient{cc}
}

func (c *healthyClient) Healthy(ctx context.Context, in *HealthyRequest, opts ...grpc.CallOption) (*HealthyReply, error) {
	out := new(HealthyReply)
	err := c.cc.Invoke(ctx, Healthy_Healthy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HealthyServer is the server API for Healthy service.
// All implementations must embed UnimplementedHealthyServer
// for forward compatibility
type HealthyServer interface {
	// 健康检查
	Healthy(context.Context, *HealthyRequest) (*HealthyReply, error)
	mustEmbedUnimplementedHealthyServer()
}

// UnimplementedHealthyServer must be embedded to have forward compatible implementations.
type UnimplementedHealthyServer struct {
}

func (UnimplementedHealthyServer) Healthy(context.Context, *HealthyRequest) (*HealthyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Healthy not implemented")
}
func (UnimplementedHealthyServer) mustEmbedUnimplementedHealthyServer() {}

// UnsafeHealthyServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HealthyServer will
// result in compilation errors.
type UnsafeHealthyServer interface {
	mustEmbedUnimplementedHealthyServer()
}

func RegisterHealthyServer(s grpc.ServiceRegistrar, srv HealthyServer) {
	s.RegisterService(&Healthy_ServiceDesc, srv)
}

func _Healthy_Healthy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HealthyServer).Healthy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Healthy_Healthy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HealthyServer).Healthy(ctx, req.(*HealthyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Healthy_ServiceDesc is the grpc.ServiceDesc for Healthy service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Healthy_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ops.codo.notice.healthy.v1.Healthy",
	HandlerType: (*HealthyServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Healthy",
			Handler:    _Healthy_Healthy_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/healthy/healthy.v1.proto",
}