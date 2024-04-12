package dao

import (
	"context"
	"database/sql"
	"github.com/Anwenya/GeekTime/webook/payment/domain"
	"time"
)

type PaymentDAO interface {
	Insert(ctx context.Context, pmt Payment) error
	UpdateTxnIDAndStatus(ctx context.Context, bizTradeNO string, txnID string, status domain.PaymentStatus) error
	FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error)
	GetPayment(ctx context.Context, bizTradeNO string) (Payment, error)
}

type Payment struct {
	Id int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	// 金额
	Amount int64
	// 货币
	Currency string
	//
	Description string
	// 业务id
	BizTradeNO string `gorm:"column:biz_trade_no;type:varchar(256);unique"`
	// 第三方支付平台的事务id
	TxnID sql.NullString `gorm:"column:txn_id;type:varchar(128);unique"`

	Status     uint8
	UpdateTime int64
	CreateTime int64
}
