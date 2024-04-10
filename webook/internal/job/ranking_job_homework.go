package job

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/google/uuid"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/hashicorp/go-multierror"
	"github.com/redis/go-redis/v9"
	"sync"
	"sync/atomic"
	"time"
)

type RankingJobV1 struct {
	svc     service.RankingService
	timeout time.Duration

	key        string
	lockClient *rlock.Client
	localLock  *sync.Mutex
	lock       *rlock.Lock

	// 节点负载
	load *atomic.Int32
	// 标识当前节点
	nodeId string
	// 与redis通信
	redisClient redis.Cmdable
	// 节点负载信息在redis中的key
	loadKey string
	// 定时器
	loadTicker *time.Ticker

	l logger.LoggerV1
}

func NewRankingJobV1(
	svc service.RankingService,
	timeout time.Duration,
	lockClient *rlock.Client,
	redisClient redis.Cmdable,
	loadInterval time.Duration,
	l logger.LoggerV1,
) *RankingJobV1 {
	res := &RankingJobV1{
		key:        "job:ranking",
		svc:        svc,
		timeout:    timeout,
		lockClient: lockClient,
		localLock:  &sync.Mutex{},

		nodeId:      uuid.NewString(),
		redisClient: redisClient,
		load:        &atomic.Int32{},
		loadTicker:  time.NewTicker(loadInterval),
		loadKey:     "ranking_job_nodes_load",

		l: l,
	}

	res.loadCycle()
	return res
}

func (r *RankingJobV1) Name() string {
	return "ranking"
}

// Run
// 使用基于redis的分布式锁 保证只有一个实例可以执行该任务
// 在拿到锁后自动续约
func (r *RankingJobV1) Run() error {
	// localLock 是用来保护 lock的
	r.localLock.Lock()

	lock := r.lock

	if lock == nil {
		// 抢分布式锁

		// 先根据负载判断是否可以抢锁
		if !r.checkMyLoad() {
			r.l.Warn("负载不满足抢锁条件 放弃抢锁")
			r.localLock.Unlock()
			return nil
		}

		// 总超时时间给4秒
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := r.lockClient.Lock(
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

		r.l.Debug("抢到分布式锁: ", logger.String("nodeId", r.nodeId))
		// 抢到分布式锁
		r.lock = lock
		r.localLock.Unlock()

		// 续约
		go func() {
			// 根据时间情况制定续约方案
			// todo 优化:能够控制续约逻辑的话 可以加上负载检查 负载不合法时停止续约
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

func (r *RankingJobV1) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var err *multierror.Error
	if lock != nil {
		err = multierror.Append(err, lock.Unlock(ctx))
	}

	// 停止上报负载
	if r.loadTicker != nil {
		r.loadTicker.Stop()
	}

	// 从redis中删除该节点的负载信息
	err = multierror.Append(err, r.redisClient.ZRem(ctx, r.loadKey, redis.Z{Member: r.nodeId}).Err())

	return lock.Unlock(ctx)
}

// 按照设置的间隔 上报负载 检查负载
func (r *RankingJobV1) loadCycle() {
	go func() {
		for range r.loadTicker.C {
			r.reportLoad()
			r.releaseLockIfNeed()
		}
	}()
}

func (r *RankingJobV1) reportLoad() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	load := r.load.Load()
	r.l.Debug("上报负载: ", logger.String("nodeId", r.nodeId), logger.Int32("load", load))
	// 插入或更新节点负载信息
	r.redisClient.ZAdd(ctx, r.loadKey, redis.Z{Member: r.nodeId, Score: float64(load)})
	cancel()
	return
}

// 检查负载情况 如果不符合规则就直接释放锁
func (r *RankingJobV1) releaseLockIfNeed() {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()

	if lock != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if !r.checkMyLoad() {
			r.l.Debug(
				"负载不满足抢锁条件 主动释放锁",
				logger.String("nodeId", r.nodeId),
			)
			r.localLock.Lock()
			r.lock = nil
			r.localLock.Unlock()

			// 释放锁
			lock.Unlock(ctx)
		}

	}
}

// 检查当前节点的负载情况 是否允许持有锁
// 1.节点负载不能超过中位数
// 2.节点负载不能超过平均值
// 返回 true 标识可以持有锁
// 返回 false 标识不能持有锁 需要释放锁
func (r *RankingJobV1) checkMyLoad() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var currentLoad float64
	var loads []float64

	for i := 0; i < 3; i++ {
		res, err := r.redisClient.ZRangeWithScores(ctx, r.loadKey, 0, -1).Result()

		// 重试3次
		if err != nil {
			r.l.Error("查询节点负载失败", logger.String("nodeId", r.nodeId), logger.Error(err))
			time.Sleep(time.Millisecond * 100)
			continue
		}

		for _, z := range res {
			// 节点名
			nodeId, _ := z.Member.(string)
			// 负载
			load := z.Score
			if nodeId == r.nodeId {
				currentLoad = load
			}
			loads = append(loads, load)
		}

	}

	// 获得节点信息异常 视为不可抢锁
	if len(loads) <= 0 {
		return false
	}

	// 计算中位数 和 平均数
	avg := calculateAverage(loads)
	mid := calculateMedian(loads)

	// 小于等于 才合法
	if currentLoad <= avg && currentLoad <= mid {
		return true
	}

	return true
}

// 计算平均数
func calculateAverage(numbers []float64) float64 {
	total := 0.0
	for _, num := range numbers {
		total += num
	}
	return total / float64(len(numbers))
}

// 计算中位数
func calculateMedian(numbers []float64) float64 {
	// 获取数组长度
	length := len(numbers)

	// 判断数组长度的奇偶
	if length%2 == 0 {
		// 偶数长度的数组，取中间两个数的平均值
		middle := length / 2
		return (numbers[middle-1] + numbers[middle]) / 2
	} else {
		// 奇数长度的数组，取中间的数
		middle := length / 2
		return numbers[middle]
	}
}
