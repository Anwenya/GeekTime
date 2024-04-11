package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/account/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type AccountRedisCache struct {
	client redis.Cmdable
}

func NewAccountRedisCache(client redis.Cmdable) AccountCache {
	return &AccountRedisCache{client: client}
}

// SetUnique 每处理一笔都进行记录 防止重复
func (a *AccountRedisCache) SetUnique(ctx context.Context, cr domain.Credit) error {
	return a.client.Set(ctx, a.key(cr.Biz, cr.BizId), "", time.Minute*30).Err()
}

func (a *AccountRedisCache) GetUnique(ctx context.Context, cr domain.Credit) error {
	res, err := a.client.Exists(ctx, a.key(cr.Biz, cr.BizId)).Result()
	if err != nil {
		return err
	}
	if res > 0 {
		return errors.New("该业务已经处理过了")
	}
	return nil
}

func (a *AccountRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("credit:%s:%d", biz, bizId)
}
