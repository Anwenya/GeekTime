package ratelimit

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"time"
)

// SlideWindowBuilder 滑动窗口限流
//
// interval 窗口大小
// rate 阈值
type SlideWindowBuilder struct {
	prefix   string
	cmd      redis.Cmdable
	interval time.Duration
	rate     int
}

func NewSlideWindowBuilder(cmd redis.Cmdable, interval time.Duration, rate int) *SlideWindowBuilder {
	return &SlideWindowBuilder{
		cmd:      cmd,
		prefix:   "ip-limiter",
		interval: interval,
		rate:     rate,
	}
}

func (builder *SlideWindowBuilder) Prefix(prefix string) *SlideWindowBuilder {
	builder.prefix = prefix
	return builder
}

func (builder *SlideWindowBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := builder.limit(ctx)
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

func (builder *SlideWindowBuilder) limit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", builder.prefix, ctx.ClientIP())
	return builder.cmd.Eval(
		ctx,
		slideWindowLuaScript,
		[]string{key},
		builder.interval.Milliseconds(),
		builder.rate,
		time.Now().UnixMilli(),
	).Bool()
}
