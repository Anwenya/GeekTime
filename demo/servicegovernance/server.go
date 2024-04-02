package servicegovernance

import (
	"context"
	"fmt"
	"github.com/Anwenya/GeekTime/demo/servicegovernance/pb"
	"github.com/Anwenya/GeekTime/webook/pkg/limiter"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type Server struct {
	pb.UnimplementedUserServiceServer
	Name string
}

func (s *Server) GetById(
	ctx context.Context,
	request *pb.GetByIdRequest,
) (*pb.GetByIdResponse, error) {
	ctx, span := otel.Tracer("server_biz").Start(ctx, "get_by_id")
	defer span.End()
	ddl, ok := ctx.Deadline()
	if ok {
		rest := ddl.Sub(time.Now())
		log.Println(rest.String())
	}
	time.Sleep(time.Millisecond * 100)
	return &pb.GetByIdResponse{
		User: &pb.User{
			Id:   123,
			Name: "from" + s.Name,
		},
	}, nil
}

// LimiterServer 限流装饰器
type LimiterServer struct {
	limiter limiter.Limiter
	pb.UserServiceServer
}

func (l *LimiterServer) GetById(
	ctx context.Context,
	req *pb.GetByIdRequest,
) (*pb.GetByIdResponse, error) {
	key := fmt.Sprintf("limiter:user:get_by_id:%d", req.Id)
	limited, err := l.limiter.Limit(ctx, key)
	if err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}

	if limited {
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
	return l.UserServiceServer.GetById(ctx, req)
}
