package balance

import (
	"context"
	"github.com/Anwenya/GeekTime/demo/balance/pb"
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

	// 替换gozero默认的算法
	zrpc.WithDialOption(
		grpc.WithDefaultServiceConfig(
			`{"loadBalancingConfig": [{"round_robin":{}}]}`,
		),
	)

	client := pb.NewUserServiceClient(zClient.Conn())

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &pb.GetByIdRequest{Id: 123})
		cancel()
		require.NoError(g.T(), err)
		g.T().Log(resp)
	}
}

func (g *GoZeroTestSuite) TestGoZeroServer() {
	go func() {
		g.startServer(":8090")
	}()
	g.startServer(":8091")
}

func (g *GoZeroTestSuite) startServer(address string) {
	config := zrpc.RpcServerConf{
		ListenOn: address,
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
