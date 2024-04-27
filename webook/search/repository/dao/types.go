package dao

import "context"

type UserDAO interface {
	InputUser(ctx context.Context, user User) error
	Search(ctx context.Context, keywords []string) ([]User, error)
}

type ArticleDAO interface {
	InputArticle(ctx context.Context, article Article) error
	Search(ctx context.Context, artIds []int64, keywords []string) ([]Article, error)
}

type TagDAO interface {
	Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error)
}

type AnyDAO interface {
	Input(ctx context.Context, index, docID, data string) error
}

const UserIndexName = "user_index"
const ArticleIndexName = "article_index"
const TagIndexName = "tags_index"

type User struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
}

type Article struct {
	Id      int64    `json:"id"`
	Title   string   `json:"title"`
	Status  int32    `json:"status"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

type BizTags struct {
	Uid   int64    `json:"uid"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Tags  []string `json:"tags"`
}
