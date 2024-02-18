package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
)

var ErrCodeVerifyTooMany = cache.ErrCodeVerifyTooMany

type CodeRepository struct {
	cache *cache.CodeCache
}

func NewCodeRepository(cc *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cache: cc,
	}
}

func (cc *CodeRepository) Set(ctx context.Context, biz, phone, code string) error {
	return cc.cache.Set(ctx, biz, phone, code)
}

func (cc *CodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return cc.cache.Verify(ctx, biz, phone, code)
}
