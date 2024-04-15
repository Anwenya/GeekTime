package grpc

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/api/proto/gen/reward/v1"
	"github.com/Anwenya/GeekTime/webook/reward/domain"
	"github.com/Anwenya/GeekTime/webook/reward/service"
	"google.golang.org/grpc"
)

type RewardServiceServer struct {
	rewardv1.UnimplementedRewardServiceServer
	svc service.RewardService
}

func NewRewardServiceServer(svc service.RewardService) *RewardServiceServer {
	return &RewardServiceServer{svc: svc}
}

func (r *RewardServiceServer) PreReward(ctx context.Context, request *rewardv1.PreRewardRequest) (*rewardv1.PreRewardResponse, error) {
	codeUrl, err := r.svc.PreReward(
		ctx,
		domain.Reward{
			Uid: request.Uid,
			Target: domain.Target{
				Biz:     request.Biz,
				BizId:   request.BizId,
				BizName: request.BizName,
				Uid:     request.Uid,
			},
			Amount: request.Amt,
		},
	)
	return &rewardv1.PreRewardResponse{
		CodeUrl: codeUrl.URL,
		Rid:     codeUrl.Rid,
	}, err
}

func (r *RewardServiceServer) GetReward(ctx context.Context, request *rewardv1.GetRewardRequest) (*rewardv1.GetRewardResponse, error) {
	rw, err := r.svc.GetReward(ctx, request.GetRid(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &rewardv1.GetRewardResponse{
		Status: rewardv1.RewardStatus(rw.Status),
	}, nil
}

func (r *RewardServiceServer) Register(server *grpc.Server) {
	rewardv1.RegisterRewardServiceServer(server, r)
}