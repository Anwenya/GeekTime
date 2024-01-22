package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
}

func (loginMiddlewareBuilder *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("uid")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
