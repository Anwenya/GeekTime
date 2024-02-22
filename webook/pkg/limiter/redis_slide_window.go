package limiter

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

// RedisSlideWindowLimiter 滑动窗口限流
//
// interval 窗口大小
// rate 阈值
type RedisSlideWindowLimiter struct {
	cmd      redis.Cmdable
	interval time.Duration
	rate     int
}

func NewRedisSlideWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSlideWindowLimiter {
	return &RedisSlideWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (rswl *RedisSlideWindowLimiter) Limit(ctx *gin.Context, key string) (bool, error) {
	return rswl.cmd.Eval(
		ctx,
		slideWindowLuaScript,
		[]string{key},
		rswl.interval.Milliseconds(),
		rswl.rate,
		time.Now().UnixMilli(),
	).Bool()
}
