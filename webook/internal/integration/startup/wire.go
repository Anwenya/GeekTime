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

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,

		// dao
		dao.NewUserDAO,
		dao.NewArticleGORMDAO,

		// 缓存
		cache.NewCodeCache,
		cache.NewUserCache,

		// repo
		repository.NewCachedCodeRepository,
		repository.NewCachedUserRepository,
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

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		dao.NewArticleGORMDAO,
		repository.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}
