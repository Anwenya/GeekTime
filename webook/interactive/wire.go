//go:build wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/interactive/events"
	"github.com/Anwenya/GeekTime/webook/interactive/grpc"
	"github.com/Anwenya/GeekTime/webook/interactive/ioc"
	"github.com/Anwenya/GeekTime/webook/interactive/repository"
	"github.com/Anwenya/GeekTime/webook/interactive/repository/cache"
	"github.com/Anwenya/GeekTime/webook/interactive/repository/dao"
	"github.com/Anwenya/GeekTime/webook/interactive/service"
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"github.com/google/wire"
)

type App struct {
	server    *grpcx.Server
	consumers []events.Consumer
}

var thirdPartySet = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
	ioc.InitSaramaClient,
	ioc.InitRedis,
)

var interactiveServiceSet = wire.NewSet(
	cache.NewRedisInteractiveCache,
	dao.NewGORMInteractiveDAO,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

var historySet = wire.NewSet(
	dao.NewGORMHistoryDAO,
	repository.NewCachedReadHistoryRepository,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,

		interactiveServiceSet,

		historySet,

		grpc.NewInteractiveServiceServer,

		events.NewInteractiveReadEventConsumer,
		events.NewHistoryRecordConsumer,
		ioc.InitConsumers,

		ioc.NewGrpcxServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
