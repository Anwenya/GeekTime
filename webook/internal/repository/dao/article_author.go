package dao

import (
	"context"
	"gorm.io/gorm"
)

type ArticleAuthorDAO interface {
	Create(ctx context.Context, art Article) (int64, error)
	Update(ctx context.Context, art Article) error
}

type ArticleGORMAuthorDAO struct {
	db *gorm.DB
}

func NewArticleGORMAuthorDAO(db *gorm.DB) *ArticleGORMAuthorDAO {
	return &ArticleGORMAuthorDAO{db: db}
}

func (a ArticleGORMAuthorDAO) Create(ctx context.Context, art Article) (int64, error) {
	panic(any("implement me"))
}

func (a ArticleGORMAuthorDAO) Update(ctx context.Context, art Article) error {
	panic(any("implement me"))
}
