//go:build wireinject

package startup

import (
	"github.com/Anwenya/GeekTime/webook/internal/ioc"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方
		ioc.InitDB, ioc.InitRedis,

		// dao
		dao.NewUserDAO,

		// 缓存
		cache.NewCodeCache,
		cache.NewUserCache,

		// repo
		repository.NewCachedCodeRepository,
		repository.NewCachedUserRepository,

		// service
		ioc.InitWechatService,
		ioc.InitSMSService,
		service.NewCodeService,
		service.NewUserService,

		// handler
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
