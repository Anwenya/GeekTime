package dao

import (
	"context"
	"database/sql"
)

type CommentDAO interface {
	Insert(ctx context.Context, comment Comment) error
	// FindByBiz 只查找一级评论
	FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]Comment, error)
	// FindCommentList Comment的id为0 获取一级评论，如果不为0获取对应的评论，和其评论的所有回复
	FindCommentList(ctx context.Context, comment Comment) ([]Comment, error)
	// FindRepliesByPid 查找回复
	FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error)
	// Delete 删除本节点及其对应的子节点
	Delete(ctx context.Context, comment Comment) error
	FindOneByIds(ctx context.Context, ids []int64) ([]Comment, error)
	FindRepliesByRid(ctx context.Context, rid int64, id int64, limit int64) ([]Comment, error)
}

type Comment struct {
	Id int64 `gorm:"autoIncrement,primaryKey"`
	// 发表评论的人
	// 如果需要查询某个人发表的所有的评论 就在这里创建一个索引
	Uid int64
	// 被评价的东西
	Biz     string `gorm:"index:biz_type_id"`
	BizID   int64  `gorm:"index:biz_type_id"`
	Content string

	// 如果这个字段是 NULL 它是根评论
	RootID sql.NullInt64 `gorm:"column:root_id;index"`

	// 这个是 NULL 也是根评论
	PID sql.NullInt64 `gorm:"column:pid;index"`

	// 通过外键来实现级联删除
	ParentComment *Comment `gorm:"ForeignKey:PID;AssociationForeignKey:ID;constraint:OnDelete:CASCADE"`

	CreateTime int64
	UpdateTime int64
}

func (*Comment) TableName() string {
	return "comments"
}
