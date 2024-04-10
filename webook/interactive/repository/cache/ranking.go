package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/interactive/domain"
	"github.com/ecodeclub/ekit/queue"
	"github.com/redis/go-redis/v9"
	"strconv"
)

var (
	//go:embed lua/ranking_incr.lua
	luaRankingIncr string

	//go:embed lua/ranking_set.lua
	luaRankingSet string
)

type InteractiveRanking interface {
	// IncrRankingIfPresent 如果排名数据存在就+1
	IncrRankingIfPresent(ctx context.Context, biz string, bizId int64) error
	// SetRankingScore 如果排名数据不存在 就把数据库的数据更新到缓存
	SetRankingScore(ctx context.Context, biz string, bizId int64, count int64) error
	// LikeTop 获得排名数据 默认前100
	LikeTop(ctx context.Context, biz string, topN int64) ([]domain.Interactive, error)
}

// IncrRankingIfPresent
// 缓存中有该key时,对应的指标 + 1  不过不会把所有数据都放在一个key中
// 根据业务id拆分到100个key
func (i *InteractiveRedisCache) IncrRankingIfPresent(ctx context.Context, biz string, bizId int64) error {
	// 如果存在就更新缓存
	res, err := i.client.Eval(ctx, luaRankingIncr, []string{i.rankingKey(biz, bizId)}, bizId).Result()
	if err != nil {
		return err
	}
	// 如果之前不存在 返回特定的错误
	if res.(int64) == 0 {
		return ErrRankingUpdate
	}
	return nil
}

// SetRankingScore 从数据库同步到缓存 前提是之前不存在
func (i *InteractiveRedisCache) SetRankingScore(ctx context.Context, biz string, bizId int64, count int64) error {
	return i.client.Eval(ctx, luaRankingSet, []string{i.rankingKey(biz, bizId)}, bizId, count).Err()
}

// BatchSetRankingScore
// 批量从数据库同步到缓存 用于初始化缓存
func (i *InteractiveRedisCache) BatchSetRankingScore(ctx context.Context, biz string, interactives []domain.Interactive) error {
	members := make([][]redis.Z, 100)
	for index := 0; index < 100; index++ {
		members[index] = make([]redis.Z, 0, len(interactives))
	}

	// 按key分
	for _, interactive := range interactives {
		key := interactive.BizId % 100
		members[key] = append(members[key], redis.Z{
			Score:  float64(interactive.LikeCnt),
			Member: interactive.BizId,
		})
	}

	for index, val := range members {
		if len(val) <= 0 {
			continue
		}

		err := i.client.ZAdd(ctx, fmt.Sprintf("top_%s_%d", biz, index), val...).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// LikeTop 从缓存中获得排行榜数据
func (i *InteractiveRedisCache) LikeTop(ctx context.Context, biz string, topN int64) ([]domain.Interactive, error) {
	topNQueue := queue.NewConcurrentPriorityQueue[domain.Interactive](
		int(topN),
		func(src domain.Interactive, dst domain.Interactive) int {
			if src.LikeCnt > dst.LikeCnt {
				return 1
			} else if src.LikeCnt == dst.LikeCnt {
				return 0
			} else {
				return -1
			}
		},
	)

	for index := 0; index < 100; index++ {
		key := fmt.Sprintf("top_%s_%d", biz, index)
		// 查出指定key的指定范围的缓存
		res, err := i.client.ZRevRangeWithScores(ctx, key, 0, topN).Result()
		if err != nil {
			return nil, err
		}
		// 遍历一次所有记录 维护一个小顶堆来取榜单
		for j := 0; j < len(res); j++ {
			id, _ := strconv.ParseInt(res[j].Member.(string), 10, 64)
			interactive := domain.Interactive{
				Biz:     biz,
				BizId:   id,
				LikeCnt: int64(res[j].Score),
			}

			// 满
			if topNQueue.Len() >= int(topN) {
				minEle, _ := topNQueue.Dequeue()
				if minEle.LikeCnt < interactive.LikeCnt {
					_ = topNQueue.Enqueue(interactive)
				} else {
					_ = topNQueue.Enqueue(minEle)
				}
			} else {
				_ = topNQueue.Enqueue(interactive)
			}
		}
	}

	// 计算结果
	interactives := make([]domain.Interactive, topNQueue.Len())
	// 小顶堆 要倒序
	for i := topNQueue.Len() - 1; i >= 0; i-- {
		ele, _ := topNQueue.Dequeue()
		interactives[i] = ele
	}

	return interactives, nil
}

func (i *InteractiveRedisCache) rankingKey(biz string, bizId int64) string {
	return fmt.Sprintf("top_%s_%d", biz, bizId%100)
}
