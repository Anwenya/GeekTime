package service

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/account/domain"
	"github.com/Anwenya/GeekTime/webook/account/repository"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
)

type accountService struct {
	repo repository.AccountRepository
	l    logger.LoggerV1
}

func NewAccountService(repo repository.AccountRepository, l logger.LoggerV1) AccountService {
	return &accountService{repo: repo, l: l}
}

func (a *accountService) Credit(ctx context.Context, cr domain.Credit) error {
	err := a.repo.CheckUnique(ctx, cr)
	if err != nil {
		return err
	}

	err = a.repo.AddCredit(ctx, cr)
	if err == nil {
		err := a.repo.SetUnique(ctx, cr)
		if err != nil {
			a.l.Error("设置缓存失败", logger.Error(err))
		}
	}
	return err
}
