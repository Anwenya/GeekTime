package startup

import (
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func InitRedis(l logger.LoggerV1) redis.Cmdable {
	type config struct {
		Address string `yaml:"address"`
	}

	cfg := config{
		Address: "192.168.2.130:6379",
	}

	return redis.NewClient(&redis.Options{
		Addr: cfg.Address,
	})
}
