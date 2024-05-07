//go:build wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"github.com/Anwenya/GeekTime/webook/tag/events"
	"github.com/Anwenya/GeekTime/webook/tag/grpc"
	"github.com/Anwenya/GeekTime/webook/tag/ioc"
	"github.com/Anwenya/GeekTime/webook/tag/repository"
	"github.com/Anwenya/GeekTime/webook/tag/repository/cache"
	"github.com/Anwenya/GeekTime/webook/tag/repository/dao"
	"github.com/Anwenya/GeekTime/webook/tag/service"
	"github.com/google/wire"
)

type App struct {
	server *grpcx.Server
}

func Init() *App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitDB,
		ioc.InitEtcdClient,
		ioc.InitKafka,
		ioc.InitRedis,

		events.NewSaramaSyncProducer,

		cache.NewTagRedisCache,
		dao.NewTagGormDao,
		repository.NewTagCachedRepository,
		service.NewTagService,

		grpc.NewTagServiceServer,
		ioc.InitGrpcxServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
