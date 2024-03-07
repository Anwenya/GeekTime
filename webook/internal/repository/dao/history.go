package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type HistoryDao interface {
	UpsertReadHistory(ctx context.Context, history ReadHistory) error
}

type GORMHistoryDAO struct {
	db *gorm.DB
}

func NewGORMHistoryDAO(db *gorm.DB) HistoryDao {
	return &GORMHistoryDAO{db: db}
}

func (g GORMHistoryDAO) UpsertReadHistory(ctx context.Context, history ReadHistory) error {
	now := time.Now().UnixMilli()
	history.UpdateTime = now
	history.CreateTime = now
	// 如果之前已经访问过就更新时间
	return g.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(
				map[string]interface{}{
					"read_time":   history.ReadTime,
					"update_time": now,
				},
			),
		},
	).Create(&history).Error

}

type ReadHistory struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Uid        int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId      int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	ReadTime   int64
	CreateTime int64
	UpdateTime int64
}

func (ReadHistory) TableName() string {
	return "read_histories"
}
