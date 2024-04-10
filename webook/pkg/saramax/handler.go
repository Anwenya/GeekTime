package saramax

import (
	"encoding/json"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type Handler[T any] struct {
	fn     func(*sarama.ConsumerMessage, T) error
	l      logger.LoggerV1
	vector *prometheus.SummaryVec
}

func NewHandler[T any](
	consumer string,
	l logger.LoggerV1,
	fn func(*sarama.ConsumerMessage, T) error,
) *Handler[T] {

	vector := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "saramax",
			Subsystem: "consumer_handler",
			Name:      consumer,
		},
		[]string{"topic", "error"})

	return &Handler[T]{fn: fn, l: l, vector: vector}
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	msgChan := claim.Messages()
	for msg := range msgChan {
		h.consumeClaim(msg)
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *Handler[T]) consumeClaim(msg *sarama.ConsumerMessage) {
	start := time.Now()
	var err error
	defer func() {
		errInfo := strconv.FormatBool(err != nil)
		// 主题 和 异常情况
		h.vector.WithLabelValues(msg.Topic, errInfo).Observe(float64(time.Since(start).Milliseconds()))
	}()

	var t T
	err = json.Unmarshal(msg.Value, &t)
	if err != nil {
		h.l.Error(
			"反序列化消息体失败",
			logger.String("topic", msg.Topic),
			logger.Int32("partition", msg.Partition),
			logger.Int64("offset", msg.Offset),
			logger.Error(err),
		)
	}

	err = h.fn(msg, t)
	if err != nil {
		h.l.Error(
			"处理消息失败",
			logger.String("topic", msg.Topic),
			logger.Int32("partition", msg.Partition),
			logger.Int64("offset", msg.Offset),
			logger.Error(err),
		)
	}

}
