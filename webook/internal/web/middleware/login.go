package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/token"
	"github.com/Anwenya/GeekTime/webook/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func (loginMiddlewareBuilder *LoginMiddlewareBuilder) CheckLogin(config *util.Config, tokenMaker token.Maker) gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if strings.HasPrefix(path, "/users/signup") || strings.HasPrefix(path, "/users/login") {
			return
		}

		if ok := checkLoginWithSession(ctx, config); ok {
			ctx.Next()
			return
		}

		if ok := checkLoginWithToken(ctx, config, tokenMaker); ok {
			ctx.Next()
			return
		}

		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}

func checkLoginWithSession(ctx *gin.Context, config *util.Config) bool {
	sess := sessions.Default(ctx)
	userId := sess.Get("uid")
	if userId == nil {
		fmt.Printf("%v", "未解析到uid")
		return false
	}
	// 因为是用redis存储的
	// 能解析到uid说明该session未过期
	now := time.Now()
	// 尝试刷新session
	const updateTimeKey = "updateTime"
	val := sess.Get(updateTimeKey)
	lastUpdateTime, ok := val.(time.Time)
	if !ok || now.Sub(lastUpdateTime) >= time.Minute {
		sess.Set(updateTimeKey, now)
		sess.Set("uid", userId)
		err := sess.Save()
		if err != nil {
			fmt.Printf("刷新session失败:%v", err)
		}
	}
	// 设置uid用于下文
	ctx.Set("uid", userId)
	return true
}

func checkLoginWithToken(ctx *gin.Context, config *util.Config, tokenMaker token.Maker) bool {
	authCode := ctx.GetHeader("Authorization")
	// 没有认证头
	if authCode == "" {
		return false
	}
	segs := strings.Split(authCode, " ")
	// 无效
	if len(segs) != 2 {
		return false
	}
	tokenStr := segs[1]
	// 过期或者其他原因导致的校验失败
	payload, err := tokenMaker.VerifyToken(tokenStr)
	if err != nil {
		return false
	}
	// 到这里已经是校验成功了
	// 尝试刷新token
	if payload.ExpiresAt.Sub(time.Now()) < time.Minute {
		newToken, _, err := tokenMaker.CreateToken(payload.Uid, payload.Username, config.AccessTokenDuration)
		if err != nil {
			fmt.Printf("刷新token失败:%v", err)
		}
		ctx.Header(config.TokenKey, newToken)
	}
	// 设置uid用于下文
	ctx.Set("uid", payload.Uid)
	return true
}
