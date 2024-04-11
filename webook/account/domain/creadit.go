package domain

type Credit struct {
	Biz   string
	BizId int64
	Items []CreditItem
}

type CreditItem struct {
	// 用户id
	Uid int64
	// 账户id
	Account int64
	// 账户类型
	AccountType AccountType
	// 余额
	Amount int64
	// 货币
	Currency string
}

type AccountType uint8

func (a AccountType) AsUint8() uint8 {
	return uint8(a)
}

const (
	AccountTypeUnknown = iota
	AccountTypeReward
	AccountTypeSystem
)
