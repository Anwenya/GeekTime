package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, uid int64) (domain.User, error)
	Set(ctx context.Context, du domain.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (ruc *RedisUserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := ruc.key(uid)
	// 查询缓存
	data, err := ruc.cmd.Get(ctx, key).Result()
	// 可能是不存在该key 也可能是redis异常 也可能是网络异常
	if err != nil {
		return domain.User{}, err
	}
	// 可能有不同的序列化方式
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (ruc *RedisUserCache) Set(ctx context.Context, du domain.User) error {
	key := ruc.key(du.Id)
	// 可能有不同的序列化方式
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}
	// 设置缓存并返回结果
	return ruc.cmd.Set(ctx, key, data, ruc.expiration).Err()
}

func (ruc *RedisUserCache) key(uid int64) string {
	return fmt.Sprintf("user:info:%d", uid)
}
