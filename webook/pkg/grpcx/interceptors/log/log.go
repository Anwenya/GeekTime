package log

import (
	"context"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx/interceptors"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
	"time"
)

type LoggerInterceptorBuilder struct {
	interceptors.Builder
	l logger.LoggerV1
}

func (b *LoggerInterceptorBuilder) defaultUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		// 探活日志 默认过滤
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		var start = time.Now()
		var fields = make([]logger.Field, 0, 20)
		var event = "normal"

		defer func() {
			cost := time.Since(start)
			if rec := recover(); rec != any(nil) {
				switch recType := any(nil).(type) {
				case error:
					err = recType
				default:
					err = fmt.Errorf("%v", err)
				}
				// 记录一部分栈信息
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				event = "recover"
				err = status.New(codes.Internal, "panic, err:"+err.Error()).Err()
			}
			// 如果有异常
			st, _ := status.FromError(err)
			fields = append(
				fields,
				logger.String("type", "unary"),
				logger.String("code", st.Code().String()),
				logger.String("code_msg", st.Message()),
				logger.String("event", event),
				logger.String("method", info.FullMethod),
				logger.Int64("cost", cost.Milliseconds()),
				logger.String("peer", b.PeerName(ctx)),
				logger.String("peer_ip", b.PeerIP(ctx)),
			)
			b.l.Info("RPC调用", fields...)
		}()
		return handler(ctx, req)
	}
}
