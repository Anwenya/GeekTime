package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error
}

// ArticleGORMDAO
// 这里的例子是读和写对应到同一个数据库的两张表
// 这样可以通过数据库事务来保证 发表文章 这个操作的原子性
type ArticleGORMDAO struct {
	db *gorm.DB
}

func NewArticleGORMDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGORMDAO{
		db: db,
	}
}

// Insert 插入到作者表
func (a *ArticleGORMDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.CreateTime = now
	art.UpdateTime = now
	// Create inserts value,
	// returning the inserted data's primary key in value's id
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

// UpdateById 更新到作者表
func (a *ArticleGORMDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := a.db.WithContext(ctx).Model(&art).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":       art.Title,
			"content":     art.Content,
			"status":      art.Status,
			"update_time": now,
		})
	if res.Error != nil {
		return res.Error
	}

	// 更新操作实际没有更新
	// 可能文章不存在 或者 作者不匹配
	if res.RowsAffected == 0 {
		return errors.New("要更新的文章不存在或者创作者不匹配")
	}
	return nil
}

// Sync 同步作者表的文章到读者表
// 传统写法 使用现成的Transaction方法
func (a *ArticleGORMDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		dao := NewArticleGORMDAO(tx)
		if id > 0 {
			// 已经存在id则为更新操作
			err = dao.UpdateById(ctx, art)
		} else {
			// 新建文章是没有id的
			id, err = dao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		// 更新或者新建成功后同步到读者表
		art.Id = id
		now := time.Now().UnixMilli()
		pa := PublishedArticle(art)
		pa.CreateTime = now
		pa.UpdateTime = now
		// 插入或者更新
		err = tx.Clauses(
			clause.OnConflict{
				// 方言
				// mysql : INSERT xxx ON DUPLICATE KEY SET `title`=?
				// sqlite: INSERT XXX ON CONFLICT DO UPDATES WHERE
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.Assignments(
					map[string]interface{}{
						"title":       pa.Title,
						"content":     pa.Content,
						"status":      pa.Status,
						"update_time": now,
					},
				),
			},
		).Create(&pa).Error
		return err
	})
	return id, err
}

// SyncStatus 更新文章状态
func (a *ArticleGORMDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	now := time.Now().UnixMilli()
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", uid, id).
			Updates(map[string]any{
				"update_time": now,
				"status":      status,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("id不存在或者创作者不匹配")
		}
		return tx.Model(&PublishedArticle{}).
			Where("id = ?", uid).
			Updates(map[string]any{
				"update_time": now,
				"status":      status,
			}).Error
	})
}

// SyncV1 同步作者表的文章到读者表
// 自己控制事务的写法
func (a *ArticleGORMDAO) SyncV1(ctx context.Context, art Article) (int64, error) {
	tx := a.db.WithContext(ctx).Begin()
	// 开启事务失败
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 如果commit成功 这里会回滚失败 但不会影响最终结果
	// 如果commit失败或其他异常发生 这里会正常回滚
	defer tx.Rollback()

	// 其余逻辑与上面的写法基本一致
	var (
		id  = art.Id
		err error
	)
	dao := NewArticleGORMDAO(tx)
	if id > 0 {
		err = dao.UpdateById(ctx, art)
	} else {
		id, err = dao.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}

	art.Id = id
	now := time.Now().UnixMilli()
	pa := PublishedArticle(art)
	pa.CreateTime = now
	pa.UpdateTime = now

	// 插入或者更新
	err = tx.Clauses(
		clause.OnConflict{
			// 方言
			// mysql : INSERT xxx ON DUPLICATE KEY SET `title`=?
			// sqlite: INSERT XXX ON CONFLICT DO UPDATES WHERE
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(
				map[string]interface{}{
					"title":       pa.Title,
					"content":     pa.Content,
					"status":      pa.Status,
					"update_time": now,
				},
			),
		},
	).Create(&pa).Error

	if err != nil {
		return 0, err
	}
	// 提交事务
	tx.Commit()
	return id, nil
}

type Article struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Title      string `gorm:"type=varchar(4096)"`
	Content    string `gorm:"type=BLOB"`
	AuthorId   int64  `gorm:"index"`
	Status     uint8
	CreateTime int64
	UpdateTime int64
}

func (Article) TableName() string {
	return "articles"
}

// PublishedArticle 文章在发布领域内的实体
type PublishedArticle Article

func (PublishedArticle) TableName() string {
	return "published_articles"
}
