package startup

import (
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: "192.168.133.128:6379",
	})
}
