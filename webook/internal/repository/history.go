package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"time"
)

type ReadHistoryRepository interface {
	AddRecord(ctx context.Context, record domain.ReadHistory) error
}

// CachedReadHistoryRepository
// todo 要不要加缓存? 感觉观看历史更新很频繁 且使用率应该不高
type CachedReadHistoryRepository struct {
	dao dao.HistoryDao
}

func NewCachedReadHistoryRepository(dao dao.HistoryDao) ReadHistoryRepository {
	return &CachedReadHistoryRepository{dao: dao}
}

func (c *CachedReadHistoryRepository) AddRecord(ctx context.Context, history domain.ReadHistory) error {
	return c.dao.UpsertReadHistory(ctx, c.toEntity(history))
}

func (c *CachedReadHistoryRepository) toEntity(history domain.ReadHistory) dao.ReadHistory {

	return dao.ReadHistory{
		Uid:      history.Uid,
		BizId:    history.BizId,
		Biz:      history.Biz,
		ReadTime: history.ReadTime.UnixMilli(),
	}

}

func (c *CachedReadHistoryRepository) toDomain(history dao.ReadHistory) domain.ReadHistory {
	return domain.ReadHistory{
		Uid:      history.Uid,
		BizId:    history.BizId,
		Biz:      history.Biz,
		ReadTime: time.UnixMilli(history.ReadTime),
	}
}
