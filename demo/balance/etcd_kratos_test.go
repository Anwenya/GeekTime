package balance

import (
	"context"
	"github.com/Anwenya/GeekTime/demo/balance/pb"
	kratosetcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/random"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

type KratosTestSuite struct {
	suite.Suite
	etcdClient *etcdv3.Client
}

func (k *KratosTestSuite) SetupSuite() {
	client, err := etcdv3.New(
		etcdv3.Config{
			Endpoints: []string{"192.168.2.130:12379"},
		},
	)
	require.NoError(k.T(), err)
	k.etcdClient = client
}

func (k *KratosTestSuite) TestClient() {
	// kratos默认是wrr算法
	r := kratosetcd.New(k.etcdClient)
	cc, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	require.NoError(k.T(), err)
	defer cc.Close()

	client := pb.NewUserServiceClient(cc)

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &pb.GetByIdRequest{
			Id: 123,
		})
		cancel()
		require.NoError(k.T(), err)
		k.T().Log(resp.User)
	}

}

func (k *KratosTestSuite) TestClientLoadBalancer() {
	// kratos默认是wrr算法
	// 指定算法
	selector.SetGlobalSelector(random.NewBuilder())

	r := kratosetcd.New(k.etcdClient)
	cc, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	require.NoError(k.T(), err)
	defer cc.Close()

	client := pb.NewUserServiceClient(cc)

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &pb.GetByIdRequest{
			Id: 123,
		})
		cancel()
		require.NoError(k.T(), err)
		k.T().Log(resp.User)
	}

}

func (k *KratosTestSuite) TestServer() {
	go func() {
		k.startServer(":8090")
	}()
	k.startServer(":8091")
}

func (k *KratosTestSuite) startServer(address string) {
	grpcServer := grpc.NewServer(
		grpc.Address(address),
		grpc.Middleware(recovery.Recovery()),
	)
	pb.RegisterUserServiceServer(grpcServer, &Server{})
	// 注册
	r := kratosetcd.New(k.etcdClient)
	app := kratos.New(
		kratos.Name("user"),
		kratos.Server(grpcServer),
		kratos.Registrar(r),
	)
	app.Run()
}

func TestKratos(t *testing.T) {
	suite.Run(t, new(KratosTestSuite))
}
