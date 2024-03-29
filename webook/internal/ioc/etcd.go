package ioc

import (
	"github.com/Anwenya/GeekTime/webook/config"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcd() *etcdv3.Client {
	client, err := etcdv3.NewFromURLs(config.Config.Etcd.Address)
	if err != nil {
		panic(any(err))
	}
	return client
}
