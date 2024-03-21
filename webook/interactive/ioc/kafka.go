package ioc

import (
	"github.com/Anwenya/GeekTime/webook/interactive/config"
	"github.com/Anwenya/GeekTime/webook/interactive/events"
	"github.com/IBM/sarama"
)

func InitSaramaClient() sarama.Client {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(config.Config.Kafka.Address, cfg)
	if err != nil {
		panic(any(err))
	}
	return client
}

func InitConsumers(
	c *events.InteractiveReadEventConsumer,
	c1 *events.HistoryRecordConsumer,
) []events.Consumer {
	return []events.Consumer{c, c1}
}
