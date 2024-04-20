package dao

import (
	"context"
)

const (
	FollowRelationStatusUnknown uint8 = iota
	FollowRelationStatusActive
	FollowRelationStatusInactive
)

type FollowDao interface {
	// FollowRelationList 获取某人的关注列表
	FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error)
	FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error)
	// CreateFollowRelation 创建联系人
	CreateFollowRelation(ctx context.Context, c FollowRelation) error
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error
	// CntFollower 统计计算关注自己的人有多少
	CntFollower(ctx context.Context, uid int64) (int64, error)
	// CntFollowee 统计自己关注了多少人
	CntFollowee(ctx context.Context, uid int64) (int64, error)
}

type FollowStatics struct {
	Id  int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid int64 `gorm:"unique"`
	// 有多少粉丝
	Followers int64
	// 关注了多少人
	Followees int64

	UpdateTime int64
	CreateTime int64
}

// FollowRelation 存储用户的关注数据
type FollowRelation struct {
	Id int64 `gorm:"column:id;autoIncrement;primaryKey;"`

	// 要在这两个列上，创建一个联合唯一索引
	// 如果你认为查询一个人关注了哪些人，是主要查询场景
	// <follower, followee>
	// 如果你认为查询一个人有哪些粉丝，是主要查询场景
	// <followee, follower>
	// 我查我关注了哪些人？ WHERE follower = 123(我的 uid)
	// 也可以额外创建一个索引
	Follower int64 `gorm:"uniqueIndex:follower_followee"`
	Followee int64 `gorm:"uniqueIndex:follower_followee"`

	// 软删除策略
	Status uint8

	// 如果你的关注有类型的，有优先级，有一些备注数据的
	// Type string
	// Priority string
	// Gid 分组ID

	CreateTime int64
	UpdateTime int64
}
