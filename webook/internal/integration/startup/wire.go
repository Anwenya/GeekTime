//go:build wireinject

package startup

import (
	"github.com/Anwenya/GeekTime/webook/internal/ioc"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	InitLogger,
	InitRedis,
	InitDB,
)

var userRepoSet = wire.NewSet(
	thirdPartySet,
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		userRepoSet,

		// dao
		dao.NewArticleGORMDAO,

		// 缓存
		cache.NewCodeCache,
		cache.NewArticleRedisCache,

		// repo
		repository.NewCachedCodeRepository,
		repository.NewCachedArticleRepository,

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
		cache.NewArticleRedisCache,
		repository.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}
