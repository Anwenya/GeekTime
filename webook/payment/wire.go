//go:build wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/payment/grpc"
	"github.com/Anwenya/GeekTime/webook/payment/ioc"
	"github.com/Anwenya/GeekTime/webook/payment/repository"
	"github.com/Anwenya/GeekTime/webook/payment/repository/dao"
	"github.com/Anwenya/GeekTime/webook/payment/web"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx"
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx"
	"github.com/google/wire"
)

type App struct {
	GRPCServer *grpcx.Server
	WebServer  *ginx.Server
}

func Init() *App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitDB,
		ioc.InitEtcdClient,
		ioc.InitKafka,
		ioc.InitProducer,
		ioc.InitWechatClient,

		dao.NewPaymentGORMDAO,
		repository.NewPaymentRepository,

		ioc.InitWechatNativeService,
		ioc.InitWechatConfig,
		ioc.InitWechatNotifyHandler,

		web.NewWechatHandler,
		ioc.InitGinServer,

		grpc.NewWechatServiceServer,
		ioc.InitGrpcxServer,

		wire.Struct(new(App), "WebServer", "GRPCServer"),
	)
	return new(App)
}
