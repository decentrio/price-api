// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             (unknown)
// source: user/query.proto

package user

import (
	grpc "google.golang.org/grpc"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

// UserQueryClient is the client API for UserQuery service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserQueryClient interface {
}

type userQueryClient struct {
	cc grpc.ClientConnInterface
}

func NewUserQueryClient(cc grpc.ClientConnInterface) UserQueryClient {
	return &userQueryClient{cc}
}

// UserQueryServer is the server API for UserQuery service.
// All implementations must embed UnimplementedUserQueryServer
// for forward compatibility
type UserQueryServer interface {
	mustEmbedUnimplementedUserQueryServer()
}

// UnimplementedUserQueryServer must be embedded to have forward compatible implementations.
type UnimplementedUserQueryServer struct {
}

func (UnimplementedUserQueryServer) mustEmbedUnimplementedUserQueryServer() {}

// UnsafeUserQueryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserQueryServer will
// result in compilation errors.
type UnsafeUserQueryServer interface {
	mustEmbedUnimplementedUserQueryServer()
}

func RegisterUserQueryServer(s grpc.ServiceRegistrar, srv UserQueryServer) {
	s.RegisterService(&UserQuery_ServiceDesc, srv)
}

// UserQuery_ServiceDesc is the grpc.ServiceDesc for UserQuery service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserQuery_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "user.UserQuery",
	HandlerType: (*UserQueryServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "user/query.proto",
}