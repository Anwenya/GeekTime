package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaProducer(client sarama.Client) (Producer, error) {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	return &SaramaProducer{
		producer: producer,
	}, nil
}

func (s *SaramaProducer) ProducePaymentEvent(ctx context.Context, evt PaymentEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	_, _, err = s.producer.SendMessage(
		&sarama.ProducerMessage{
			Key:   sarama.StringEncoder(evt.BizTradeNO),
			Topic: evt.Topic(),
			Value: sarama.ByteEncoder(data),
		},
	)
	return err
}