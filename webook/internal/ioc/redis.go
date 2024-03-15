package ioc

import (
	"github.com/Anwenya/GeekTime/webook/config"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Address,
	})
}

func InitRlockClient(client redis.Cmdable) *rlock.Client {
	return rlock.NewClient(client)
}
