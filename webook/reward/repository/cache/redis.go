package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/reward/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type RewardRedisCache struct {
	client redis.Cmdable
}

func NewRewardRedisCache(client redis.Cmdable) RewardCache {
	return &RewardRedisCache{client: client}
}

func (rr *RewardRedisCache) GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	key := rr.codeURLKey(r)
	data, err := rr.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.CodeURL{}, err
	}
	var res domain.CodeURL
	err = json.Unmarshal(data, &res)
	return res, err
}

func (rr *RewardRedisCache) CachedCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error {
	key := rr.codeURLKey(r)
	data, err := json.Marshal(cu)
	if err != nil {
		return err
	}
	// 支付过期时间是30分钟 缓存设置29分钟
	return rr.client.Set(ctx, key, data, time.Minute*29).Err()
}

func (rr *RewardRedisCache) codeURLKey(r domain.Reward) string {
	return fmt.Sprintf("reward:code_url:%s:%d:%d", r.Target.Biz, r.Target.BizId, r.Uid)
}
