package async

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"math"

	"time"
)

type Service struct {
	svc    sms.SMService
	repo   repository.AsyncSMSRepository
	record *record
	l      logger.LoggerV1
}

func NewService(
	svc sms.SMService,
	repo repository.AsyncSMSRepository,
	l logger.LoggerV1,
) *Service {
	s := &Service{
		svc:  svc,
		repo: repo,
		l:    l,
		// 响应时间不能连续3次超过3秒
		// 近5次的平均响应时间不能超过3秒
		record: newRecord(time.Second*3, 3, 5, time.Second*3),
	}
	go func() {
		s.StartAsyncCycle()
	}()
	return s
}

// StartAsyncCycle
// 异步发送消息
// 最简单的抢占式调度
func (s *Service) StartAsyncCycle() {
	// 防止在运行测试时立刻去抢占任务
	// 会有偶发性异常
	time.Sleep(time.Second * 3)
	for {
		s.AsyncSend()
	}
}

func (s *Service) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ast, err := s.repo.PreemptWaitingTask(ctx)
	cancel()
	switch err {
	case nil:
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = s.svc.Send(ctx, ast.TplId, ast.Args, ast.Numbers...)
		if err != nil {
			s.l.Error(
				"执行异步发送短信失败",
				logger.Error(err),
				logger.Field{
					Key: "task",
					Val: ast,
				},
			)
		}
		res := err == nil
		// 更新状态
		err = s.repo.ReportScheduleResult(ctx, ast.Id, res)
		if err != nil {
			s.l.Error(
				"执行异步发送短信成功 但是标记数据库失败",
				logger.Error(err),
				logger.Bool("res", res),
				logger.Field{
					Key: "task",
					Val: ast,
				},
			)
		}
	case repository.ErrWaitingSMSNotFound:
		time.Sleep(time.Second)
	default:
		// 正常来说应该是数据库那边出了问题
		// 睡眠的话可以规避掉短时间的网络抖动问题
		s.l.Error(
			"抢占异步发送短信任务失败",
			logger.Error(err),
		)
		time.Sleep(time.Second)
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	if s.needAsync() {
		// 需要异步发送 直接转储到数据库
		err := s.repo.CreateTask(
			ctx,
			domain.AsyncSMS{
				TplId:    tplId,
				Args:     args,
				Numbers:  numbers,
				RetryMax: 3,
			},
		)
		return err
	}
	start := time.Now()
	// 非异步
	err := s.svc.Send(ctx, tplId, args, numbers...)
	if err != nil {
		return err
	}
	s.record.Record(time.Since(start))
	return nil
}

// 根据具体的规则判断是否需要异步发送
func (s *Service) needAsync() bool {
	// 这边就是你要设计的，各种判定要不要触发异步的方案
	// 1. 基于响应时间的，平均响应时间
	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超过了 500ms，然后后续请求转异步
	// 1.2 变化趋势，比如说当前一秒钟内的所有请求的响应时间比上一秒钟增长了 X%，就转异步
	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步

	// 什么时候退出异步
	// 1. 进入异步 N 分钟后
	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率
	return s.record.Judge()
}

type record struct {

	// consecutiveTimeoutTimes 连续超时次数
	ctt int
	// thresholdTimeout 超时阈值
	tto time.Duration
	// maxConsecutiveTimeoutTimes 最大连续超时次数
	mctt int

	// responseTimeoutList 记录响应时间
	rtl []int64
	// averageQuantity 参与计算平均响应时间的次数
	aq int
	// thresholdAverageTimeout 平均响应时间阈值
	tat time.Duration
	// sumResponseTime 近几次总响应时间
	srt int64
	// 总响应时间阈值
	tsrt int64

	// 最大/小响应时间
	maxrt int64
	minrt int64
}

func newRecord(tto time.Duration, mctt int, aq int, tat time.Duration) *record {
	return &record{
		tto:   tto,
		mctt:  mctt,
		aq:    aq,
		tat:   tat,
		rtl:   make([]int64, aq, aq),
		maxrt: math.MinInt,
		minrt: math.MaxInt,
		tsrt:  int64(aq) * tat.Milliseconds(),
	}
}

func (r *record) Record(rt time.Duration) {
	rtInt64 := rt.Milliseconds()

	// 计算总响应时间
	r.srt += rtInt64
	r.srt -= r.rtl[0]

	// 记录近几次的响应时间 从后往前添加
	copy(r.rtl, r.rtl[1:r.aq])
	r.rtl[r.aq-1] = rtInt64

	// 记录连续超时次数
	if rt > r.tto {
		r.ctt += 1
	} else {
		r.ctt = 0
	}

	// 最大与最小响应时间
	if rtInt64 > r.maxrt {
		r.maxrt = rtInt64
	}
	if rtInt64 < r.minrt {
		r.minrt = rtInt64
	}
}

func (r *record) Judge() bool {
	// 平均响应时间是否达到阈值
	if r.srt > r.tsrt {
		return true
	}

	// 连续超时次数是否达到阈值
	if r.ctt >= r.mctt {
		return true
	}
	return false
}
