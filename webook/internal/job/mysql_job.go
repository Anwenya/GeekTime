package job

import (
	"context"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"golang.org/x/sync/semaphore"
	"time"
)

type Executor interface {
	Name() string
	Exec(ctx context.Context, job domain.Job) error
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() Executor {
	return &LocalFuncExecutor{
		funcs: map[string]func(ctx context.Context, j domain.Job) error{},
	}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, job domain.Job) error {
	fn, ok := l.funcs[job.Name]
	if !ok {
		return fmt.Errorf("未注册本地执行器:%s", job.Name)
	}
	return fn(ctx, job)
}

func (l *LocalFuncExecutor) RegisterFunc(
	name string,
	fn func(ctx context.Context, j domain.Job) error,
) {
	l.funcs[name] = fn
}

// Scheduler 调度器
type Scheduler struct {
	dbTimeout time.Duration
	svc       service.CronJobService

	executors map[string]Executor

	l logger.LoggerV1

	limiter *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{
		svc:       svc,
		dbTimeout: time.Second,
		limiter:   semaphore.NewWeighted(100),
		l:         l,
		executors: map[string]Executor{},
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		// 主动放弃调度
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 信号量限制同时执行的任务数
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		// 抢任务
		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		job, err := s.svc.Preempt(dbCtx)
		cancel()

		if err != nil {
			// 有异常 睡眠之后继续新一轮
			s.l.Error(
				"抢占任务异常",
				logger.Error(err),
			)
			time.Sleep(time.Second * 3)
			continue
		}

		exec, ok := s.executors[job.Executor]
		if !ok {
			// 没有对应的处理函数 继续新一轮
			s.l.Error(
				"找不到执行器",
				logger.Int64("jid", job.Id),
				logger.String("executor", job.Executor),
			)
			continue
		}

		// 执行具体任务
		go func() {
			// 释放
			defer func() {
				s.limiter.Release(1)
				job.CancelFun()
			}()

			// 执行
			err := exec.Exec(ctx, job)
			if err != nil {
				s.l.Error(
					"执行任务失败",
					logger.Int64("jid", job.Id),
					logger.Error(err),
				)
				return
			}

			// 执行完成后更新下一次执行的时间
			err = s.svc.ResetNextTime(ctx, job)
			if err != nil {
				s.l.Error(
					"重置下次执行时间失败",
					logger.Int64("jid", job.Id),
					logger.Error(err),
				)
			}
		}()
	}
}
