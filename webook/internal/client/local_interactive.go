package client

import (
	"context"
	interactivev1 "github.com/Anwenya/GeekTime/webook/api/proto/gen/interactive/v1"
	"github.com/Anwenya/GeekTime/webook/interactive/domain"
	"github.com/Anwenya/GeekTime/webook/interactive/service"
	"google.golang.org/grpc"
)

type LocalInteractiveServiceAdapter struct {
	svc service.InteractiveService
}

func NewLocalInteractiveServiceAdapter(svc service.InteractiveService) *LocalInteractiveServiceAdapter {
	return &LocalInteractiveServiceAdapter{svc: svc}
}

func (l *LocalInteractiveServiceAdapter) IncrReadCnt(ctx context.Context, in *interactivev1.IncrReadCntRequest, opts ...grpc.CallOption) (*interactivev1.IncrReadCntResponse, error) {
	err := l.svc.IncrReadCnt(ctx, in.GetBiz(), in.GetBizId())
	return &interactivev1.IncrReadCntResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) Like(ctx context.Context, in *interactivev1.LikeRequest, opts ...grpc.CallOption) (*interactivev1.LikeResponse, error) {
	err := l.svc.Like(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &interactivev1.LikeResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) CancelLike(ctx context.Context, in *interactivev1.CancelLikeRequest, opts ...grpc.CallOption) (*interactivev1.CancelLikeResponse, error) {
	err := l.svc.CancelLike(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &interactivev1.CancelLikeResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) Collect(ctx context.Context, in *interactivev1.CollectRequest, opts ...grpc.CallOption) (*interactivev1.CollectResponse, error) {
	err := l.svc.Collect(ctx, in.GetBiz(), in.GetBizId(), in.GetCid(), in.GetUid())
	return &interactivev1.CollectResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) Get(ctx context.Context, in *interactivev1.GetRequest, opts ...grpc.CallOption) (*interactivev1.GetResponse, error) {
	intr, err := l.svc.Get(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	if err != nil {
		return nil, err
	}
	return &interactivev1.GetResponse{
		Interactive: l.toDTO(intr),
	}, err
}

func (l *LocalInteractiveServiceAdapter) GetByIds(ctx context.Context, in *interactivev1.GetByIdsRequest, opts ...grpc.CallOption) (*interactivev1.GetByIdsResponse, error) {
	res, err := l.svc.GetByIds(ctx, in.GetBiz(), in.GetIds())
	if err != nil {
		return nil, err
	}
	intrs := make(map[int64]*interactivev1.Interactive, len(res))
	for k, v := range res {
		intrs[k] = l.toDTO(v)
	}
	return &interactivev1.GetByIdsResponse{
		Interactives: intrs,
	}, nil
}

func (l *LocalInteractiveServiceAdapter) toDTO(interactive domain.Interactive) *interactivev1.Interactive {
	return &interactivev1.Interactive{
		Biz:        interactive.Biz,
		BizId:      interactive.BizId,
		ReadCnt:    interactive.ReadCnt,
		CollectCnt: interactive.CollectCnt,
		Collected:  interactive.Collected,
		Liked:      interactive.Liked,
		LikeCnt:    interactive.LikeCnt,
	}
}
