package ioc

import (
	interactivev1 "github.com/Anwenya/GeekTime/webook/api/proto/gen/interactive/v1"
	"github.com/Anwenya/GeekTime/webook/interactive/service"
	"github.com/Anwenya/GeekTime/webook/internal/client"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitInteractiveClient 从配置文件读服务地址
func InitInteractiveClient(svc service.InteractiveService) interactivev1.InteractiveServiceClient {
	type Config struct {
		Address   string `yaml:"address"`
		Secure    bool
		Threshold int32
	}

	var config Config

	err := viper.UnmarshalKey("grpc.client.interactive", &config)
	if err != nil {
		panic(any(err))
	}

	var opts []grpc.DialOption
	if !config.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	cc, err := grpc.Dial(config.Address, opts...)
	if err != nil {
		panic(any(err))
	}

	remote := interactivev1.NewInteractiveServiceClient(cc)
	local := client.NewLocalInteractiveServiceAdapter(svc)
	interactiveClient := client.NewInteractiveClient(remote, local)
	viper.OnConfigChange(
		func(in fsnotify.Event) {
			config := Config{}
			err := viper.UnmarshalKey("grpc.client.interactive", &config)
			if err != nil {
				panic(any(err))
			}
			interactiveClient.UpdateThreshold(config.Threshold)
		},
	)
	return interactiveClient

}

// InitInteractiveClientV1 从注册中心读取服务地址
func InitInteractiveClientV1(client *etcdv3.Client) interactivev1.InteractiveServiceClient {
	type Config struct {
		Addr   string `yaml:"addr"`
		Secure bool   `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.interactive", &cfg)
	if err != nil {
		panic(any(err))
	}
	builder, err := resolver.NewBuilder(client)
	if err != nil {
		panic(any(err))
	}

	opts := []grpc.DialOption{grpc.WithResolvers(builder)}
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	cc, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(any(err))
	}

	remote := interactivev1.NewInteractiveServiceClient(cc)
	return remote
}
