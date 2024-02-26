//go:build wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/internal/ioc"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	itoken "github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// log
		ioc.InitLogger,

		// 第三方
		ioc.InitDB,
		ioc.InitRedis,

		// dao
		dao.NewUserDAO,

		// 缓存
		cache.NewCodeCache,
		cache.NewUserCache,

		// repo
		repository.NewCachedCodeRepository,
		repository.NewCachedUserRepository,

		// service
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewCodeService,
		service.NewUserService,

		// handler
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		itoken.NewRedisTokenHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
