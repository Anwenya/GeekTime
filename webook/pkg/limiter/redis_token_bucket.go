package limiter

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

// RedisTokenBucketLimiter 令牌桶限流
//
// maxTokens:桶中最大令牌数
// tokenCreateRate:每秒生成数
type RedisTokenBucketLimiter struct {
	cmd             redis.Cmdable
	maxTokens       int
	tokenCreateRate int
}

func NewRedisTokenBucketLimiter(cmd redis.Cmdable, maxTokens int, tokenCreateRate int) *RedisTokenBucketLimiter {
	return &RedisTokenBucketLimiter{
		cmd:             cmd,
		maxTokens:       maxTokens,
		tokenCreateRate: tokenCreateRate,
	}
}

func (rtbl *RedisTokenBucketLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return rtbl.cmd.Eval(
		ctx,
		tokenBucketLuaScript,
		[]string{key},
		rtbl.maxTokens,
		rtbl.tokenCreateRate,
		time.Now().UnixMilli(),
	).Bool()
}
