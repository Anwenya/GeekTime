package domain

type Amount struct {
	// 货币
	Currency string
	// 金额
	Total int64
}

type Payment struct {
	Amt Amount
	// 交易id
	BizTradeNO string
	// 描述
	Description string
	// 状态
	Status PaymentStatus

	TxnID string
}

type PaymentStatus uint8

func (p PaymentStatus) AsUint8() uint8 {
	return uint8(p)
}

const (
	PaymentStatusUnknown = iota
	PaymentStatusInit
	// PaymentStatusSuccess 成功
	PaymentStatusSuccess
	// PaymentStatusFailed 失败
	PaymentStatusFailed
	// PaymentStatusRefund 退款
	PaymentStatusRefund
)
