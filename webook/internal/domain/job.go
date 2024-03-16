package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

type Job struct {
	Id      int64
	Version int
	Name    string
	// cron表达式 控制执行时机
	Expression string
	// 简单理解就是负责执行该任务的函数
	Executor string
	// 任务配置
	Config    string
	CancelFun func()
}

func (j Job) NextTime() time.Time {
	c := cron.NewParser(
		cron.Second | cron.Minute |
			cron.Hour | cron.Dom |
			cron.Month | cron.Dow |
			cron.Descriptor,
	)
	s, _ := c.Parse(j.Expression)
	return s.Next(time.Now())
}
