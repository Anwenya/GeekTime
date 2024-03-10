package ioc

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/middleware"
	itoken "github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/decorator"
	mprometheus "github.com/Anwenya/GeekTime/webook/pkg/ginx/middleware/prometheus"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/middleware/ratelimit"
	"github.com/Anwenya/GeekTime/webook/pkg/limiter"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(
	middlewares []gin.HandlerFunc,
	userHandler *web.UserHandler,
	wechatHandler *web.OAuth2WechatHandler,
	articleHandler *web.ArticleHandler,
) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	userHandler.RegisterRoutes(server)
	wechatHandler.RegisterRoutes(server)
	articleHandler.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(
	redisClient redis.Cmdable,
	th itoken.TokenHandler,
	l logger.LoggerV1,
) []gin.HandlerFunc {
	// 自定义gin插件
	pb := &mprometheus.Builder{
		Namespace: "GeekTime",
		Subsystem: "webook",
		Name:      "_gin_http",
		Help:      "统计GIN的HTTP接口",
	}

	// gin业务打点
	decorator.InitCounter(
		prometheus.CounterOpts{
			Namespace: "GeekTime",
			Subsystem: "webook",
			Name:      "biz_code",
			Help:      "统计业务状态码",
		},
	)

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
		pb.BuildResponseTime(),
		pb.BuildActiveRequest(),
		middleware.NewLoginTokenMiddlewareBuilder(th).Build(),
	}
}
