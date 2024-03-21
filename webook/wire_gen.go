// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/interactive/events"
	repository2 "github.com/Anwenya/GeekTime/webook/interactive/repository"
	cache2 "github.com/Anwenya/GeekTime/webook/interactive/repository/cache"
	dao2 "github.com/Anwenya/GeekTime/webook/interactive/repository/dao"
	service2 "github.com/Anwenya/GeekTime/webook/interactive/service"
	events2 "github.com/Anwenya/GeekTime/webook/internal/events"
	"github.com/Anwenya/GeekTime/webook/internal/events/article"
	"github.com/Anwenya/GeekTime/webook/internal/ioc"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/robfig/cron/v3"
)

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Injectors from wire.go:

func InitWebServer() *App {
	cmdable := ioc.InitRedis()
	tokenHandler := token.NewRedisTokenHandler(cmdable)
	loggerV1 := ioc.InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, tokenHandler, loggerV1)
	db := ioc.InitDB(loggerV1)
	userDAO := dao.NewGORMUserDAO(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCachedCodeRepository(codeCache)
	smService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smService)
	userHandler := web.NewUserHandler(userService, codeService, tokenHandler)
	wechatService := ioc.InitWechatService(loggerV1)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, tokenHandler)
	articleDAO := dao.NewGORMArticleDAO(db)
	articleCache := cache.NewRedisArticleCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, articleCache, userRepository, loggerV1)
	client := ioc.InitSaramaClient()
	syncProducer := ioc.InitSyncProducer(client)
	producer := article.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	interactiveDAO := dao2.NewGORMInteractiveDAO(db)
	interactiveCache := cache2.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository2.NewCachedInteractiveRepository(interactiveDAO, interactiveCache, loggerV1)
	interactiveService := service2.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(loggerV1, articleService, interactiveService)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	interactiveReadEventConsumer := events.NewInteractiveReadEventConsumer(interactiveRepository, client, loggerV1)
	historyDao := dao2.NewGORMHistoryDAO(db)
	readHistoryRepository := repository2.NewCachedReadHistoryRepository(historyDao)
	historyRecordConsumer := events.NewHistoryRecordConsumer(readHistoryRepository, client, loggerV1)
	v2 := ioc.InitConsumers(interactiveReadEventConsumer, historyRecordConsumer)
	rankingService := service.NewBatchRankingService(interactiveService, articleService)
	rlockClient := ioc.InitRlockClient(cmdable)
	job := ioc.InitRankingJob(rankingService, rlockClient, loggerV1)
	cron := ioc.InitJobs(loggerV1, job)
	app := &App{
		server:    engine,
		consumers: v2,
		cron:      cron,
	}
	return app
}

// wire.go:

type App struct {
	server    *gin.Engine
	consumers []events2.Consumer
	cron      *cron.Cron
}

var interactiveServiceSet = wire.NewSet(dao2.NewGORMInteractiveDAO, cache2.NewRedisInteractiveCache, repository2.NewCachedInteractiveRepository, service2.NewInteractiveService)

var rankingServiceSet = wire.NewSet(cache.NewRedisRankingCache, repository.NewCachedRankingRepository, service.NewBatchRankingService)
