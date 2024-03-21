package service

import (
	"context"
	interactivev1 "github.com/Anwenya/GeekTime/webook/api/proto/gen/interactive/v1"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
)

type RankingService interface {
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	// 查询文章的点赞数
	intrSvc interactivev1.InteractiveServiceClient
	// 查询时间段内的文章
	artSvc ArticleService

	batchSize int
	scoreFunc func(likeCnt int64, updateTime time.Time) float64
	// topN
	n int

	repo repository.RankingRepository
}

func NewBatchRankingService(
	intrSvc interactivev1.InteractiveServiceClient,
	artSvc ArticleService,
) RankingService {
	return &BatchRankingService{
		intrSvc:   intrSvc,
		artSvc:    artSvc,
		batchSize: 100,
		n:         100,
		// 积分计算规则
		scoreFunc: func(likeCnt int64, updateTime time.Time) float64 {
			// 如果频繁更新 该时间就会比较小
			// 这在该计算规则下不平衡
			duration := time.Since(updateTime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	return b.repo.ReplaceTopN(ctx, arts)
}

func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	// 只统计前七天的文章
	ddl := start.Add(-7 * 24 * time.Hour)

	type Score struct {
		score float64
		art   domain.Article
	}
	topNQueue := queue.NewConcurrentPriorityQueue[Score](
		b.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		},
	)

	for {
		// 取数据
		arts, err := b.artSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return nil, err
		}

		// 提前退出
		if len(arts) == 0 {
			break
		}

		ids := slice.Map[domain.Article, int64](
			arts,
			func(idx int, art domain.Article) int64 {
				return art.Id
			},
		)

		// 取点赞数
		intrResp, err := b.intrSvc.GetByIds(ctx, &interactivev1.GetByIdsRequest{
			Biz: "article",
			Ids: ids,
		})
		if err != nil {
			return nil, err
		}

		intrMap := intrResp.GetInteractives()
		for _, art := range arts {
			intr := intrMap[art.Id]

			score := b.scoreFunc(intr.LikeCnt, art.UpdateTime)
			ele := Score{
				score: score,
				art:   art,
			}

			// 满
			if topNQueue.Len() >= b.n {
				minEle, _ := topNQueue.Dequeue()
				if minEle.score < score {
					_ = topNQueue.Enqueue(ele)
				} else {
					_ = topNQueue.Enqueue(minEle)
				}
			} else {
				_ = topNQueue.Enqueue(ele)
			}
		}

		offset += len(arts)
		// 最后一页 或者 有文章超过的更新时间超过了时间范围
		if len(arts) < b.batchSize || arts[len(arts)-1].UpdateTime.Before(ddl) {
			break
		}
	}

	// 计算结果
	res := make([]domain.Article, topNQueue.Len())
	// 小顶堆 要倒序
	for i := topNQueue.Len() - 1; i >= 0; i-- {
		ele, _ := topNQueue.Dequeue()
		res[i] = ele.art
	}
	return res, nil

}

func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.repo.GetTopN(ctx)
}
