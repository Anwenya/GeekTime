package ioc

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/middleware"
	itoken "github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/middleware/ratelimit"
	"github.com/Anwenya/GeekTime/webook/pkg/limiter"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
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

func InitGinMiddlewares(
	redisClient redis.Cmdable,
	th itoken.TokenHandler,
	l logger.LoggerV1,
) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.NewCorsMiddlewareBuilder().Build(),
		//ratelimit.NewSlideWindowBuilder(redisClient, time.Second, 1).Build(),
		ratelimit.NewBuilder(
			limiter.NewRedisTokenBucketLimiter(
				redisClient,
				10,
				1),
		).Build(),
		middleware.NewLogMiddlewareBuilder(
			func(ctx context.Context, al middleware.AccessLog) {
				l.Debug("", logger.Field{Key: "req", Val: al})
			},
		).AllowReqBody().AllowRespBody().Build(),
		middleware.NewLoginTokenMiddlewareBuilder(th).Build(),
	}
}
