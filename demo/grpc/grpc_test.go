package grpc

import (
	"context"
	"github.com/Anwenya/GeekTime/demo/grpc/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	gs := grpc.NewServer()
	us := &Server{}
	pb.RegisterUserServiceServer(gs, us)
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	err = gs.Serve(l)
	t.Log(err)
}

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := pb.NewUserServiceClient(cc)
	resp, err := client.GetById(context.Background(), &pb.GetByIdRequest{Id: 123})
	require.NoError(t, err)
	t.Log(resp)
}
