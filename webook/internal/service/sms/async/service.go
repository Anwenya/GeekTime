package async

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"

	"time"
)

type Service struct {
	svc  sms.SMService
	repo repository.AsyncSMSRepository
	l    logger.LoggerV1
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
	// 非异步
	return s.svc.Send(ctx, tplId, args, numbers...)
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
	return true
}
