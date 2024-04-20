package cache

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/follow/domain"
)

const (
	// 被多少人关注
	fieldFollowerCnt = "follower_cnt"
	// 关注了多少人
	fieldFolloweeCnt = "followee_cnt"
)

type FollowCache interface {
	StaticsInfo(ctx context.Context, uid int64) (domain.FollowStatics, error)
	SetStaticsInfo(ctx context.Context, uid int64, statics domain.FollowStatics) error
	Follow(ctx context.Context, follower, followee int64) error
	CancelFollow(ctx context.Context, follower, followee int64) error
}
