package ioc

import (
	"context"
	_ "embed"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/olivere/elastic/v7"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"time"
)

func InitES(l logger.LoggerV1) *elastic.Client {
	type config struct {
		Url   string `yaml:"url"`
		Sniff bool   `yaml:"sniff"`
	}

	var cfg config
	err := viper.UnmarshalKey("es", &cfg)
	if err != nil {
		l.Error("读取es配置失败", logger.Error(err))
		panic(any(err))
	}

	const timeout = 100 * time.Second
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(cfg.Url),
		elastic.SetHealthcheckTimeoutStartup(timeout),
	}
	client, err := elastic.NewClient(opts...)
	if err != nil {
		l.Error("es连接失败", logger.Error(err))
		panic(any(err))
	}
	l.Info("es连接成功")
	err = initESIndex(client)
	if err != nil {
		l.Error("eses索引创建失败", logger.Error(err))
		panic(any(err))
	}
	l.Info("es索引创建成功")
	return client
}

var (
	//go:embed es_user_index.json
	userIndex string
	//go:embed es_article_index.json
	articleIndex string
	//go:embed es_tag_index.json
	tagIndex string
)

const UserIndexName = "user_index"
const ArticleIndexName = "article_index"
const TagIndexName = "tags_index"

// InitESIndex 创建索引
func initESIndex(client *elastic.Client) error {
	const timeout = time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var eg errgroup.Group
	eg.Go(func() error {
		return tryCreateIndex(ctx, client, UserIndexName, userIndex)
	})
	eg.Go(func() error {
		return tryCreateIndex(ctx, client, ArticleIndexName, articleIndex)
	})
	eg.Go(func() error {
		return tryCreateIndex(ctx, client, TagIndexName, tagIndex)
	})

	return eg.Wait()
}

func tryCreateIndex(
	ctx context.Context,
	client *elastic.Client,
	idxName, idxCfg string,
) error {
	// 索引可能已经建好了
	ok, err := client.IndexExists(idxName).Do(ctx)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = client.CreateIndex(idxName).Body(idxCfg).Do(ctx)
	return err
}
