package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/tag/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type TagRedisCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewTagRedisCache(client redis.Cmdable) TagCache {
	return &TagRedisCache{client: client}
}

func (t *TagRedisCache) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	data, err := t.client.HGetAll(ctx, t.userTagsKey(uid)).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrKeyNotExist
	}

	res := make([]domain.Tag, 0, len(data))
	for _, val := range data {
		var t domain.Tag
		err = json.Unmarshal([]byte(val), &t)
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (t *TagRedisCache) Append(ctx context.Context, uid int64, tags ...domain.Tag) error {
	key := t.userTagsKey(uid)
	pip := t.client.Pipeline()
	for _, tag := range tags {
		val, err := json.Marshal(tag)
		if err != nil {
			return err
		}

		pip.HMSet(ctx, key, strconv.FormatInt(tag.Id, 10), val)
	}
	pip.Expire(ctx, key, t.expiration)
	_, err := pip.Exec(ctx)
	return err
}

func (t *TagRedisCache) DelTags(ctx context.Context, uid int64) error {
	return t.client.Del(ctx, t.userTagsKey(uid)).Err()
}

func (t *TagRedisCache) userTagsKey(uid int64) string {
	return fmt.Sprintf("tag:user_tags:%d", uid)
}
