// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: network.proto

package transport

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

// GameConfiguratorClient is the client API for GameConfigurator service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GameConfiguratorClient interface {
	GetListOfGames(ctx context.Context, in *GameFilter, opts ...grpc.CallOption) (GameConfigurator_GetListOfGamesClient, error)
	CreateGame(ctx context.Context, opts ...grpc.CallOption) (GameConfigurator_CreateGameClient, error)
	JoinGame(ctx context.Context, opts ...grpc.CallOption) (GameConfigurator_JoinGameClient, error)
}

type gameConfiguratorClient struct {
	cc grpc.ClientConnInterface
}

func NewGameConfiguratorClient(cc grpc.ClientConnInterface) GameConfiguratorClient {
	return &gameConfiguratorClient{cc}
}

func (c *gameConfiguratorClient) GetListOfGames(ctx context.Context, in *GameFilter, opts ...grpc.CallOption) (GameConfigurator_GetListOfGamesClient, error) {
	stream, err := c.cc.NewStream(ctx, &GameConfigurator_ServiceDesc.Streams[0], "/transport.GameConfigurator/GetListOfGames", opts...)
	if err != nil {
		return nil, err
	}
	x := &gameConfiguratorGetListOfGamesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GameConfigurator_GetListOfGamesClient interface {
	Recv() (*ListItem, error)
	grpc.ClientStream
}

type gameConfiguratorGetListOfGamesClient struct {
	grpc.ClientStream
}

func (x *gameConfiguratorGetListOfGamesClient) Recv() (*ListItem, error) {
	m := new(ListItem)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gameConfiguratorClient) CreateGame(ctx context.Context, opts ...grpc.CallOption) (GameConfigurator_CreateGameClient, error) {
	stream, err := c.cc.NewStream(ctx, &GameConfigurator_ServiceDesc.Streams[1], "/transport.GameConfigurator/CreateGame", opts...)
	if err != nil {
		return nil, err
	}
	x := &gameConfiguratorCreateGameClient{stream}
	return x, nil
}

type GameConfigurator_CreateGameClient interface {
	Send(*CreateRequest) error
	Recv() (*CreateResponse, error)
	grpc.ClientStream
}

type gameConfiguratorCreateGameClient struct {
	grpc.ClientStream
}

func (x *gameConfiguratorCreateGameClient) Send(m *CreateRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gameConfiguratorCreateGameClient) Recv() (*CreateResponse, error) {
	m := new(CreateResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gameConfiguratorClient) JoinGame(ctx context.Context, opts ...grpc.CallOption) (GameConfigurator_JoinGameClient, error) {
	stream, err := c.cc.NewStream(ctx, &GameConfigurator_ServiceDesc.Streams[2], "/transport.GameConfigurator/JoinGame", opts...)
	if err != nil {
		return nil, err
	}
	x := &gameConfiguratorJoinGameClient{stream}
	return x, nil
}

type GameConfigurator_JoinGameClient interface {
	Send(*JoinRequest) error
	Recv() (*JoinResponse, error)
	grpc.ClientStream
}

type gameConfiguratorJoinGameClient struct {
	grpc.ClientStream
}

func (x *gameConfiguratorJoinGameClient) Send(m *JoinRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gameConfiguratorJoinGameClient) Recv() (*JoinResponse, error) {
	m := new(JoinResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GameConfiguratorServer is the server API for GameConfigurator service.
// All implementations must embed UnimplementedGameConfiguratorServer
// for forward compatibility
type GameConfiguratorServer interface {
	GetListOfGames(*GameFilter, GameConfigurator_GetListOfGamesServer) error
	CreateGame(GameConfigurator_CreateGameServer) error
	JoinGame(GameConfigurator_JoinGameServer) error
	mustEmbedUnimplementedGameConfiguratorServer()
}

// UnimplementedGameConfiguratorServer must be embedded to have forward compatible implementations.
type UnimplementedGameConfiguratorServer struct {
}

func (UnimplementedGameConfiguratorServer) GetListOfGames(*GameFilter, GameConfigurator_GetListOfGamesServer) error {
	return status.Errorf(codes.Unimplemented, "method GetListOfGames not implemented")
}
func (UnimplementedGameConfiguratorServer) CreateGame(GameConfigurator_CreateGameServer) error {
	return status.Errorf(codes.Unimplemented, "method CreateGame not implemented")
}
func (UnimplementedGameConfiguratorServer) JoinGame(GameConfigurator_JoinGameServer) error {
	return status.Errorf(codes.Unimplemented, "method JoinGame not implemented")
}
func (UnimplementedGameConfiguratorServer) mustEmbedUnimplementedGameConfiguratorServer() {}

// UnsafeGameConfiguratorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GameConfiguratorServer will
// result in compilation errors.
type UnsafeGameConfiguratorServer interface {
	mustEmbedUnimplementedGameConfiguratorServer()
}

func RegisterGameConfiguratorServer(s grpc.ServiceRegistrar, srv GameConfiguratorServer) {
	s.RegisterService(&GameConfigurator_ServiceDesc, srv)
}

func _GameConfigurator_GetListOfGames_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GameFilter)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GameConfiguratorServer).GetListOfGames(m, &gameConfiguratorGetListOfGamesServer{stream})
}

type GameConfigurator_GetListOfGamesServer interface {
	Send(*ListItem) error
	grpc.ServerStream
}

type gameConfiguratorGetListOfGamesServer struct {
	grpc.ServerStream
}

func (x *gameConfiguratorGetListOfGamesServer) Send(m *ListItem) error {
	return x.ServerStream.SendMsg(m)
}

func _GameConfigurator_CreateGame_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GameConfiguratorServer).CreateGame(&gameConfiguratorCreateGameServer{stream})
}

type GameConfigurator_CreateGameServer interface {
	Send(*CreateResponse) error
	Recv() (*CreateRequest, error)
	grpc.ServerStream
}

type gameConfiguratorCreateGameServer struct {
	grpc.ServerStream
}

func (x *gameConfiguratorCreateGameServer) Send(m *CreateResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gameConfiguratorCreateGameServer) Recv() (*CreateRequest, error) {
	m := new(CreateRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _GameConfigurator_JoinGame_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GameConfiguratorServer).JoinGame(&gameConfiguratorJoinGameServer{stream})
}

type GameConfigurator_JoinGameServer interface {
	Send(*JoinResponse) error
	Recv() (*JoinRequest, error)
	grpc.ServerStream
}

type gameConfiguratorJoinGameServer struct {
	grpc.ServerStream
}

func (x *gameConfiguratorJoinGameServer) Send(m *JoinResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gameConfiguratorJoinGameServer) Recv() (*JoinRequest, error) {
	m := new(JoinRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GameConfigurator_ServiceDesc is the grpc.ServiceDesc for GameConfigurator service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GameConfigurator_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "transport.GameConfigurator",
	HandlerType: (*GameConfiguratorServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetListOfGames",
			Handler:       _GameConfigurator_GetListOfGames_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "CreateGame",
			Handler:       _GameConfigurator_CreateGame_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "JoinGame",
			Handler:       _GameConfigurator_JoinGame_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "network.proto",
}