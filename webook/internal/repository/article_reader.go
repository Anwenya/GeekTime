package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
)

type ArticleReaderRepository interface {
	// Save 插入或更新
	Save(ctx context.Context, art domain.Article) error
}
