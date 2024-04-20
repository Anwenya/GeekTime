package service

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/follow/domain"
	"github.com/Anwenya/GeekTime/webook/follow/repository"
)

type followService struct {
	repo repository.FollowRepository
}

func NewFollowService(repo repository.FollowRepository) FollowService {
	return &followService{repo: repo}
}

func (f *followService) GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error) {
	return f.repo.GetFollowee(ctx, follower, offset, limit)
}

func (f *followService) FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error) {
	val, err := f.repo.FollowInfo(ctx, follower, followee)
	return val, err
}

func (f *followService) Follow(ctx context.Context, follower, followee int64) error {
	return f.repo.AddFollowRelation(ctx, domain.FollowRelation{
		Followee: followee,
		Follower: follower,
	})
}

func (f *followService) CancelFollow(ctx context.Context, follower, followee int64) error {
	return f.repo.InactiveFollowRelation(ctx, follower, followee)
}
