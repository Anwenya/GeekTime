package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/ecodeclub/ekit/sqlx"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

type AsyncSMSRepository interface {
	CreateTask(ctx context.Context, as domain.AsyncSMS) error
	PreemptWaitingTask(ctx context.Context) (domain.AsyncSMS, error)
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}

type asyncSmsRepository struct {
	dao dao.AsyncSMSDAO
}

func NewAsyncSMSRepository(dao dao.AsyncSMSDAO) AsyncSMSRepository {
	return &asyncSmsRepository{
		dao: dao,
	}
}

// CreateTask 创建一个任务
func (a *asyncSmsRepository) CreateTask(ctx context.Context, as domain.AsyncSMS) error {
	return a.dao.CreateTask(
		ctx,
		dao.AsyncSMSTask{
			Config: sqlx.JsonColumn[dao.SMSConfig]{
				Val: dao.SMSConfig{
					TplId:   as.TplId,
					Args:    as.Args,
					Numbers: as.Numbers,
				},
				Valid: true,
			},
			RetryMax: as.RetryMax,
		},
	)
}

// PreemptWaitingTask 抢占式获得一个任务
func (a *asyncSmsRepository) PreemptWaitingTask(ctx context.Context) (domain.AsyncSMS, error) {
	ast, err := a.dao.GetWaitingTask(ctx)
	if err != nil {
		return domain.AsyncSMS{}, err
	}
	return domain.AsyncSMS{
		Id:       ast.Id,
		TplId:    ast.Config.Val.TplId,
		Numbers:  ast.Config.Val.Numbers,
		Args:     ast.Config.Val.Args,
		RetryMax: ast.RetryMax,
	}, nil
}

// ReportScheduleResult 更新任务状态
func (a *asyncSmsRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return a.dao.MarkSuccess(ctx, id)
	}
	return a.dao.MarkFailed(ctx, id)
}
