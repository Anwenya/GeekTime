//go:build wireinject

package startup

import (
	"github.com/Anwenya/GeekTime/webook/account/grpc"
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

func Init() *grpc.AccountServiceServer {
	wire.Build(
		InitLogger,
		InitDB,
		InitRedis,

		cache.NewAccountRedisCache,
		dao.NewAccountGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,
	)
	return new(grpc.AccountServiceServer)
}
