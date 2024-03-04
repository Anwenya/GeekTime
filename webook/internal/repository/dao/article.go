package dao

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
}

// GORMArticleDAO
// 这里的例子是读和写对应到同一个数据库的两张表
// 这样可以通过数据库事务来保证 发表文章 这个操作的原子性
type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

// GetByAuthor 查询文章列表
func (a *GORMArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := a.db.WithContext(ctx).Where("author_id = ?", uid).
		Offset(offset).Limit(limit).Order("update_time desc").
		Find(&arts).Error

	return arts, err
}

// Insert 插入到作者表
func (a *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.CreateTime = now
	art.UpdateTime = now
	// Create inserts value,
	// returning the inserted data's primary key in value's id
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

// UpdateById 更新到作者表
func (a *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
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
func (a *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		dao := NewGORMArticleDAO(tx)
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
func (a *GORMArticleDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
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

func (a *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := a.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	return art, err
}

func (a *GORMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pa PublishedArticle
	err := a.db.WithContext(ctx).Where("id = ?", id).First(&pa).Error
	return pa, err
}

// SyncV1 同步作者表的文章到读者表
// 自己控制事务的写法
func (a *GORMArticleDAO) SyncV1(ctx context.Context, art Article) (int64, error) {
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
	dao := NewGORMArticleDAO(tx)
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
	Id         int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title      string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content    string `gorm:"type=BLOB" bson:"content,omitempty"`
	AuthorId   int64  `gorm:"index" bson:"author_id,omitempty"`
	Status     uint8  `bson:"status,omitempty"`
	CreateTime int64  `bson:"create_time,omitempty"`
	UpdateTime int64  `bson:"update_time,omitempty"`
}

func (Article) TableName() string {
	return "articles"
}

// PublishedArticle 文章在发布领域内的实体
type PublishedArticle Article

func (PublishedArticle) TableName() string {
	return "published_articles"
}

// ArticleMongoDBDAO mongodb版本
type ArticleMongoDBDAO struct {
	sf      *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func NewArticleMongoDBDAO(db *mongo.Database, sf *snowflake.Node) ArticleDAO {
	return &ArticleMongoDBDAO{
		sf:      sf,
		liveCol: db.Collection("published_articles"),
		col:     db.Collection("articles"),
	}
}

// Insert 插入到作者集合
func (a *ArticleMongoDBDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.CreateTime = now
	art.UpdateTime = now
	art.Id = a.sf.Generate().Int64()
	_, err := a.col.InsertOne(ctx, &art)
	return art.Id, err
}

func (a *ArticleMongoDBDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	filter := bson.D{
		bson.E{Key: "id", Value: art.Id},
		bson.E{Key: "author_id", Value: art.AuthorId},
	}

	set := bson.D{
		bson.E{
			Key: "$set",
			Value: bson.M{
				"title":       art.Title,
				"content":     art.Content,
				"status":      art.Status,
				"update_time": now,
			},
		},
	}

	res, err := a.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("要更新的文章不存在或者创作者不匹配")
	}
	return nil
}

func (a *ArticleMongoDBDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		// 已经存在id则为更新操作
		err = a.UpdateById(ctx, art)
	} else {
		// 新建文章是没有id的
		id, err = a.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	// 更新或者新建成功后同步到读者表
	art.Id = id
	now := time.Now().UnixMilli()
	art.UpdateTime = now
	// 插入或者更新
	filter := bson.D{
		bson.E{Key: "id", Value: art.Id},
		bson.E{Key: "author_id", Value: art.AuthorId},
	}

	set := bson.D{
		bson.E{Key: "$set", Value: art},
		bson.E{
			Key: "$setOnInsert",
			Value: bson.D{
				bson.E{Key: "create_time", Value: now},
			},
		},
	}

	// 如果设置为true 则在没有文档与查询条件匹配时创建新文档 默认的 value 是false
	_, err = a.liveCol.UpdateOne(ctx, filter, set, options.Update().SetUpsert(true))

	return id, err
}

func (a *ArticleMongoDBDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	now := time.Now().UnixMilli()

	filter := bson.D{
		bson.E{Key: "id", Value: id},
		bson.E{Key: "author_id", Value: uid},
	}

	sets := bson.D{
		bson.E{
			Key: "$set",
			Value: bson.D{
				bson.E{Key: "status", Value: status},
				bson.E{Key: "update_time", Value: now},
			},
		},
	}
	res, err := a.col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return errors.New("id不存在或者创作者不匹配")
	}
	_, err = a.liveCol.UpdateOne(ctx, filter, sets)
	return err
}

func (a *ArticleMongoDBDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	panic(any("implement me"))
}

func (a *ArticleMongoDBDAO) GetById(ctx context.Context, id int64) (Article, error) {
	panic(any("implement me"))
}

func (a *ArticleMongoDBDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	panic(any("implement me"))
}
