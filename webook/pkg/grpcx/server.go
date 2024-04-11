package grpcx

import (
	"context"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/Anwenya/GeekTime/webook/pkg/netx"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"time"
)

type Server struct {
	*grpc.Server
	Port int
	Name string

	TTL    int64
	Client *etcdv3.Client
	key    string
	em     endpoints.Manager
	cancel func()

	L logger.LoggerV1
}

func (s *Server) Serve() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	// 服务监听的地址
	addr := fmt.Sprintf(":%d", s.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// 注册服务
	err = s.register(ctx)
	if err != nil {
		return err
	}
	return s.Server.Serve(listener)
}

func (s *Server) register(ctx context.Context) error {
	// etcd客户端
	client := s.Client
	em, err := endpoints.NewManager(client, fmt.Sprintf("service/%s", s.Name))
	// 服务端注册的地址
	addr := fmt.Sprintf("%s:%d", netx.GetOutboundIP(), s.Port)
	// 在etcd中的key
	s.key = fmt.Sprintf("service/%s/%s", s.Name, addr)

	// 租期
	leaseResp, err := client.Grant(ctx, s.TTL)
	if err != nil {
		return err
	}

	// 向etcd注册key 并绑定租期
	err = em.AddEndpoint(
		ctx, s.key,
		endpoints.Endpoint{
			Addr: addr,
		},
		etcdv3.WithLease(leaseResp.ID),
	)

	if err != nil {
		return err
	}

	// 续约
	ch, err := client.KeepAlive(ctx, leaseResp.ID)
	go func() {
		for kaResp := range ch {
			s.L.Debug("续约", logger.String("resp", kaResp.String()))
		}
	}()
	return err
}

func (s *Server) Close() error {
	// 停止续约
	s.cancel()

	// 删除key
	if s.em != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := s.em.DeleteEndpoint(ctx, s.key)
		if err != nil {
			return err
		}
	}

	// 关闭etcd客户端
	if s.Client != nil {
		err := s.Client.Close()
		if err != nil {
			return err
		}
	}

	// grpc优雅停机
	s.GracefulStop()
	return nil
}
