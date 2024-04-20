//go:build wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/comment/grpc"
	"github.com/Anwenya/GeekTime/webook/comment/ioc"
	"github.com/Anwenya/GeekTime/webook/comment/repository"
	"github.com/Anwenya/GeekTime/webook/comment/repository/dao"
	"github.com/Anwenya/GeekTime/webook/comment/service"
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
		ioc.InitEtcdClient,
		ioc.InitGrpcxServer,

		dao.NewCommentGORMDAO,
		repository.NewCommentRepository,
		service.NewCommentService,
		grpc.NewCommentServiceServer,

		wire.Struct(new(App), "server"),
	)
	return new(App)
}
