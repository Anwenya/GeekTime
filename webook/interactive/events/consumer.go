package events

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/interactive/domain"
	"github.com/Anwenya/GeekTime/webook/interactive/repository"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/Anwenya/GeekTime/webook/pkg/saramax"
	"github.com/IBM/sarama"
	"time"
)

type InteractiveReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	l      logger.LoggerV1
}

func NewInteractiveReadEventConsumer(
	repo repository.InteractiveRepository,
	client sarama.Client,
	l logger.LoggerV1,
) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		repo:   repo,
		client: client,
		l:      l,
	}
}

func (i *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient(TopicGroupID, i.client)
	if err != nil {
		return err
	}

	go func() {
		err := cg.Consume(
			context.Background(),
			[]string{TopicReadEvent},
			saramax.NewHandler[ReadEvent](i.l, i.Consume),
		)
		if err != nil {
			i.l.Error("退出消费", logger.Error(err))
		}
	}()

	return nil
}

func (i *InteractiveReadEventConsumer) StartBatch() error {
	cg, err := sarama.NewConsumerGroupFromClient(TopicGroupID, i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(
			context.Background(),
			[]string{TopicReadEvent},
			saramax.NewBatchHandler[ReadEvent](i.l, i.BatchConsume),
		)
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (i *InteractiveReadEventConsumer) Consume(
	msg *sarama.ConsumerMessage,
	event ReadEvent,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, Biz, event.Aid)
}

func (i *InteractiveReadEventConsumer) BatchConsume(
	msgs []*sarama.ConsumerMessage,
	events []ReadEvent,
) error {
	bizs := make([]string, 0, len(events))
	bizIds := make([]int64, 0, len(events))

	for _, event := range events {
		bizs = append(bizs, Biz)
		bizIds = append(bizIds, event.Aid)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.BatchIncrReadCnt(ctx, bizs, bizIds)
}

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
