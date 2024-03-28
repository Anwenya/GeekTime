package etcd

import (
	"context"
	"github.com/Anwenya/GeekTime/demo/etcd/pb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type EtcdTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (e *EtcdTestSuite) SetupSuite() {
	// 初始化etcd客户端
	cli, err := etcdv3.NewFromURL("192.168.2.130:12379")
	require.NoError(e.T(), err)
	e.cli = cli
}

func (e *EtcdTestSuite) TestClient() {
	t := e.T()
	builder, err := resolver.NewBuilder(e.cli)
	require.NoError(t, err)

	// 后续客户端根据该解析器从etcd对应的key解析可用的服务地址
	cc, err := grpc.Dial(
		"etcd:///server/user",
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	require.NoError(t, err)

	client := pb.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetById(
		ctx,
		&pb.GetByIdRequest{
			Id: 123,
		},
	)
	require.NoError(t, err)
	t.Log(resp)
}

func (e *EtcdTestSuite) TestServer() {
	t := e.T()
	em, err := endpoints.NewManager(e.cli, "server/user")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 该服务端监听的地址
	addr := "127.0.0.1:8090"
	// 在etcd中对应的key
	key := "server/user/" + addr

	// grpc服务端 监听的地址
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)

	// 租期
	var ttl int64 = 5
	leaseResp, err := e.cli.Grant(ctx, ttl)
	require.NoError(t, err)

	// 向 key 中存入一个值 addr 表示可用的服务端的地址
	err = em.AddEndpoint(
		ctx, key,
		endpoints.Endpoint{Addr: addr},
		// 该key对应的租约
		etcdv3.WithLease(leaseResp.ID),
	)
	require.NoError(t, err)

	// 续约 默认续约间隔是 ttl/3
	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		ch, err := e.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err)
		for kaRep := range ch {
			t.Log(kaRep.String())
		}
	}()

	// 模拟更新注册信息
	go func() {
		ticker := time.NewTicker(time.Second)
		for now := range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := em.AddEndpoint(
				ctx, key,
				endpoints.Endpoint{
					Addr:     addr,
					Metadata: now.String(),
				},
				etcdv3.WithLease(leaseResp.ID),
			)
			cancel()
			if err != nil {
				t.Log(err)
			}
		}
	}()

	// 启动grpc服务端
	server := grpc.NewServer()
	pb.RegisterUserServiceServer(server, &Server{})
	server.Serve(l)

	// 停止续约
	kaCancel()
	// 删除注册信息
	err = em.DeleteEndpoint(ctx, key)
	if err != nil {
		t.Log(err)
	}
	// grpc 优雅停机
	server.GracefulStop()
	// 关闭etcd客户端
	e.cli.Close()
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
