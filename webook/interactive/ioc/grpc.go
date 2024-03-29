package ioc

import (
	"github.com/Anwenya/GeekTime/webook/interactive/config"
	localgrpc "github.com/Anwenya/GeekTime/webook/interactive/grpc"
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"google.golang.org/grpc"
)

func NewGrpcxServer(interactive *localgrpc.InteractiveServiceServer, l logger.LoggerV1) *grpcx.Server {
	server := grpc.NewServer()
	interactive.Register(server)
	return &grpcx.Server{
		Server:   server,
		EtcdAddr: config.Config.Grpc.EtcdAddr,
		Port:     config.Config.Grpc.Port,
		Name:     config.Config.Grpc.Name,
		L:        l,
	}
}
