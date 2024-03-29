package middleware

import (
	"github.com/Anwenya/GeekTime/webook/config"
	itoken "github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

type LoginTokenMiddlewareBuilder struct {
	itoken.TokenHandler
}

func NewLoginTokenMiddlewareBuilder(th itoken.TokenHandler) *LoginTokenMiddlewareBuilder {
	return &LoginTokenMiddlewareBuilder{
		TokenHandler: th,
	}
}

func (ltmb *LoginTokenMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if strings.HasPrefix(path, "/users/signup") ||
			strings.HasPrefix(path, "/users/login") ||
			strings.HasPrefix(path, "/oauth2") {
			return
		}

		tokenStr := ltmb.ExtractToken(ctx)
		var uc itoken.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(t *jwt.Token) (interface{}, error) {
			return config.Config.SecretKey.Token, nil
		})

		// 非法
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
		// 过期
		if token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
		// 校验是否退出
		err = ltmb.CheckSession(ctx, uc.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("user", uc)
	}
}
