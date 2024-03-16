package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jid int64, version int) error
	UpdateTime(ctx context.Context, jid int64) error
	UpdateNextTime(ctx context.Context, jid int64, t time.Time) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func NewGORMJobDAO(db *gorm.DB) *GORMJobDAO {
	return &GORMJobDAO{db: db}
}

func (j *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := j.db.WithContext(ctx)
	// 乐观锁
	for {
		var job Job
		now := time.Now().UnixMilli()
		// 拿一个等待执行并且可以执行的任务
		// 或者 一个正在执行但续约失败的任务(status = 1 AND update_time < now - 2 * 续约间隔)
		// 认为连续两次续约失败
		err := db.Where("(status = ? AND next_time < ?) OR (status = ? AND update_time < ?)",
			jobStatusWaiting, now, now-(time.Minute*2).Milliseconds()).
			First(&job).Error
		if err != nil {
			return job, err
		}

		// 如果该任务的版本号没变就更新其版本号和状态
		// 表示抢到了该任务
		res := db.WithContext(ctx).
			Model(&Job{}).
			Where("id = ? AND version = ?", job.Id, job.Version).
			Updates(map[string]any{
				"status":      jobStatusRunning,
				"version":     job.Version + 1,
				"update_time": now,
			})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 没抢到
			continue
		}
		return job, err
	}
}

func (j *GORMJobDAO) Release(ctx context.Context, jid int64, version int) error {
	now := time.Now().UnixMilli()
	return j.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ? AND version = ?", jid, version).
		Updates(map[string]any{
			"status":      jobStatusWaiting,
			"update_time": now,
		}).Error
}

func (j *GORMJobDAO) UpdateTime(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return j.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ?", jid).
		Updates(map[string]any{
			"update_time": now,
		}).Error
}

func (j *GORMJobDAO) UpdateNextTime(ctx context.Context, jid int64, t time.Time) error {
	now := time.Now().UnixMilli()
	return j.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ?", jid).
		Updates(map[string]any{
			"update_time": now,
			"next_time":   t.UnixMilli(),
		}).Error
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Executor   string
	Expression string
	Config     string
	Status     int
	Version    int
	NextTime   int64 `gorm:"index"`

	UpdateTime int64
	CreateTime int64
}

const (
	// jobStatusWaiting 没人抢
	jobStatusWaiting = iota
	// jobStatusRunning 已经被人抢了
	jobStatusRunning
	// jobStatusPaused 不再需要调度了
	jobStatusPaused
)
