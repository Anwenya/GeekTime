package grpc

import (
	"context"
	"errors"
	commentv1 "github.com/Anwenya/GeekTime/webook/api/proto/gen/comment/v1"
)

type RateLimitComment struct {
	CommentServiceServer
}

func (c *RateLimitComment) GetMoreReplies(ctx context.Context, req *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	// 限流或降级时
	if ctx.Value("limited") == "true" || ctx.Value("downgrade") == "true" {
		return &commentv1.GetMoreRepliesResponse{}, errors.New("触发限流或降级")
	}
	return c.CommentServiceServer.GetMoreReplies(ctx, req)
}

func (c *RateLimitComment) GetCommentList(ctx context.Context, request *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	// 区分热门或分热门 使用不同的限速规则
	isHotBiz := c.isHotBiz(request.Biz, request.BizId)
	if isHotBiz {
		// 限速400/s
	} else {
		// 限速100/s
	}
	return c.CommentServiceServer.GetCommentList(ctx, request)
}

func (c *RateLimitComment) GetCommentListV1(ctx context.Context, request *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	isHotBiz := c.isHotBiz(request.Biz, request.BizId)
	if !isHotBiz && ctx.Value("downgrade") == "true" {
		return &commentv1.CommentListResponse{}, errors.New("触发限流或降级")
	}
	return c.CommentServiceServer.GetCommentList(ctx, request)
}

func (c *RateLimitComment) CreateComment(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	if ctx.Value("limited") == "true" || ctx.Value("downgrade") == "true" {
		// 限速降级时 可以转异步
		return &commentv1.CreateCommentResponse{}, nil
	}
	err := c.svc.CreateComment(ctx, convertToDomain(request.GetComment()))
	return &commentv1.CreateCommentResponse{}, err
}

func (c *RateLimitComment) isHotBiz(biz string, bizid int64) bool {
	return true
}
