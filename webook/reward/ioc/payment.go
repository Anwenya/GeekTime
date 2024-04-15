package ioc

import (
	pmtv1 "github.com/Anwenya/GeekTime/webook/api/proto/gen/payment"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitPaymentClient(l logger.LoggerV1, etcdClient *etcdv3.Client) pmtv1.WechatPaymentServiceClient {
	type Config struct {
		Target string `json:"target"`
		Secure bool   `json:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.payment", &cfg)
	if err != nil {
		l.Error("读取payment配置失败", logger.Error(err))
		panic(any(err))
	}
	rs, err := resolver.NewBuilder(etcdClient)
	if err != nil {
		l.Error("创建etcd客户端失败", logger.Error(err))
		panic(any(err))
	}
	opts := []grpc.DialOption{grpc.WithResolvers(rs)}
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.Dial(cfg.Target, opts...)
	if err != nil {
		l.Error("连接payment服务失败", logger.Error(err))
		panic(any(err))
	}
	return pmtv1.NewWechatPaymentServiceClient(cc)
}
