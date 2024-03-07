//go:build wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/internal/events"
	"github.com/Anwenya/GeekTime/webook/internal/events/article"
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

type App struct {
	server    *gin.Engine
	consumers []events.Consumer
}

var interactiveServiceSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitWebServer() *App {
	wire.Build(
		// log
		ioc.InitLogger,

		// 第三方
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,

		// dao
		dao.NewGORMUserDAO,
		dao.NewGORMArticleDAO,

		interactiveServiceSet,

		// 消息
		article.NewSaramaSyncProducer,
		article.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,

		// 缓存
		cache.NewRedisCodeCache,
		cache.NewRedisUserCache,
		cache.NewRedisArticleCache,

		// repo
		repository.NewCachedCodeRepository,
		repository.NewCachedUserRepository,
		repository.NewCachedArticleRepository,

		// service
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewCodeService,
		service.NewUserService,
		service.NewArticleService,

		// handler
		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		itoken.NewRedisTokenHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
