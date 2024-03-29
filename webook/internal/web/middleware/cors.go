package middleware

import (
	"github.com/Anwenya/GeekTime/webook/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
)

type CorsMiddlewareBuilder struct {
}

func NewCorsMiddlewareBuilder() *CorsMiddlewareBuilder {
	return &CorsMiddlewareBuilder{}
}

func (corsMiddlewareBuilder *CorsMiddlewareBuilder) Build() gin.HandlerFunc {

	return cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		// 允许前端访问后端响应中带的头部
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: config.Config.Duration.Cors,
	})
}
