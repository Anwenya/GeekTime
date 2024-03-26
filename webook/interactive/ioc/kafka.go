package ioc

import (
	"github.com/Anwenya/GeekTime/webook/interactive/config"
	"github.com/Anwenya/GeekTime/webook/interactive/events"
	"github.com/Anwenya/GeekTime/webook/interactive/repository/dao"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator/events/fixer"
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

func InitSaramaSyncProducer(client sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(any(err))
	}
	return p
}

func InitConsumers(
	c *events.InteractiveReadEventConsumer,
	c1 *events.HistoryRecordConsumer,
	fixConsumer *fixer.Consumer[dao.Interactive],
) []events.Consumer {
	return []events.Consumer{c, c1, fixConsumer}
}
