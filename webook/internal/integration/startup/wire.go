//go:build wireinject

package startup

import (
	repo2 "github.com/Anwenya/GeekTime/webook/interactive/repository"
	cache2 "github.com/Anwenya/GeekTime/webook/interactive/repository/cache"
	dao2 "github.com/Anwenya/GeekTime/webook/interactive/repository/dao"
	service2 "github.com/Anwenya/GeekTime/webook/interactive/service"
	"github.com/Anwenya/GeekTime/webook/internal/events/article"
	"github.com/Anwenya/GeekTime/webook/internal/ioc"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms/async"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	InitLogger,
	InitRedis,
	InitDB,
	InitSaramaClient,
	InitSyncProducer,
)

var userRepoSet = wire.NewSet(
	thirdPartySet,
	dao.NewGORMUserDAO,
	cache.NewRedisUserCache,
	repository.NewCachedUserRepository,
)

var interactiveServiceSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewRedisInteractiveCache,
	repo2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		userRepoSet,
		interactiveServiceSet,
		// dao
		dao.NewGORMArticleDAO,

		// 缓存
		cache.NewRedisCodeCache,
		cache.NewRedisArticleCache,

		// repo
		repository.NewCachedCodeRepository,
		repository.NewCachedArticleRepository,

		article.NewSaramaSyncProducer,

		// service
		ioc.InitWechatService,
		ioc.InitSMSService,
		service.NewCodeService,
		service.NewUserService,
		service.NewArticleService,

		// handler
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		token.NewRedisTokenHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		userRepoSet,
		interactiveServiceSet,
		cache.NewRedisArticleCache,
		repository.NewCachedArticleRepository,
		article.NewSaramaSyncProducer,
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}

func InitAsyncSMSService(svc sms.SMService) *async.Service {
	wire.Build(
		thirdPartySet,
		repository.NewAsyncSMSRepository,
		dao.NewGORMAsyncSMSDAO,
		async.NewService,
	)
	return &async.Service{}
}
