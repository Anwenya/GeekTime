package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type AccountGORMDAO struct {
	db *gorm.DB
}

func NewAccountGORMDAO(db *gorm.DB) AccountDAO {
	return &AccountGORMDAO{db: db}
}

func (a *AccountGORMDAO) AddActivities(ctx context.Context, activities ...AccountActivity) error {
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()

		// 按正常业务流程 在执行此处的相关操作之前就已经为该用户创建了对应的账户

		for _, activity := range activities {
			err := a.db.Clauses(clause.OnConflict{
				// 根据活动记录更新账户余额
				DoUpdates: clause.Assignments(map[string]interface{}{
					"balance":     gorm.Expr("`balance` + ?", activity.Amount),
					"update_time": now,
				}),
				// 账户不存在时创建账户
			}).Create(&Account{
				Uid:        activity.Uid,
				Account:    activity.Account,
				Type:       activity.AccountType,
				Balance:    activity.Amount,
				Currency:   activity.Currency,
				UpdateTime: now,
				CreateTime: now,
			}).Error
			if err != nil {
				return err
			}
		}

		// 插入活动记录
		return tx.Create(&activities).Error
	})
}
