package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/reward/domain"
	"github.com/Anwenya/GeekTime/webook/reward/repository/cache"
	"github.com/Anwenya/GeekTime/webook/reward/repository/dao"
)

type rewardRepository struct {
	dao   dao.RewardDAO
	cache cache.RewardCache
}

func NewRewardRepository(dao dao.RewardDAO, cache cache.RewardCache) RewardRepository {
	return &rewardRepository{dao: dao, cache: cache}
}

func (rr *rewardRepository) CreateReward(ctx context.Context, reward domain.Reward) (int64, error) {
	return rr.dao.Insert(ctx, rr.toEntity(reward))
}

func (rr *rewardRepository) GetReward(ctx context.Context, rid int64) (domain.Reward, error) {
	r, err := rr.dao.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	return rr.toDomain(r), nil
}

func (rr *rewardRepository) GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	return rr.cache.GetCachedCodeURL(ctx, r)
}

func (rr *rewardRepository) CachedCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error {
	return rr.cache.CachedCodeURL(ctx, cu, r)
}

func (rr *rewardRepository) UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error {
	return rr.dao.UpdateStatus(ctx, rid, status.AsUint8())
}

func (rr *rewardRepository) toEntity(r domain.Reward) dao.Reward {
	return dao.Reward{
		Status:    r.Status.AsUint8(),
		Biz:       r.Target.Biz,
		BizName:   r.Target.BizName,
		BizId:     r.Target.BizId,
		TargetUid: r.Target.Uid,
		Uid:       r.Uid,
		Amount:    r.Amount,
	}
}

func (rr *rewardRepository) toDomain(r dao.Reward) domain.Reward {
	return domain.Reward{
		Id:  r.Id,
		Uid: r.Uid,
		Target: domain.Target{
			Biz:     r.Biz,
			BizId:   r.BizId,
			BizName: r.BizName,
			Uid:     r.Uid,
		},
		Amount: r.Amount,
		Status: domain.RewardStatus(r.Status),
	}
}
