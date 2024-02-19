package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
)

var ErrCodeVerifyTooMany = cache.ErrCodeVerifyTooMany

type CodeRepository interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type CachedCodeRepository struct {
	cc cache.CodeCache
}

func NewCachedCodeRepository(cc cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cc: cc,
	}
}

func (ccr *CachedCodeRepository) Set(ctx context.Context, biz, phone, code string) error {
	return ccr.cc.Set(ctx, biz, phone, code)
}

func (ccr *CachedCodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return ccr.cc.Verify(ctx, biz, phone, code)
}
