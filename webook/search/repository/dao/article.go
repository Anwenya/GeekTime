package dao

import (
	"context"
	"encoding/json"
	"github.com/ecodeclub/ekit/slice"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

type ArticleElasticDao struct {
	client *elastic.Client
}

func NewArticleElasticDao(client *elastic.Client) ArticleDAO {
	return &ArticleElasticDao{client: client}
}

func (a *ArticleElasticDao) InputArticle(ctx context.Context, article Article) error {
	_, err := a.client.
		Index().
		// 索引 相当于MySQL的数据库
		Index(ArticleIndexName).
		// 文档的ID
		Id(strconv.FormatInt(article.Id, 10)).
		// 文档 相当于MySQL的记录
		BodyJson(article).
		Do(ctx)
	return err
}

func (a *ArticleElasticDao) Search(ctx context.Context, artIds []int64, keywords []string) ([]Article, error) {
	queryString := strings.Join(keywords, " ")
	// 文章的状态 2是已发布 表示只能搜索到已发布的文章
	status := elastic.NewTermsQuery("status", 2)
	// 文章标题命中关键词
	title := elastic.NewMatchQuery("title", queryString)
	// 文章内容命中关键词
	content := elastic.NewMatchQuery("content", queryString)
	// id命中 高权重
	tag := elastic.NewTermsQuery("id",
		slice.Map[int64](artIds, func(idx int, src int64) any {
			return src
		}),
	).Boost(2)

	// 标题/内容/标签  任何一个命中即可
	or := elastic.NewBoolQuery().Should(title, content, tag)
	// 状态必须命中
	query := elastic.NewBoolQuery().Must(status, or)

	// 执行查询
	resp, err := a.client.Search(ArticleIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}

	// 命中的结果
	res := make([]Article, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var article Article
		err = json.Unmarshal(hit.Source, &article)
		if err != nil {
			return nil, err
		}
		res = append(res, article)
	}
	return res, nil
}
