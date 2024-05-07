package dao

import "context"

type TagDAO interface {
	CreateTag(ctx context.Context, tag Tag) (int64, error)
	CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error
	GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error)
	GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error)
	GetTags(ctx context.Context, offset, limit int) ([]Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]Tag, error)
}

type Tag struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 联合唯一索引 <uid, name>
	Name string `gorm:"type=varchar(4096)"`
	// 有一个典型的场景，是查出一个人有什么标签
	Uid        int64 `gorm:"index"`
	CreateTime int64
	UpdateTime int64
}

type TagBiz struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	BizId int64  `gorm:"index:biz_type_id"`
	Biz   string `gorm:"index:biz_type_id"`
	// 冗余字段，加快查询和删除
	Uid int64 `gorm:"index"`
	Tid int64

	// TagName string
	Tag        *Tag  `gorm:"ForeignKey:Tid;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	CreateTime int64 `bson:"create_time,omitempty"`
	UpdateTime int64 `bson:"update_time,omitempty"`
}
