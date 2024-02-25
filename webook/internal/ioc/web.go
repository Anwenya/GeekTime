package ioc

import (
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/middleware"
	itoken "github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/middleware/ratelimit"
	"github.com/Anwenya/GeekTime/webook/pkg/limiter"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(
	middlewares []gin.HandlerFunc,
	userHandler *web.UserHandler,
	wechatHandler *web.OAuth2WechatHandler,
) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	userHandler.RegisterRoutes(server)
	wechatHandler.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable, th itoken.TokenHandler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		(&middleware.CorsMiddlewareBuilder{}).Cors(),
		//ratelimit.NewSlideWindowBuilder(redisClient, time.Second, 1).Build(),
		ratelimit.NewBuilder(limiter.NewRedisTokenBucketLimiter(redisClient, 10, 1)).Build(),
		middleware.NewLoginTokenMiddlewareBuilder(th).CheckLogin(),
	}
}
