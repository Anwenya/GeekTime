package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRanking(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "192.168.2.130:6379",
		Password: "",
		DB:       0,
	})

	interactiveCache := NewInteractiveRedisCache(client)

	biz := "article"

	// 第一次设置缓存
	interactiveCache.SetRankingScore(context.Background(), biz, 1, 10)
	interactiveCache.SetRankingScore(context.Background(), biz, 2, 20)
	interactiveCache.SetRankingScore(context.Background(), biz, 3, 30)
	// 已存在更新缓存
	interactiveCache.SetRankingScore(context.Background(), biz, 3, 30)

	// 正常+1
	interactiveCache.IncrRankingIfPresent(context.Background(), biz, 2)

	// 不存在的key + 1 应返回特定异常
	err := interactiveCache.IncrRankingIfPresent(context.Background(), biz, 4)
	require.Equal(t, ErrRankingUpdate, err)

	// 获得排名
	top, err := interactiveCache.LikeTop(context.Background(), biz, 2)
	require.NoError(t, err)

	t.Log(top)

	// 删除
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("top_%s_%d", biz, i)
		client.Del(context.Background(), key)
	}
}

func TestBatch(t *testing.T) {
	type Interactive struct {
		BizId   int64
		LikeCnt int64
	}

	interactives := []Interactive{
		{
			BizId:   10,
			LikeCnt: 10,
		},
		{
			BizId:   8,
			LikeCnt: 8,
		},
		{
			BizId:   7,
			LikeCnt: 7,
		},
		{
			BizId:   6,
			LikeCnt: 6,
		},
		{
			BizId:   5,
			LikeCnt: 5,
		},
		{
			BizId:   4,
			LikeCnt: 4,
		}, {
			BizId:   3,
			LikeCnt: 3,
		}, {
			BizId:   2,
			LikeCnt: 2,
		}, {
			BizId:   1,
			LikeCnt: 1,
		},
	}

	members := make([][]redis.Z, 100)
	for i := 0; i < 100; i++ {
		members[i] = make([]redis.Z, 0, len(interactives))
	}

	for _, interactive := range interactives {
		key := interactive.BizId % 100
		members[key] = append(
			members[key],
			redis.Z{
				Score:  float64(interactive.LikeCnt),
				Member: interactive.BizId,
			},
		)
	}

	for index, val := range members {
		//err := i.client.ZAdd(ctx, fmt.Sprintf("top_%s_%d", biz, index), val...).Err()
		//if err != nil {
		//    return err
		//}
		t.Log(index, val)
	}

}
