package cache

import (
	"context"
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"time"
)

// LocalRankingCache 本地缓存
// 对于并发要求很高的场景
// 可以使用本地+redis双缓存
// 默认先走本地缓存
type LocalRankingCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func (l *LocalRankingCache) Set(ctx context.Context, arts []domain.Article) error {
	l.topN.Store(arts)
	// 过期时间
	l.ddl.Store(time.Now().Add(l.expiration))
	// 本地缓存的操作 基本不可能失败
	return nil
}

func (l *LocalRankingCache) Get(ctx context.Context) ([]domain.Article, error) {
	ddl := l.ddl.Load()
	arts := l.topN.Load()
	if len(arts) == 0 || ddl.Before(time.Now()) {
		return nil, errors.New("本地缓存失效")
	}
	return arts, nil
}

// ForceGet
// 极端情况下不检查有效时间
// 过期的榜单也比没有强
func (l *LocalRankingCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	arts := l.topN.Load()
	if len(arts) == 0 {
		return nil, errors.New("本地缓存失效")
	}
	return arts, nil
}
