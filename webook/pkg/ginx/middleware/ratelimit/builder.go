package ratelimit

import (
	"fmt"
	"github.com/Anwenya/GeekTime/webook/pkg/limiter"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Builder struct {
	prefix  string
	limiter limiter.Limiter
}

func NewBuilder(limiter limiter.Limiter) *Builder {
	return &Builder{
		prefix:  "ip-limiter",
		limiter: limiter,
	}
}

func (builder *Builder) Prefix(prefix string) *Builder {
	builder.prefix = prefix
	return builder
}

func (builder *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := builder.limiter.Limit(ctx, fmt.Sprintf("%s:%s", builder.prefix, ctx.ClientIP()))
		if err != nil {
			log.Printf("限流失败:%v", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			// 尽量服务正常的用户 可以选择放行
			// ctx.Next()
			return
		}
		if limited {
			log.Printf("限流成功:%s 超出限制", ctx.ClientIP())
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}
