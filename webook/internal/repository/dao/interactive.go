package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{db: db}
}

func (dao *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(
				map[string]interface{}{
					"read_cnt":    gorm.Expr("`read_cnt` + 1"),
					"update_time": now,
				},
			),
		},
	).Create(
		&Interactive{
			Biz:        biz,
			BizId:      bizId,
			ReadCnt:    1,
			CreateTime: now,
			UpdateTime: now,
		},
	).Error
}

func (dao *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	// 先插入点赞记录 如果已存在将status置为1,代表点赞
	// 之后更新互动量
	return dao.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			err := tx.Clauses(
				clause.OnConflict{
					DoUpdates: clause.Assignments(
						map[string]interface{}{
							"update_time": now,
							"status":      1,
						},
					),
				},
			).Create(
				&UserLikeBiz{
					Uid:        uid,
					Biz:        biz,
					BizId:      id,
					Status:     1,
					UpdateTime: now,
					CreateTime: now,
				},
			).Error

			if err != nil {
				return err
			}

			return tx.WithContext(ctx).Clauses(
				clause.OnConflict{
					DoUpdates: clause.Assignments(
						map[string]interface{}{
							"like_cnt":    gorm.Expr("`like_cnt` + 1"),
							"update_time": now,
						},
					),
				},
			).Create(
				&Interactive{
					Biz:        biz,
					BizId:      id,
					LikeCnt:    1,
					CreateTime: now,
					UpdateTime: now,
				},
			).Error
		},
	)
}

func (dao *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 软删除
			err := tx.Model(&UserLikeBiz{}).
				Where("uid = ? AND biz_id =? AND biz = ?", uid, id, biz).
				Updates(
					map[string]interface{}{
						"update_time": now,
						"status":      0,
					},
				).Error

			if err != nil {
				return err
			}

			return tx.Model(&Interactive{}).
				Where("biz = ? AND biz_id = ?", biz, id).
				Updates(
					map[string]interface{}{
						"like_cnt":    gorm.Expr("`like_cnt` - 1"),
						"update_time": now,
					},
				).Error
		},
	)
}

type UserLikeBiz struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Uid        int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId      int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	Status     int
	UpdateTime int64
	CreateTime int64
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	UpdateTime int64
	CreateTime int64
}
