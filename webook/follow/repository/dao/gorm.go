package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type FollowGORMDAO struct {
	db *gorm.DB
}

func NewFollowGORMDAO(db *gorm.DB) FollowDao {
	return &FollowGORMDAO{db: db}
}

// FollowRelationList 关注列表
func (f *FollowGORMDAO) FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error) {
	var res []FollowRelation
	err := f.db.WithContext(ctx).
		Where("follower = ? AND status = ?", follower, FollowRelationStatusActive).
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&res).
		Error
	return res, err
}

// FollowRelationDetail 是否关注某个人
func (f *FollowGORMDAO) FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error) {
	var res FollowRelation
	err := f.db.WithContext(ctx).
		Where("follower = ? AND followee = ? AND status = ?", follower, followee, FollowRelationStatusActive).
		First(&res).
		Error
	return res, err
}

// CreateFollowRelation 关注
func (f *FollowGORMDAO) CreateFollowRelation(ctx context.Context, c FollowRelation) error {
	// insert or update 语义  第一次关注插入 后续更新
	now := time.Now().UnixMilli()
	c.CreateTime = now
	c.UpdateTime = now
	c.Status = FollowRelationStatusActive
	return f.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				// 这代表的是关注了-取消了-再关注了
				"status":      FollowRelationStatusActive,
				"update_time": now,
			}),
		}).Create(&f).Error
	// 在这里更新 FollowStats 的计数（也是 upsert）
}

// UpdateStatus 更新状态
func (f *FollowGORMDAO) UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error {
	return f.db.WithContext(ctx).
		Where("follower = ? AND followee = ?", follower, followee).
		Updates(map[string]any{
			"status":      status,
			"update_time": time.Now().UnixMilli(),
		}).Error
}

// CntFollower 查询关注数
func (f *FollowGORMDAO) CntFollower(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := f.db.WithContext(ctx).
		Select("count(follower)").
		// 如果要是没有额外索引，不用怀疑，全表扫描
		// 可以考虑在 followee 额外创建一个索引
		Where("followee = ? AND status = ?", uid, FollowRelationStatusActive).
		Count(&res).Error
	return res, err
}

// CntFollowee 查询粉丝数
func (f *FollowGORMDAO) CntFollowee(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := f.db.WithContext(ctx).
		Select("count(followee)").
		// <follower, followee>
		Where("follower = ? AND status = ?", uid, FollowRelationStatusActive).
		Count(&res).Error
	return res, err
}
