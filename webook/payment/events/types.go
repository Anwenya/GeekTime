package events

import "context"

type Producer interface {
	ProducePaymentEvent(ctx context.Context, evt PaymentEvent)
}

type PaymentEvent struct {
	BizTradeNO string
	Status     uint8
}

func (PaymentEvent) Topic() string {
	return "payment_events"
}
