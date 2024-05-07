package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/Anwenya/GeekTime/webook/tag/domain"
	"github.com/Anwenya/GeekTime/webook/tag/repository/cache"
	"github.com/Anwenya/GeekTime/webook/tag/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type TagCachedRepository struct {
	dao   dao.TagDAO
	cache cache.TagCache
	l     logger.LoggerV1
}

func NewTagCachedRepository(dao dao.TagDAO, cache cache.TagCache, l logger.LoggerV1) TagRepository {
	return &TagCachedRepository{dao: dao, cache: cache, l: l}
}

func (t *TagCachedRepository) CreateTag(ctx context.Context, tag domain.Tag) (int64, error) {
	id, err := t.dao.CreateTag(ctx, t.toEntity(tag))
	if err != nil {
		return 0, err
	}

	err = t.cache.Append(ctx, tag.Uid, tag)
	if err != nil {

	}
	return id, nil
}

func (t *TagCachedRepository) BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error {
	return t.dao.CreateTagBiz(ctx,
		slice.Map[int64, dao.TagBiz](tags, func(idx int, src int64) dao.TagBiz {
			return dao.TagBiz{
				Tid:   src,
				BizId: bizId,
				Biz:   biz,
				Uid:   uid,
			}
		}))
}

func (t *TagCachedRepository) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	res, err := t.cache.GetTags(ctx, uid)
	if err == nil {
		return res, nil
	}

	tags, err := t.dao.GetTagsByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	res = slice.Map[dao.Tag, domain.Tag](tags,
		func(idx int, src dao.Tag) domain.Tag {
			return t.toDomain(src)
		})
	err = t.cache.Append(ctx, uid, res...)
	if err != nil {

	}
	return res, nil
}

func (t *TagCachedRepository) GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error) {
	tags, err := t.dao.GetTagsById(ctx, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.Tag, domain.Tag](tags, func(idx int, src dao.Tag) domain.Tag {
		return t.toDomain(src)
	}), nil
}

func (t *TagCachedRepository) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	tags, err := t.dao.GetTagsByBiz(ctx, uid, biz, bizId)
	if err != nil {
		return nil, err
	}

	return slice.Map[dao.Tag, dao.Tag](
		tags,
		func(idx int, src dao.Tag) domain.Tag {
			return t.toDomain(src)
		},
	), nil
}

func (t *TagCachedRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}

func (t *TagCachedRepository) toEntity(tag domain.Tag) dao.Tag {
	return dao.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}
