package dao

import (
	"context"
	"gorm.io/gorm"
)

type CommentGORMDAO struct {
	db *gorm.DB
}

func NewCommentGORMDAO(db *gorm.DB) CommentDAO {
	return &CommentGORMDAO{db: db}
}

func (c *CommentGORMDAO) Insert(ctx context.Context, comment Comment) error {
	return c.db.WithContext(ctx).Create(&comment).Error
}

func (c *CommentGORMDAO) FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND id < ? AND pid IS NULL", biz, bizId, minId).
		Limit(int(limit)).
		Find(&res).Error
	return res, err
}

// FindCommentList Comment的id为0 获取一级评论 如果不为0获取对应的评论及所有回复
func (c *CommentGORMDAO) FindCommentList(ctx context.Context, comment Comment) ([]Comment, error) {
	var res []Comment
	builder := c.db.WithContext(ctx)
	// 一级评论
	if comment.Id == 0 {
		builder = builder.Where("biz = ? AND biz_id = ? AND root_id is null", comment.Biz, comment.BizID)
	} else {
		builder = builder.Where("root_id = ? OR id = ?", comment.Id, comment.Id)
	}
	err := builder.Find(&res).Error
	return res, err
}

// FindRepliesByPid 查找回复 也就是二级评论
func (c *CommentGORMDAO) FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("pid = ?", pid).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&res).
		Error
	return res, err
}

func (c *CommentGORMDAO) Delete(ctx context.Context, comment Comment) error {
	return c.db.WithContext(ctx).Delete(&Comment{Id: comment.Id}).Error
}

func (c *CommentGORMDAO) FindOneByIds(ctx context.Context, ids []int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("id in ?", ids).
		First(&res).
		Error
	return res, err
}

func (c *CommentGORMDAO) FindRepliesByRid(ctx context.Context, rid int64, id int64, limit int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("root_id = ? AND id > ?", rid, id).
		Order("id ASC").
		Limit(int(limit)).
		Find(&res).
		Error
	return res, err
}
