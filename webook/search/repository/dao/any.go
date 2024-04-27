package dao

import (
	"context"
	"github.com/olivere/elastic/v7"
)

type AnyEsDao struct {
	client *elastic.Client
}

func NewAnyEsDao(client *elastic.Client) AnyDAO {
	return &AnyEsDao{client: client}
}

func (a *AnyEsDao) Input(ctx context.Context, index, docID, data string) error {
	_, err := a.client.Index().Index(index).Id(docID).BodyString(data).Do(ctx)
	return err
}
