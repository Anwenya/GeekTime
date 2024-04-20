//go:build wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/follow/grpc"
	"github.com/Anwenya/GeekTime/webook/follow/ioc"
	"github.com/Anwenya/GeekTime/webook/follow/repository"
	"github.com/Anwenya/GeekTime/webook/follow/repository/cache"
	"github.com/Anwenya/GeekTime/webook/follow/repository/dao"
	"github.com/Anwenya/GeekTime/webook/follow/service"
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"github.com/google/wire"
)

type App struct {
	server *grpcx.Server
}

func Init() *App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitEtcdClient,
		ioc.InitGrpcxServer,

		cache.NewFollowRedisCache,
		dao.NewFollowGORMDAO,
		repository.NewFollowRepository,
		service.NewFollowService,
		grpc.NewFollowServiceServer,

		wire.Struct(new(App), "server"),
	)
	return new(App)
}
