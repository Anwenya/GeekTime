package ioc

import (
	"github.com/Anwenya/GeekTime/webook/util"
	"github.com/redis/go-redis/v9"
)

func InitRedis(config *util.Config) redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: config.RedisAddress,
	})
}
