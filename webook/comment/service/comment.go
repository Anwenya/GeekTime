package service

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/comment/domain"
	"github.com/Anwenya/GeekTime/webook/comment/repository"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
)

type commentService struct {
	repo repository.CommentRepository
	l    logger.LoggerV1
}

func NewCommentService(repo repository.CommentRepository, l logger.LoggerV1) CommentService {
	return &commentService{repo: repo, l: l}
}

func (c *commentService) GetCommentList(ctx context.Context, biz string, bizId, minId, limit int64) ([]domain.Comment, error) {
	list, err := c.repo.FindByBiz(ctx, biz, bizId, minId, limit)
	if err != nil {
		return nil, err
	}
	return list, err
}

func (c *commentService) DeleteComment(ctx context.Context, id int64) error {
	return c.repo.DeleteComment(ctx, domain.Comment{Id: id})
}

func (c *commentService) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.repo.CreateComment(ctx, comment)
}

func (c *commentService) GetMoreReplies(ctx context.Context, rid int64, maxId int64, limit int64) ([]domain.Comment, error) {
	return c.repo.GetMoreReplies(ctx, rid, maxId, limit)
}
