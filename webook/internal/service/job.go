package service

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"time"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, job domain.Job) error
}

type cronJobService struct {
	repo            repository.CronJobRepository
	refreshInterval time.Duration
	l               logger.LoggerV1
}

func newCronJobService(
	repo repository.CronJobRepository,
	l logger.LoggerV1,
) CronJobService {
	return &cronJobService{
		repo: repo,
		l:    l,

		refreshInterval: time.Minute,
	}
}

func (c cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	job, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(job.Id)
		}
	}()

	job.CancelFun = func() {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := c.repo.Release(ctx, job.Id, job.Version)
		if err != nil {
			c.l.Error(
				"释放 job 失败",
				logger.Error(err),
				logger.Int64("jib", job.Id),
				logger.Int("version", job.Version),
			)
		}
	}

	return job, err
}

func (c cronJobService) ResetNextTime(ctx context.Context, job domain.Job) error {
	nextTime := job.NextTime()
	return c.repo.UpdateNextTime(ctx, job.Id, nextTime)
}

func (c *cronJobService) refresh(id int64) {
	// 续约本质上就是更新一下更新时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateTime(ctx, id)
	if err != nil {
		c.l.Error(
			"续约失败",
			logger.Error(err),
			logger.Int64("jid", id),
		)
	}
}
