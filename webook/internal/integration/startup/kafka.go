package startup

import "github.com/IBM/sarama"

func InitSaramaClient() sarama.Client {
	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"192.168.2.128:9094"}, scfg)
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
