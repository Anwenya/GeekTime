package ioc

import (
	"github.com/Anwenya/GeekTime/webook/util"
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: util.Config.RedisAddress,
	})
}
