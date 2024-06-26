package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/interactive/domain"
	"github.com/Anwenya/GeekTime/webook/interactive/repository/cache"
	"github.com/Anwenya/GeekTime/webook/interactive/repository/dao"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.LoggerV1
}

func NewCachedInteractiveRepository(
	dao dao.InteractiveDAO,
	cache cache.InteractiveCache,
	l logger.LoggerV1,
) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (c *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	intrs, err := c.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.Interactive, domain.Interactive](
		intrs,
		func(idx int, src dao.Interactive) domain.Interactive {
			return c.toDomain(src)
		},
	), nil
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	// 更新缓存失败会造成数据与缓存不一致
	// 从实际使用上来说无关紧要
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error {
	err := c.dao.BatchIncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}

	// 更新缓存
	// todo 改为管道
	go func() {
		for i := 0; i < len(biz); i++ {
			err := c.cache.IncrReadCntIfPresent(ctx, biz[i], bizId[i])
			if err != nil {
				c.l.Error(
					"回写缓存失败",
					logger.String("biz", biz[i]),
					logger.Int64("bizId", bizId[i]),
					logger.Error(err),
				)
			}
		}
	}()
	return nil
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.InsertLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	// 更新缓存失败会造成数据与缓存不一致
	// 从实际使用上来说无关紧要

	// 更新文章互动量缓存中的点赞数
	err = c.cache.IncrLikeCntIfPresent(ctx, biz, id)
	if err != nil {
		return err
	}
	// 增加文章排行榜的文章点赞数
	err = c.cache.IncrRankingIfPresent(ctx, biz, id)
	if err == cache.ErrRankingUpdate {
		// 从互动量缓存拿到文章点赞数
		val, err := c.dao.Get(ctx, biz, id)
		if err != nil {
			return err
		}
		// 设置点赞数到文章排行缓存
		return c.cache.SetRankingScore(ctx, biz, id, val.LikeCnt)
	}
	return err
}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	// 更新缓存失败会造成数据与缓存不一致
	// 从实际使用上来说无关紧要
	return c.cache.DecrLikeCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(
		ctx,
		dao.UserCollectionBiz{
			Biz:   biz,
			BizId: id,
			Cid:   cid,
			Uid:   uid,
		},
	)
	if err != nil {
		return err
	}

	return c.cache.IncrCollectCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	interactive, err := c.cache.Get(ctx, biz, id)
	if err == nil {
		return interactive, nil
	}
	ie, err := c.dao.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, err
	}

	if err == nil {
		res := c.toDomain(ie)
		err = c.cache.Set(ctx, biz, id, res)
		if err != nil {
			c.l.Error(
				"回写缓存失败",
				logger.String("biz", biz),
				logger.Int64("bizId", id),
				logger.Error(err),
			)
		}
		return res, nil
	}
	return interactive, err
}

func (c *CachedInteractiveRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) LikeTop(ctx context.Context, biz string) ([]domain.Interactive, error) {
	return c.cache.LikeTop(ctx, biz, 100)
}

func (c *CachedInteractiveRepository) toDomain(ie dao.Interactive) domain.Interactive {
	return domain.Interactive{
		ReadCnt:    ie.ReadCnt,
		LikeCnt:    ie.LikeCnt,
		CollectCnt: ie.CollectCnt,
	}
}
