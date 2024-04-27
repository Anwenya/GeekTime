package ioc

import (
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis(l logger.LoggerV1) redis.Cmdable {
	type config struct {
		Address string `yaml:"address"`
	}

	var cfg config

	err := viper.UnmarshalKey("redis", &cfg)
	if err != nil {
		l.Error("读取redis配置失败", logger.Error(err))
		panic(any(err))
	}

	return redis.NewClient(&redis.Options{
		Addr: cfg.Address,
	})
}
