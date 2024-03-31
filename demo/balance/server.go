package balance

import (
	"context"
	"github.com/Anwenya/GeekTime/demo/balance/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type Server struct {
	pb.UnimplementedUserServiceServer
	name string
}

func (s *Server) GetById(
	ctx context.Context,
	req *pb.GetByIdRequest,
) (*pb.GetByIdResponse, error) {
	log.Println("命中服务器", s.name)
	return &pb.GetByIdResponse{
		User: &pb.User{
			Id:   req.Id,
			Name: s.name,
		},
	}, nil
}

// FailoverServer 用于模拟服务异常的情况
type FailoverServer struct {
	pb.UnimplementedUserServiceServer
	code codes.Code
}

func (fs *FailoverServer) GetById(
	ctx context.Context,
	req *pb.GetByIdRequest,
) (*pb.GetByIdResponse, error) {
	log.Println("命中了failover服务器")
	return &pb.GetByIdResponse{
		User: &pb.User{
			Name: "failover",
		},
	}, status.Error(fs.code, "模拟failover")
}
