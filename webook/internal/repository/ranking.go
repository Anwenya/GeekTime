package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	cache cache.RankingCache
}

func NewCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{cache: cache}
}

// ReplaceTopN
// 设置缓存 或者 更新缓存
func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return c.cache.Set(ctx, arts)
}

// GetTopN
// 查询缓存
func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return c.cache.Get(ctx)
}

// CachedDoubleRankingRepository
// 双缓存方案
type CachedDoubleRankingRepository struct {
	redisCache cache.RankingCache
	localCache *cache.LocalRankingCache
}

func NewCachedDoubleRankingRepository(
	redisCache cache.RankingCache,
	localCache *cache.LocalRankingCache,
) *CachedDoubleRankingRepository {
	return &CachedDoubleRankingRepository{
		redisCache: redisCache,
		localCache: localCache,
	}
}

func (c *CachedDoubleRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	// 本地缓存基本不可能出错
	_ = c.localCache.Set(ctx, arts)
	return c.redisCache.Set(ctx, arts)
}

func (c *CachedDoubleRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	res, err := c.localCache.Get(ctx)
	if err == nil {
		return res, nil
	}
	res, err = c.redisCache.Get(ctx)
	// 本地缓存异常 + redis异常 (可能是没有缓存或者未知异常)
	// 时 强制从本地读缓存并且不考虑过期时间
	if err != nil {
		return c.localCache.ForceGet(ctx)
	}

	// 回写本地缓存
	_ = c.localCache.Set(ctx, res)
	return res, nil
}
