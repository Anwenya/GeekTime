// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/interactive/events"
	"github.com/Anwenya/GeekTime/webook/interactive/grpc"
	"github.com/Anwenya/GeekTime/webook/interactive/ioc"
	"github.com/Anwenya/GeekTime/webook/interactive/repository"
	"github.com/Anwenya/GeekTime/webook/interactive/repository/cache"
	"github.com/Anwenya/GeekTime/webook/interactive/repository/dao"
	"github.com/Anwenya/GeekTime/webook/interactive/service"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx"
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"github.com/google/wire"
)

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Injectors from wire.go:

func InitApp() *App {
	loggerV1 := ioc.InitLogger()
	srcDB := ioc.InitSrcDB(loggerV1)
	dstDB := ioc.InitDstDB(loggerV1)
	doubleWritePool := ioc.InitDoubleWritePool(srcDB, dstDB, loggerV1)
	db := ioc.InitBizDB(doubleWritePool)
	interactiveDAO := dao.NewGORMInteractiveDAO(db)
	cmdable := ioc.InitRedis()
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, interactiveCache, loggerV1)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	interactiveServiceServer := grpc.NewInteractiveServiceServer(interactiveService)
	server := ioc.NewGrpcxServer(interactiveServiceServer, loggerV1)
	client := ioc.InitSaramaClient()
	interactiveReadEventConsumer := events.NewInteractiveReadEventConsumer(interactiveRepository, client, loggerV1)
	historyDao := dao.NewGORMHistoryDAO(db)
	readHistoryRepository := repository.NewCachedReadHistoryRepository(historyDao)
	historyRecordConsumer := events.NewHistoryRecordConsumer(readHistoryRepository, client, loggerV1)
	consumer := ioc.InitFixerConsumer(client, loggerV1, srcDB, dstDB)
	v := ioc.InitConsumers(interactiveReadEventConsumer, historyRecordConsumer, consumer)
	syncProducer := ioc.InitSaramaSyncProducer(client)
	producer := ioc.InitInteractiveProducer(syncProducer)
	ginxServer := ioc.InitGinxServer(srcDB, dstDB, doubleWritePool, producer, loggerV1)
	app := &App{
		server:      server,
		consumers:   v,
		adminServer: ginxServer,
	}
	return app
}

// wire.go:

type App struct {
	server      *grpcx.Server
	consumers   []events.Consumer
	adminServer *ginx.Server
}

var thirdPartySet = wire.NewSet(ioc.InitLogger, ioc.InitDstDB, ioc.InitSrcDB, ioc.InitDoubleWritePool, ioc.InitBizDB, ioc.InitSaramaClient, ioc.InitSaramaSyncProducer, ioc.InitRedis)

var interactiveServiceSet = wire.NewSet(cache.NewInteractiveRedisCache, dao.NewGORMInteractiveDAO, repository.NewCachedInteractiveRepository, service.NewInteractiveService)

var historySet = wire.NewSet(dao.NewGORMHistoryDAO, repository.NewCachedReadHistoryRepository)
