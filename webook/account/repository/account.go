package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/account/domain"
	"github.com/Anwenya/GeekTime/webook/account/repository/cache"
	"github.com/Anwenya/GeekTime/webook/account/repository/dao"
	"time"
)

type accountRepository struct {
	dao   dao.AccountDAO
	cache cache.AccountCache
}

func NewAccountRepository(dao dao.AccountDAO, cache cache.AccountCache) AccountRepository {
	return &accountRepository{dao: dao, cache: cache}
}

func (a *accountRepository) AddCredit(ctx context.Context, c domain.Credit) error {
	activities := make([]dao.AccountActivity, 0, len(c.Items))
	now := time.Now().UnixMilli()

	for _, item := range c.Items {
		activities = append(
			activities,
			dao.AccountActivity{
				Uid:         item.Uid,
				Biz:         c.Biz,
				BizId:       c.BizId,
				Account:     item.Account,
				AccountType: item.AccountType.AsUint8(),
				Amount:      item.Amount,
				Currency:    item.Currency,
				CreateTime:  now,
				UpdateTime:  now,
			},
		)
	}
	return a.dao.AddActivities(ctx, activities...)
}

func (a *accountRepository) CheckUnique(ctx context.Context, c domain.Credit) error {
	return a.cache.GetUnique(ctx, c)
}

func (a *accountRepository) SetUnique(ctx context.Context, c domain.Credit) error {
	return a.cache.SetUnique(ctx, c)
}
