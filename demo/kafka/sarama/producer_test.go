package sarama

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

var addr = []string{"192.168.2.128:9094"}

// kafka-console-consumer -topic=test_topic -brokers=192.168.2.128:9094

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	// If enabled, successfully delivered messages will be returned on the
	// successes channel (default disabled).
	cfg.Producer.Return.Successes = true
	// walks through the available partitions one at a time.
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner

	producer, err := sarama.NewSyncProducer(addr, cfg)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		_, _, err := producer.SendMessage(
			&sarama.ProducerMessage{
				Topic: "test_topic",
				Value: sarama.StringEncoder("这是消息"),
				Headers: []sarama.RecordHeader{
					{
						Key:   []byte("key1"),
						Value: []byte("value1"),
					},
				},
				Metadata: "这是metadata",
			},
		)

		assert.NoError(t, err)
	}
}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	// If enabled, successfully delivered messages will be returned on the
	// successes channel (default disabled).
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	// walks through the available partitions one at a time.
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner

	producer, err := sarama.NewAsyncProducer(addr, cfg)
	assert.NoError(t, err)

	msgChan := producer.Input()

	msgChan <- &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("这是消息"),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("key1"),
				Value: []byte("value1"),
			},
		},
		Metadata: "这是metadata",
	}

	select {
	case msg := <-producer.Successes():
		t.Log("发送成功", string(msg.Value.(sarama.StringEncoder)))
	case err := <-producer.Errors():
		t.Log("发送失败", err.Err, err.Msg)
	}
}
