package dao

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

type AsyncSMSDAO interface {
	CreateTask(ctx context.Context, as AsyncSMSTask) error
	GetWaitingTask(ctx context.Context) (AsyncSMSTask, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

const (
	asyncStatusWaiting = iota
	// 失败并且超过了重试次数
	asyncStatusFailed
	asyncStatusSuccess
)

type GORMAsyncSMSDAO struct {
	db *gorm.DB
}

func NewGORMAsyncSMSDAO(db *gorm.DB) AsyncSMSDAO {
	return &GORMAsyncSMSDAO{
		db: db,
	}
}

func (g *GORMAsyncSMSDAO) CreateTask(ctx context.Context, ast AsyncSMSTask) error {
	now := time.Now().UnixMilli()
	ast.CreateTime = now
	ast.UpdateTime = now
	return g.db.Create(&ast).Error
}

func (g *GORMAsyncSMSDAO) GetWaitingTask(ctx context.Context) (AsyncSMSTask, error) {
	// 如果在高并发情况下 SELECT for UPDATE 对数据库的压力很大
	// 但是我们不是高并发 因为部署N台机器 才有 N 个goroutine 来查询
	var ast AsyncSMSTask
	err := g.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 为了避开一些偶发性的失败 只找 1 分钟前的异步短信发送
			now := time.Now().UnixMilli()

			endTime := now - time.Minute.Milliseconds()
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where(
					"update_time < ? and status = ?",
					endTime,
					asyncStatusWaiting,
				).First(&ast).Error
			// SELECT xx FROM xxx WHERE xx FOR UPDATE，锁住了
			if err != nil {
				return err
			}

			// 只要更新了更新时间 根据前面的规则 就不可能被别的节点抢占了
			err = tx.Model(&AsyncSMSTask{}).
				Where("id = ?", ast.Id).
				Updates(map[string]any{
					"retry_cnt": gorm.Expr("retry_cnt + 1"),
					// 更新成了当前时间戳 确保我在发送过程中 没人会再次抢到它
					// 也相当于重试间隔一分钟
					"update_time": now,
				}).Error
			return err
		})
	return ast, err
}

// MarkSuccess 标记任务成功
func (g *GORMAsyncSMSDAO) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).
		Model(&AsyncSMSTask{}).
		Where("id =?", id).
		Updates(map[string]any{
			"update_time": now,
			"status":      asyncStatusSuccess,
		}).Error
}

// MarkFailed 标记任务失败
func (g *GORMAsyncSMSDAO) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).
		Model(&AsyncSMSTask{}).
		Where("id = ? AND `retry_cnt`>=`retry_max`", id).
		Updates(map[string]any{
			"update_time": now,
			"status":      asyncStatusFailed,
		}).Error
}

// AsyncSMSTask 发短信的任务
type AsyncSMSTask struct {
	Id     int64
	Config sqlx.JsonColumn[SMSConfig]
	// 当前重试次数
	RetryCnt int
	// 最大重试次数
	RetryMax   int
	Status     uint8
	CreateTime int64
	UpdateTime int64 `gorm:"index"`
}

func (AsyncSMSTask) TableName() string {
	return "async_sms_tasks"
}

type SMSConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}
