package events

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/Anwenya/GeekTime/webook/pkg/saramax"
	"github.com/Anwenya/GeekTime/webook/reward/service"
	"github.com/IBM/sarama"
	"strings"
	"time"
)

type PaymentEventConsumer struct {
	client sarama.Client
	svc    service.RewardService
	l      logger.LoggerV1
}

func NewPaymentEventConsumer(client sarama.Client, svc service.RewardService, l logger.LoggerV1) saramax.Consumer {
	return &PaymentEventConsumer{client: client, svc: svc, l: l}
}

func (r *PaymentEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("reward", r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(
			context.Background(),
			[]string{"payment_events"},
			saramax.NewHandler[PaymentEvent](r.l, r.Consume),
		)
		if err != nil {
			r.l.Error("消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (r *PaymentEventConsumer) Consume(msg *sarama.ConsumerMessage, evt PaymentEvent) error {
	if !strings.HasPrefix(evt.BizTradeNO, "reward") {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return r.svc.UpdateReward(ctx, evt.BizTradeNO, evt.ToDomainStatus())
}
