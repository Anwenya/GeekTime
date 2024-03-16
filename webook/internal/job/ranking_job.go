package job

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
)

type RankingJob struct {
	svc     service.RankingService
	timeout time.Duration

	key       string
	client    *rlock.Client
	localLock *sync.Mutex
	lock      *rlock.Lock

	l logger.LoggerV1
}

func NewRankingJob(
	svc service.RankingService,
	timeout time.Duration,
	client *rlock.Client,
	l logger.LoggerV1,
) *RankingJob {
	return &RankingJob{
		key:       "job:ranking",
		svc:       svc,
		timeout:   timeout,
		client:    client,
		localLock: &sync.Mutex{},
		l:         l,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

// Run
// 使用基于redis的分布式锁 保证只有一个实例可以执行该任务
// 在拿到锁后自动续约
func (r *RankingJob) Run() error {
	// localLock 是用来保护 lock的
	r.localLock.Lock()

	lock := r.lock

	if lock == nil {
		// 抢分布式锁
		// 总超时时间给4秒
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := r.client.Lock(
			ctx,
			r.key,
			// 锁过期时间
			r.timeout,
			// 抢锁失败 重试机制 该时间包含在总超时时间内
			&rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
			},
			// 超时时间
			time.Second,
		)
		if err != nil {
			r.localLock.Unlock()
			r.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}

		// 抢到分布式锁
		r.lock = lock
		r.localLock.Unlock()

		// 续约
		go func() {
			// 根据时间情况制定续约方案
			err := lock.AutoRefresh(r.timeout/2, r.timeout)
			if err != nil {
				// 续约失败
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}

	// 拿到锁后执行任务
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
