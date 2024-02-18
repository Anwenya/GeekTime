package ratelimit

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"time"
)

// TokenBucketBuilder 令牌桶限流
//
// maxTokens:桶中最大令牌数
// tokenCreateRate:每秒生成数
type TokenBucketBuilder struct {
	prefix          string
	cmd             redis.Cmdable
	maxTokens       int
	tokenCreateRate int
}

func NewTokenBucketBuilder(cmd redis.Cmdable, maxTokens int, tokenCreateRate int) *TokenBucketBuilder {
	return &TokenBucketBuilder{
		cmd:             cmd,
		prefix:          "ip-limiter",
		maxTokens:       maxTokens,
		tokenCreateRate: tokenCreateRate,
	}
}

func (builder *TokenBucketBuilder) Prefix(prefix string) *TokenBucketBuilder {
	builder.prefix = prefix
	return builder
}

func (builder *TokenBucketBuilder) Build() gin.HandlerFunc {
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

func (builder *TokenBucketBuilder) limit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", builder.prefix, ctx.ClientIP())
	return builder.cmd.Eval(
		ctx,
		tokenBucketLuaScript,
		[]string{key},
		builder.maxTokens,
		builder.tokenCreateRate,
		time.Now().UnixMilli(),
	).Bool()
}
