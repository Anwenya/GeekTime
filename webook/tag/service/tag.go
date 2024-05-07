package service

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/Anwenya/GeekTime/webook/tag/domain"
	"github.com/Anwenya/GeekTime/webook/tag/events"
	"github.com/Anwenya/GeekTime/webook/tag/repository"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type tagService struct {
	repo     repository.TagRepository
	logger   logger.LoggerV1
	producer events.Producer
}

func NewTagService(repo repository.TagRepository, logger logger.LoggerV1, producer events.Producer) TagService {
	return &tagService{repo: repo, logger: logger, producer: producer}
}

func (t *tagService) CreateTag(ctx context.Context, uid int64, name string) (int64, error) {
	return t.repo.CreateTag(ctx, domain.Tag{
		Uid:  uid,
		Name: name,
	})
}

func (t *tagService) AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error {
	err := t.repo.BindTagToBiz(ctx, uid, biz, bizId, tags)
	if err != nil {
		return err
	}

	// 异步发送  由搜索服务来消费
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		tags, err := t.repo.GetTagsById(ctx, tags)
		cancel()
		if err != nil {
			return
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		err = t.producer.ProduceSyncEvent(ctx, events.BizTags{
			Biz:   biz,
			BizId: bizId,
			Uid:   uid,
			Tags: slice.Map[domain.Tag](tags, func(idx int, src domain.Tag) string {
				return src.Name
			}),
		})
		cancel()
		if err != nil {
			// 记录一下日志
		}
	}()
	return err

}

func (t *tagService) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	return t.repo.GetTags(ctx, uid)
}

func (t *tagService) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	return t.repo.GetBizTags(ctx, uid, biz, bizId)
}
