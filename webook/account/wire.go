//go:build wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/account/grpc"
	"github.com/Anwenya/GeekTime/webook/account/ioc"
	"github.com/Anwenya/GeekTime/webook/account/repository"
	"github.com/Anwenya/GeekTime/webook/account/repository/cache"
	"github.com/Anwenya/GeekTime/webook/account/repository/dao"
	"github.com/Anwenya/GeekTime/webook/account/service"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx"
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"github.com/Anwenya/GeekTime/webook/pkg/saramax"
	"github.com/google/wire"
)

type App struct {
	server      *grpcx.Server
	consumers   []saramax.Consumer
	adminServer *ginx.Server
}

func Init() *App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitEtcdClient,
		ioc.InitGrpcxServer,

		cache.NewAccountRedisCache,
		dao.NewAccountGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,

		wire.Struct(new(App), "server"),
	)
	return new(App)
}
