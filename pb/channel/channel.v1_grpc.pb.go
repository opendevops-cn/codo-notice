// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.27.1
// source: pb/channel/channel.v1.proto

package channel

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
	Channel_ListChannel_FullMethodName        = "/ops.codo.notice.channel.v1.Channel/ListChannel"
	Channel_GetChannel_FullMethodName         = "/ops.codo.notice.channel.v1.Channel/GetChannel"
	Channel_CreateChannel_FullMethodName      = "/ops.codo.notice.channel.v1.Channel/CreateChannel"
	Channel_UpdateChannel_FullMethodName      = "/ops.codo.notice.channel.v1.Channel/UpdateChannel"
	Channel_UpdateChannelBatch_FullMethodName = "/ops.codo.notice.channel.v1.Channel/UpdateChannelBatch"
	Channel_DeleteChannel_FullMethodName      = "/ops.codo.notice.channel.v1.Channel/DeleteChannel"
)

// ChannelClient is the client API for Channel service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChannelClient interface {
	ListChannel(ctx context.Context, in *ListChannelRequest, opts ...grpc.CallOption) (*ListChannelReply, error)
	GetChannel(ctx context.Context, in *GetChannelRequest, opts ...grpc.CallOption) (*ChannelDTO, error)
	CreateChannel(ctx context.Context, in *CreateChannelRequest, opts ...grpc.CallOption) (*ChannelDTO, error)
	UpdateChannel(ctx context.Context, in *UpdateChannelRequest, opts ...grpc.CallOption) (*UpdateChannelReply, error)
	UpdateChannelBatch(ctx context.Context, in *UpdateChannelBatchRequest, opts ...grpc.CallOption) (*UpdateChannelBatchReply, error)
	DeleteChannel(ctx context.Context, in *DeleteChannelRequest, opts ...grpc.CallOption) (*DeleteChannelReply, error)
}

type channelClient struct {
	cc grpc.ClientConnInterface
}

func NewChannelClient(cc grpc.ClientConnInterface) ChannelClient {
	return &channelClient{cc}
}

func (c *channelClient) ListChannel(ctx context.Context, in *ListChannelRequest, opts ...grpc.CallOption) (*ListChannelReply, error) {
	out := new(ListChannelReply)
	err := c.cc.Invoke(ctx, Channel_ListChannel_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelClient) GetChannel(ctx context.Context, in *GetChannelRequest, opts ...grpc.CallOption) (*ChannelDTO, error) {
	out := new(ChannelDTO)
	err := c.cc.Invoke(ctx, Channel_GetChannel_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelClient) CreateChannel(ctx context.Context, in *CreateChannelRequest, opts ...grpc.CallOption) (*ChannelDTO, error) {
	out := new(ChannelDTO)
	err := c.cc.Invoke(ctx, Channel_CreateChannel_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelClient) UpdateChannel(ctx context.Context, in *UpdateChannelRequest, opts ...grpc.CallOption) (*UpdateChannelReply, error) {
	out := new(UpdateChannelReply)
	err := c.cc.Invoke(ctx, Channel_UpdateChannel_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelClient) UpdateChannelBatch(ctx context.Context, in *UpdateChannelBatchRequest, opts ...grpc.CallOption) (*UpdateChannelBatchReply, error) {
	out := new(UpdateChannelBatchReply)
	err := c.cc.Invoke(ctx, Channel_UpdateChannelBatch_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelClient) DeleteChannel(ctx context.Context, in *DeleteChannelRequest, opts ...grpc.CallOption) (*DeleteChannelReply, error) {
	out := new(DeleteChannelReply)
	err := c.cc.Invoke(ctx, Channel_DeleteChannel_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChannelServer is the server API for Channel service.
// All implementations must embed UnimplementedChannelServer
// for forward compatibility
type ChannelServer interface {
	ListChannel(context.Context, *ListChannelRequest) (*ListChannelReply, error)
	GetChannel(context.Context, *GetChannelRequest) (*ChannelDTO, error)
	CreateChannel(context.Context, *CreateChannelRequest) (*ChannelDTO, error)
	UpdateChannel(context.Context, *UpdateChannelRequest) (*UpdateChannelReply, error)
	UpdateChannelBatch(context.Context, *UpdateChannelBatchRequest) (*UpdateChannelBatchReply, error)
	DeleteChannel(context.Context, *DeleteChannelRequest) (*DeleteChannelReply, error)
	mustEmbedUnimplementedChannelServer()
}

// UnimplementedChannelServer must be embedded to have forward compatible implementations.
type UnimplementedChannelServer struct {
}

func (UnimplementedChannelServer) ListChannel(context.Context, *ListChannelRequest) (*ListChannelReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListChannel not implemented")
}
func (UnimplementedChannelServer) GetChannel(context.Context, *GetChannelRequest) (*ChannelDTO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChannel not implemented")
}
func (UnimplementedChannelServer) CreateChannel(context.Context, *CreateChannelRequest) (*ChannelDTO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateChannel not implemented")
}
func (UnimplementedChannelServer) UpdateChannel(context.Context, *UpdateChannelRequest) (*UpdateChannelReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateChannel not implemented")
}
func (UnimplementedChannelServer) UpdateChannelBatch(context.Context, *UpdateChannelBatchRequest) (*UpdateChannelBatchReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateChannelBatch not implemented")
}
func (UnimplementedChannelServer) DeleteChannel(context.Context, *DeleteChannelRequest) (*DeleteChannelReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteChannel not implemented")
}
func (UnimplementedChannelServer) mustEmbedUnimplementedChannelServer() {}

// UnsafeChannelServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChannelServer will
// result in compilation errors.
type UnsafeChannelServer interface {
	mustEmbedUnimplementedChannelServer()
}

func RegisterChannelServer(s grpc.ServiceRegistrar, srv ChannelServer) {
	s.RegisterService(&Channel_ServiceDesc, srv)
}

func _Channel_ListChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelServer).ListChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Channel_ListChannel_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelServer).ListChannel(ctx, req.(*ListChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Channel_GetChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelServer).GetChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Channel_GetChannel_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelServer).GetChannel(ctx, req.(*GetChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Channel_CreateChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelServer).CreateChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Channel_CreateChannel_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelServer).CreateChannel(ctx, req.(*CreateChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Channel_UpdateChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelServer).UpdateChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Channel_UpdateChannel_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelServer).UpdateChannel(ctx, req.(*UpdateChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Channel_UpdateChannelBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateChannelBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelServer).UpdateChannelBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Channel_UpdateChannelBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelServer).UpdateChannelBatch(ctx, req.(*UpdateChannelBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Channel_DeleteChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelServer).DeleteChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Channel_DeleteChannel_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelServer).DeleteChannel(ctx, req.(*DeleteChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Channel_ServiceDesc is the grpc.ServiceDesc for Channel service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Channel_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ops.codo.notice.channel.v1.Channel",
	HandlerType: (*ChannelServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListChannel",
			Handler:    _Channel_ListChannel_Handler,
		},
		{
			MethodName: "GetChannel",
			Handler:    _Channel_GetChannel_Handler,
		},
		{
			MethodName: "CreateChannel",
			Handler:    _Channel_CreateChannel_Handler,
		},
		{
			MethodName: "UpdateChannel",
			Handler:    _Channel_UpdateChannel_Handler,
		},
		{
			MethodName: "UpdateChannelBatch",
			Handler:    _Channel_UpdateChannelBatch_Handler,
		},
		{
			MethodName: "DeleteChannel",
			Handler:    _Channel_DeleteChannel_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/channel/channel.v1.proto",
}
