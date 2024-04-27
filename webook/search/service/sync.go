package service

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/search/domain"
	"github.com/Anwenya/GeekTime/webook/search/repository"
)

type syncService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
	anyRepo     repository.AnyRepository
}

func NewSyncService(
	userRepo repository.UserRepository,
	articleRepo repository.ArticleRepository,
	anyRepo repository.AnyRepository,
) SyncService {
	return &syncService{
		userRepo:    userRepo,
		articleRepo: articleRepo,
		anyRepo:     anyRepo,
	}
}

func (s *syncService) InputArticle(ctx context.Context, article domain.Article) error {
	return s.articleRepo.InputArticle(ctx, article)
}

func (s *syncService) InputUser(ctx context.Context, user domain.User) error {
	return s.userRepo.InputUser(ctx, user)
}

func (s *syncService) InputAny(ctx context.Context, idxName, docID, data string) error {
	return s.anyRepo.Input(ctx, idxName, docID, data)
}
