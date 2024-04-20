package service

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/comment/domain"
)

type CommentService interface {
	// GetCommentList Comment的id为0 获取一级评论
	// 按照 ID 倒序排序
	GetCommentList(ctx context.Context, biz string, bizId, minId, limit int64) ([]domain.Comment, error)
	// DeleteComment 删除评论及子评论
	DeleteComment(ctx context.Context, id int64) error
	// CreateComment 创建评论
	CreateComment(ctx context.Context, comment domain.Comment) error
	GetMoreReplies(ctx context.Context, rid int64, maxId int64, limit int64) ([]domain.Comment, error)
}
