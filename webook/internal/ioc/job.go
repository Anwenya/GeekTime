package ioc

import (
	"github.com/Anwenya/GeekTime/webook/internal/job"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(
	svc service.RankingService,
	client *rlock.Client,
	l logger.LoggerV1,
) job.Job {
	return job.NewRankingJob(svc, time.Second*30, client, l)
}

func InitJobs(l logger.LoggerV1, rjob job.Job) *cron.Cron {
	builder := job.NewCronJobBuilder(
		prometheus.SummaryOpts{
			Namespace: "GeekTime",
			Subsystem: "webook",
			Name:      "cron_job",
			Help:      "定时任务执行",
			Objectives: map[float64]float64{
				0.5:   0.01,
				0.75:  0.01,
				0.9:   0.01,
				0.99:  0.001,
				0.999: 0.0001,
			},
		},
		l,
	)
	expr := cron.New(cron.WithSeconds())
	_, err := expr.AddJob("@every 30m", builder.Build(rjob))
	if err != nil {
		panic(any(err))
	}
	return expr
}
