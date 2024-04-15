//go:build wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"github.com/Anwenya/GeekTime/webook/pkg/saramax"
	"github.com/Anwenya/GeekTime/webook/reward/grpc"
	"github.com/Anwenya/GeekTime/webook/reward/ioc"
	"github.com/Anwenya/GeekTime/webook/reward/repository"
	"github.com/Anwenya/GeekTime/webook/reward/repository/cache"
	"github.com/Anwenya/GeekTime/webook/reward/repository/dao"
	"github.com/Anwenya/GeekTime/webook/reward/service"
	"github.com/google/wire"
)

type App struct {
	GRPCServer *grpcx.Server
	Consumer   saramax.Consumer
}

func Init() *App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitRedis,
		ioc.InitDB,
		ioc.InitEtcdClient,
		ioc.InitKafka,
		ioc.InitConsumer,

		dao.NewRewardGORMDAO,
		cache.NewRewardRedisCache,
		repository.NewRewardRepository,
		service.NewWechatNativeRewardService,

		grpc.NewRewardServiceServer,

		ioc.InitAccountClient,
		ioc.InitPaymentClient,
		ioc.InitGrpcxServer,

		wire.Struct(new(App), "GRPCServer", "Consumer"),
	)
	return new(App)
}
