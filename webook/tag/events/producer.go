package events

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaSyncProducer(client sarama.Client) Producer {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil
	}
	return &SaramaSyncProducer{producer: producer}
}

func (p *SaramaSyncProducer) ProduceSyncEvent(ctx context.Context, tags BizTags) error {
	val, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	evt := SyncDataEvent{
		IndexName: "tags_index",
		DocID:     fmt.Sprintf("%d_%s_%d", tags.Uid, tags.Biz, tags.BizId),
		Data:      string(val),
	}
	val, err = json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "sync_search_data",
		Value: sarama.ByteEncoder(val),
	})
	return err
}

type BizTags struct {
	Tags  []string `json:"tags"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Uid   int64    `json:"uid"`
}
