package servicegovernance

import (
	"context"
	"fmt"
	"github.com/Anwenya/GeekTime/demo/servicegovernance/pb"
	"github.com/Anwenya/GeekTime/webook/pkg/limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	limiter limiter.Limiter
	key     string
}

func NewInterceptorBuilder(limiter limiter.Limiter, key string) *InterceptorBuilder {
	return &InterceptorBuilder{
		limiter: limiter,
		key:     key,
	}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptorBiz() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		if getReq, ok := req.(pb.GetByIdRequest); ok {
			key := fmt.Sprintf("limiter:user:get_by_id:%d", getReq.Id)
			limited, err := b.limiter.Limit(ctx, key)
			if err != nil {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
		}
		return handler(ctx, req)
	}
}
