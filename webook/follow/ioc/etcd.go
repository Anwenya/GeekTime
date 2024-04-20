package ioc

import (
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcdClient(l logger.LoggerV1) *clientv3.Client {
	var cfg clientv3.Config
	err := viper.UnmarshalKey("etcd", &cfg)
	if err != nil {
		l.Error("读取etcd配置失败", logger.Error(err))
		panic(any(err))
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		l.Error("初始化etcd客户端失败", logger.Error(err))
		panic(any(err))
	}
	return client
}
