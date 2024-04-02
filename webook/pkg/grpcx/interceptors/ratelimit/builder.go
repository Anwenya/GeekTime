package ratelimit

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/pkg/limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type InterceptorBuilder struct {
	limiter limiter.Limiter
	key     string
}

// NewInterceptor
// key应明确标识出限流的范围
func NewInterceptor(limiter limiter.Limiter, key string) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {

		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			// 保守做法
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}
		if limited {
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}
		return handler(ctx, req)
	}
}

// BuildServerUnaryInterceptorService 针对服务的限流
func (b *InterceptorBuilder) BuildServerUnaryInterceptorService() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		// limiter:UserService
		if strings.HasPrefix(info.FullMethod, "/UserService") {
			limited, err := b.limiter.Limit(ctx, b.key)
			if err != nil {
				// 保守做法
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
		}
		return handler(ctx, req)
	}
}

// BuildServerUnaryInterceptorMethod 针对方法的限流
func (b *InterceptorBuilder) BuildServerUnaryInterceptorMethod() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil || limited {
			// 具体的方法再去解析该值 之后做出相应的处理
			ctx = context.WithValue(ctx, "downgrade", "true")
		}
		return handler(ctx, req)
	}
}
