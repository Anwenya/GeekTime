package article

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/Anwenya/GeekTime/webook/pkg/saramax"
	"github.com/IBM/sarama"
	"time"
)

type HistoryRecordConsumer struct {
	repo   repository.ReadHistoryRepository
	client sarama.Client
	l      logger.LoggerV1
}

func NewHistoryRecordConsumer(
	repo repository.ReadHistoryRepository,
	client sarama.Client,
	l logger.LoggerV1,
) *HistoryRecordConsumer {
	return &HistoryRecordConsumer{repo: repo, client: client, l: l}
}

func (i *HistoryRecordConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(
			context.Background(),
			[]string{TopicReadEvent},
			saramax.NewHandler[ReadEvent](i.l, i.Consume),
		)
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (i *HistoryRecordConsumer) Consume(
	msg *sarama.ConsumerMessage,
	event ReadEvent,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.AddRecord(
		ctx,
		domain.ReadHistory{
			BizId:    event.Aid,
			Biz:      "article",
			Uid:      event.Uid,
			ReadTime: time.UnixMilli(event.ReadTime),
		},
	)
}
