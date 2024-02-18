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

type UserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewUserCache(cmd redis.Cmdable) *UserCache {
	return &UserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (uc *UserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := uc.key(uid)
	// 查询缓存
	data, err := uc.cmd.Get(ctx, key).Result()
	// 可能是不存在该key 也可能是redis异常 也可能是网络异常
	if err != nil {
		return domain.User{}, err
	}
	// 可能有不同的序列化方式
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (uc *UserCache) Set(ctx context.Context, du domain.User) error {
	key := uc.key(du.Id)
	// 可能有不同的序列化方式
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}
	// 设置缓存并返回结果
	return uc.cmd.Set(ctx, key, data, uc.expiration).Err()
}

func (uc *UserCache) key(uid int64) string {
	return fmt.Sprintf("user:info:%d", uid)
}
