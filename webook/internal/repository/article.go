package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
}

// CachedArticleRepository
// 这里提供了两种思路及写法
// 不要被类似的字段搞混了
type CachedArticleRepository struct {
	// 同库不同表写法
	dao dao.ArticleDAO

	// 分库的写法
	rDAO dao.ArticleReaderDAO
	aDAO dao.ArticleAuthorDAO

	// 在这里控制事务就要耦合db
	db *gorm.DB
}

// NewCachedArticleRepository
// 同库不同表
func NewCachedArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{dao: dao}
}

// NewCachedArticleRepositoryV2
// 分库写法 仅用于跑测试
func NewCachedArticleRepositoryV2(
	rDAO dao.ArticleReaderDAO,
	aDAO dao.ArticleAuthorDAO,
) *CachedArticleRepository {
	return &CachedArticleRepository{
		rDAO: rDAO,
		aDAO: aDAO,
	}
}

// Create
// 同库不同表
func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, c.toEntity(art))
}

// Update
// 同库不同表
func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, c.toEntity(art))
}

// Sync
// 同库不同表
func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Sync(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
}

// SyncV1
// 分库写法 无事务
func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)

	if id > 0 {
		err = c.aDAO.Update(ctx, artn)
	} else {
		id, err = c.aDAO.Create(ctx, artn)
	}

	if err != nil {
		return 0, err
	}

	artn.Id = id
	// 如果写入读者库失败可以引入重试
	err = c.rDAO.UpsertV2(ctx, artn)
	return id, err
}

// SyncV2
// 分库写法 有事务
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	defer tx.Rollback()

	aDAO := dao.NewArticleGORMAuthorDAO(tx)
	rDAO := dao.NewArticleGORMReaderDAO(tx)

	artn := c.toEntity(art)

	var (
		id  = art.Id
		err error
	)

	if id > 0 {
		err = aDAO.Update(ctx, artn)
	} else {
		id, err = aDAO.Create(ctx, artn)
	}

	if err != nil {
		return 0, err
	}

	artn.Id = id
	err = rDAO.UpsertV2(ctx, artn)
	if err != nil {
		return 0, err
	}

	tx.Commit()
	return id, nil
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}
