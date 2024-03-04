package service

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(c context.Context, biz string, id int64, uid int64) error
	CancelLike(c context.Context, biz string, id int64, uid int64) error
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{repo: repo}
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

func (i *interactiveService) Like(c context.Context, biz string, id int64, uid int64) error {
	return i.repo.IncrLike(c, biz, id, uid)
}

func (i *interactiveService) CancelLike(c context.Context, biz string, id int64, uid int64) error {
	return i.repo.DecrLike(c, biz, id, uid)
}
