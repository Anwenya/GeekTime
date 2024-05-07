package ioc

import (
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

func InitKafka(l logger.LoggerV1) sarama.Client {
	type Config struct {
		Address []string `yaml:"addresses"`
	}

	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		l.Error("读取kafka配置失败", logger.Error(err))
		panic(any(err))
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner

	client, err := sarama.NewClient(cfg.Address, saramaCfg)
	if err != nil {
		l.Error("创建kafka客户端失败", logger.Error(err))
		panic(any(err))
	}
	return client
}
