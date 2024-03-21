package client

import (
	"context"
	interactivev1 "github.com/Anwenya/GeekTime/webook/api/proto/gen/interactive/v1"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"google.golang.org/grpc"
	"math/rand"
)

type InteractiveClient struct {
	remote interactivev1.InteractiveServiceClient
	local  interactivev1.InteractiveServiceClient

	threshold *atomicx.Value[int32]
}

func NewInteractiveClient(
	remote interactivev1.InteractiveServiceClient,
	local interactivev1.InteractiveServiceClient,
) *InteractiveClient {
	return &InteractiveClient{remote: remote, local: local}
}

func (i *InteractiveClient) selectClient() interactivev1.InteractiveServiceClient {
	num := rand.Int31n(100)
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}

func (i *InteractiveClient) UpdateThreshold(threshold int32) {
	i.threshold.Store(threshold)
}

func (i *InteractiveClient) IncrReadCnt(ctx context.Context, in *interactivev1.IncrReadCntRequest, opts ...grpc.CallOption) (*interactivev1.IncrReadCntResponse, error) {

	return i.selectClient().IncrReadCnt(ctx, in, opts...)
}

func (i *InteractiveClient) Like(ctx context.Context, in *interactivev1.LikeRequest, opts ...grpc.CallOption) (*interactivev1.LikeResponse, error) {
	return i.selectClient().Like(ctx, in, opts...)
}

func (i *InteractiveClient) CancelLike(ctx context.Context, in *interactivev1.CancelLikeRequest, opts ...grpc.CallOption) (*interactivev1.CancelLikeResponse, error) {
	return i.selectClient().CancelLike(ctx, in, opts...)
}

func (i *InteractiveClient) Collect(ctx context.Context, in *interactivev1.CollectRequest, opts ...grpc.CallOption) (*interactivev1.CollectResponse, error) {
	return i.selectClient().Collect(ctx, in, opts...)
}

func (i *InteractiveClient) Get(ctx context.Context, in *interactivev1.GetRequest, opts ...grpc.CallOption) (*interactivev1.GetResponse, error) {
	return i.selectClient().Get(ctx, in, opts...)
}

func (i *InteractiveClient) GetByIds(ctx context.Context, in *interactivev1.GetByIdsRequest, opts ...grpc.CallOption) (*interactivev1.GetByIdsResponse, error) {
	return i.selectClient().GetByIds(ctx, in, opts...)
}
