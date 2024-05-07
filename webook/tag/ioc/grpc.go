package ioc

import (
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	localgrpc "github.com/Anwenya/GeekTime/webook/tag/grpc"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func InitGrpcxServer(
	tag *localgrpc.TagServiceServer,
	etcdCli *etcdv3.Client,
	l logger.LoggerV1,
) *grpcx.Server {

	type config struct {
		Port int    `yaml:"port"`
		Name string `yaml:"name"`
		TTL  int64  `yaml:"ttl"`
	}

	var cfg config
	err := viper.UnmarshalKey("grpc.server", &cfg)

	if err != nil {
		l.Error("读取grpc配置失败", logger.Error(err))
		panic(any(err))
	}

	server := grpc.NewServer()
	tag.Register(server)
	return &grpcx.Server{
		Server: server,
		Client: etcdCli,
		TTL:    cfg.TTL,
		Port:   cfg.Port,
		Name:   cfg.Name,
		L:      l,
	}
}
