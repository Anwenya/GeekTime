package repository

import (
	"context"
	"database/sql"
	"github.com/Anwenya/GeekTime/webook/comment/domain"
	"github.com/Anwenya/GeekTime/webook/comment/repository/dao"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"golang.org/x/sync/errgroup"
	"time"
)

type commentRepository struct {
	dao dao.CommentDAO
	l   logger.LoggerV1
}

func NewCommentRepository(dao dao.CommentDAO, l logger.LoggerV1) CommentRepository {
	return &commentRepository{dao: dao, l: l}
}

func (c *commentRepository) FindByBiz(ctx context.Context, biz string, bizId, minID, limit int64) ([]domain.Comment, error) {
	daoComments, err := c.dao.FindByBiz(ctx, biz, bizId, minID, limit)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Comment, 0, len(daoComments))
	// 查找三条子评论

	// 触发降级就不要查询子评论了
	downgrade := ctx.Value("downgrade") == "true"
	if downgrade {
		return res, nil
	}

	var eg errgroup.Group
	for _, dc := range daoComments {
		cm := c.toDomain(dc)
		res = append(res, cm)
		eg.Go(
			func() error {
				subComments, err := c.dao.FindRepliesByPid(ctx, dc.Id, 0, 3)
				if err != nil {
					return err
				}
				cm.Children = make([]domain.Comment, 0, len(subComments))
				for _, sc := range subComments {
					cm.Children = append(cm.Children, c.toDomain(sc))
				}
				return nil
			},
		)
	}
	return res, eg.Wait()
}

func (c *commentRepository) DeleteComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Delete(ctx, dao.Comment{Id: comment.Id})
}

func (c *commentRepository) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Insert(ctx, c.toEntity(comment))
}

func (c *commentRepository) GetCommentByIds(ctx context.Context, ids []int64) ([]domain.Comment, error) {
	vals, err := c.dao.FindOneByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	comments := make([]domain.Comment, 0, len(vals))
	for _, v := range vals {
		comment := c.toDomain(v)
		comments = append(comments, comment)
	}
	return comments, nil
}

func (c *commentRepository) GetMoreReplies(ctx context.Context, rid int64, maxId int64, limit int64) ([]domain.Comment, error) {
	cs, err := c.dao.FindRepliesByRid(ctx, rid, maxId, limit)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Comment, 0, len(cs))
	for _, cm := range cs {
		res = append(res, c.toDomain(cm))
	}

	return res, nil
}

func (c *commentRepository) toDomain(daoComment dao.Comment) domain.Comment {
	val := domain.Comment{
		Id: daoComment.Id,
		Commentator: domain.User{
			ID: daoComment.Uid,
		},
		Biz:        daoComment.Biz,
		BizId:      daoComment.BizID,
		Content:    daoComment.Content,
		CreateTime: time.UnixMilli(daoComment.CreateTime),
		UpdateTime: time.UnixMilli(daoComment.UpdateTime),
	}
	if daoComment.PID.Valid {
		val.ParentComment = &domain.Comment{
			Id: daoComment.PID.Int64,
		}
	}
	if daoComment.RootID.Valid {
		val.RootComment = &domain.Comment{
			Id: daoComment.RootID.Int64,
		}
	}
	return val
}

func (c *commentRepository) toEntity(domainComment domain.Comment) dao.Comment {
	daoComment := dao.Comment{
		Id:      domainComment.Id,
		Uid:     domainComment.Commentator.ID,
		Biz:     domainComment.Biz,
		BizID:   domainComment.BizId,
		Content: domainComment.Content,
	}
	if domainComment.RootComment != nil {
		daoComment.RootID = sql.NullInt64{
			Valid: true,
			Int64: domainComment.RootComment.Id,
		}
	}
	if domainComment.ParentComment != nil {
		daoComment.PID = sql.NullInt64{
			Valid: true,
			Int64: domainComment.ParentComment.Id,
		}
	}
	daoComment.CreateTime = time.Now().UnixMilli()
	daoComment.UpdateTime = time.Now().UnixMilli()
	return daoComment
}
