package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/follow/domain"
	"github.com/Anwenya/GeekTime/webook/follow/repository/cache"
	"github.com/Anwenya/GeekTime/webook/follow/repository/dao"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
)

type followRepository struct {
	dao   dao.FollowDao
	cache cache.FollowCache
	l     logger.LoggerV1
}

func NewFollowRepository(dao dao.FollowDao, cache cache.FollowCache, l logger.LoggerV1) FollowRepository {
	return &followRepository{dao: dao, cache: cache, l: l}
}

func (f *followRepository) GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error) {
	// 可以考虑在这里缓存关注者列表的第一页
	followerList, err := f.dao.FollowRelationList(ctx, follower, offset, limit)
	if err != nil {
		return nil, err
	}
	return f.genFollowRelationList(followerList), nil
}

func (f *followRepository) genFollowRelationList(followerList []dao.FollowRelation) []domain.FollowRelation {
	res := make([]domain.FollowRelation, 0, len(followerList))
	for _, c := range followerList {
		res = append(res, f.toDomain(c))
	}
	return res
}

func (f *followRepository) FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error) {
	c, err := f.dao.FollowRelationDetail(ctx, follower, followee)
	if err != nil {
		return domain.FollowRelation{}, err
	}
	return f.toDomain(c), nil
}

func (f *followRepository) AddFollowRelation(ctx context.Context, fr domain.FollowRelation) error {
	// 关注
	err := f.dao.CreateFollowRelation(ctx, f.toEntity(fr))
	if err != nil {
		return err
	}
	// 更新缓存里面的关注了多少人 以及有多少粉丝的计数 +1
	return f.cache.Follow(ctx, fr.Follower, fr.Followee)
}

func (f *followRepository) InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error {
	// 取关
	err := f.dao.UpdateStatus(ctx, followee, follower, dao.FollowRelationStatusInactive)
	if err != nil {
		return err
	}
	// -1
	return f.cache.CancelFollow(ctx, follower, followee)
}

func (f *followRepository) GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	// 先看缓存
	res, err := f.cache.StaticsInfo(ctx, uid)
	if err == nil {
		return res, nil
	}
	// 没有我就去数据库里查询
	res.Followers, err = f.dao.CntFollower(ctx, uid)
	if err != nil {
		return domain.FollowStatics{}, err
	}
	res.Followees, err = f.dao.CntFollowee(ctx, uid)
	if err != nil {
		return domain.FollowStatics{}, err
	}
	err = f.cache.SetStaticsInfo(ctx, uid, res)
	if err != nil {
		f.l.Error("设置缓存失败", logger.Error(err))
	}
	return res, nil
}

func (f *followRepository) toDomain(fr dao.FollowRelation) domain.FollowRelation {
	return domain.FollowRelation{
		Followee: fr.Followee,
		Follower: fr.Follower,
	}
}

func (f *followRepository) toEntity(c domain.FollowRelation) dao.FollowRelation {
	return dao.FollowRelation{
		Followee: c.Followee,
		Follower: c.Follower,
	}
}
