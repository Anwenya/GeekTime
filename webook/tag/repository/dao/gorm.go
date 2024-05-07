package dao

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

type TagGormDao struct {
	db *gorm.DB
}

func NewTagGormDao(db *gorm.DB) TagDAO {
	return &TagGormDao{db: db}
}

func (t *TagGormDao) CreateTag(ctx context.Context, tag Tag) (int64, error) {
	now := time.Now().UnixMilli()
	tag.CreateTime = now
	tag.UpdateTime = now
	err := t.db.WithContext(ctx).Create(&tag).Error
	return tag.Id, err
}

func (t *TagGormDao) CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error {
	if len(tagBiz) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	for _, tag := range tagBiz {
		tag.CreateTime = now
		tag.UpdateTime = now
	}
	return t.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			first := tagBiz[0]
			err := tx.Model(&TagBiz{}).
				Delete("uid = > AND biz = ? AND biz_id = ?", first.Uid, first.Biz, first.BizId).Error
			if err != nil {
				return err
			}
			return tx.Create(&tagBiz).Error
		},
	)
}

func (t *TagGormDao) GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error) {
	var res []Tag
	err := t.db.WithContext(ctx).Where("uid = ?", uid).Find(&res).Error
	return res, err
}

func (t *TagGormDao) GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error) {
	// join查询的写法
	var tagBizs []TagBiz
	err := t.db.WithContext(ctx).
		Model(&TagBiz{}).
		InnerJoins("Tag", t.db.Model(&Tag{})).
		Where("Tag.uid = ? AND biz = ? AND biz_id = ?", uid, biz, bizId).
		Find(&tagBizs).Error
	if err != nil {
		return nil, err
	}
	return slice.Map[TagBiz, Tag](tagBizs, func(idx int, src TagBiz) Tag {
		return *src.Tag
	}), nil
}

func (t *TagGormDao) GetTags(ctx context.Context, offset, limit int) ([]Tag, error) {
	var res []Tag
	err := t.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (t *TagGormDao) GetTagsById(ctx context.Context, ids []int64) ([]Tag, error) {
	var res []Tag
	err := t.db.WithContext(ctx).Where("id IN ?", ids).Find(res).Error
	return res, err
}
