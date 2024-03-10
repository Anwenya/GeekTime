package ioc

import (
	"github.com/Anwenya/GeekTime/webook/config"
	"github.com/Anwenya/GeekTime/webook/internal/events"
	"github.com/Anwenya/GeekTime/webook/internal/events/article"
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

func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(any(err))
	}
	return p
}

func InitConsumers(
	c *article.InteractiveReadEventConsumer,
	c1 *article.HistoryRecordConsumer,
) []events.Consumer {
	return []events.Consumer{c, c1}
}