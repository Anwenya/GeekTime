package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/payment/domain"
	"time"
)

type PaymentRepository interface {
	AddPayment(ctx context.Context, pmt domain.Payment) error
	UpdatePayment(ctx context.Context, pmt domain.Payment) error
	FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error)
	GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error)
}
