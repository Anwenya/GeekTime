//go:build wireinject

package main

import (
	events2 "github.com/Anwenya/GeekTime/webook/interactive/events"
	repo2 "github.com/Anwenya/GeekTime/webook/interactive/repository"
	cache2 "github.com/Anwenya/GeekTime/webook/interactive/repository/cache"
	dao2 "github.com/Anwenya/GeekTime/webook/interactive/repository/dao"
	service2 "github.com/Anwenya/GeekTime/webook/interactive/service"
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
	"github.com/robfig/cron/v3"
)

type App struct {
	server    *gin.Engine
	consumers []events.Consumer
	cron      *cron.Cron
}

var interactiveServiceSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewRedisInteractiveCache,
	repo2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

var rankingServiceSet = wire.NewSet(
	cache.NewRedisRankingCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
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
		ioc.InitRlockClient,

		// dao
		dao.NewGORMUserDAO,
		dao.NewGORMArticleDAO,
		dao2.NewGORMHistoryDAO,

		interactiveServiceSet,
		rankingServiceSet,
		ioc.InitRankingJob,
		ioc.InitJobs,

		ioc.InitInteractiveClient,

		// 消息
		article.NewSaramaSyncProducer,
		events2.NewInteractiveReadEventConsumer,
		events2.NewHistoryRecordConsumer,
		ioc.InitConsumers,

		// 缓存
		cache.NewRedisCodeCache,
		cache.NewRedisUserCache,
		cache.NewRedisArticleCache,

		// repo
		repository.NewCachedCodeRepository,
		repository.NewCachedUserRepository,
		repository.NewCachedArticleRepository,
		repo2.NewCachedReadHistoryRepository,

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
