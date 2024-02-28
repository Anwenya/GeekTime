package service

import (
	"context"
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
}

type articleService struct {
	repo repository.ArticleRepository

	// 分库写法
	rRepo repository.ArticleReaderRepository
	aRepo repository.ArticleAuthorRepository

	l logger.LoggerV1
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func (a articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

func (a articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

// NewArticleServiceV1 分库写法
func NewArticleServiceV1(
	rRepo repository.ArticleReaderRepository,
	aRepo repository.ArticleAuthorRepository,
	l logger.LoggerV1,
) *articleService {
	return &articleService{
		rRepo: rRepo,
		aRepo: aRepo,
		l:     l,
	}
}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)

	// 操作作者库
	if art.Id > 0 {
		err = a.aRepo.Update(ctx, art)
	} else {
		id, err = a.aRepo.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}

	// 操作读者库
	art.Id = id
	for i := 0; i < 3; i++ {
		// 可能线上库已经有数据了
		// 也可能没有
		err = a.rRepo.Save(ctx, art)
		if err != nil {
			// 多接入一些 tracing 的工具
			a.l.Error("保存到作者库成功但是到读者库失败",
				logger.Int64("aid", art.Id),
				logger.Error(err))
		} else {
			return id, nil
		}
	}
	a.l.Error("保存到作者库成功但是到读者库失败 重试次数耗尽",
		logger.Int64("aid", art.Id),
		logger.Error(err))
	return id, errors.New("保存到读者库失败 重试次数耗尽")
}
