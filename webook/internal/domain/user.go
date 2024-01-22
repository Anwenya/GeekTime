package domain

import "time"

type User struct {
	Id         int64
	Email      string
	Password   string
	Nickname   string
	Bio        string
	Birthday   time.Time
	CreateTime time.Time
}
