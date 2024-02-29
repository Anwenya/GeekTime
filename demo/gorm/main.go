package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Article struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Title      string `gorm:"type=varchar(4096)"`
	Content    string `gorm:"type=BLOB"`
	AuthorId   int64  `gorm:"index"`
	Status     uint8
	CreateTime int64
	UpdateTime int64
}

func main() {
	// 输出建表语句
	db, err := gorm.Open(
		mysql.Open("root:root@tcp(192.168.2.128:13306)/webook"),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		},
	)
	if err != nil {
		panic(any(err))
	}

	err = db.AutoMigrate(&Article{})
	if err != nil {
		panic(any(err))
	}

}
