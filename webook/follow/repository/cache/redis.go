package cache

import (
	"context"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/follow/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
)

var ErrKeyNotExist = redis.Nil

type FollowRedisCache struct {
	client redis.Cmdable
}

func NewFollowRedisCache(client redis.Cmdable) FollowCache {
	return &FollowRedisCache{client: client}
}

func (r *FollowRedisCache) Follow(ctx context.Context, follower, followee int64) error {
	return r.updateStaticsInfo(ctx, follower, followee, 1)
}

func (r *FollowRedisCache) CancelFollow(ctx context.Context, follower, followee int64) error {
	return r.updateStaticsInfo(ctx, follower, followee, -1)
}

func (r *FollowRedisCache) updateStaticsInfo(ctx context.Context, follower, followee int64, delta int64) error {
	tx := r.client.TxPipeline()
	// 这两个操作，只是记录了一下，还没发过去 redis
	tx.HIncrBy(ctx, r.staticsKey(follower), fieldFolloweeCnt, delta)
	tx.HIncrBy(ctx, r.staticsKey(followee), fieldFollowerCnt, delta)
	// 发过去了 Redis 执行，并且返回了结果
	_, err := tx.Exec(ctx)
	return err
}

func (r *FollowRedisCache) StaticsInfo(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	data, err := r.client.HGetAll(ctx, r.staticsKey(uid)).Result()
	if err != nil {
		return domain.FollowStatics{}, err
	}
	if len(data) == 0 {
		return domain.FollowStatics{}, ErrKeyNotExist
	}
	var res domain.FollowStatics
	res.Followers, _ = strconv.ParseInt(data[fieldFollowerCnt], 10, 64)
	res.Followees, _ = strconv.ParseInt(data[fieldFolloweeCnt], 10, 64)
	return res, nil
}

func (r *FollowRedisCache) SetStaticsInfo(ctx context.Context, uid int64, statics domain.FollowStatics) error {
	return r.client.HMSet(ctx, r.staticsKey(uid), fieldFollowerCnt, statics.Followers, fieldFolloweeCnt, statics.Followees).Err()
}

func (r *FollowRedisCache) staticsKey(uid int64) string {
	return fmt.Sprintf("follow:statics:%d", uid)
}
