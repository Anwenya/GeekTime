package servicegovernance

import (
	"context"
	"github.com/Anwenya/GeekTime/demo/servicegovernance/pb"
	"github.com/Anwenya/GeekTime/webook/pkg/grpcx/interceptors/trace"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type InterceptorTestSuite struct {
	suite.Suite
}

func (s *InterceptorTestSuite) TestClient() {
	t := s.T()
	initZipkin()
	cc, err := grpc.Dial("127.0.0.1:8090",
		grpc.WithChainUnaryInterceptor(
			trace.NewOTELInterceptorBuilder("client_test", nil, nil).
				BuildUnaryClientInterceptor(),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	require.NoError(t, err)

	// 创建客户端
	client := pb.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	time.Sleep(time.Millisecond * 100)
	resp, err := client.GetById(
		ctx,
		&pb.GetByIdRequest{
			Id: 123,
		},
	)
	require.NoError(t, err)
	t.Log(resp.User)
	time.Sleep(time.Second)
}

func (s *InterceptorTestSuite) TestServer() {
	initZipkin()
	t := s.T()
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			NewLogInterceptor(t),
			trace.NewOTELInterceptorBuilder("server_test", nil, nil).
				BuildUnaryServerInterceptor(),
		),
	)

	pb.RegisterUserServiceServer(
		server,
		&Server{
			Name: "interceptor_test",
		},
	)

	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	err = server.Serve(l)
	require.NoError(t, err)
}

func NewLogInterceptor(t *testing.T) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		t.Log("请求处理前", req, info)
		resp, err = handler(ctx, req)
		t.Log("请求处理后", resp, err)
		return
	}
}

func TestInterceptorTestSuite(t *testing.T) {
	suite.Run(t, new(InterceptorTestSuite))
}
