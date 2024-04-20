package domain

import "time"

type Comment struct {
	Id int64 `json:"id"`
	// 评论者
	Commentator User `json:"user"`
	// 业务
	Biz   string `json:"biz"`
	BizId int64  `json:"biz_id"`
	// 评论内容
	Content string `json:"content"`
	// 根评论
	RootComment *Comment `json:"root_comment"`
	// 父评论
	ParentComment *Comment `json:"parent_comment"`
	// 子评论
	Children   []Comment `json:"children"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
