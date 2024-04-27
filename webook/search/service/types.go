package service

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/search/domain"
)

type SearchService interface {
	Search(ctx context.Context, uid int64, expression string) (domain.SearchResult, error)
}

type SyncService interface {
	InputArticle(ctx context.Context, article domain.Article) error
	InputUser(ctx context.Context, user domain.User) error
	InputAny(ctx context.Context, idxName, docID, data string) error
}
