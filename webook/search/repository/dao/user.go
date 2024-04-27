package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

type UserEsDao struct {
	client *elastic.Client
}

func NewUserEsDao(client *elastic.Client) UserDAO {
	return &UserEsDao{client: client}
}

func (u *UserEsDao) InputUser(ctx context.Context, user User) error {
	_, err := u.client.
		Index().
		Index(UserIndexName).
		Id(strconv.FormatInt(user.Id, 10)).
		BodyJson(user).
		Do(ctx)
	return err
}

func (u *UserEsDao) Search(ctx context.Context, keywords []string) ([]User, error) {
	queryString := strings.Join(keywords, " ")
	// 昵称命中关键词
	query := elastic.NewMatchQuery("nickname", queryString)
	resp, err := u.client.Search(UserIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]User, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var user User
		err = json.Unmarshal(hit.Source, &user)
		if err != nil {
			return nil, err
		}
		res = append(res, user)
	}
	return nil, err
}
