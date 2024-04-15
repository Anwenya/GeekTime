package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type RewardGORMDAO struct {
	db *gorm.DB
}

func NewRewardGORMDAO(db *gorm.DB) RewardDAO {
	return &RewardGORMDAO{db: db}
}

func (rr *RewardGORMDAO) Insert(ctx context.Context, r Reward) (int64, error) {
	now := time.Now().UnixMilli()
	r.CreateTime = now
	r.UpdateTime = now
	err := rr.db.WithContext(ctx).Create(&r).Error
	return r.Id, err
}

func (rr *RewardGORMDAO) GetReward(ctx context.Context, rid int64) (Reward, error) {
	// todo:校验uid
	var r Reward
	err := rr.db.WithContext(ctx).Where("id = ?", rid).First(&r).Error
	return r, err
}

func (rr *RewardGORMDAO) UpdateStatus(ctx context.Context, rid int64, status uint8) error {
	return rr.db.WithContext(ctx).
		Where("id = ?", rid).
		Updates(
			map[string]any{
				"status":      status,
				"update_time": time.Now().UnixMilli(),
			},
		).Error
}
