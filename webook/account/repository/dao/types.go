package dao

import "context"

type AccountDAO interface {
	AddActivities(ctx context.Context, activities ...AccountActivity) error
}

// Account 一个账户
type Account struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 用户id
	Uid int64
	// 账户id
	Account int64 `gorm:"uniqueIndex:account_type"`
	// 账户类型
	Type uint8 `gorm:"uniqueIndex:account_type"`
	// 余额
	Balance int64
	// 货币
	Currency string

	UpdateTime int64
	CreateTime int64
}

func (Account) TableName() string {
	return "accounts"
}

// AccountActivity 账户的使用记录
type AccountActivity struct {
	Id  int64 `gorm:"primaryKey,autoIncrement"`
	Uid int64

	// 联合索引可以唯一标识该记录
	Biz   string `gorm:"uniqueIndex:biz_type_id;type:varchar(32)"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id"`

	Account     int64 `gorm:"index:account_type;uniqueIndex:biz_type_id"`
	AccountType uint8 `gorm:"index:account_type;uniqueIndex:biz_type_id"`

	// 入账还是出账
	Amount   int64
	Currency string

	UpdateTime int64
	CreateTime int64
}

func (AccountActivity) TableName() string {
	return "account_activities"
}
