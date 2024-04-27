package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

type TagEsDao struct {
	client *elastic.Client
}

func NewTagEsDao(client *elastic.Client) TagDAO {
	return &TagEsDao{client: client}
}

func (t *TagEsDao) Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error) {
	// 必须是用户自己打的标签
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("uid", uid),
		elastic.NewTermQuery("biz", biz),
		elastic.NewTermsQueryFromStrings("tags", keywords...),
	)

	// 搜索
	resp, err := t.client.Search(TagIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var bt BizTags
		err = json.Unmarshal(hit.Source, &bt)
		if err != nil {
			return nil, err
		}
		res = append(res, bt.BizId)
	}
	return res, nil
}
