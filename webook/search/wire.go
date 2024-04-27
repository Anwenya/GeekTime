//go:build wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"github.com/Anwenya/GeekTime/webook/search/events"
	"github.com/Anwenya/GeekTime/webook/search/grpc"
	"github.com/Anwenya/GeekTime/webook/search/ioc"
	"github.com/Anwenya/GeekTime/webook/search/repository"
	"github.com/Anwenya/GeekTime/webook/search/repository/dao"
	"github.com/Anwenya/GeekTime/webook/search/service"
	"github.com/google/wire"
)

type App struct {
	server    *grpcx.Server
	consumers []events.Consumer
}

func Init() *App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitES,
		ioc.InitEtcdClient,
		ioc.InitGrpcxServer,
		ioc.InitKafka,

		dao.NewArticleElasticDao,
		dao.NewTagEsDao,
		dao.NewUserEsDao,
		dao.NewAnyEsDao,

		repository.NewUserRepository,
		repository.NewArticleRepository,
		repository.NewAnyRepository,

		service.NewSyncService,
		service.NewSearchService,

		events.NewArticleConsumer,
		events.NewUserConsumer,
		ioc.InitConsumer,

		grpc.NewSyncServiceServer,
		grpc.NewSearchServiceServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
