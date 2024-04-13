package wechat

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/payment/domain"
)

type PaymentService interface {
	// Prepay 预支付 对应于创建微信订单的步骤
	Prepay(ctx context.Context, pmt domain.Payment) (string, error)
}
