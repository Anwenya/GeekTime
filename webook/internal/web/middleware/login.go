package middleware

import (
	"context"
	"encoding/gob"
	"github.com/Anwenya/GeekTime/webook/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"google.golang.org/appengine/log"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func (loginMiddlewareBuilder *LoginMiddlewareBuilder) CheckLogin(config *util.Config) gin.HandlerFunc {
	gob.Register(time.Now())
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
		now := time.Now()
		// 刷新
		const updateTimeKey = "updateTime"
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if !ok || now.Sub(lastUpdateTime) >= time.Second*10 {
			sess.Set(updateTimeKey, now)
			sess.Set("uid", userId)
			err := sess.Save()
			if err != nil {
				log.Errorf(context.Background(), "%v", err)
			}
		}

	}
}
