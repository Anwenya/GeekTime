package job

import (
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"strconv"
	"time"
)

type CronJobBuilder struct {
	vector *prometheus.SummaryVec
	l      logger.LoggerV1
}

func NewCronJobBuilder(opts prometheus.SummaryOpts, l logger.LoggerV1) *CronJobBuilder {
	vector := prometheus.NewSummaryVec(
		opts,
		[]string{"job", "success"},
	)

	return &CronJobBuilder{vector: vector, l: l}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJonAdapterFun(
		func() {
			start := time.Now()
			b.l.Debug(
				"开始运行",
				logger.String("name", name),
			)
			err := job.Run()
			if err != nil {
				b.l.Error(
					"执行失败",
					logger.Error(err),
					logger.String("name", name),
				)
			}

			b.l.Debug(
				"结束运行",
				logger.String("name", name),
			)
			duration := time.Since(start)
			b.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).
				Observe(float64(duration.Milliseconds()))
		},
	)
}

type cronJonAdapterFun func()

func (c cronJonAdapterFun) Run() {
	c()
}
