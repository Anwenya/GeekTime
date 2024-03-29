package etcd

import (
	"context"
	"github.com/Anwenya/GeekTime/demo/etcd/pb"
	kratosetcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
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
	r := kratosetcd.New(k.etcdClient)
	cc, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	require.NoError(k.T(), err)
	defer cc.Close()

	client := pb.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetById(ctx, &pb.GetByIdRequest{
		Id: 123,
	})
	require.NoError(k.T(), err)
	k.T().Log(resp.User)
}

func (k *KratosTestSuite) TestServer() {
	grpcServer := grpc.NewServer(
		grpc.Address(":8090"),
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
