package dao

import (
	"context"
	"gorm.io/gorm"
)

type ArticleReaderDAO interface {
	// Upsert 同库不同表
	Upsert(ctx context.Context, art PublishedArticle) error
	// UpsertV2 分库
	UpsertV2(ctx context.Context, art Article) error
}

type ArticleGORMReaderDAO struct {
	db *gorm.DB
}

func (a ArticleGORMReaderDAO) Upsert(ctx context.Context, art PublishedArticle) error {
	panic(any("implement me"))
}

func (a ArticleGORMReaderDAO) UpsertV2(ctx context.Context, art Article) error {
	panic(any("implement me"))
}

func NewArticleGORMReaderDAO(db *gorm.DB) *ArticleGORMReaderDAO {
	return &ArticleGORMReaderDAO{db: db}
}
