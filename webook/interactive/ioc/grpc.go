package ioc

import (
	"github.com/Anwenya/GeekTime/webook/interactive/config"
	localgrpc "github.com/Anwenya/GeekTime/webook/interactive/grpc"
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"google.golang.org/grpc"
)

func NewGrpcxServer(interactive *localgrpc.InteractiveServiceServer) *grpcx.Server {
	server := grpc.NewServer()
	interactive.Register(server)
	return &grpcx.Server{
		Server: server,
		Addr:   config.Config.Grpc.Address,
	}
}
