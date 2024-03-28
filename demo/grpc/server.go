package grpc

import (
	"context"
	"github.com/Anwenya/GeekTime/demo/grpc/pb"
)

type Server struct {
	pb.UnimplementedUserServiceServer
}

func (s *Server) GetById(context.Context, *pb.GetByIdRequest) (*pb.GetByIdResponse, error) {
	return &pb.GetByIdResponse{
		User: &pb.User{
			Id:   123,
			Name: "wlq",
		},
	}, nil
}
