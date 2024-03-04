package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, res []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, res domain.Article) error
}

type RedisArticleCache struct {
	client redis.Cmdable
}

func NewRedisArticleCache(client redis.Cmdable) ArticleCache {
	return &RedisArticleCache{client: client}
}

// GetFirstPage 获得用户文章列表的缓存
func (a RedisArticleCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := a.authorFirstPageKey(uid)
	val, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

// SetFirstPage 设置用户文章列表的缓存
func (a RedisArticleCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	// 只缓存文章的摘要 列表页一般也只展示摘要
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	key := a.authorFirstPageKey(uid)
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, key, val, time.Minute*10).Err()
}

// DelFirstPage 删除缓存
func (a RedisArticleCache) DelFirstPage(ctx context.Context, uid int64) error {
	return a.client.Del(ctx, a.authorFirstPageKey(uid)).Err()
}

func (a RedisArticleCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.authorDetailKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a RedisArticleCache) Set(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.authorDetailKey(art.Id), val, time.Minute*10).Err()
}

func (a RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.pubDetailKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a RedisArticleCache) SetPub(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.pubDetailKey(art.Id), val, time.Minute*10).Err()
}

// 作者库第一页缓存的key
func (a *RedisArticleCache) pubDetailKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}

// 作者库文章详情的key
func (a *RedisArticleCache) authorDetailKey(id int64) string {
	return fmt.Sprintf("article:author:detail:%d", id)
}

// 作者库第一页缓存的key
func (a *RedisArticleCache) authorFirstPageKey(uid int64) string {
	return fmt.Sprintf("article:author:first_page:%d", uid)
}
