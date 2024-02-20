package ioc

import (
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/middleware"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/middleware/ratelimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHandler *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	userHandler.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		(&middleware.CorsMiddlewareBuilder{}).Cors(),
		//ratelimit.NewSlideWindowBuilder(redisClient, time.Second, 1).Build(),
		ratelimit.NewTokenBucketBuilder(redisClient, 10, 1).Build(),
		(&middleware.SessionMiddlewareBuilder{}).Session(),
		(&middleware.LoginMiddlewareBuilder{}).CheckLogin(),
	}
}
