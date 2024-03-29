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

	EtcdAddr string
	client   *etcdv3.Client
	kaCancel func()

	L logger.LoggerV1
}

func (s *Server) Serve() error {
	addr := fmt.Sprintf(":%d", s.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// 注册服务
	err = s.register()
	if err != nil {
		return err
	}
	return s.Server.Serve(listener)
}

func (s *Server) register() error {
	// etcd客户端
	client, err := etcdv3.NewFromURL(s.EtcdAddr)
	if err != nil {
		return err
	}

	s.client = client
	em, err := endpoints.NewManager(client, fmt.Sprintf("service/%s", s.Name))
	// 服务端注册的地址
	addr := fmt.Sprintf("%s:%d", netx.GetOutboundIP(), s.Port)
	// 在etcd中的key
	key := fmt.Sprintf("service/%s/%s", s.Name, addr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 租期
	var ttl int64 = 5
	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	// 向etcd注册key 并绑定租期
	err = em.AddEndpoint(
		ctx, key,
		endpoints.Endpoint{
			Addr: addr,
		},
		etcdv3.WithLease(leaseResp.ID),
	)

	if err != nil {
		return err
	}

	// 续约
	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	ch, err := client.KeepAlive(kaCtx, leaseResp.ID)
	go func() {
		for kaResp := range ch {
			s.L.Debug(kaResp.String())
		}
	}()
	return err
}

func (s *Server) Close() error {
	// 停止续约
	if s.kaCancel != nil {
		s.kaCancel()
	}

	// 关闭etcd客户端
	if s.client != nil {
		err := s.client.Close()
		if err != nil {
			return err
		}
	}
	// grpc优雅停机
	s.GracefulStop()
	return nil
}
