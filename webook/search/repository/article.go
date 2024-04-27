package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/search/domain"
	"github.com/Anwenya/GeekTime/webook/search/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type articleRepository struct {
	dao dao.ArticleDAO
	tag dao.TagDAO
}

func NewArticleRepository(dao dao.ArticleDAO, tag dao.TagDAO) ArticleRepository {
	return &articleRepository{dao: dao, tag: tag}
}

func (a *articleRepository) InputArticle(ctx context.Context, msg domain.Article) error {
	return a.dao.InputArticle(
		ctx,
		dao.Article{
			Id:      msg.Id,
			Title:   msg.Title,
			Status:  msg.Status,
			Content: msg.Content,
		},
	)
}

func (a *articleRepository) SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.Article, error) {
	// 先查询标签满足要求的
	artIds, err := a.tag.Search(ctx, uid, "article", keywords)
	if err != nil {
		return nil, err
	}
	// 再查询文章满足要求的 命中标签的搜索结果的给高点的权重
	arts, err := a.dao.Search(ctx, artIds, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.Article](arts,
		func(idx int, src dao.Article) domain.Article {
			return domain.Article{
				Id:      src.Id,
				Title:   src.Title,
				Status:  src.Status,
				Content: src.Content,
				Tags:    src.Tags,
			}
		},
	), nil
}
