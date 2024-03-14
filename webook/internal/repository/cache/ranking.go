package cache

import (
	"context"
	"encoding/json"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RedisRankingCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

func NewRedisRankingCache(client redis.Cmdable) RankingCache {
	return &RedisRankingCache{
		client: client,
		key:    "ranking:top_n",
		// 所有对榜单的访问都是走缓存的
		// 也可以考虑将过期时间设置很大
		// 或者不设置过期时间
		expiration: time.Minute * 3,
	}
}

func (r *RedisRankingCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := range arts {
		arts[i].Content = arts[i].Abstract()
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key, val, r.expiration).Err()
}

func (r *RedisRankingCache) Get(ctx context.Context) ([]domain.Article, error) {
	val, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}
