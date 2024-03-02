package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
}

// CachedArticleRepository
// 这里提供了两种思路及写法
// 不要被类似的字段搞混了
type CachedArticleRepository struct {
	// 同库不同表写法
	dao      dao.ArticleDAO
	cache    cache.ArticleCache
	userRepo UserRepository

	// 分库的写法
	rDAO dao.ArticleReaderDAO
	aDAO dao.ArticleAuthorDAO

	// 在这里控制事务就要耦合db
	db *gorm.DB

	l logger.LoggerV1
}

// NewCachedArticleRepository
// 同库不同表
func NewCachedArticleRepository(
	dao dao.ArticleDAO,
	cache cache.ArticleCache,
	userRepo UserRepository,
	l logger.LoggerV1,
) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
		l:        l,
	}
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

func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 假设默认limit为100 查询数量<=100的都可以走该缓存
	if offset == 0 && limit <= 100 {
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return res[:limit], err
		} else {
			c.l.Error("查询文章缓存失败", logger.Error(err))
		}
	}

	// 查数据库
	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}

	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})

	// 回写缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit == 100 {
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				c.l.Error("设置文章缓存失败", logger.Error(err))
			}
		}
	}()

	// 预写详情的缓存
	// 查询完列表后 大概率会点进去第一条文章的详情页
	// 这里可以尝试提前写入缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, res)
	}()
	return res, nil
}

// Create
// 同库不同表
func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(art))
	// 数据库有更新 需要删除缓存
	// 这也是 先写入数据库 后删除缓存的策略
	if err == nil {
		err := c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			c.l.Error("删除文章缓存失败", logger.Error(err))
		}
	}
	return id, err
}

// Update
// 同库不同表
func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	// 数据库有更新 需要删除缓存
	// 这也是 先写入数据库 后删除缓存的策略
	if err == nil {
		err := c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			c.l.Error("删除文章缓存失败", logger.Error(err))
		}
	}
	return err
}

// Sync
// 同库不同表
func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	// 数据库有更新 需要删除缓存
	// 这也是 先写入数据库 后删除缓存的策略
	if err == nil {
		err := c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			c.l.Error("删除文章缓存失败", logger.Error(err))
		}
	}

	// 在更发布后 直接尝试设置缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 可以灵活设置缓存的过期时间
		// 如影响力,预期流量大小,持续时间等
		user, err := c.userRepo.FindById(ctx, art.Author.Id)
		if err != nil {
			c.l.Error("查询用户信息失败", logger.Error(err))
			return
		}

		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}

		err = c.cache.SetPub(ctx, art)
		if err != nil {
			c.l.Error("设置文章缓存失败", logger.Error(err))
		}

	}()

	return id, err
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
	// 数据库有更新 需要删除缓存
	// 这也是 先写入数据库 后删除缓存的策略
	if err == nil {
		err := c.cache.DelFirstPage(ctx, uid)
		if err != nil {
			c.l.Error("删除文章缓存失败", logger.Error(err))
		}
	}
	return err
}

func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	// 查缓存
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	} else {
		c.l.Error("查询文章缓存失败", logger.Error(err))
	}
	// 查文章
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}

	res = c.toDomain(art)
	go func() {
		err := c.cache.Set(ctx, res)
		if err != nil {
			c.l.Error("设置文章缓存失败", logger.Error(err))
		}
	}()

	return res, nil
}

func (c *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	// 查缓存
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	} else {
		c.l.Error("查询发布文章缓存失败", logger.Error(err))
	}

	// 查已发布文章
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}

	// 查询用户信息
	res = c.toDomain(dao.Article(art))
	author, err := c.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
	}

	res.Author.Name = author.Nickname

	// 回写缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := c.cache.SetPub(ctx, res)
		if err != nil {
			c.l.Error("设置发布文章缓存失败", logger.Error(err))
		}
	}()

	return res, nil
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

func (c *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		CreateTime: time.UnixMilli(art.CreateTime),
		UpdateTime: time.UnixMilli(art.UpdateTime),
		Status:     domain.ArticleStatus(art.Status),
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			c.l.Error("预写文章缓存失败", logger.Error(err))
		}
	}
}
