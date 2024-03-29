package etcd

import (
	"context"
	"github.com/Anwenya/GeekTime/demo/etcd/pb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"testing"
	"time"
)

type GoZeroTestSuite struct {
	suite.Suite
}

func (g *GoZeroTestSuite) TestGoZeroClient() {
	// 本质还是从etcd中获得可用的服务地址
	// 再连接上服务端
	zClient := zrpc.MustNewClient(
		zrpc.RpcClientConf{
			Etcd: discov.EtcdConf{
				Hosts: []string{"192.168.2.130:12379"},
				Key:   "user",
			},
		},
	)

	client := pb.NewUserServiceClient(zClient.Conn())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetById(ctx, &pb.GetByIdRequest{Id: 123})
	require.NoError(g.T(), err)
	g.T().Log(resp)
}

func (g *GoZeroTestSuite) TestGoZeroServer() {
	config := zrpc.RpcServerConf{
		ListenOn: ":8090",
		Etcd: discov.EtcdConf{
			Hosts: []string{"192.168.2.130:12379"},
			Key:   "user",
		},
	}

	// 创建实例并注册
	server := zrpc.MustNewServer(
		config,
		func(server *grpc.Server) {
			pb.RegisterUserServiceServer(server, &Server{})
		},
	)

	server.Start()
}

func TestGoZero(t *testing.T) {
	suite.Run(t, new(GoZeroTestSuite))
}
