package balance

import (
	"context"
	"fmt"
	"github.com/Anwenya/GeekTime/demo/balance/pb"
	_ "github.com/Anwenya/GeekTime/webook/pkg/grpcx/balancer/wrr"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/weightedroundrobin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type GrpcTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (e *GrpcTestSuite) SetupSuite() {
	// 初始化etcd客户端
	cli, err := etcdv3.NewFromURL("192.168.133.128:12379")
	require.NoError(e.T(), err)
	e.cli = cli
}

// 测试服务存在异常时的负载均衡
func (e *GrpcTestSuite) TestServerFailover() {
	// 启动会返回异常的服务
	go func() {
		e.startFailoverServer(8090)
	}()
	// 启动正常的服务
	e.startWeightServer(8091, 20)
}

// 启动带权重的正常服务
func (e *GrpcTestSuite) TestServer() {
	go func() {
		e.startWeightServer(8090, 10)
	}()
	go func() {
		e.startWeightServer(8092, 20)
	}()
	e.startWeightServer(8091, 30)
}

func (e *GrpcTestSuite) startFailoverServer(port int) {
	t := e.T()
	em, err := endpoints.NewManager(e.cli, "service/user")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	address := fmt.Sprintf("127.0.0.1:%d", port)
	key := fmt.Sprintf("service/user/%s", address)

	// 租约
	var ttl int64 = 5
	leaseResp, err := e.cli.Grant(ctx, ttl)

	// 注册服务
	err = em.AddEndpoint(
		ctx, key,
		endpoints.Endpoint{
			Addr: address,
		},
		etcdv3.WithLease(leaseResp.ID),
	)
	require.NoError(t, err)

	// 续约
	kaCtx, _ := context.WithCancel(context.Background())
	go func() {
		_, err := e.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err)
	}()

	// 启动服务端
	server := grpc.NewServer()
	pb.RegisterUserServiceServer(server, &FailoverServer{code: codes.Unavailable})
	listen, err := net.Listen("tcp", address)
	require.NoError(t, err)
	server.Serve(listen)
}

func (e *GrpcTestSuite) startWeightServer(port int, weight int) {
	t := e.T()
	em, err := endpoints.NewManager(e.cli, "service/user")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	address := fmt.Sprintf("127.0.0.1:%d", port)
	key := fmt.Sprintf("service/user/%s", address)

	// 租约
	var ttl int64 = 5
	leaseResp, err := e.cli.Grant(ctx, ttl)

	// 注册
	err = em.AddEndpoint(
		ctx, key,
		endpoints.Endpoint{
			Addr: address,
			Metadata: map[string]any{
				"weight": weight,
			},
		},
		etcdv3.WithLease(leaseResp.ID),
	)
	require.NoError(t, err)

	// 续约
	kaCtx, _ := context.WithCancel(context.Background())
	go func() {
		_, err := e.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err)
	}()

	// 启动服务
	server := grpc.NewServer()
	pb.RegisterUserServiceServer(server, &Server{name: address})
	listen, err := net.Listen("tcp", address)
	require.NoError(t, err)
	server.Serve(listen)
}

// 测试 grpc 默认情况
// 默认是没有负载均衡的 每次都会选第一个
func (e *GrpcTestSuite) TestClientDefault() {
	t := e.T()
	etcdResolver, err := resolver.NewBuilder(e.cli)
	require.NoError(t, err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	userClient := pb.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := userClient.GetById(
			ctx,
			&pb.GetByIdRequest{
				Id: 123,
			},
		)
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}

}

// 测试 grpc 自带的普通轮询算法
// 轮流选择
func (e *GrpcTestSuite) TestClientRoundRobin() {
	t := e.T()
	etcdResolver, err := resolver.NewBuilder(e.cli)
	require.NoError(t, err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(
			`{"loadBalancingConfig": [{"round_robin":{}}]}`,
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	userClient := pb.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := userClient.GetById(
			ctx,
			&pb.GetByIdRequest{
				Id: 123,
			},
		)
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}

}

// 测试 grpc 自带的普通轮询算法
// 额外加上重试 切换 机制
// 异常时会重试 切换节点 等等
func (e *GrpcTestSuite) TestClientRoundRobinFailover() {
	t := e.T()
	etcdResolver, err := resolver.NewBuilder(e.cli)
	require.NoError(t, err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(
			`{
				  "loadBalancingConfig": [{"round_robin":{}}],
				  "methodConfig": [{
					"name": [{"service": "UserService"}],
					"retryPolicy": {
					  "maxAttempts": 4,
					  "initialBackoff": "0.01s",
					  "maxBackoff": "0.1s",
					  "backoffMultiplier": 2.0,
					  "retryableStatusCodes": [ "UNAVAILABLE" ]
					}
				  }]
				}`,
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	userClient := pb.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := userClient.GetById(
			ctx,
			&pb.GetByIdRequest{
				Id: 123,
			},
		)
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}

}

// 测试 grpc 自带的加权轮询算法
// 该算法会自己统计服务的权重
func (e *GrpcTestSuite) TestClientWeightedRoundRobin() {
	t := e.T()
	etcdResolver, err := resolver.NewBuilder(e.cli)
	require.NoError(t, err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(
			`{"loadBalancingConfig": [{"weighted_round_robin":{}}]}`,
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	userClient := pb.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := userClient.GetById(
			ctx,
			&pb.GetByIdRequest{
				Id: 123,
			},
		)
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}

}

// 测试 自定义的算法
// 平滑加权轮询
func (e *GrpcTestSuite) TestClientCustomWeightedRondRobin() {
	t := e.T()
	etcdResolver, err := resolver.NewBuilder(e.cli)
	require.NoError(t, err)
	cc, err := grpc.Dial(
		"etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(
			`{"loadBalancingConfig": [{"custom_weighted_round_robin":{}}]}`,
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	userClient := pb.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := userClient.GetById(
			ctx,
			&pb.GetByIdRequest{
				Id: 123,
			},
		)
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}

}

func TestBalance(t *testing.T) {
	suite.Run(t, new(GrpcTestSuite))
}
